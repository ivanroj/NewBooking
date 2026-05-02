package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/example/coworking/internal/models"
)

// GetAdminByEmail returns an admin row including password_hash (not exposed via JSON elsewhere).
func (r *Repo) GetAdminByEmail(email string) (models.User, error) {
	const q = `SELECT id, telegram_id, role, email, password_hash FROM users WHERE email = $1 AND role = 'admin' LIMIT 1`
	var u models.User
	switch err := r.db.Get(&u, q, email); {
	case errors.Is(err, sql.ErrNoRows):
		return models.User{}, err
	case err != nil:
		return models.User{}, fmt.Errorf("get admin: %w", err)
	default:
		return u, nil
	}
}

// UpsertStudentTelegramUser creates or updates the student keyed by telegram_id.
func (r *Repo) UpsertStudentTelegramUser(telegramNumericID int64) (models.User, error) {
	tg := strconv.FormatInt(telegramNumericID, 10)
	const q = `
INSERT INTO users (telegram_id, role, email, password_hash)
VALUES ($1, 'student', NULL, NULL)
ON CONFLICT (telegram_id)
DO UPDATE SET role = excluded.role
RETURNING id, telegram_id, role, email, password_hash`

	var u models.User
	if err := r.db.Get(&u, q, tg); err != nil {
		return models.User{}, fmt.Errorf("upsert student: %w", err)
	}
	return u, nil
}
