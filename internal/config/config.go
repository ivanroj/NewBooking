package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port             string
	DBDSN            string
	JWTSecret        []byte
	TelegramBotToken string
	JWTExpiry        time.Duration
}

func LoadConfig() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	} else if port[0] != ':' {
		port = fmt.Sprintf(":%s", port)
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@127.0.0.1:5432/coworking_db?sslmode=disable"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-jwt-secret-change-me-please-32b!!"
	}
	if len(secret) < 24 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 24 characters")
	}

	bot := os.Getenv("TELEGRAM_BOT_TOKEN")

	exp := 168 * time.Hour
	if v := os.Getenv("JWT_EXPIRES_HOURS"); v != "" {
		h, err := strconv.Atoi(v)
		if err != nil || h < 1 || h > 24*365 {
			return nil, fmt.Errorf("invalid JWT_EXPIRES_HOURS=%q", v)
		}
		exp = time.Duration(h) * time.Hour
	}

	return &Config{
		Port:             port,
		DBDSN:            dsn,
		JWTSecret:        []byte(secret),
		TelegramBotToken: bot,
		JWTExpiry:        exp,
	}, nil
}
