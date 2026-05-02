package handlers

import (
	"github.com/example/coworking/internal/auth"
	"github.com/example/coworking/internal/repository"
)

// Handlers wires HTTP controllers to auth and persistence.
type Handlers struct {
	Auth *auth.Service
	Repo *repository.Repo
}

func New(auth *auth.Service, repo *repository.Repo) *Handlers {
	return &Handlers{Auth: auth, Repo: repo}
}
