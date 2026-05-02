package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/example/coworking/internal/middleware"
	"github.com/example/coworking/internal/repository"
	"github.com/example/coworking/internal/service"
	"github.com/gorilla/mux"
)

type createBookingReq struct {
	WorkspaceID int64  `json:"workspace_id"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
}

// CreateBooking POST /api/bookings — student.
func (h *Handlers) CreateBooking(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.Claims(r.Context())
	if !ok {
		_ = writeJSON(w, http.StatusUnauthorized, errResp{Error: "unauthorized"})
		return
	}

	var req createBookingReq
	if err := readJSONBody(r, &req); err != nil {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "bad_json"})
		return
	}
	if req.WorkspaceID == 0 {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "workspace_id_required"})
		return
	}

	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid_start_time"})
		return
	}
	end, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid_end_time"})
		return
	}

	if !service.ValidateBookingWindow(start, end, time.Now()) {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid_time_window"})
		return
	}

	booking, err := h.Repo.CreateBooking(claims.UserID, req.WorkspaceID, start, end)
	switch {
	case errors.Is(err, repository.ErrConflict):
		_ = writeJSON(w, http.StatusConflict, errResp{Error: "conflict"})
	case errors.Is(err, repository.ErrLimitExceeded):
		_ = writeJSON(w, http.StatusUnprocessableEntity, errResp{Error: "limit_exceeded"})
	case err != nil:
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
	default:
		_ = writeJSON(w, http.StatusCreated, booking)
	}
}

// CancelMyBooking PATCH /api/bookings/{id}/cancel — student (owner only).
func (h *Handlers) CancelMyBooking(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.Claims(r.Context())
	if !ok {
		_ = writeJSON(w, http.StatusUnauthorized, errResp{Error: "unauthorized"})
		return
	}
	id, err := parseID(mux.Vars(r)["id"])
	if err != nil {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid_id"})
		return
	}
	b, err := h.Repo.CancelMyBooking(claims.UserID, id)
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

// ListMyBookings GET /api/bookings/my — student/admin.
// Query: ?status=active|canceled|completed&page=1&limit=20
func (h *Handlers) ListMyBookings(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.Claims(r.Context())
	if !ok {
		_ = writeJSON(w, http.StatusUnauthorized, errResp{Error: "unauthorized"})
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	result, err := h.Repo.ListMyBookings(repository.BookingListParams{
		UserID: claims.UserID,
		Status: q.Get("status"),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
		return
	}
	_ = writeJSON(w, http.StatusOK, result)
}
