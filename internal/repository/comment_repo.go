package repository

import (
	"context"
	"fmt"

	"blog-platform/internal/model"

	"github.com/jmoiron/sqlx"
)

type CommentRepo struct{ db *sqlx.DB }

func NewCommentRepo(db *sqlx.DB) *CommentRepo { return &CommentRepo{db: db} }

func (r *CommentRepo) Create(ctx context.Context, postID, userID int, content string) (*model.Comment, error) {
	const q = `INSERT INTO comments (post_id, user_id, content)
	           VALUES ($1, $2, $3)
	           RETURNING id, post_id, user_id, content, created_at`
	c := &model.Comment{}
	if err := r.db.QueryRowxContext(ctx, q, postID, userID, content).StructScan(c); err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}
	return c, nil
}

func (r *CommentRepo) GetByPostID(ctx context.Context, postID int) ([]model.Comment, error) {
	const q = `SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, u.username
	           FROM comments c
	           JOIN users u ON u.id = c.user_id
	           WHERE c.post_id = $1
	           ORDER BY c.created_at DESC`
	comments := []model.Comment{}
	if err := r.db.SelectContext(ctx, &comments, q, postID); err != nil {
		return nil, fmt.Errorf("get comments: %w", err)
	}
	return comments, nil
}