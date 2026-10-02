package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"blog-platform/internal/model"

	"github.com/jmoiron/sqlx"
)

type PostRepo struct{ db *sqlx.DB }

func NewPostRepo(db *sqlx.DB) *PostRepo { return &PostRepo{db: db} }

const postCols = `id, user_id, title, content, status, publish_at, created_at, updated_at`

func (r *PostRepo) Create(ctx context.Context, post *model.Post) (*model.Post, error) {
	const q = `INSERT INTO posts (user_id, title, content, status, publish_at)
	           VALUES ($1, $2, $3, $4, $5)
	           RETURNING ` + postCols
	res := &model.Post{}
	err := r.db.QueryRowxContext(ctx, q, post.UserID, post.Title, post.Content, post.Status, post.PublishAt).StructScan(res)
	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}
	return res, nil
}

func (r *PostRepo) GetByID(ctx context.Context, id int) (*model.Post, error) {
	const q = `SELECT ` + postCols + ` FROM posts WHERE id = $1 AND status = 'published'`
	return r.getOne(ctx, q, id)
}

func (r *PostRepo) GetByIDAny(ctx context.Context, id int) (*model.Post, error) {
	const q = `SELECT ` + postCols + ` FROM posts WHERE id = $1`
	return r.getOne(ctx, q, id)
}

func (r *PostRepo) getOne(ctx context.Context, q string, id int) (*model.Post, error) {
	post := &model.Post{}
	if err := r.db.GetContext(ctx, post, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get post: %w", err)
	}
	return post, nil
}

func (r *PostRepo) GetAll(ctx context.Context, limit, offset int) ([]model.Post, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM posts WHERE status = 'published'`); err != nil {
		return nil, 0, fmt.Errorf("count posts: %w", err)
	}
	const q = `SELECT ` + postCols + ` FROM posts WHERE status = 'published'
	           ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	posts := []model.Post{}
	if err := r.db.SelectContext(ctx, &posts, q, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("list posts: %w", err)
	}
	return posts, total, nil
}

func (r *PostRepo) Update(ctx context.Context, id int, req model.UpdatePostRequest) (*model.Post, error) {
	const q = `UPDATE posts SET title = $1, content = $2, updated_at = NOW()
	           WHERE id = $3 RETURNING ` + postCols
	post := &model.Post{}
	if err := r.db.QueryRowxContext(ctx, q, req.Title, req.Content, id).StructScan(post); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("update post: %w", err)
	}
	return post, nil
}

func (r *PostRepo) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM posts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *PostRepo) GetDuePosts(ctx context.Context, limit int) ([]model.Post, error) {
	const q = `SELECT ` + postCols + ` FROM posts
	           WHERE status = 'draft' AND publish_at IS NOT NULL AND publish_at <= NOW()
	           ORDER BY publish_at ASC LIMIT $1`
	posts := []model.Post{}
	if err := r.db.SelectContext(ctx, &posts, q, limit); err != nil {
		return nil, fmt.Errorf("get due posts: %w", err)
	}
	return posts, nil
}

func (r *PostRepo) Publish(ctx context.Context, id int) error {
	if _, err := r.db.ExecContext(ctx,
		`UPDATE posts SET status = 'published', updated_at = NOW() WHERE id = $1`, id); err != nil {
		return fmt.Errorf("publish post: %w", err)
	}
	return nil
}

func (r *PostRepo) GetByAuthorID(ctx context.Context, authorID, limit, offset int) ([]model.Post, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM posts WHERE user_id = $1 AND status = 'published'`
	if err := r.db.GetContext(ctx, &total, countQuery, authorID); err != nil {
		return nil, 0, fmt.Errorf("count posts by author: %w", err)
	}

	const q = `SELECT ` + postCols + ` FROM posts 
	           WHERE user_id = $1 AND status = 'published' 
	           ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	posts := []model.Post{}
	if err := r.db.SelectContext(ctx, &posts, q, authorID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("list posts by author: %w", err)
	}
	return posts, total, nil
}
