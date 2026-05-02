package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/example/coworking/internal/repository"
	"github.com/gorilla/mux"
)

type adminUpdateBookingReq struct {
	Status string `json:"status"`
}

// AdminListBookings GET /api/admin/bookings — admin only.
func (h *Handlers) AdminListBookings(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	status := q.Get("status")
	roomID, _ := strconv.ParseInt(q.Get("room_id"), 10, 64)

	result, err := h.Repo.AdminListBookings(page, limit, status, roomID)
	if err != nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
		return
	}
	_ = writeJSON(w, http.StatusOK, result)
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
