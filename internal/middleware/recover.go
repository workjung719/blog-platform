package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"blog-platform/internal/respond"
)

func RecoverMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic_recovered", "panic", rec, "stack", string(debug.Stack()))
					respond.Error(w, http.StatusInternalServerError, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}