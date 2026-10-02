package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"blog-platform/internal/model"

	"github.com/jmoiron/sqlx"
)

type CommentRepo struct{ db *sqlx.DB }

func NewCommentRepo(db *sqlx.DB) *CommentRepo { return &CommentRepo{db: db} }

func (r *CommentRepo) Create(ctx context.Context, postID, userID int, content string) (*model.Comment, error) {
	const q = `INSERT INTO comments (post_id, user_id, content) VALUES ($1, $2, $3) RETURNING id, post_id, user_id, content, created_at`
	c := &model.Comment{}
	if err := r.db.QueryRowxContext(ctx, q, postID, userID, content).StructScan(c); err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}
	return c, nil
}

func (r *CommentRepo) GetByPostID(ctx context.Context, postID int) ([]model.Comment, error) {
	const q = `SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, u.username 
	           FROM comments c JOIN users u ON u.id = c.user_id 
	           WHERE c.post_id = $1 ORDER BY c.created_at DESC`
	comments := []model.Comment{}
	if err := r.db.SelectContext(ctx, &comments, q, postID); err != nil {
		return nil, fmt.Errorf("get comments: %w", err)
	}
	return comments, nil
}

func (r *CommentRepo) GetByID(ctx context.Context, id int) (*model.Comment, error) {
	const q = `SELECT id, post_id, user_id, content, created_at FROM comments WHERE id = $1`
	c := &model.Comment{}
	if err := r.db.GetContext(ctx, c, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get comment by id: %w", err)
	}
	return c, nil
}

func (r *CommentRepo) Update(ctx context.Context, id int, content string) (*model.Comment, error) {
	const q = `UPDATE comments SET content = $1 WHERE id = $2 RETURNING id, post_id, user_id, content, created_at`
	c := &model.Comment{}
	if err := r.db.QueryRowxContext(ctx, q, content, id).StructScan(c); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("update comment: %w", err)
	}
	return c, nil
}

func (r *CommentRepo) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM comments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}
