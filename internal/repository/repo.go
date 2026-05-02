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

// --- Room CRUD (stub) ---
func (r *Repo) GetAllRooms() ([]models.Room, error) {
    var rooms []models.Room
    err := r.db.Select(&rooms, "SELECT id, name, description FROM rooms")
    return rooms, err
}

// Additional methods for Workspace, User, Booking, Settings can be added later
