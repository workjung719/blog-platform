package repository

import (
	"context"

	"blog-platform/internal/model"
)

// Контракты репозиториев. Сервисы зависят от абстракций (DIP),
// что позволяет подменять их моками в unit-тестах без реальной БД.

type UserRepository interface {
	Create(ctx context.Context, email, username, passwordHash string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id int) (*model.User, error)
}

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) (*model.Post, error)
	GetByID(ctx context.Context, id int) (*model.Post, error)    // только published
	GetByIDAny(ctx context.Context, id int) (*model.Post, error) // любой статус
	GetAll(ctx context.Context, limit, offset int) ([]model.Post, int, error)
	Update(ctx context.Context, id int, req model.UpdatePostRequest) (*model.Post, error)
	Delete(ctx context.Context, id int) error
	GetDuePosts(ctx context.Context, limit int) ([]model.Post, error)
	Publish(ctx context.Context, id int) error
	GetByAuthorID(ctx context.Context, authorID, limit, offset int) ([]model.Post, int, error)
}

type CommentRepository interface {
	Create(ctx context.Context, postID, userID int, content string) (*model.Comment, error)
	GetByPostID(ctx context.Context, postID int) ([]model.Comment, error)
	GetByID(ctx context.Context, id int) (*model.Comment, error)
	Update(ctx context.Context, id int, content string) (*model.Comment, error)
	Delete(ctx context.Context, id int) error
}
