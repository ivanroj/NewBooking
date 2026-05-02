package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var errNoUserPayload = errors.New("init data missing user")

func buildDataCheckString(values url.Values, excludeKeys ...string) string {
	skip := map[string]bool{"hash": true}
	for _, k := range excludeKeys {
		skip[k] = true
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		if skip[k] {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(values.Get(k))
	}
	return b.String()
}

func computeWebAppSignature(botToken string, dataCheckString string) (string, error) {
	skMac := hmac.New(sha256.New, []byte("WebAppData"))
	if _, err := skMac.Write([]byte(botToken)); err != nil {
		return "", err
	}
	secretKey := skMac.Sum(nil)

	sigMac := hmac.New(sha256.New, secretKey)
	if _, err := sigMac.Write([]byte(dataCheckString)); err != nil {
		return "", err
	}
	sum := sigMac.Sum(nil)
	return hex.EncodeToString(sum), nil
}

// ValidateTelegramInitData parses and validates Telegram Web App init_data per Mini Apps hashing rules.
func ValidateTelegramInitData(initDataRaw, botToken string, maxAge time.Duration) (int64, error) {
	if botToken == "" {
		return 0, fmt.Errorf("%w: empty bot token", ErrInvalidTelegramInitData)
	}
	values, err := url.ParseQuery(initDataRaw)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidTelegramInitData, err)
	}
	wantHash := values.Get("hash")
	if wantHash == "" {
		return 0, fmt.Errorf("%w: missing hash", ErrInvalidTelegramInitData)
	}
	ad := values.Get("auth_date")
	if ad == "" {
		return 0, fmt.Errorf("%w: missing auth_date", ErrInvalidTelegramInitData)
	}
	ts, err := strconv.ParseInt(ad, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: bad auth_date", ErrInvalidTelegramInitData)
	}
	if maxAge > 0 {
		if age := time.Since(time.Unix(ts, 0)); age > maxAge || age < -5*time.Minute {
			return 0, fmt.Errorf("%w: stale auth_date", ErrInvalidTelegramInitData)
		}
	}

	// For bot-token HMAC validation, only "hash" is excluded from data_check_string.
	// "signature" (if present) must be INCLUDED — it is only excluded for Ed25519 public-key validation.
	dataCheck := buildDataCheckString(values)
	log.Printf("[auth-debug] data_check_string keys: %v", func() []string {
		ks := make([]string, 0)
		for k := range values {
			if k != "hash" {
				ks = append(ks, k)
			}
		}
		sort.Strings(ks)
		return ks
	}())
	log.Printf("[auth-debug] wantHash=%s", wantHash[:min(16, len(wantHash))])

	got, err := computeWebAppSignature(botToken, dataCheck)
	if err != nil {
		return 0, err
	}
	log.Printf("[auth-debug] gotHash=%s", got[:min(16, len(got))])

	if subtle.ConstantTimeCompare([]byte(strings.ToLower(got)), []byte(strings.ToLower(wantHash))) != 1 {
		// Check if signature-based validation is needed (Bot API 7.x+)
		sig := values.Get("signature")
		if sig != "" {
			log.Printf("[auth-debug] signature field present, length=%d", len(sig))
		}
		return 0, fmt.Errorf("%w: hash mismatch", ErrInvalidTelegramInitData)
	}

	payload := values.Get("user")
	if payload == "" {
		return 0, fmt.Errorf("%w: %w", ErrInvalidTelegramInitData, errNoUserPayload)
	}
	var u struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(payload), &u); err != nil || u.ID == 0 {
		return 0, fmt.Errorf("%w: bad user payload", ErrInvalidTelegramInitData)
	}
	return u.ID, nil
}
