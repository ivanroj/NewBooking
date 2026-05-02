package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/example/coworking/internal/models"
)

// CreateRoom inserts a new room and returns it with assigned ID.
func (r *Repo) CreateRoom(name, description string) (models.Room, error) {
	const q = `INSERT INTO rooms (name, description) VALUES ($1, $2) RETURNING id, name, description`
	var room models.Room
	if err := r.db.Get(&room, q, name, description); err != nil {
		return models.Room{}, fmt.Errorf("create room: %w", err)
	}
	return room, nil
}

// UpdateRoom partially updates name and/or description of a room by ID.
func (r *Repo) UpdateRoom(id int64, name, description string) (models.Room, error) {
	const q = `
UPDATE rooms SET name = $1, description = $2 WHERE id = $3
RETURNING id, name, description`
	var room models.Room
	switch err := r.db.Get(&room, q, name, description, id); {
	case errors.Is(err, sql.ErrNoRows):
		return models.Room{}, ErrNotFound
	case err != nil:
		return models.Room{}, fmt.Errorf("update room: %w", err)
	}
	return room, nil
}

// DeleteRoom removes a room by ID (cascades to workspaces and bookings).
func (r *Repo) DeleteRoom(id int64) error {
	const q = `DELETE FROM rooms WHERE id = $1`
	res, err := r.db.Exec(q, id)
	if err != nil {
		return fmt.Errorf("delete room: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// GetRoomByID returns a single room or ErrNotFound.
func (r *Repo) GetRoomByID(id int64) (models.Room, error) {
	const q = `SELECT id, name, description FROM rooms WHERE id = $1`
	var room models.Room
	switch err := r.db.Get(&room, q, id); {
	case errors.Is(err, sql.ErrNoRows):
		return models.Room{}, ErrNotFound
	case err != nil:
		return models.Room{}, fmt.Errorf("get room: %w", err)
	}
	return room, nil
}
