package middleware

import (
	"net/http"
)

// RequireAdmin rejects requests whose JWT role is not "admin".
// Must be used after BearerAuth (which injects Claims into context).
func RequireAdmin(next http.Handler) http.Handler {
	forbidden := []byte(`{"error":"forbidden"}`)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cc, ok := Claims(r.Context())
		if !ok || cc.Role != "admin" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write(forbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
