package config

import (
	"testing"
	"time"
)

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
		AITunnelAPIKey:        "sk-aitunnel-test123",
		AITunnelModel:         "qwen3-235b-a22b-2507",
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
		// AI-поля не заданы — это допустимо
	}

	if err := cfg.ValidateServer(); err != nil {
		t.Fatalf("ожидали успешную валидацию без AI, получили: %v", err)
	}
}
