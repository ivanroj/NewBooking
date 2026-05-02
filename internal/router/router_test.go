package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/coworking/internal/auth"
	"github.com/example/coworking/internal/db"
	"github.com/example/coworking/internal/handlers"
	"github.com/example/coworking/internal/repository"
)

func TestHealth_OK(t *testing.T) {
	t.Parallel()

	var nilConn *db.DB
	authSvc := auth.NewService(
		repository.NewRepo(nilConn),
		[]byte("dev-jwt-secret-change-me-please-32"),
		time.Hour,
		"",
	)
	h := handlers.New(authSvc, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)

	NewRouter(h).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want %d, got %d", http.StatusOK, rec.Code)
	}
	if body := rec.Body.String(); body != "OK" {
		t.Fatalf("body: want %q, got %q", "OK", body)
	}
}
