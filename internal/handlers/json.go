package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

var errEmptyBody = errors.New("empty request body")

const maxJSONBodyBytes = 1 << 17

func readJSONBody(r *http.Request, dst any) error {
	defer func() {
		if err := r.Body.Close(); err != nil {
			_ = err
		}
	}()
	raw, err := io.ReadAll(io.LimitReader(r.Body, maxJSONBodyBytes+1))
	if err != nil {
		return err
	}
	if len(raw) > maxJSONBodyBytes {
		return io.ErrUnexpectedEOF
	}
	if strings.TrimSpace(string(raw)) == "" {
		return errEmptyBody
	}
	return json.Unmarshal(raw, dst)
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
