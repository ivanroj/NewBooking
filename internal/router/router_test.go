package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/coworking/internal/db"
	"github.com/example/coworking/internal/repository"
)

func TestHealth_OK(t *testing.T) {
	t.Parallel()

	var nilDB *db.DB
	repo := repository.NewRepo(nilDB)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)

	NewRouter(repo).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want %d, got %d", http.StatusOK, rec.Code)
	}
	if body := rec.Body.String(); body != "OK" {
		t.Fatalf("body: want %q, got %q", "OK", body)
	}
}
