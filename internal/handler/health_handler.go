package handler

import (
	"net/http"

	"blog-platform/internal/respond"

	"github.com/jmoiron/sqlx"
)

type HealthHandler struct{ db *sqlx.DB }

func NewHealthHandler(db *sqlx.DB) *HealthHandler { return &HealthHandler{db: db} }

// Маршрут зарегистрирован в chi только на GET, поэтому неверные методы
// отклоняются роутером (405) до попадания сюда.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	if err := h.db.Ping(); err != nil {
		respond.Error(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}