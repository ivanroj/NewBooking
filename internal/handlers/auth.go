package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/example/coworking/internal/auth"
)

type adminLoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type telegramAuthReq struct {
	InitData string `json:"init_data"`
}

type tokenResp struct {
	AccessToken string `json:"access_token"`
}

type errResp struct {
	Error string `json:"error"`
}

// AdminLogin exchanges admin credentials for a JWT with role admin.
func (h *Handlers) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var req adminLoginReq
	switch err := readJSONBody(r, &req); {
	case errors.Is(err, errEmptyBody):
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "empty_body"})
		return
	case err != nil:
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "bad_json"})
		return
	}

	tok, err := h.Auth.AdminLogin(req.Email, req.Password)
	if err != nil || tok == "" {
		_ = writeJSON(w, http.StatusUnauthorized, errResp{Error: "unauthorized"})
		return
	}

	_ = writeJSON(w, http.StatusOK, tokenResp{AccessToken: tok})
}

// StudentTelegramAuth validates Telegram Mini App init_data and returns a JWT for the student workspace.
func (h *Handlers) StudentTelegramAuth(w http.ResponseWriter, r *http.Request) {
	var req telegramAuthReq
	switch err := readJSONBody(r, &req); {
	case errors.Is(err, errEmptyBody):
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "empty_body"})
		return
	case err != nil:
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "bad_json"})
		return
	}

	tok, err := h.Auth.UpsertStudentFromTelegramInitData(req.InitData)
	switch {
	case err == nil:
		_ = writeJSON(w, http.StatusOK, tokenResp{AccessToken: tok})
	case errors.Is(err, auth.ErrTelegramNotConfigured):
		log.Printf("[auth] telegram not configured")
		_ = writeJSON(w, http.StatusServiceUnavailable, errResp{Error: "telegram_not_configured"})
	case errors.Is(err, auth.ErrInvalidTelegramInitData):
		log.Printf("[auth] invalid init_data: %v", err)
		_ = writeJSON(w, http.StatusUnauthorized, errResp{Error: "invalid_telegram_init_data"})
	default:
		log.Printf("[auth] unexpected error: %v", err)
		_ = writeJSON(w, http.StatusUnauthorized, errResp{Error: "unauthorized"})
	}
}
