package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"blog-platform/internal/model"
)

// PostPublisher — узкая абстракция для планировщика (DIP):
// позволяет тестировать логику публикации без БД.
type PostPublisher interface {
	GetDuePosts(ctx context.Context, limit int) ([]model.Post, error)
	Publish(ctx context.Context, id int) error
}

type Publisher struct {
	posts     PostPublisher
	interval  time.Duration
	workers   int
	batchSize int
	logger    *slog.Logger
}

func NewPublisher(posts PostPublisher, intervalSec, workers int, logger *slog.Logger) *Publisher {
	return &Publisher{
		posts:     posts,
		interval:  time.Duration(intervalSec) * time.Second,
		workers:   workers,
		batchSize: 100,
		logger:    logger,
	}
}

// Run крутит ticker до отмены контекста (graceful shutdown).
func (p *Publisher) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	p.logger.Info("scheduler_started", "interval", p.interval.String(), "workers", p.workers)
	for {
		select {
		case <-ctx.Done():
			p.logger.Info("scheduler_stopped_gracefully")
			return
		case <-ticker.C:
			p.processDuePosts(ctx)
		}
	}
}

// processDuePosts распределяет посты по worker pool и публикует конкурентно.
func (p *Publisher) processDuePosts(ctx context.Context) {
	posts, err := p.posts.GetDuePosts(ctx, p.batchSize)
	if err != nil {
		p.logger.Error("scheduler_get_due_posts_failed", "error", err)
		return
	}
	if len(posts) == 0 {
		return
	}
	p.logger.Info("scheduler_found_posts_to_publish", "count", len(posts))

	jobs := make(chan model.Post, len(posts))
	var wg sync.WaitGroup

	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for post := range jobs {
				if ctx.Err() != nil {
					return
				}
				if err := p.posts.Publish(ctx, post.ID); err != nil {
					p.logger.Error("scheduler_publish_failed", "post_id", post.ID, "worker", id, "error", err)
					continue
				}
				p.logger.Info("post_published", "post_id", post.ID, "worker", id)
			}
		}(i)
	}

	for _, post := range posts {
		jobs <- post
	}
	close(jobs)
	wg.Wait()
}