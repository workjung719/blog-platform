package handler

import (
	"log/slog"
	"net/http"

	"blog-platform/internal/model"
	"blog-platform/internal/respond"
	"blog-platform/internal/service"
)

type UserHandler struct {
	users  *service.UserService
	logger *slog.Logger
}

func NewUserHandler(users *service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{users: users, logger: logger}
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  model.User  `json:"user"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := respond.ParseJSON(r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	user, token, err := h.users.Register(r.Context(), req)
	if err != nil {
		status, msg := respond.MapError(err, h.logger)
		respond.Error(w, status, msg)
		return
	}
	respond.JSON(w, http.StatusCreated, AuthResponse{Token: token, User: *user})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := respond.ParseJSON(r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	user, token, err := h.users.Login(r.Context(), req)
	if err != nil {
		status, msg := respond.MapError(err, h.logger)
		// Маскируем именно 401 (неверные креды); 500 при сбое БД остаётся 500.
		if status == http.StatusUnauthorized {
			msg = "invalid email or password"
		}
		respond.Error(w, status, msg)
		return
	}
	respond.JSON(w, http.StatusOK, AuthResponse{Token: token, User: *user})
}