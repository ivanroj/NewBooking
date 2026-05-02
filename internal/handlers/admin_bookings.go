package handlers

import (
	"errors"
	"net/http"

	"github.com/example/coworking/internal/repository"
	"github.com/gorilla/mux"
)

type adminUpdateBookingReq struct {
	Status string `json:"status"`
}

// AdminUpdateBooking PATCH /api/admin/bookings/{id} — admin only.
func (h *Handlers) AdminUpdateBooking(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(mux.Vars(r)["id"])
	if err != nil {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid_id"})
		return
	}

	var req adminUpdateBookingReq
	if err = readJSONBody(r, &req); err != nil {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "bad_json"})
		return
	}
	switch req.Status {
	case "active", "canceled", "completed":
		// valid
	default:
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid_status"})
		return
	}

	b, err := h.Repo.AdminUpdateBooking(id, req.Status)
	if errors.Is(err, repository.ErrNotFound) {
		_ = writeJSON(w, http.StatusNotFound, errResp{Error: "not_found"})
		return
	}
	if err != nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
		return
	}
	_ = writeJSON(w, http.StatusOK, b)
}
