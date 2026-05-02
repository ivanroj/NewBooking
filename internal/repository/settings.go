package repository

import (
	"fmt"
)

// GetBookingLimit returns the global active-booking limit.
func (r *Repo) GetBookingLimit() (int, error) {
	var raw string
	const q = `SELECT value FROM global_settings WHERE key = 'booking_limit'`
	if err := r.db.Get(&raw, q); err != nil {
		return 3, nil
	}
	var limit int
	if _, err := fmt.Sscanf(raw, "%d", &limit); err != nil || limit < 0 {
		return 3, nil
	}
	return limit, nil
}

// SetBookingLimit updates the global active-booking limit.
func (r *Repo) SetBookingLimit(limit int) error {
	const q = `
INSERT INTO global_settings (key, value) VALUES ('booking_limit', $1)
ON CONFLICT (key) DO UPDATE SET value = excluded.value`
	if _, err := r.db.Exec(q, fmt.Sprintf("%d", limit)); err != nil {
		return fmt.Errorf("set booking limit: %w", err)
	}
	return nil
}
