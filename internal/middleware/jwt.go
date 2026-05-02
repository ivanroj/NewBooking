package middleware

import (
	"context"
	"net/http"

	"github.com/example/coworking/internal/auth"
)

type ctxKey int

const claimsKey ctxKey = 1

// Claims returns JWT claims injected by BearerAuth middleware.
func Claims(ctx context.Context) (*auth.Claims, bool) {
	v, ok := ctx.Value(claimsKey).(*auth.Claims)
	return v, ok
}

// BearerAuth rejects requests missing a valid Bearer token.
func BearerAuth(svc *auth.Service) func(http.Handler) http.Handler {
	unauth := []byte(`{"error":"unauthorized"}`)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cc, err := svc.ParseAuthorizationHeader(r.Header.Get("Authorization"))
			if err != nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write(unauth)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsKey, cc)))
		})
	}
}
