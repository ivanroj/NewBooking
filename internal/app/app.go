package app

import (
	"github.com/example/coworking/internal/config"
	"github.com/example/coworking/internal/db"
	"github.com/example/coworking/internal/repository"
	"github.com/example/coworking/internal/router"
)

type App struct {
	cfg  *config.Config
	db   *db.DB
	repo *repository.Repo
	r    *router.Router
}

func NewApp(cfg *config.Config) (*App, error) {
	dbConn, err := db.NewDB(cfg.DBDSN)
	if err != nil {
		return nil, err
	}
	repo := repository.NewRepo(dbConn)
	r := router.NewRouter(repo)
	return &App{cfg: cfg, db: dbConn, repo: repo, r: r}, nil
}

func (a *App) Router() *router.Router {
	return a.r
}
