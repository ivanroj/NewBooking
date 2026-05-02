package main

import (
    "log"
    "net/http"

    "github.com/example/coworking/internal/app"
    "github.com/example/coworking/internal/config"
)

func main() {
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    a, err := app.NewApp(cfg)
    if err != nil {
        log.Fatalf("failed to initialize app: %v", err)
    }

    log.Printf("starting server on %s", cfg.Port)
    if err := http.ListenAndServe(cfg.Port, a.Router()); err != nil {
        log.Fatalf("server error: %v", err)
    }
}
