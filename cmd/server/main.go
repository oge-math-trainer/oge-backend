package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oge-math-trainer/oge-backend.git/ai"
	"github.com/oge-math-trainer/oge-backend.git/internal/api"
	"github.com/oge-math-trainer/oge-backend.git/internal/auth"
	"github.com/oge-math-trainer/oge-backend.git/internal/config"
	"github.com/oge-math-trainer/oge-backend.git/internal/db"
	"github.com/oge-math-trainer/oge-backend.git/internal/diagnostic"
	"github.com/oge-math-trainer/oge-backend.git/internal/progress"
	"github.com/oge-math-trainer/oge-backend.git/internal/tasks"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	envPath := findDotEnvPath()
	if envPath != "" {
		log.Printf("Loaded .env from: %s", envPath)
	} else {
		log.Print("No .env file found; using process environment")
	}

	cfg, err := config.Load(envPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := cfg.ValidateServer(); err != nil {
		log.Fatal(err)
	}

	store, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	aiClient := ai.NewConfiguredClient(cfg.AITunnelBaseURL, cfg.AITunnelAPIKey, cfg.AITunnelModel)

	log.Printf("DEBUG: AITunnelAPIKey length: %d", len(cfg.AITunnelAPIKey))
	if aiClient.IsConfigured() {
		log.Print("AI CLIENT: configured (key is present)")
	} else {
		log.Print("AI CLIENT: not configured (key is missing or empty)")
	}

	authService := auth.NewService(store, cfg.AuthSecret, cfg.TokenTTL)
	authService.ConfigureOAuth(auth.OAuthConfig{
		Google: auth.OAuthProviderConfig{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
		},
		Yandex: auth.OAuthProviderConfig{
			ClientID:     cfg.YandexClientID,
			ClientSecret: cfg.YandexClientSecret,
			RedirectURL:  cfg.YandexRedirectURL,
		},
	})
	taskService := tasks.NewService(store, aiClient)
	diagnosticService := diagnostic.NewService(store, taskService, aiClient)
	progressService := progress.NewService(store)

	if aiClient.IsConfigured() && cfg.PreparedTasksMin > 0 {
		worker := tasks.NewWorker(taskService, store, cfg.PreparedTasksMin, cfg.PreparedTasksInterval)
		go func() {
			log.Printf("embedded prepared task worker starting: min_per_target=%d interval=%s", cfg.PreparedTasksMin, cfg.PreparedTasksInterval)
			if err := worker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("embedded prepared task worker stopped: %v", err)
			}
		}()
	} else {
		log.Print("embedded prepared task worker disabled: AI client is not configured or PREPARED_TASKS_MIN <= 0")
	}

	router := api.NewRouter(api.Dependencies{
		Auth:                  authService,
		Tasks:                 taskService,
		Diagnostic:            diagnosticService,
		Progress:              progressService,
		Health:                store,
		CORSAllowedOrigins:    cfg.CORSAllowedOrigins,
		AuthRateLimitRequests: cfg.AuthRateLimitRequests,
		AuthRateLimitWindow:   cfg.AuthRateLimitWindow,
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      300 * time.Second,
		IdleTimeout:       300 * time.Second,
	}

	log.Printf("server listening on :%s", cfg.Port)
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatal(err)
		}
		if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}
}

func findDotEnvPath() string {
	for _, path := range []string{".env", "../.env", "../../.env", "../../../.env"} {
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			return path
		}
	}
	return ""
}
