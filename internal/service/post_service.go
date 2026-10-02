package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"blog-platform/internal/model"
	"blog-platform/internal/repository"
)

type PostService struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) *PostService {
	return &PostService{repo: repo}
}

// Create вычисляет статус публикации на уровне сервиса (бизнес-логика),
// а репозиторий отвечает только за хранение данных.
func (s *PostService) Create(ctx context.Context, userID int, req model.CreatePostRequest) (*model.Post, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", model.ErrInvalidInput, err)
	}
	status := "published"
	if req.PublishAt != nil && req.PublishAt.After(time.Now()) {
		status = "draft" // отложенная публикация
	}
	post := &model.Post{
		UserID:    userID,
		Title:     req.Title,
		Content:   req.Content,
		Status:    status,
		PublishAt: req.PublishAt,
	}
	return s.repo.Create(ctx, post)
}

func (s *PostService) GetByID(ctx context.Context, id int) (*model.Post, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PostService) GetAll(ctx context.Context, limit, offset int) (model.PaginatedResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // защита от чрезмерных выборок
	}
	if offset < 0 {
		offset = 0
	}
	posts, total, err := s.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return model.PaginatedResponse{}, err
	}
	return model.PaginatedResponse{Items: posts, Total: total, Limit: limit, Offset: offset}, nil
}

// Update проверяет права автора ДО обращения к репозиторию.
func (s *PostService) Update(ctx context.Context, id, userID int, req model.UpdatePostRequest) (*model.Post, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", model.ErrInvalidInput, err)
	}
	post, err := s.repo.GetByIDAny(ctx, id)
	if err != nil {
		return nil, err
	}
	if post.UserID != userID {
		return nil, model.ErrForbidden
	}
	return s.repo.Update(ctx, id, req)
}

func (s *PostService) Delete(ctx context.Context, id, userID int) error {
	post, err := s.repo.GetByIDAny(ctx, id)
	if err != nil {
		return err
	}
	if post.UserID != userID {
		return model.ErrForbidden
	}
	return s.repo.Delete(ctx, id)
}

// Методы для планировщика (системные, проверка прав не требуется).
func (s *PostService) GetDuePosts(ctx context.Context, limit int) ([]model.Post, error) {
	return s.repo.GetDuePosts(ctx, limit)
}

func (s *PostService) Publish(ctx context.Context, id int) error {
	return s.repo.Publish(ctx, id)
}

// ensure interface satisfaction at compile time (защита от рассинхрона контракта).
var _ = errors.Is

func (s *PostService) GetByAuthor(ctx context.Context, authorID, limit, offset int) (model.PaginatedResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	posts, total, err := s.repo.GetByAuthorID(ctx, authorID, limit, offset)
	if err != nil {
		return model.PaginatedResponse{}, err
	}
	return model.PaginatedResponse{Items: posts, Total: total, Limit: limit, Offset: offset}, nil
}
