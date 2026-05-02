package handlers

import "net/http"

type setLimitReq struct {
	MaxActive int `json:"max_active"`
}

// AdminSetBookingLimit PATCH /api/admin/settings/booking-limit — admin only.
func (h *Handlers) AdminSetBookingLimit(w http.ResponseWriter, r *http.Request) {
	var req setLimitReq
	if err := readJSONBody(r, &req); err != nil {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "bad_json"})
		return
	}
	if req.MaxActive < 0 {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "max_active_must_be_non_negative"})
		return
	}
	if err := h.Repo.SetBookingLimit(req.MaxActive); err != nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
		return
	}
	_ = writeJSON(w, http.StatusOK, map[string]int{"max_active": req.MaxActive})
}

// AdminGetBookingLimit GET /api/admin/settings/booking-limit — admin only.
func (h *Handlers) AdminGetBookingLimit(w http.ResponseWriter, _ *http.Request) {
	limit, err := h.Repo.GetBookingLimit()
	if err != nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
		return
	}
	_ = writeJSON(w, http.StatusOK, map[string]int{"max_active": limit})
}
