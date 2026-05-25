package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oge-math-trainer/oge-backend.git/internal/config"
	"github.com/oge-math-trainer/oge-backend.git/internal/migrations"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	envPath := findDotEnvPath()
	cfg, err := config.Load(envPath)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	migrateCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	result, err := migrations.Apply(migrateCtx, pool, envOrDefault("MIGRATIONS_DIR", "migrations"))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("migrations complete: applied=%d skipped=%d", len(result.Applied), len(result.Skipped))
	for _, name := range result.Applied {
		log.Printf("migration applied: %s", name)
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

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
