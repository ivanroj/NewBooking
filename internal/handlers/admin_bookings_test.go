package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestAdminUpdateBooking_InvalidID(t *testing.T) {
	t.Parallel()
	h := &Handlers{}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/bookings/abc", http.NoBody)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	h.AdminUpdateBooking(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: want %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "invalid_id") {
		t.Fatalf("body should contain invalid_id, got: %s", rec.Body.String())
	}
}

func TestAdminUpdateBooking_BadJSON(t *testing.T) {
	t.Parallel()
	h := &Handlers{}

	rec := httptest.NewRecorder()
	body := strings.NewReader("")
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/bookings/1", body)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	h.AdminUpdateBooking(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: want %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestAdminUpdateBooking_InvalidStatus(t *testing.T) {
	t.Parallel()
	h := &Handlers{}

	rec := httptest.NewRecorder()
	body := strings.NewReader(`{"status":"invalid_status"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/bookings/1", body)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	h.AdminUpdateBooking(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: want %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "invalid_status") {
		t.Fatalf("body should contain invalid_status, got: %s", rec.Body.String())
	}
}

func TestAdminListBookings_NoRepo(t *testing.T) {
	t.Parallel()
	h := &Handlers{}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/bookings", http.NoBody)

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic with nil repo: %v", r)
		}
	}()
	h.AdminListBookings(rec, req)
}
