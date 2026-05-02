package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/example/coworking/internal/dbmigrations"
)

func main() {
	cmd := flag.String("command", "up", `migration command: "up", "down", "down-all"`)
	dir := flag.String("dir", "migrations", "path to migration files")
	flag.Parse()

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, `DB_DSN is required (e.g. postgres://user:pass@localhost:5432/db?sslmode=disable)`)
		os.Exit(1)
	}

	var err error
	switch *cmd {
	case "up":
		err = dbmigrations.Up(dsn, *dir)
	case "down":
		err = dbmigrations.Down(dsn, *dir)
	case "down-all":
		err = dbmigrations.DownAll(dsn, *dir)
	default:
		log.Fatalf("unknown -command=%q", *cmd)
	}
	if err != nil {
		log.Fatal(err)
	}
}
