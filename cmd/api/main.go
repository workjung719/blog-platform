package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blog-platform/internal/config"
	"blog-platform/internal/handler"
	"blog-platform/internal/middleware"
	"blog-platform/internal/repository"
	"blog-platform/internal/scheduler"
	"blog-platform/internal/service"
	"blog-platform/migrations"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config_load_failed", "error", err)
		os.Exit(1)
	}

	db, err := sqlx.Connect("postgres", cfg.DB.ConnectionString())
	if err != nil {
		logger.Error("db_connect_failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database_connected")

	if err := migrations.Apply(db); err != nil {
		logger.Error("migration_failed", "error", err)
		os.Exit(1)
	}
	logger.Info("migrations_applied")

	// Repositories
	userRepo := repository.NewUserRepo(db)
	postRepo := repository.NewPostRepo(db)
	commentRepo := repository.NewCommentRepo(db)

	// Services
	userService := service.NewUserService(userRepo, cfg.JWT)
	postService := service.NewPostService(postRepo)
	// Передаем userRepo в CommentService для будущей проверки прав (если понадобится)
	commentService := service.NewCommentService(commentRepo, postRepo, userRepo)

	// Handlers
	userHandler := handler.NewUserHandler(userService, logger)
	postHandler := handler.NewPostHandler(postService, logger)
	commentHandler := handler.NewCommentHandler(commentService, logger)
	healthHandler := handler.NewHealthHandler(db)

	r := chi.NewRouter()
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.RecoverMiddleware(logger))

	// Public Routes
	r.Get("/api/health", healthHandler.Check)
	r.Post("/api/register", userHandler.Register)
	r.Post("/api/login", userHandler.Login)
	r.Get("/api/posts", postHandler.List)
	r.Get("/api/posts/{id}", postHandler.GetByID)
	r.Get("/api/users/{userId}/posts", postHandler.GetByAuthor)
	r.Get("/api/posts/{postId}/comments", commentHandler.ListByPost)

	// Protected Routes
	r.Group(func(pr chi.Router) {
		pr.Use(middleware.AuthMiddleware(userService))
		pr.Post("/api/posts", postHandler.Create)
		pr.Put("/api/posts/{id}", postHandler.Update)
		pr.Delete("/api/posts/{id}", postHandler.Delete)

		pr.Post("/api/posts/{postId}/comments", commentHandler.Create)
		pr.Put("/api/comments/{commentId}", commentHandler.Update)
		pr.Delete("/api/comments/{commentId}", commentHandler.Delete)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	publisher := scheduler.NewPublisher(postService, cfg.Sched.IntervalSec, cfg.Sched.Workers, logger)
	go publisher.Run(ctx)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server_starting", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server_error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutdown_signal_received")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server_shutdown_error", "error", err)
	}
	logger.Info("server_stopped_gracefully")
}
