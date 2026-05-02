//go:build integration

package testsupport

import (
	"os"
	"path/filepath"
	"testing"

	dbmigrations "github.com/example/coworking/internal/dbmigrations"
)

func MigrationDir(tb testing.TB) string {
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
			tb.Fatal("migrations not found")
		}
		dir = parent
	}
}

func DSN(tb testing.TB) string {
	tb.Helper()
	if v := os.Getenv("INTEGRATION_DB_DSN"); v != "" {
		return v
	}
	if v := os.Getenv("DB_DSN"); v != "" {
		return v
	}
	return "postgres://postgres:postgres@127.0.0.1:5432/coworking_db?sslmode=disable"
}

func ResetSchema(tb testing.TB) {
	tb.Helper()
	RunIfEnabled(tb)

	dsn := DSN(tb)
	dir := MigrationDir(tb)
	if err := dbmigrations.DownAll(dsn, dir); err != nil {
		tb.Fatalf("down-all: %v", err)
	}
	if err := dbmigrations.Up(dsn, dir); err != nil {
		tb.Fatalf("up: %v", err)
	}
}

func RunIfEnabled(tb testing.TB) {
	tb.Helper()
	if os.Getenv("CI") != "true" && os.Getenv("RUN_INTEGRATION") != "1" {
		tb.Skip(`set RUN_INTEGRATION=1 or run in CI (CI=true)`)
	}
}
