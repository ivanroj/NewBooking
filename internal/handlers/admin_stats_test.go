package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminGetStats_NoRepo(t *testing.T) {
	t.Parallel()
	h := &Handlers{}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/stats", http.NoBody)

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic with nil repo: %v", r)
		}
	}()
	h.AdminGetStats(rec, req)
}
