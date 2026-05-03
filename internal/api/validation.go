package api

import (
	"net/mail"
	"strings"

	"oge-ai-trainer/backend/internal/app"
)

const (
	minPasswordBytes = 8
	maxPasswordBytes = 72
)

func validateCredentials(email, password string) error {
	if strings.TrimSpace(email) == "" {
		return app.Validation("Email обязателен")
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(email)); err != nil {
		return app.Validation("Некорректный email")
	}
	// MVP password policy:
	// - minimum 8 bytes;
	// - maximum 72 bytes because bcrypt ignores bytes after that limit;
	// - no complexity rules for a student-facing MVP.
	if len(password) < minPasswordBytes {
		return app.Validation("Пароль должен быть не короче 8 символов")
	}
	if len(password) > maxPasswordBytes {
		return app.Validation("Пароль должен быть не длиннее 72 байт")
	}
	return nil
}

func validatePositiveID(id int64, field string) error {
	if id <= 0 {
		return app.Validation(field + " должен быть положительным")
	}
	return nil
}

func validateOgeNumber(number int) error {
	if number < 6 || number > 19 {
		return app.Validation("oge_number должен быть от 6 до 19")
	}
	return nil
}
