package auth

import "errors"

var (
	// ErrUnauthorized is returned when credentials or tokens are rejected.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrTelegramNotConfigured means TELEGRAM_BOT_TOKEN is empty.
	ErrTelegramNotConfigured = errors.New("telegram bot token not configured")

	// ErrInvalidTelegramInitData is returned on failed init_data verification.
	ErrInvalidTelegramInitData = errors.New("invalid telegram init data")
)
