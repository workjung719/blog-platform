package respond

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"blog-platform/internal/model"
)

// Response — единый формат JSON-ответа для всего API.
type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Response{Data: data})
}

func Error(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Response{Error: message})
}

// ParseJSON строго разбирает тело запроса (неизвестные поля отклоняются).
func ParseJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return errors.New("request body is empty")
	}
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// MapError — центральная обработка ошибок: доменное исключение -> HTTP-статус.
// Неизвестные ошибки логируются и маскируются в 500 (не раскрываем internals клиенту).
func MapError(err error, logger *slog.Logger) (int, string) {
	switch {
	case errors.Is(err, model.ErrNotFound):
		return http.StatusNotFound, "not found"
	case errors.Is(err, model.ErrAlreadyExists):
		return http.StatusConflict, "resource already exists"
	case errors.Is(err, model.ErrForbidden):
		return http.StatusForbidden, "access denied"
	case errors.Is(err, model.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, model.ErrInvalidInput):
		return http.StatusBadRequest, err.Error()
	default:
		logger.Error("unhandled error", "error", err)
		return http.StatusInternalServerError, "internal server error"
	}
}