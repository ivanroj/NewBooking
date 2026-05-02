package handlers

import (
	"net/http"

	"github.com/example/coworking/internal/models"
)

// ListRooms returns all coworking rooms (student or admin token required).
func (h *Handlers) ListRooms(w http.ResponseWriter, _ *http.Request) {
	if h.Repo == nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "misconfigured_repository"})
		return
	}

	rooms, err := h.Repo.GetAllRooms()
	if err != nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
		return
	}

	out := rooms
	if out == nil {
		out = []models.Room{}
	}
	_ = writeJSON(w, http.StatusOK, out)
}
