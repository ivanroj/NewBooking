package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/example/coworking/internal/models"
	"github.com/jmoiron/sqlx"
)

// BookingListParams holds optional filters for listing bookings.
type BookingListParams struct {
	UserID int64
	Status string // empty = all
	Page   int    // 1-based
	Limit  int
}

// BookingListResult is a paginated list of bookings.
type BookingListResult struct {
	Items []models.Booking `json:"items"`
	Total int              `json:"total"`
}

// CreateBooking inserts a new booking inside a transaction:
//   - checks for time-slot conflicts (FOR UPDATE)
//   - checks active-booking limit against the global setting
//
// Returns ErrConflict or ErrLimitExceeded on business-rule violations.
func (r *Repo) CreateBooking(userID, workspaceID int64, start, end time.Time) (models.Booking, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return models.Booking{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Conflict check (lock conflicting rows).
	var conflictCount int
	const qConflict = `
SELECT COUNT(*) FROM bookings
WHERE workspace_id = $1
  AND status       = 'active'
  AND start_time   < $2
  AND end_time     > $3
FOR UPDATE`
	if err = tx.Get(&conflictCount, qConflict, workspaceID, end, start); err != nil {
		return models.Booking{}, fmt.Errorf("conflict check: %w", err)
	}
	if conflictCount > 0 {
		return models.Booking{}, ErrConflict
	}

	// 2. Limit check.
	limit, err := getGlobalLimitTx(tx)
	if err != nil {
		return models.Booking{}, err
	}

	var activeCount int
	const qActive = `SELECT COUNT(*) FROM bookings WHERE user_id = $1 AND status = 'active'`
	if err = tx.Get(&activeCount, qActive, userID); err != nil {
		return models.Booking{}, fmt.Errorf("limit check: %w", err)
	}
	if activeCount >= limit {
		return models.Booking{}, ErrLimitExceeded
	}

	// 3. Insert.
	const qInsert = `
INSERT INTO bookings (user_id, workspace_id, start_time, end_time, status)
VALUES ($1, $2, $3, $4, 'active')
RETURNING id, user_id, workspace_id, start_time, end_time, status`
	var b models.Booking
	if err = tx.Get(&b, qInsert, userID, workspaceID, start, end); err != nil {
		return models.Booking{}, fmt.Errorf("insert booking: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return models.Booking{}, fmt.Errorf("commit tx: %w", err)
	}
	return b, nil
}

// CancelMyBooking sets status='canceled' only if the booking belongs to userID.
func (r *Repo) CancelMyBooking(userID, bookingID int64) (models.Booking, error) {
	const q = `
UPDATE bookings SET status = 'canceled'
WHERE id = $1 AND user_id = $2 AND status = 'active'
RETURNING id, user_id, workspace_id, start_time, end_time, status`
	var b models.Booking
	switch err := r.db.Get(&b, q, bookingID, userID); {
	case errors.Is(err, sql.ErrNoRows):
		return models.Booking{}, ErrNotFound
	case err != nil:
		return models.Booking{}, fmt.Errorf("cancel booking: %w", err)
	}
	return b, nil
}

// AdminUpdateBooking sets an arbitrary status on any booking (admin use).
func (r *Repo) AdminUpdateBooking(bookingID int64, status string) (models.Booking, error) {
	const q = `
UPDATE bookings SET status = $1
WHERE id = $2
RETURNING id, user_id, workspace_id, start_time, end_time, status`
	var b models.Booking
	switch err := r.db.Get(&b, q, status, bookingID); {
	case errors.Is(err, sql.ErrNoRows):
		return models.Booking{}, ErrNotFound
	case err != nil:
		return models.Booking{}, fmt.Errorf("admin update booking: %w", err)
	}
	return b, nil
}

// ListMyBookings returns a paginated list of bookings for the given user.
func (r *Repo) ListMyBookings(p BookingListParams) (BookingListResult, error) {
	if p.Limit <= 0 {
		p.Limit = 20
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	offset := (p.Page - 1) * p.Limit

	args := []any{p.UserID}
	statusFilter := ""
	if p.Status != "" {
		args = append(args, p.Status)
		statusFilter = fmt.Sprintf("AND status = $%d", len(args))
	}

	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM bookings WHERE user_id = $1 %s`, statusFilter)
	var total int
	if err := r.db.Get(&total, countQ, args...); err != nil {
		return BookingListResult{}, fmt.Errorf("count bookings: %w", err)
	}

	args = append(args, p.Limit, offset)
	listQ := fmt.Sprintf(`
SELECT id, user_id, workspace_id, start_time, end_time, status
FROM bookings
WHERE user_id = $1 %s
ORDER BY start_time DESC
LIMIT $%d OFFSET $%d`, statusFilter, len(args)-1, len(args))

	var items []models.Booking
	if err := r.db.Select(&items, listQ, args...); err != nil {
		return BookingListResult{}, fmt.Errorf("list bookings: %w", err)
	}
	if items == nil {
		items = []models.Booking{}
	}
	return BookingListResult{Items: items, Total: total}, nil
}

// AdminBookingView is a booking with joined workspace and room names.
type AdminBookingView struct {
	ID            int64  `db:"id" json:"id"`
	UserID        int64  `db:"user_id" json:"user_id"`
	WorkspaceID   int64  `db:"workspace_id" json:"workspace_id"`
	WorkspaceName string `db:"workspace_name" json:"workspace_name"`
	RoomName      string `db:"room_name" json:"room_name"`
	StartTime     string `db:"start_time" json:"start_time"`
	EndTime       string `db:"end_time" json:"end_time"`
	Status        string `db:"status" json:"status"`
}

// AdminBookingListResult is a paginated list of admin booking views.
type AdminBookingListResult struct {
	Items []AdminBookingView `json:"items"`
	Total int                `json:"total"`
}

// AdminListBookings returns all bookings with joined workspace/room names for admin.
func (r *Repo) AdminListBookings(page, limit int, status string, roomID int64) (AdminBookingListResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	where := "WHERE 1=1"
	args := []any{}
	argN := 0

	if status != "" {
		argN++
		where += fmt.Sprintf(" AND b.status = $%d", argN)
		args = append(args, status)
	}
	if roomID > 0 {
		argN++
		where += fmt.Sprintf(" AND w.room_id = $%d", argN)
		args = append(args, roomID)
	}

	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM bookings b
		JOIN workspaces w ON w.id = b.workspace_id %s`, where)
	var total int
	if err := r.db.Get(&total, countQ, args...); err != nil {
		return AdminBookingListResult{}, fmt.Errorf("count admin bookings: %w", err)
	}

	argN++
	limitIdx := argN
	argN++
	offsetIdx := argN
	args = append(args, limit, offset)

	listQ := fmt.Sprintf(`
SELECT b.id, b.user_id, b.workspace_id, w.name AS workspace_name,
       r.name AS room_name, b.start_time, b.end_time, b.status
FROM bookings b
JOIN workspaces w ON w.id = b.workspace_id
JOIN rooms r ON r.id = w.room_id
%s
ORDER BY b.start_time DESC
LIMIT $%d OFFSET $%d`, where, limitIdx, offsetIdx)

	var items []AdminBookingView
	if err := r.db.Select(&items, listQ, args...); err != nil {
		return AdminBookingListResult{}, fmt.Errorf("list admin bookings: %w", err)
	}
	if items == nil {
		items = []AdminBookingView{}
	}
	return AdminBookingListResult{Items: items, Total: total}, nil
}

// getGlobalLimitTx reads booking_limit from global_settings inside an existing transaction.
func getGlobalLimitTx(tx *sqlx.Tx) (int, error) { //nolint:unparam // error kept for future use
	var raw string
	const q = `SELECT value FROM global_settings WHERE key = 'booking_limit'`
	if err := tx.Get(&raw, q); err != nil {
		return 3, nil // safe default if not found
	}
	var limit int
	if _, err := fmt.Sscanf(raw, "%d", &limit); err != nil || limit < 0 {
		return 3, nil
	}
	return limit, nil
}
