package config

import (
	"bufio"
	"errors"
	stdmail "net/mail"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port                    string
	DatabaseURL             string
	AuthSecret              string
	TokenTTL                time.Duration
	AuthRateLimitRequests   int
	AuthRateLimitWindow     time.Duration
	CORSAllowedOrigins      []string
	GoogleClientID          string
	GoogleClientSecret      string
	GoogleRedirectURL       string
	YandexClientID          string
	YandexClientSecret      string
	YandexRedirectURL       string
	OAuthSuccessRedirectURL string
	SMTPHost                string
	SMTPPort                int
	SMTPUsername            string
	SMTPPassword            string
	SMTPFromEmail           string
	SMTPFromName            string
	PasswordResetCodeTTL    time.Duration
	AITunnelBaseURL         string
	AITunnelAPIKey          string
	AITunnelModel           string
	PreparedTasksMin        int
	PreparedTasksInterval   time.Duration
}

func Load(path string) (Config, error) {
	if path != "" {
		if err := loadDotEnv(path); err != nil {
			return Config{}, err
		}
	}

	tokenTTL := 24 * time.Hour
	if raw := strings.TrimSpace(os.Getenv("TOKEN_TTL")); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, err
		}
		tokenTTL = parsed
	}

	rateLimitRequests := 10
	if raw := strings.TrimSpace(os.Getenv("AUTH_RATE_LIMIT_REQUESTS")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, err
		}
		rateLimitRequests = parsed
	}

	rateLimitWindow := time.Minute
	if raw := strings.TrimSpace(os.Getenv("AUTH_RATE_LIMIT_WINDOW")); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, err
		}
		rateLimitWindow = parsed
	}

	passwordResetCodeTTL := 15 * time.Minute
	if raw := strings.TrimSpace(os.Getenv("PASSWORD_RESET_CODE_TTL")); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, err
		}
		passwordResetCodeTTL = parsed
	}

	origins := splitCSV(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if len(origins) == 0 {
		origins = []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:5174",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
			"http://127.0.0.1:5174",
		}
	}

	preparedTasksMin := 50
	if raw := strings.TrimSpace(os.Getenv("PREPARED_TASKS_MIN")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, err
		}
		preparedTasksMin = parsed
	}
	if preparedTasksMin > 0 && preparedTasksMin < 50 {
		preparedTasksMin = 50
	}

	preparedTasksInterval := 5 * time.Minute
	if raw := strings.TrimSpace(os.Getenv("PREPARED_TASKS_INTERVAL")); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, err
		}
		preparedTasksInterval = parsed
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	smtpHost := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	smtpUsername := strings.TrimSpace(os.Getenv("SMTP_USERNAME"))
	smtpPassword := strings.TrimSpace(os.Getenv("SMTP_PASSWORD"))
	smtpFromEmail := strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL"))
	if smtpFromEmail == "" && strings.Contains(smtpUsername, "@") {
		smtpFromEmail = smtpUsername
	}
	smtpFromName := envOrDefault("SMTP_FROM_NAME", "ОГЭ по математике")
	smtpPort := 0
	if smtpHost != "" || smtpUsername != "" || smtpPassword != "" || smtpFromEmail != "" {
		smtpPort = 587
	}
	if raw := strings.TrimSpace(os.Getenv("SMTP_PORT")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, err
		}
		smtpPort = parsed
	}

	return Config{
		Port:                    port,
		DatabaseURL:             strings.TrimSpace(os.Getenv("DATABASE_URL")),
		AuthSecret:              strings.TrimSpace(os.Getenv("AUTH_SECRET")),
		TokenTTL:                tokenTTL,
		AuthRateLimitRequests:   rateLimitRequests,
		AuthRateLimitWindow:     rateLimitWindow,
		CORSAllowedOrigins:      origins,
		GoogleClientID:          strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")),
		GoogleClientSecret:      strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET")),
		GoogleRedirectURL:       strings.TrimSpace(os.Getenv("GOOGLE_REDIRECT_URL")),
		YandexClientID:          strings.TrimSpace(os.Getenv("YANDEX_CLIENT_ID")),
		YandexClientSecret:      strings.TrimSpace(os.Getenv("YANDEX_CLIENT_SECRET")),
		YandexRedirectURL:       strings.TrimSpace(os.Getenv("YANDEX_REDIRECT_URL")),
		OAuthSuccessRedirectURL: strings.TrimSpace(os.Getenv("OAUTH_SUCCESS_REDIRECT_URL")),
		SMTPHost:                smtpHost,
		SMTPPort:                smtpPort,
		SMTPUsername:            smtpUsername,
		SMTPPassword:            smtpPassword,
		SMTPFromEmail:           smtpFromEmail,
		SMTPFromName:            smtpFromName,
		PasswordResetCodeTTL:    passwordResetCodeTTL,
		AITunnelBaseURL:         envOrDefault("AITUNNEL_BASE_URL", "https://api.aitunnel.ru/v1"),
		AITunnelAPIKey:          strings.TrimSpace(os.Getenv("AITUNNEL_API_KEY")),
		AITunnelModel:           strings.TrimSpace(os.Getenv("AITUNNEL_MODEL")),
		PreparedTasksMin:        preparedTasksMin,
		PreparedTasksInterval:   preparedTasksInterval,
	}, nil
}

