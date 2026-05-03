package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"oge-ai-trainer/backend/internal/app"
)

const maxJSONBodyBytes = 1 << 20

func decodeStrictJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	return decodeJSON(w, r, dst, true)
}

func decodeAuthJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	return decodeJSON(w, r, dst, false)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any, requireContentType bool) error {
	contentType := r.Header.Get("Content-Type")
	if requireContentType && contentType == "" {
		return app.Validation("Content-Type должен быть application/json")
	}
	if contentType != "" && !strings.Contains(strings.ToLower(contentType), "application/json") {
		return app.Validation("Content-Type должен быть application/json")
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return jsonDecodeError(err)
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return app.Validation("JSON должен содержать только один объект")
	}
	return nil
}

func jsonDecodeError(err error) error {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError

	switch {
	case errors.As(err, &syntaxErr):
		return app.Validation("Некорректный JSON")
	case errors.As(err, &typeErr):
		return app.Validation("Некорректный тип поля " + typeErr.Field)
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		return app.Validation(fmt.Sprintf("Неизвестное поле %s", strings.TrimPrefix(err.Error(), "json: unknown field ")))
	case errors.Is(err, io.EOF):
		return app.Validation("Тело запроса не должно быть пустым")
	case strings.Contains(err.Error(), "http: request body too large"):
		return app.Validation("JSON слишком большой")
	default:
		return app.Validation("Некорректные входные данные")
	}
}
