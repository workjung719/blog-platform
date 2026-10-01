package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"blog-platform/internal/model"
	"blog-platform/internal/repository"
)

// mockPostRepo — мок репозитория постов для изолированных unit-тестов сервиса.
type mockPostRepo struct {
	created       *model.Post
	anyByID       *model.Post
	anyErr        error
	updatedCalled bool
	deletedCalled bool
	updateErr     error
}

func (m *mockPostRepo) Create(_ context.Context, post *model.Post) (*model.Post, error) {
	m.created = post
	out := *post
	out.ID = 1
	return &out, nil
}
func (m *mockPostRepo) GetByID(context.Context, int) (*model.Post, error)        { return nil, nil }
func (m *mockPostRepo) GetByIDAny(context.Context, int) (*model.Post, error)     { return m.anyByID, m.anyErr }
func (m *mockPostRepo) GetAll(_ context.Context, limit, offset int) ([]model.Post, int, error) {
	return []model.Post{}, 0, nil
}
func (m *mockPostRepo) Update(context.Context, int, model.UpdatePostRequest) (*model.Post, error) {
	m.updatedCalled = true
	return &model.Post{ID: 1}, m.updateErr
}
func (m *mockPostRepo) Delete(context.Context, int) error {
	m.deletedCalled = true
	return nil
}
func (m *mockPostRepo) GetDuePosts(context.Context, int) ([]model.Post, error) { return nil, nil }
func (m *mockPostRepo) Publish(context.Context, int) error                     { return nil }

var _ repository.PostRepository = (*mockPostRepo)(nil)

// Тест корректности СТАТУСА публикации (draft vs published).
func TestPostService_Create_Status(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name        string
		publishAt   *time.Time
		wantStatus  string
	}{
		{"без publish_at -> published", nil, "published"},
		{"publish_at в прошлом -> published", &past, "published"},
		{"publish_at в будущем -> draft (отложенная)", &future, "draft"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPostRepo{}
			svc := NewPostService(repo)
			_, err := svc.Create(context.Background(), 7, model.CreatePostRequest{
				Title: "t", Content: "c", PublishAt: tt.publishAt,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if repo.created.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", repo.created.Status, tt.wantStatus)
			}
			if repo.created.UserID != 7 {
				t.Errorf("user_id = %d, want 7", repo.created.UserID)
			}
		})
	}
}

// Тест прав доступа: чужой пост -> 403 (ErrForbidden), репозиторий не трогается.
func TestPostService_Update_Forbidden(t *testing.T) {
	repo := &mockPostRepo{anyByID: &model.Post{ID: 1, UserID: 100}}
	svc := NewPostService(repo)

	_, err := svc.Update(context.Background(), 1, 999, model.UpdatePostRequest{Title: "x", Content: "y"})
	if !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("want ErrForbidden, got %v", err)
	}
	if repo.updatedCalled {
		t.Error("repo.Update не должен вызываться для чужого поста")
	}
}

func TestPostService_Delete_Forbidden(t *testing.T) {
	repo := &mockPostRepo{anyByID: &model.Post{ID: 1, UserID: 100}}
	svc := NewPostService(repo)

	err := svc.Delete(context.Background(), 1, 999)
	if !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("want ErrForbidden, got %v", err)
	}
	if repo.deletedCalled {
		t.Error("repo.Delete не должен вызываться для чужого поста")
	}
}

func TestPostService_Update_OwnerAllowed(t *testing.T) {
	repo := &mockPostRepo{anyByID: &model.Post{ID: 1, UserID: 42}}
	svc := NewPostService(repo)

	if _, err := svc.Update(context.Background(), 1, 42, model.UpdatePostRequest{Title: "x", Content: "y"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.updatedCalled {
		t.Error("repo.Update должен вызываться для своего поста")
	}
}

// Тест нормализации пагинации.
func TestPostService_GetAll_PaginationNormalization(t *testing.T) {
	svc := NewPostService(&mockPostRepo{})
	cases := []struct{ inLimit, inOffset, wantLimit, wantOffset int }{
		{0, 0, 10, 0},
		{-5, -5, 10, 0},
		{200, 10, 100, 10},
		{25, 5, 25, 5},
	}
	for _, c := range cases {
		resp, err := svc.GetAll(context.Background(), c.inLimit, c.inOffset)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Limit != c.wantLimit || resp.Offset != c.wantOffset {
			t.Errorf("limit/offset = %d/%d, want %d/%d", resp.Limit, resp.Offset, c.wantLimit, c.wantOffset)
		}
	}
}

func TestPostService_Create_InvalidInput(t *testing.T) {
	svc := NewPostService(&mockPostRepo{})
	_, err := svc.Create(context.Background(), 1, model.CreatePostRequest{Title: "", Content: ""})
	if !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}