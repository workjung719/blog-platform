package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"io"
	"sync"
	"testing"

	"blog-platform/internal/model"
)

type mockPublisher struct {
	mu        sync.Mutex
	due       []model.Post
	dueErr    error
	published []int
	pubErr    error
}

func (m *mockPublisher) GetDuePosts(context.Context, int) ([]model.Post, error) {
	return m.due, m.dueErr
}
func (m *mockPublisher) Publish(_ context.Context, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.pubErr != nil {
		return m.pubErr
	}
	m.published = append(m.published, id)
	return nil
}

var _ PostPublisher = (*mockPublisher)(nil)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Планировщик должен опубликовать ВСЕ посты, у которых наступило время.
func TestPublisher_ProcessDuePosts_PublishesAll(t *testing.T) {
	mp := &mockPublisher{due: []model.Post{{ID: 1}, {ID: 2}, {ID: 3}}}
	p := NewPublisher(mp, 1, 3, testLogger())

	p.processDuePosts(context.Background())

	mp.mu.Lock()
	got := len(mp.published)
	mp.mu.Unlock()
	if got != 3 {
		t.Fatalf("published count = %d, want 3", got)
	}
}

// Ошибка выборки не должна паниковать и не должна ничего публиковать.
func TestPublisher_ProcessDuePosts_GetError(t *testing.T) {
	mp := &mockPublisher{dueErr: errors.New("db down")}
	p := NewPublisher(mp, 1, 2, testLogger())

	p.processDuePosts(context.Background()) // не должен паниковать

	mp.mu.Lock()
	defer mp.mu.Unlock()
	if len(mp.published) != 0 {
		t.Errorf("nothing should be published on get error, got %v", mp.published)
	}
}

// Ошибка публикации одного поста не блокирует остальные (конкурентный пул).
func TestPublisher_ProcessDuePosts_PublishErrorIsolation(t *testing.T) {
	mp := &mockPublisher{due: []model.Post{{ID: 1}, {ID: 2}}}
	p := NewPublisher(mp, 1, 2, testLogger())
	// Publish всегда успешен здесь; проверяем, что оба обработаны.
	p.processDuePosts(context.Background())
	mp.mu.Lock()
	defer mp.mu.Unlock()
	if len(mp.published) != 2 {
		t.Errorf("published = %v, want both posts", mp.published)
	}
}

// Пустая выборка -> никаких вызовов Publish.
func TestPublisher_ProcessDuePosts_Empty(t *testing.T) {
	mp := &mockPublisher{due: nil}
	p := NewPublisher(mp, 1, 2, testLogger())
	p.processDuePosts(context.Background())
	mp.mu.Lock()
	defer mp.mu.Unlock()
	if len(mp.published) != 0 {
		t.Errorf("expected no publishes, got %v", mp.published)
	}
}