package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/oge-math-trainer/oge-backend.git/ai"
	"github.com/oge-math-trainer/oge-backend.git/internal/config"
	"github.com/oge-math-trainer/oge-backend.git/internal/db"
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
	if err := cfg.ValidateWorker(); err != nil {
		log.Fatal(err)
	}

	store, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	aiClient := ai.NewConfiguredClient(cfg.AITunnelBaseURL, cfg.AITunnelAPIKey, cfg.AITunnelModel)
	if !aiClient.IsConfigured() {
		log.Fatal("AI client is not configured; set AITUNNEL_API_KEY and AITUNNEL_MODEL")
	}

	taskService := tasks.NewService(store, aiClient)
	worker := tasks.NewWorker(taskService, store, cfg.PreparedTasksMin, cfg.PreparedTasksInterval)

	log.Printf("prepared task worker starting: min_per_target=%d interval=%s", cfg.PreparedTasksMin, cfg.PreparedTasksInterval)
	if err := worker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
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
