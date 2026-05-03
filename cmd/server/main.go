package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"oge-ai-trainer/backend/internal/ai"
	"oge-ai-trainer/backend/internal/api"
	"oge-ai-trainer/backend/internal/auth"
	"oge-ai-trainer/backend/internal/config"
	"oge-ai-trainer/backend/internal/db"
	"oge-ai-trainer/backend/internal/diagnostic"
	"oge-ai-trainer/backend/internal/progress"
	"oge-ai-trainer/backend/internal/tasks"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(".env")
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

	aiClient := ai.NewClient(cfg.OpenRouterBaseURL, cfg.OpenRouterAPIKey, cfg.OpenRouterModel)
	authService := auth.NewService(store, cfg.AuthSecret, cfg.TokenTTL)
	taskService := tasks.NewService(store, aiClient)
	diagnosticService := diagnostic.NewService(store, taskService, aiClient)
	progressService := progress.NewService(store)

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
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
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
