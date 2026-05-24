package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
)

type successResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

type errorResponse struct {
	Success bool        `json:"success"`
	Error   errorObject `json:"error"`
}

type errorObject struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeSuccess(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(successResponse{
		Success: true,
		Data:    data,
	})
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	code := app.CodeInternal
	message := "Внутренняя ошибка сервера"
	requestID := ""
	if r != nil {
		requestID = requestIDFromContext(r.Context())
	}

	var appErr *app.Error
	if errors.As(err, &appErr) {
		code = appErr.Code
		message = appErr.Message
		status = statusForCode(appErr.Code)
		if appErr.Err != nil {
			log.Printf("request failed: request_id=%s code=%s err=%v", requestID, appErr.Code, appErr.Err)
		}
	} else if err != nil {
		log.Printf("request failed: request_id=%s err=%v", requestID, err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{
		Success: false,
		Error: errorObject{
			Code:    code,
			Message: message,
		},
	})
}

func statusForCode(code string) int {
	switch code {
	case app.CodeValidation:
		return http.StatusBadRequest
	case app.CodeUnauthorized:
		return http.StatusUnauthorized
	case app.CodeNotFound:
		return http.StatusNotFound
	case app.CodeConflict:
		return http.StatusConflict
	case app.CodeAIUnavailable, app.CodeTaskUnavailable, app.CodeDBUnavailable:
		return http.StatusServiceUnavailable
	case app.CodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
