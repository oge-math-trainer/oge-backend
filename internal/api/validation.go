package api

import (
	"net/mail"
	"strings"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
)

const (
	minPasswordBytes = 8
	maxPasswordBytes = 72
)

func validateCredentials(email, password string) error {
	if err := validateEmail(email); err != nil {
		return err
	}
	return validatePassword(password)
}

func validateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return app.Validation("Email обязателен")
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(email)); err != nil {
		return app.Validation("Некорректный email")
	}
	return nil
}

func validatePassword(password string) error {
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

func validateResetCode(code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return app.Validation("Код обязателен")
	}
	if len(code) != 6 {
		return app.Validation("Код должен состоять из 6 цифр")
	}
	for _, ch := range code {
		if ch < '0' || ch > '9' {
			return app.Validation("Код должен состоять из 6 цифр")
		}
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
