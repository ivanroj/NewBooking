//go:build integration

package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	dbmigrations "github.com/example/coworking/internal/dbmigrations"
)

func migrationDir(tb testing.TB) string {
	tb.Helper()
	dir, err := os.Getwd()
	if err != nil {
		tb.Fatal(err)
	}
	for {
		marker := filepath.Join(dir, "migrations", "001_initial.up.sql")
		if fi, ferr := os.Stat(marker); ferr == nil && !fi.IsDir() {
			return filepath.Join(dir, "migrations")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			tb.Fatal("migration files not found (expected migrations/001_initial.up.sql reachable from cwd)")
		}
		dir = parent
	}
}

func integrationDSN(tb testing.TB) string {
	tb.Helper()
	dsn := os.Getenv("INTEGRATION_DB_DSN")
	if dsn == "" {
		dsn = os.Getenv("DB_DSN")
	}
	if dsn == "" {
		dsn = "postgres://postgres:postgres@127.0.0.1:5432/coworking_db?sslmode=disable"
	}
	return dsn
}

func TestIntegration_MigrationsAndSchema(t *testing.T) {
	if os.Getenv("CI") != "true" && os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip(`set RUN_INTEGRATION=1 with PostgreSQL or run in CI (CI=true)`)
	}

	dsn := integrationDSN(t)
	migDir := migrationDir(t)
	if err := dbmigrations.DownAll(dsn, migDir); err != nil {
		t.Fatalf("down-all: %v", err)
	}
	if err := dbmigrations.Up(dsn, migDir); err != nil {
		t.Fatalf("up: %v", err)
	}

	dbConn, err := NewDB(dsn)
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer func() {
		if cerr := dbConn.Close(); cerr != nil {
			t.Errorf("Close: %v", cerr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dbConn.DB.PingContext(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}

	const q = `
SELECT COUNT(*) FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN ('rooms', 'workspaces', 'users', 'bookings', 'booking_settings')`

	var cnt int
	if err := dbConn.DB.QueryRowContext(ctx, q).Scan(&cnt); err != nil {
		t.Fatalf("schema query: %v", err)
	}
	if cnt != 5 {
		t.Fatalf("expected all 5 core tables migrated, got %d", cnt)
	}
}
