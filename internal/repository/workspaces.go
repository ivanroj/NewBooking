package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/example/coworking/internal/models"
)

// WorkspaceAvailability extends Workspace with an availability flag for a given time slot.
type WorkspaceAvailability struct {
	models.Workspace
	Available bool `db:"available" json:"available"`
}

// CreateWorkspace inserts a new workspace in the given room.
func (r *Repo) CreateWorkspace(roomID int64, name string) (models.Workspace, error) {
	const q = `INSERT INTO workspaces (room_id, name) VALUES ($1, $2) RETURNING id, room_id, name`
	var ws models.Workspace
	if err := r.db.Get(&ws, q, roomID, name); err != nil {
		return models.Workspace{}, fmt.Errorf("create workspace: %w", err)
	}
	return ws, nil
}

// ListWorkspaces returns all workspaces for a room with availability flag for the given slot.
// If start.IsZero() no availability check is performed (available = true always).
func (r *Repo) ListWorkspaces(roomID int64, start, end time.Time) ([]WorkspaceAvailability, error) {
	const qSimple = `SELECT id, room_id, name, TRUE AS available FROM workspaces WHERE room_id = $1 ORDER BY id`
	const qAvail = `
SELECT w.id, w.room_id, w.name,
       NOT EXISTS (
           SELECT 1 FROM bookings b
           WHERE b.workspace_id = w.id
             AND b.status = 'active'
             AND b.start_time < $2
             AND b.end_time   > $1
       ) AS available
FROM workspaces w
WHERE w.room_id = $3
ORDER BY w.id`

	var out []WorkspaceAvailability
	var err error
	if start.IsZero() {
		err = r.db.Select(&out, qSimple, roomID)
	} else {
		err = r.db.Select(&out, qAvail, start, end, roomID)
	}
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	return out, nil
}

// UpdateWorkspace renames a workspace.
func (r *Repo) UpdateWorkspace(roomID, wsID int64, name string) (models.Workspace, error) {
	const q = `
UPDATE workspaces SET name = $1 WHERE id = $2 AND room_id = $3
RETURNING id, room_id, name`
	var ws models.Workspace
	switch err := r.db.Get(&ws, q, name, wsID, roomID); {
	case errors.Is(err, sql.ErrNoRows):
		return models.Workspace{}, ErrNotFound
	case err != nil:
		return models.Workspace{}, fmt.Errorf("update workspace: %w", err)
	}
	return ws, nil
}

// DeleteWorkspace removes a workspace (cascades to bookings).
func (r *Repo) DeleteWorkspace(roomID, wsID int64) error {
	const q = `DELETE FROM workspaces WHERE id = $1 AND room_id = $2`
	res, err := r.db.Exec(q, wsID, roomID)
	if err != nil {
		return fmt.Errorf("delete workspace: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
