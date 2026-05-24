package app

import "fmt"

const (
	CodeValidation      = "validation_error"
	CodeUnauthorized    = "unauthorized"
	CodeNotFound        = "not_found"
	CodeConflict        = "conflict"
	CodeAIUnavailable   = "ai_unavailable"
	CodeTaskUnavailable = "task_unavailable"
	CodeDBUnavailable   = "db_unavailable"
	CodeRateLimited     = "rate_limited"
	CodeInternal        = "internal_error"
)

type Error struct {
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err == nil {
		return e.Code
	}
	return fmt.Sprintf("%s: %v", e.Code, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func NewError(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

func WrapError(code, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

func Validation(message string) *Error {
	if message == "" {
		message = "Некорректные входные данные"
	}
	return NewError(CodeValidation, message)
}

func Unauthorized(message string) *Error {
	if message == "" {
		message = "Требуется авторизация"
	}
	return NewError(CodeUnauthorized, message)
}

func NotFound(message string) *Error {
	if message == "" {
		message = "Ресурс не найден"
	}
	return NewError(CodeNotFound, message)
}

func Conflict(message string) *Error {
	if message == "" {
		message = "Конфликт данных"
	}
	return NewError(CodeConflict, message)
}

func AIUnavailable(err error) *Error {
	return WrapError(CodeAIUnavailable, "AI-сервис недоступен", err)
}

func TaskUnavailable(message string) *Error {
	if message == "" {
		message = "Готовая задача для этой категории еще готовится"
	}
	return NewError(CodeTaskUnavailable, message)
}

func DBUnavailable(err error) *Error {
	return WrapError(CodeDBUnavailable, "База данных недоступна", err)
}

func RateLimited() *Error {
	return NewError(CodeRateLimited, "Слишком много запросов, попробуйте позже")
}

func Internal(err error) *Error {
	return WrapError(CodeInternal, "Внутренняя ошибка сервера", err)
}
