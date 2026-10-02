package service

import (
	"context"
	"errors"
	"fmt"

	"blog-platform/internal/model"
	"blog-platform/internal/repository"
)

type CommentService struct {
	comments repository.CommentRepository
	posts    repository.PostRepository
	users    repository.UserRepository
}

func NewCommentService(c repository.CommentRepository, p repository.PostRepository, u repository.UserRepository) *CommentService {
	return &CommentService{comments: c, posts: p, users: u}
}

func (s *CommentService) Create(ctx context.Context, postID, userID int, req model.CreateCommentRequest) (*model.Comment, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", model.ErrInvalidInput, err)
	}
	// Проверяем существование поста
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, fmt.Errorf("%w: post", model.ErrNotFound)
		}
		return nil, err
	}
	return s.comments.Create(ctx, postID, userID, req.Content)
}

func (s *CommentService) GetByPostID(ctx context.Context, postID int) ([]model.Comment, error) {
	return s.comments.GetByPostID(ctx, postID)
}

func (s *CommentService) Update(ctx context.Context, id, userID int, req model.UpdateCommentRequest) (*model.Comment, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", model.ErrInvalidInput, err)
	}
	comment, err := s.comments.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Проверка прав: может редактировать только автор комментария
	if comment.UserID != userID {
		return nil, model.ErrForbidden
	}
	return s.comments.Update(ctx, id, req.Content)
}

func (s *CommentService) Delete(ctx context.Context, id, userID int) error {
	comment, err := s.comments.GetByID(ctx, id)
	if err != nil {
		return err
	}
	// Проверка прав: может удалить только автор комментария
	if comment.UserID != userID {
		return model.ErrForbidden
	}
	return s.comments.Delete(ctx, id)
}
