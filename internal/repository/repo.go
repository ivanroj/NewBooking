package repository

import (
	"github.com/example/coworking/internal/db"
	"github.com/example/coworking/internal/models"
)

type Repo struct {
	db *db.DB
}

func NewRepo(db *db.DB) *Repo {
	return &Repo{db: db}
}

// GetAllRooms returns all rooms (stub used until Room API is implemented).
func (r *Repo) GetAllRooms() ([]models.Room, error) {
	var rooms []models.Room
	err := r.db.Select(&rooms, "SELECT id, name, description FROM rooms")
	return rooms, err
}
