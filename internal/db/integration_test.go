//go:build integration

package db

import (
	"context"
	"testing"
	"time"

	dbmigrations "github.com/example/coworking/internal/dbmigrations"
	"github.com/example/coworking/internal/testsupport"
)

func TestIntegration_MigrationsAndSchema(t *testing.T) {
	testsupport.RunIfEnabled(t)

	dsn := testsupport.DSN(t)
	dir := testsupport.MigrationDir(t)

	if err := dbmigrations.DownAll(dsn, dir); err != nil {
		t.Fatalf("down-all: %v", err)
	}
	if err := dbmigrations.Up(dsn, dir); err != nil {
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
