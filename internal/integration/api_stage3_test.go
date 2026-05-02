//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/coworking/internal/auth"
	"github.com/example/coworking/internal/db"
	"github.com/example/coworking/internal/handlers"
	"github.com/example/coworking/internal/models"
	"github.com/example/coworking/internal/repository"
	"github.com/example/coworking/internal/router"
	"github.com/example/coworking/internal/testsupport"
)

const (
	stage3JWTSecret = "test-jwt-secret-change-me-min-24-ch!!!"
	stage3BotTok    = "1111111111:AAAAAAAAAABbbbbbbbbBBBBbbbbbbbbBBB"
)

// stage3Env holds a running test server with admin and student tokens ready.
type stage3Env struct {
	srv        *httptest.Server
	cl         *http.Client
	adminToken string
	studToken  string
	repo       *repository.Repo
}

func setupStage3(t *testing.T) *stage3Env {
	t.Helper()
	testsupport.ResetSchema(t)

	conn, err := db.NewDB(testsupport.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	repo := repository.NewRepo(conn)
	svc := auth.NewService(repo, []byte(stage3JWTSecret), 24*time.Hour, stage3BotTok)
	h := handlers.New(svc, repo)
	srv := httptest.NewServer(router.NewRouter(h))
	t.Cleanup(srv.Close)

	// Admin token via login
	body, _ := json.Marshal(map[string]string{
		"email":    "admin@cowork.local",
		"password": "AdminDevPass#1",
	})
	res, err := srv.Client().Post(srv.URL+"/api/auth/admin/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("admin login: got %d", res.StatusCode)
	}
	var adTok tokenPayload
	if err := json.NewDecoder(res.Body).Decode(&adTok); err != nil {
		t.Fatal(err)
	}

	// Student token via Telegram init_data
	tgID := int64(555444333)
	userJSON := fmt.Sprintf(`{"id":%d,"username":"teststu"}`, tgID)
	initRaw, err := auth.SignedInitDataForTests(stage3BotTok, userJSON, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	studBody, _ := json.Marshal(map[string]string{"init_data": initRaw})
	resSt, err := srv.Client().Post(srv.URL+"/api/auth/student/telegram", "application/json", bytes.NewReader(studBody))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resSt.Body.Close() }()
	var stTok tokenPayload
	if err := json.NewDecoder(resSt.Body).Decode(&stTok); err != nil {
		t.Fatal(err)
	}

	return &stage3Env{
		srv:        srv,
		cl:         srv.Client(),
		adminToken: adTok.AccessToken,
		studToken:  stTok.AccessToken,
		repo:       repo,
	}
}

// helper: do a JSON request with optional bearer token.
func (e *stage3Env) do(t *testing.T, method, path string, body any, token string) *http.Response {
	t.Helper()
	var b *bytes.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		b = bytes.NewReader(data)
	} else {
		b = bytes.NewReader(nil)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, e.srv.URL+path, b)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := e.cl.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

// ────────────────────────────── Rooms CRUD ──────────────────────────────────

func TestAdminCreateRoom_OK(t *testing.T) {
	e := setupStage3(t)

	res := e.do(t, http.MethodPost, "/api/rooms",
		map[string]string{"name": "Room Alpha", "description": "test"}, e.adminToken)
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("want 201, got %d", res.StatusCode)
	}
	var room models.Room
	if err := json.NewDecoder(res.Body).Decode(&room); err != nil {
		t.Fatal(err)
	}
	if room.ID == 0 || room.Name != "Room Alpha" {
		t.Fatalf("unexpected room: %+v", room)
	}
}

func TestAdminCreateRoom_StudentForbidden(t *testing.T) {
	e := setupStage3(t)

	res := e.do(t, http.MethodPost, "/api/rooms",
		map[string]string{"name": "X"}, e.studToken)
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("want 403, got %d", res.StatusCode)
	}
}

func TestAdminCreateRoom_EmptyName(t *testing.T) {
	e := setupStage3(t)

	res := e.do(t, http.MethodPost, "/api/rooms",
		map[string]string{"name": ""}, e.adminToken)
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", res.StatusCode)
	}
}

