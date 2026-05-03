package config

import (
	"testing"
	"time"
)

func TestValidateServerRequiresModelWhenOpenRouterKeySet(t *testing.T) {
	cfg := Config{
		Port:                  "8080",
		DatabaseURL:           "postgres://example",
		AuthSecret:            "12345678901234567890123456789012",
		TokenTTL:              time.Hour,
		AuthRateLimitRequests: 10,
		AuthRateLimitWindow:   time.Minute,
		OpenRouterAPIKey:      "secret",
	}

	if err := cfg.ValidateServer(); err == nil {
		t.Fatal("expected error when OPENROUTER_MODEL is empty")
	}
}