func (c Config) ValidateServer() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	if c.AuthSecret == "" {
		return errors.New("AUTH_SECRET is required")
	}
	if len(c.AuthSecret) < 32 {
		return errors.New("AUTH_SECRET must be at least 32 characters")
	}
	if c.TokenTTL <= 0 {
		return errors.New("TOKEN_TTL must be positive")
	}
	if c.AuthRateLimitRequests <= 0 {
		return errors.New("AUTH_RATE_LIMIT_REQUESTS must be positive")
	}
	if c.AuthRateLimitWindow <= 0 {
		return errors.New("AUTH_RATE_LIMIT_WINDOW must be positive")
	}
	if c.AITunnelAPIKey != "" && c.AITunnelModel == "" {
		return errors.New("AITUNNEL_MODEL is required when AITUNNEL_API_KEY is set")
	}
	if err := validateOAuthProvider("GOOGLE", c.GoogleClientID, c.GoogleClientSecret, c.GoogleRedirectURL); err != nil {
		return err
	}
	if err := validateOAuthProvider("YANDEX", c.YandexClientID, c.YandexClientSecret, c.YandexRedirectURL); err != nil {
		return err
	}
	if c.OAuthSuccessRedirectURL != "" && !strings.HasPrefix(c.OAuthSuccessRedirectURL, "http://") && !strings.HasPrefix(c.OAuthSuccessRedirectURL, "https://") {
		return errors.New("OAUTH_SUCCESS_REDIRECT_URL must start with http:// or https://")
	}
	if c.PasswordResetCodeTTL <= 0 {
		return errors.New("PASSWORD_RESET_CODE_TTL must be positive")
	}
	if err := validateSMTP(c); err != nil {
		return err
	}
	if _, err := strconv.Atoi(c.Port); err != nil {
		return errors.New("PORT must be a number")
	}
	return nil
}

func (c Config) ValidateWorker() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	if c.AITunnelAPIKey != "" && c.AITunnelModel == "" {
		return errors.New("AITUNNEL_MODEL is required when AITUNNEL_API_KEY is set")
	}
	if c.PreparedTasksMin < 0 {
		return errors.New("PREPARED_TASKS_MIN must be non-negative")
	}
	if c.PreparedTasksInterval <= 0 {
		return errors.New("PREPARED_TASKS_INTERVAL must be positive")
	}
	return nil
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func validateOAuthProvider(prefix, clientID, clientSecret, redirectURL string) error {
	values := []string{clientID, clientSecret, redirectURL}
	filled := 0
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			filled++
		}
	}
	if filled == 0 {
		return nil
	}
	if filled != len(values) {
		return errors.New(prefix + "_CLIENT_ID, " + prefix + "_CLIENT_SECRET, and " + prefix + "_REDIRECT_URL must be set together")
	}
	if !strings.HasPrefix(redirectURL, "http://") && !strings.HasPrefix(redirectURL, "https://") {
		return errors.New(prefix + "_REDIRECT_URL must start with http:// or https://")
	}
	return nil
}

func validateSMTP(c Config) error {
	values := []string{c.SMTPHost, c.SMTPUsername, c.SMTPPassword, c.SMTPFromEmail}
	filled := 0
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			filled++
		}
	}
	if filled == 0 && c.SMTPPort == 0 {
		return nil
	}
	if c.SMTPHost == "" {
		return errors.New("SMTP_HOST is required when password reset email is enabled")
	}
	if c.SMTPPort <= 0 {
		return errors.New("SMTP_PORT must be positive")
	}
	if (c.SMTPUsername == "") != (c.SMTPPassword == "") {
		return errors.New("SMTP_USERNAME and SMTP_PASSWORD must be set together")
	}
	if c.SMTPFromEmail == "" {
		return errors.New("SMTP_FROM_EMAIL is required when password reset email is enabled")
	}
	if _, err := stdmail.ParseAddress(c.SMTPFromEmail); err != nil {
		return errors.New("SMTP_FROM_EMAIL must be a valid email")
	}
	return nil
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
