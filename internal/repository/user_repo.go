package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"blog-platform/internal/model"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type UserRepo struct{ db *sqlx.DB }

func NewUserRepo(db *sqlx.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, email, username, passwordHash string) (*model.User, error) {
	const q = `INSERT INTO users (email, username, password_hash)
	           VALUES ($1, $2, $3)
	           RETURNING id, email, username, created_at`
	user := &model.User{}
	err := r.db.QueryRowxContext(ctx, q, email, username, passwordHash).StructScan(user)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return nil, fmt.Errorf("%w: email or username already taken", model.ErrAlreadyExists)
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	const q = `SELECT id, email, username, password_hash, created_at FROM users WHERE email = $1`
	user := &model.User{}
	if err := r.db.GetContext(ctx, user, q, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int) (*model.User, error) {
	const q = `SELECT id, email, username, password_hash, created_at FROM users WHERE id = $1`
	user := &model.User{}
	if err := r.db.GetContext(ctx, user, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}