func TestAdminDeleteRoom_OK(t *testing.T) {
	e := setupStage3(t)

	room, err := e.repo.CreateRoom("ToDelete", "")
	if err != nil {
		t.Fatal(err)
	}
	res := e.do(t, http.MethodDelete, fmt.Sprintf("/api/rooms/%d", room.ID), nil, e.adminToken)
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("want 204, got %d", res.StatusCode)
	}
}

// ────────────────────────────── Bookings ────────────────────────────────────

func createTestRoom(t *testing.T, repo *repository.Repo) models.Room {
	t.Helper()
	room, err := repo.CreateRoom("Test Room", "integration")
	if err != nil {
		t.Fatal(err)
	}
	return room
}

func createTestWorkspace(t *testing.T, repo *repository.Repo, roomID int64) models.Workspace {
	t.Helper()
	ws, err := repo.CreateWorkspace(roomID, "Desk 1")
	if err != nil {
		t.Fatal(err)
	}
	return ws
}

func futureSlot() (string, string) {
	start := time.Now().UTC().Add(2 * time.Hour).Truncate(time.Hour)
	end := start.Add(2 * time.Hour)
	return start.Format(time.RFC3339), end.Format(time.RFC3339)
}

func TestCreateBooking_OK(t *testing.T) {
	e := setupStage3(t)
	room := createTestRoom(t, e.repo)
	ws := createTestWorkspace(t, e.repo, room.ID)

	start, end := futureSlot()
	res := e.do(t, http.MethodPost, "/api/bookings", map[string]any{
		"workspace_id": ws.ID,
		"start_time":   start,
		"end_time":     end,
	}, e.studToken)
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("want 201, got %d", res.StatusCode)
	}
}

func TestCreateBooking_Conflict(t *testing.T) {
	e := setupStage3(t)
	room := createTestRoom(t, e.repo)
	ws := createTestWorkspace(t, e.repo, room.ID)

	start, end := futureSlot()
	payload := map[string]any{
		"workspace_id": ws.ID,
		"start_time":   start,
		"end_time":     end,
	}
	// First booking succeeds
	r1 := e.do(t, http.MethodPost, "/api/bookings", payload, e.studToken)
	_ = r1.Body.Close()
	if r1.StatusCode != http.StatusCreated {
		t.Fatalf("first booking want 201, got %d", r1.StatusCode)
	}

	// Second booking on same slot → 409
	r2 := e.do(t, http.MethodPost, "/api/bookings", payload, e.studToken)
	defer func() { _ = r2.Body.Close() }()
	if r2.StatusCode != http.StatusConflict {
		t.Fatalf("conflict booking want 409, got %d", r2.StatusCode)
	}
}

func TestCreateBooking_LimitExceeded(t *testing.T) {
	e := setupStage3(t)

	// Set limit to 1
	res := e.do(t, http.MethodPatch, "/api/admin/settings/booking-limit",
		map[string]int{"max_active": 1}, e.adminToken)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("set limit want 200, got %d", res.StatusCode)
	}

	room := createTestRoom(t, e.repo)
	ws1 := createTestWorkspace(t, e.repo, room.ID)
	ws2 := createTestWorkspace(t, e.repo, room.ID)

	// First booking OK
	s1 := time.Now().UTC().Add(2 * time.Hour).Truncate(time.Hour)
	e1 := s1.Add(time.Hour)
	r1 := e.do(t, http.MethodPost, "/api/bookings", map[string]any{
		"workspace_id": ws1.ID,
		"start_time":   s1.Format(time.RFC3339),
		"end_time":     e1.Format(time.RFC3339),
	}, e.studToken)
	_ = r1.Body.Close()
	if r1.StatusCode != http.StatusCreated {
		t.Fatalf("first booking want 201, got %d", r1.StatusCode)
	}

	// Second booking hits limit → 422
	s2 := e1.Add(time.Hour)
	e2 := s2.Add(time.Hour)
	r2 := e.do(t, http.MethodPost, "/api/bookings", map[string]any{
		"workspace_id": ws2.ID,
		"start_time":   s2.Format(time.RFC3339),
		"end_time":     e2.Format(time.RFC3339),
	}, e.studToken)
	defer func() { _ = r2.Body.Close() }()
	if r2.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("limit exceeded want 422, got %d", r2.StatusCode)
	}
}

