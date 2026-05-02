package handlers

import "net/http"

// AdminGetStats GET /api/admin/stats — admin only.
func (h *Handlers) AdminGetStats(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	dateFrom := q.Get("from")
	dateTo := q.Get("to")

	stats, err := h.Repo.GetStats(dateFrom, dateTo)
	if err != nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
		return
	}
	_ = writeJSON(w, http.StatusOK, stats)
}
