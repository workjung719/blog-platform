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
}

func NewCommentService(c repository.CommentRepository, p repository.PostRepository) *CommentService {
	return &CommentService{comments: c, posts: p}
}

func (s *CommentService) Create(ctx context.Context, postID, userID int, req model.CreateCommentRequest) (*model.Comment, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", model.ErrInvalidInput, err)
	}
	// Комментировать можно только опубликованный пост.
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