func TestCancelBooking_OK(t *testing.T) {
	e := setupStage3(t)
	room := createTestRoom(t, e.repo)
	ws := createTestWorkspace(t, e.repo, room.ID)

	start, end := futureSlot()
	r1 := e.do(t, http.MethodPost, "/api/bookings", map[string]any{
		"workspace_id": ws.ID,
		"start_time":   start,
		"end_time":     end,
	}, e.studToken)
	defer func() { _ = r1.Body.Close() }()
	if r1.StatusCode != http.StatusCreated {
		t.Fatalf("create booking want 201, got %d", r1.StatusCode)
	}
	var b models.Booking
	if err := json.NewDecoder(r1.Body).Decode(&b); err != nil {
		t.Fatal(err)
	}

	r2 := e.do(t, http.MethodPatch, fmt.Sprintf("/api/bookings/%d/cancel", b.ID), nil, e.studToken)
	defer func() { _ = r2.Body.Close() }()
	if r2.StatusCode != http.StatusOK {
		t.Fatalf("cancel want 200, got %d", r2.StatusCode)
	}
}

func TestCancelBooking_NotOwner(t *testing.T) {
	e := setupStage3(t)
	room := createTestRoom(t, e.repo)
	ws := createTestWorkspace(t, e.repo, room.ID)

	start, end := futureSlot()
	r1 := e.do(t, http.MethodPost, "/api/bookings", map[string]any{
		"workspace_id": ws.ID,
		"start_time":   start,
		"end_time":     end,
	}, e.studToken)
	defer func() { _ = r1.Body.Close() }()
	var b models.Booking
	_ = json.NewDecoder(r1.Body).Decode(&b)

	// Admin tries to cancel via student endpoint → 404 (not their booking)
	r2 := e.do(t, http.MethodPatch, fmt.Sprintf("/api/bookings/%d/cancel", b.ID), nil, e.adminToken)
	defer func() { _ = r2.Body.Close() }()
	if r2.StatusCode != http.StatusNotFound {
		t.Fatalf("not-owner cancel want 404, got %d", r2.StatusCode)
	}
}

func TestListMyBookings_Pagination(t *testing.T) {
	e := setupStage3(t)
	room := createTestRoom(t, e.repo)

	// Create 3 workspaces and book each
	base := time.Now().UTC().Add(3 * time.Hour).Truncate(time.Hour)
	for i := range 3 {
		ws := createTestWorkspace(t, e.repo, room.ID)
		s := base.Add(time.Duration(i*3) * time.Hour)
		end := s.Add(time.Hour)
		r := e.do(t, http.MethodPost, "/api/bookings", map[string]any{
			"workspace_id": ws.ID,
			"start_time":   s.Format(time.RFC3339),
			"end_time":     end.Format(time.RFC3339),
		}, e.studToken)
		_ = r.Body.Close()
	}

	// Get page 1, limit 2
	res := e.do(t, http.MethodGet, "/api/bookings/my?page=1&limit=2", nil, e.studToken)
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("list bookings want 200, got %d", res.StatusCode)
	}

	var result repository.BookingListResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Total != 3 {
		t.Fatalf("total want 3, got %d", result.Total)
	}
	if len(result.Items) != 2 {
		t.Fatalf("items want 2, got %d", len(result.Items))
	}
}

func TestAdminSetLimit_OK(t *testing.T) {
	e := setupStage3(t)

	res := e.do(t, http.MethodPatch, "/api/admin/settings/booking-limit",
		map[string]int{"max_active": 10}, e.adminToken)
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", res.StatusCode)
	}
	var out map[string]int
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out["max_active"] != 10 {
		t.Fatalf("want max_active=10, got %v", out)
	}
}
