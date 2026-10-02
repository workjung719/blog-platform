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

type PostHandler struct {
	posts  *service.PostService
	logger *slog.Logger
}

func NewPostHandler(posts *service.PostService, logger *slog.Logger) *PostHandler {
	return &PostHandler{posts: posts, logger: logger}
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.GetUserID(r)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req model.CreatePostRequest
	if err := respond.ParseJSON(r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	post, err := h.posts.Create(r.Context(), uid, req)
	if err != nil {
		status, msg := respond.MapError(err, h.logger)
		respond.Error(w, status, msg)
		return
	}
	respond.JSON(w, http.StatusCreated, post)
}

func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	resp, err := h.posts.GetAll(r.Context(), limit, offset)
	if err != nil {
		status, msg := respond.MapError(err, h.logger)
		respond.Error(w, status, msg)
		return
	}
	respond.JSON(w, http.StatusOK, resp)
}

func (h *PostHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid post id")
		return
	}
	post, err := h.posts.GetByID(r.Context(), id)
	if err != nil {
		status, msg := respond.MapError(err, h.logger)
		respond.Error(w, status, msg)
		return
	}
	respond.JSON(w, http.StatusOK, post)
}

func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.GetUserID(r)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid post id")
		return
	}
	var req model.UpdatePostRequest
	if err := respond.ParseJSON(r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	post, err := h.posts.Update(r.Context(), id, uid, req)
	if err != nil {
		status, msg := respond.MapError(err, h.logger)
		respond.Error(w, status, msg)
		return
	}
	respond.JSON(w, http.StatusOK, post)
}

func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.GetUserID(r)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid post id")
		return
	}
	if err := h.posts.Delete(r.Context(), id, uid); err != nil {
		status, msg := respond.MapError(err, h.logger)
		respond.Error(w, status, msg)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"message": "post deleted"})
}

func (h *PostHandler) GetByAuthor(w http.ResponseWriter, r *http.Request) {
	authorID, err := strconv.Atoi(chi.URLParam(r, "userId"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	resp, err := h.posts.GetByAuthor(r.Context(), authorID, limit, offset)
	if err != nil {
		status, msg := respond.MapError(err, h.logger)
		respond.Error(w, status, msg)
		return
	}
	respond.JSON(w, http.StatusOK, resp)
}
