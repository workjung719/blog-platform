package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"blog-platform/internal/middleware"
	"blog-platform/internal/model"
	"blog-platform/internal/respond"
	"blog-platform/internal/service"

	"github.com/go-chi/chi/v5"
)

type CommentHandler struct {
	comments *service.CommentService
	logger   *slog.Logger
}

func NewCommentHandler(comments *service.CommentService, logger *slog.Logger) *CommentHandler {
	return &CommentHandler{comments: comments, logger: logger}
}

func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.GetUserID(r)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	postID, err := strconv.Atoi(chi.URLParam(r, "postId"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid post id")
		return
	}
	var req model.CreateCommentRequest
	if err := respond.ParseJSON(r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	comment, err := h.comments.Create(r.Context(), postID, uid, req)
	if err != nil {
		status, msg := respond.MapError(err, h.logger)
		respond.Error(w, status, msg)
		return
	}
	respond.JSON(w, http.StatusCreated, comment)
}

func (h *CommentHandler) ListByPost(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(chi.URLParam(r, "postId"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid post id")
		return
	}
	comments, err := h.comments.GetByPostID(r.Context(), postID)
	if err != nil {
		status, msg := respond.MapError(err, h.logger)
		respond.Error(w, status, msg)
		return
	}
	if comments == nil {
		comments = []model.Comment{}
	}
	respond.JSON(w, http.StatusOK, comments)
}

func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.GetUserID(r)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.Atoi(chi.URLParam(r, "commentId"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid comment id")
		return
	}
	var req model.UpdateCommentRequest
	if err := respond.ParseJSON(r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	comment, err := h.comments.Update(r.Context(), id, uid, req)
	if err != nil {
		status, msg := respond.MapError(err, h.logger)
		respond.Error(w, status, msg)
		return
	}
	respond.JSON(w, http.StatusOK, comment)
}

func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.GetUserID(r)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.Atoi(chi.URLParam(r, "commentId"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid comment id")
		return
	}
	if err := h.comments.Delete(r.Context(), id, uid); err != nil {
		status, msg := respond.MapError(err, h.logger)
		respond.Error(w, status, msg)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"message": "comment deleted"})
}
