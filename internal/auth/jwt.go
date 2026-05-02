package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64  `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (s *Service) signToken(userID int64, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtTTL)),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(s.jwtSecret)
}

func (s *Service) ParseToken(token string) (*Claims, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrUnauthorized
	}

	var cc Claims
	parsed, err := jwt.ParseWithClaims(token, &cc, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, ErrUnauthorized
	}
	return &cc, nil
}

// ParseAuthorizationHeader reads a Bearer JWT from Authorization.
func (s *Service) ParseAuthorizationHeader(raw string) (*Claims, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ErrUnauthorized
	}

	parts := strings.SplitN(raw, " ", 2)
	if len(parts) != 2 || strings.ToLower(strings.TrimSpace(parts[0])) != "bearer" {
		return nil, ErrUnauthorized
	}
	return s.ParseToken(strings.TrimSpace(parts[1]))
}
