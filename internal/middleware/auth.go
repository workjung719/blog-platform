package middleware

import (
	"context"
	"net/http"
	"strings"

	"blog-platform/internal/respond"
	"blog-platform/internal/service"
)

// Типизированный ключ контекста — защита от коллизий с другими пакетами.
type contextKey string

const userIDKey contextKey = "user_id"

func AuthMiddleware(userService *service.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				respond.Error(w, http.StatusUnauthorized, "authorization header is required")
				return
			}
			tokenString := strings.TrimPrefix(header, "Bearer ")
			if tokenString == header { // префикс не найден
				respond.Error(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}
			claims, err := userService.ValidateToken(tokenString)
			if err != nil {
				respond.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(r *http.Request) (int, bool) {
	id, ok := r.Context().Value(userIDKey).(int)
	return id, ok
}