package config

import (
	"testing"
	"time"
)

func TestLoadDefaultsPreparedTasksMinTo50(t *testing.T) {
	t.Setenv("PREPARED_TASKS_MIN", "")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.PreparedTasksMin != 50 {
		t.Fatalf("expected PreparedTasksMin 50, got %d", cfg.PreparedTasksMin)
	}
}

func TestLoadFloorsPreparedTasksMinAt50(t *testing.T) {
	t.Setenv("PREPARED_TASKS_MIN", "20")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.PreparedTasksMin != 50 {
		t.Fatalf("expected PreparedTasksMin floor 50, got %d", cfg.PreparedTasksMin)
	}
}

// TestValidateServerRequiresModelWhenAITunnelKeySet проверяет,
// что при заданном ключе AI-Tunnel обязательно указывать модель.
func TestValidateServerRequiresModelWhenAITunnelKeySet(t *testing.T) {
	cfg := Config{
		Port:                  "8080",
		DatabaseURL:           "postgres://example",
		AuthSecret:            "12345678901234567890123456789012",
		TokenTTL:              time.Hour,
		AuthRateLimitRequests: 10,
		AuthRateLimitWindow:   time.Minute,
		AITunnelAPIKey:        "sk-aitunnel-test123",
		// AITunnelModel намеренно не задан — ожидаем ошибку
	}

	err := cfg.ValidateServer()
	if err == nil {
		t.Fatal("ожидали ошибку, когда AITUNNEL_MODEL не задан при установленном AITUNNEL_API_KEY")
	}

	// Проверяем, что ошибка содержит ожидаемое сообщение
	expectedMsg := "AITUNNEL_MODEL is required when AITUNNEL_API_KEY is set"
	if err.Error() != expectedMsg {
		t.Errorf("неверное сообщение ошибки:\n  получили: %q\n  ожидали: %q", err.Error(), expectedMsg)
	}
}

// TestValidateServerOKWithAITunnelConfig проверяет успешную валидацию
// при полном наборе AI-переменных.
func TestValidateServerOKWithAITunnelConfig(t *testing.T) {
	cfg := Config{
		Port:                  "8080",
		DatabaseURL:           "postgres://example",
		AuthSecret:            "12345678901234567890123456789012",
		TokenTTL:              time.Hour,
		AuthRateLimitRequests: 10,
		AuthRateLimitWindow:   time.Minute,
		PasswordResetCodeTTL:  15 * time.Minute,
		AITunnelAPIKey:        "sk-aitunnel-test123",
		AITunnelModel:         "gpt-4o-mini",
		AITunnelBaseURL:       "https://api.aitunnel.ru/v1",
	}

	if err := cfg.ValidateServer(); err != nil {
		t.Fatalf("ожидали успешную валидацию, получили ошибку: %v", err)
	}
}

// TestValidateServerOKWithoutAI проверяет, что конфигурация валидна
// даже без AI-настроек (для режимов без генерации).
func TestValidateServerOKWithoutAI(t *testing.T) {
	cfg := Config{
		Port:                  "8080",
		DatabaseURL:           "postgres://example",
		AuthSecret:            "12345678901234567890123456789012",
		TokenTTL:              time.Hour,
		AuthRateLimitRequests: 10,
		AuthRateLimitWindow:   time.Minute,
		PasswordResetCodeTTL:  15 * time.Minute,
		// AI-поля не заданы — это допустимо
	}

	if err := cfg.ValidateServer(); err != nil {
		t.Fatalf("ожидали успешную валидацию без AI, получили: %v", err)
	}
}
