package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

type DB struct {
	*sqlx.DB
}

func NewDB(dsn string) (*DB, error) {
	dbx, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &DB{DB: dbx}, nil
}
