//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/example/coworking/internal/auth"
	"github.com/example/coworking/internal/db"
	"github.com/example/coworking/internal/handlers"
	"github.com/example/coworking/internal/repository"
	"github.com/example/coworking/internal/router"
	"github.com/example/coworking/internal/testsupport"
)

type tokenPayload struct {
	AccessToken string `json:"access_token"`
}

func TestAPI_AuthRoomsFlow(t *testing.T) {
	const (
		jwtSecret = "test-jwt-secret-change-me-min-24-ch!!!"
		botTok    = "1111111111:AAAAAAAAAABbbbbbbbbBBBBbbbbbbbbBBB"
	)

	testsupport.ResetSchema(t)

	conn, err := db.NewDB(testsupport.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	})

	ctx := context.Background()
	if _, err := conn.DB.ExecContext(ctx, `INSERT INTO rooms (name, description) VALUES ('Lab A', 'demo')`); err != nil {
		t.Fatal(err)
	}

	repo := repository.NewRepo(conn)
	svc := auth.NewService(repo, []byte(jwtSecret), 24*time.Hour, botTok)
	h := handlers.New(svc, repo)
	ts := httptest.NewServer(router.NewRouter(h))
	t.Cleanup(ts.Close)
	cl := ts.Client()

	body, _ := json.Marshal(map[string]string{
		"email":    "admin@cowork.local",
		"password": "AdminDevPass#1",
	})
	res, err := cl.Post(ts.URL+"/api/auth/admin/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("admin login: want status %d, got %d", http.StatusOK, res.StatusCode)
	}
	defer func() { _ = res.Body.Close() }()
	var adminTok tokenPayload
	if err := json.NewDecoder(res.Body).Decode(&adminTok); err != nil {
		t.Fatal(err)
	}
	if adminTok.AccessToken == "" {
		t.Fatal("empty admin jwt")
	}

	bodyBad, _ := json.Marshal(map[string]string{
		"email":    "admin@cowork.local",
		"password": "wrong-password",
	})
	resBad, err := cl.Post(ts.URL+"/api/auth/admin/login", "application/json", bytes.NewReader(bodyBad))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resBad.Body.Close() }()
	if resBad.StatusCode != http.StatusUnauthorized {
		t.Fatalf("admin bad password: want %d got %d", http.StatusUnauthorized, resBad.StatusCode)
	}

	noAuthRes, err := cl.Get(ts.URL + "/api/rooms")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = noAuthRes.Body.Close() }()
	if noAuthRes.StatusCode != http.StatusUnauthorized {
		t.Fatalf("rooms without bearer: want %d got %d", http.StatusUnauthorized, noAuthRes.StatusCode)
	}

	reqRooms, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/api/rooms", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	reqRooms.Header.Set("Authorization", "Bearer "+adminTok.AccessToken)
	resRooms, err := cl.Do(reqRooms)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resRooms.Body.Close() }()
	if resRooms.StatusCode != http.StatusOK {
		t.Fatalf("rooms authorized: want %d got %d", http.StatusOK, resRooms.StatusCode)
	}

	tgID := int64(999888777)
	userJSON := `{"id":` + strconv.FormatInt(tgID, 10) + `,"username":"stu"}`
	initRaw, err := auth.SignedInitDataForTests(botTok, userJSON, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	studBody, _ := json.Marshal(map[string]string{"init_data": initRaw})
	resSt, err := cl.Post(ts.URL+"/api/auth/student/telegram", "application/json", bytes.NewReader(studBody))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resSt.Body.Close() }()
	if resSt.StatusCode != http.StatusOK {
		t.Fatalf("student auth: want %d got %d", http.StatusOK, resSt.StatusCode)
	}
	var stTok tokenPayload
	if err := json.NewDecoder(resSt.Body).Decode(&stTok); err != nil {
		t.Fatal(err)
	}
	if stTok.AccessToken == "" {
		t.Fatal("empty student jwt")
	}

	badInit := initRaw + "x"
	studBad, _ := json.Marshal(map[string]string{"init_data": badInit})
	resBadSt, err := cl.Post(ts.URL+"/api/auth/student/telegram", "application/json", bytes.NewReader(studBad))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resBadSt.Body.Close() }()
	if resBadSt.StatusCode != http.StatusUnauthorized {
		t.Fatalf("bad student init: want %d got %d", http.StatusUnauthorized, resBadSt.StatusCode)
	}
}
