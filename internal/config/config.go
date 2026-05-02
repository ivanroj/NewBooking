package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port  string
	DBDSN string
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
		// Default for running API on host with Postgres forwarded from docker-compose.
		dsn = "postgres://postgres:postgres@127.0.0.1:5432/coworking_db?sslmode=disable"
	}
	return &Config{Port: port, DBDSN: dsn}, nil
}
