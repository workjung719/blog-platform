package model

import (
	"errors"
	"net/mail"
	"strings"
	"time"
)

// Sentinel-ошибки домена — распознаются через errors.Is в центральной обработке.
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrForbidden     = errors.New("forbidden")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrInvalidInput  = errors.New("invalid input")
)

type User struct {
	ID           int       `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"` // никогда не сериализуется
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Post struct {
	ID        int        `json:"id" db:"id"`
	UserID    int        `json:"user_id" db:"user_id"`
	Title     string     `json:"title" db:"title"`
	Content   string     `json:"content" db:"content"`
	Status    string     `json:"status" db:"status"`
	PublishAt *time.Time `json:"publish_at,omitempty" db:"publish_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

type Comment struct {
	ID        int       `json:"id" db:"id"`
	PostID    int       `json:"post_id" db:"post_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Username  string    `json:"username,omitempty" db:"username"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// --- DTO запросов ---

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreatePostRequest struct {
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	PublishAt *time.Time `json:"publish_at,omitempty"`
}

type UpdatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CreateCommentRequest struct {
	Content string `json:"content"`
}

type PaginatedResponse struct {
	Items  interface{} `json:"items"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

// --- Валидация ---

// ValidateEmail выполняет реальную проверку: корректность по RFC + точка в домене.
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("invalid email format")
	}
	at := strings.LastIndex(email, "@")
	domain := email[at+1:]
	if !strings.Contains(domain, ".") {
		return errors.New("invalid email format: domain must contain a dot")
	}
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") || strings.Contains(domain, "..") {
		return errors.New("invalid email format: invalid domain")
	}
	return nil
}

func (r *RegisterRequest) Validate() error {
	if err := ValidateEmail(r.Email); err != nil {
		return err
	}
	if r.Username == "" {
		return errors.New("username is required")
	}
	// 3..50 — соответствует VARCHAR(50) в схеме БД (защита от 500 на вставке)
	if len(r.Username) < 3 || len(r.Username) > 50 {
		return errors.New("username must be between 3 and 50 characters")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

func (r *LoginRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

func (r *CreatePostRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	if len(r.Title) > 255 {
		return errors.New("title must be at most 255 characters")
	}
	if strings.TrimSpace(r.Content) == "" {
		return errors.New("content is required")
	}
	return nil
}

func (r *UpdatePostRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(r.Content) == "" {
		return errors.New("content is required")
	}
	return nil
}

func (r *CreateCommentRequest) Validate() error {
	if strings.TrimSpace(r.Content) == "" {
		return errors.New("content is required")
	}
	return nil
}