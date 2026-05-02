package app

import (
	"github.com/example/coworking/internal/auth"
	"github.com/example/coworking/internal/config"
	"github.com/example/coworking/internal/db"
	"github.com/example/coworking/internal/handlers"
	"github.com/example/coworking/internal/repository"
	"github.com/example/coworking/internal/router"
)

type App struct {
	cfg  *config.Config
	conn *db.DB
	repo *repository.Repo
	auth *auth.Service
	r    *router.Router
}

func NewApp(cfg *config.Config) (*App, error) {
	dbConn, err := db.NewDB(cfg.DBDSN)
	if err != nil {
		return nil, err
	}
	repo := repository.NewRepo(dbConn)
	svc := auth.NewService(repo, cfg.JWTSecret, cfg.JWTExpiry, cfg.TelegramBotToken)
	h := handlers.New(svc, repo)

	return &App{cfg: cfg, conn: dbConn, repo: repo, auth: svc, r: router.NewRouter(h)}, nil
}

func (a *App) Router() *router.Router {
	return a.r
}
