package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"blog-platform/internal/config"
	"blog-platform/internal/model"
	"blog-platform/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type UserService struct {
	repo repository.UserRepository
	jwt  config.JWTConfig
}

func NewUserService(repo repository.UserRepository, jwtCfg config.JWTConfig) *UserService {
	return &UserService{repo: repo, jwt: jwtCfg}
}

func (s *UserService) Register(ctx context.Context, req model.RegisterRequest) (*model.User, string, error) {
	if err := req.Validate(); err != nil {
		return nil, "", fmt.Errorf("%w: %v", model.ErrInvalidInput, err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}
	user, err := s.repo.Create(ctx, req.Email, req.Username, string(hash))
	if err != nil {
		return nil, "", err // ErrAlreadyExists пробрасывается как есть
	}
	token, err := s.generateToken(*user)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *UserService) Login(ctx context.Context, req model.LoginRequest) (*model.User, string, error) {
	if err := req.Validate(); err != nil {
		return nil, "", fmt.Errorf("%w: %v", model.ErrInvalidInput, err)
	}
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			// Не раскрываем существование email — единое сообщение с неверным паролем.
			return nil, "", model.ErrUnauthorized
		}
		return nil, "", err // реальная ошибка БД -> 500 (не маскируем под 401)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, "", model.ErrUnauthorized
	}
	token, err := s.generateToken(*user)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

// ValidateToken строго проверяет алгоритм HS256 (принцип минимально разрешённой конфигурации).
func (s *UserService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing algorithm: %v", t.Method.Alg())
		}
		return s.jwt.Secret, nil
	})
	if err != nil || !token.Valid {
		return nil, model.ErrUnauthorized
	}
	return claims, nil
}

func (s *UserService) generateToken(user model.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwt.ExpirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwt.Secret)
}

func (s *UserService) GetByID(ctx context.Context, id int) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}