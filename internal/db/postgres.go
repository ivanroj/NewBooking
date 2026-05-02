package db

import (
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    "log"
)

type DB struct {
    *sqlx.DB
}

func NewDB(dsn string) (*DB, error) {
    db, err := sqlx.Connect("postgres", dsn)
    if err != nil {
        return nil, err
    }
    // Optionally run migrations here (placeholder)
    log.Println("Connected to Postgres")
    return &DB{DB: db}, nil
}
