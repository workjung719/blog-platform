package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"blog-platform/internal/config"
	"blog-platform/internal/middleware"
	"blog-platform/internal/model"
	"blog-platform/internal/repository"
	"blog-platform/internal/service"

	"github.com/golang-jwt/jwt/v5"
)

// mockUserRepo не используется в ValidateToken, но нужен для конструктора сервиса.
type mockUserRepo struct{}

func (mockUserRepo) Create(context.Context, string, string, string) (*model.User, error) { return nil, nil }
func (mockUserRepo) GetByEmail(context.Context, string) (*model.User, error)             { return nil, nil }
func (mockUserRepo) GetByID(context.Context, int) (*model.User, error)                   { return nil, nil }

var _ repository.UserRepository = mockUserRepo{}

var secret = []byte("unit-test-secret-key-that-is-long-enough-32-chars")

func newUserService() *service.UserService {
	return service.NewUserService(mockUserRepo{}, config.JWTConfig{Secret: secret, ExpirationHours: 1})
}

func signToken(t *testing.T, userID int) string {
	t.Helper()
	claims := &service.Claims{
		UserID:   userID,
		Email:    "u@e.com",
		Username: "u",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return s
}

// next-обработчик, запоминающий userID из контекста.
func captureNext(got *int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, ok := middleware.GetUserID(r); ok {
			*got = id
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	mw := middleware.AuthMiddleware(newUserService())
	got := 0
	
	// Создаем next-обработчик, который запишет userID в переменную got
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, ok := middleware.GetUserID(r); ok {
			got = id
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/posts", nil)
	req.Header.Set("Authorization", "Bearer "+signToken(t, 42))
	rec := httptest.NewRecorder()

	// Правильный вызов: mw(nextHandler) вернет http.Handler, у которого есть ServeHTTP
	handler := mw(nextHandler)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got != 42 {
		t.Errorf("userID from context = %d, want 42", got)
	}
}

func TestAuthMiddleware_Rejects(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"нет заголовка", ""},
		{"неверный формат (Basic)", "Basic abc"},
		{"битый токен", "Bearer not.a.jwt"},
		{"подписан другим ключом", "Bearer " + foreignToken(t)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := middleware.AuthMiddleware(newUserService())
			got := 0
			req := httptest.NewRequest(http.MethodGet, "/api/posts", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			mw(captureNext(&got)).ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", rec.Code)
			}
			if got != 0 {
				t.Error("next не должен был отработать")
			}
		})
	}
}

func foreignToken(t *testing.T) string {
	t.Helper()
	claims := &service.Claims{UserID: 1, RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}}
	s, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("a-completely-different-secret-key-32-chars!!"))
	return s
}