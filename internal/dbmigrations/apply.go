package dbmigrations

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver for migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"       // migrations from filesystem
)

func fileSourceURL(dir string) (string, error) {
	if !filepath.IsAbs(dir) {
		var err error
		dir, err = filepath.Abs(dir)
		if err != nil {
			return "", err
		}
	}
	return "file://" + filepath.ToSlash(dir), nil
}

// Up applies all migrations. ErrNoChange is treated as success.
func Up(dsn, migrationsDir string) error {
	src, err := fileSourceURL(migrationsDir)
	if err != nil {
		return err
	}
	m, err := migrate.New(src, dsn)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()
	if verr := m.Up(); verr != nil && verr != migrate.ErrNoChange {
		return verr
	}
	return nil
}

// Down rolls back the last migration.
func Down(dsn, migrationsDir string) error {
	src, err := fileSourceURL(migrationsDir)
	if err != nil {
		return err
	}
	m, err := migrate.New(src, dsn)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()
	e := m.Steps(-1)
	if e != nil && e != migrate.ErrNoChange {
		return e
	}
	return nil
}

// DownAll rolls back migrations until baseline (no-change).
func DownAll(dsn, migrationsDir string) error {
	src, err := fileSourceURL(migrationsDir)
	if err != nil {
		return err
	}
	for {
		m, err := migrate.New(src, dsn)
		if err != nil {
			return fmt.Errorf("migrate.New: %w", err)
		}
		e := m.Steps(-1)
		_, _ = m.Close()
		if e == migrate.ErrNoChange {
			return nil
		}
		if e != nil {
			return e
		}
	}
}
