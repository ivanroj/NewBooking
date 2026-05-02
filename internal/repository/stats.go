package repository

import "fmt"

// RoomLoad holds booking count per room.
type RoomLoad struct {
	RoomID   int64  `db:"room_id" json:"room_id"`
	RoomName string `db:"room_name" json:"room_name"`
	Count    int    `db:"cnt" json:"count"`
}

// PopularWorkspace holds booking count per workspace.
type PopularWorkspace struct {
	WorkspaceID   int64  `db:"workspace_id" json:"workspace_id"`
	WorkspaceName string `db:"workspace_name" json:"workspace_name"`
	RoomName      string `db:"room_name" json:"room_name"`
	Count         int    `db:"cnt" json:"count"`
}

// DayDistribution holds booking count per day of week (0=Sun..6=Sat).
type DayDistribution struct {
	Day   int `db:"dow" json:"day"`
	Count int `db:"cnt" json:"count"`
}

// StatsResult aggregates analytics data.
type StatsResult struct {
	TotalBookings  int                `json:"total_bookings"`
	ActiveBookings int                `json:"active_bookings"`
	RoomLoad       []RoomLoad         `json:"room_load"`
	TopWorkspaces  []PopularWorkspace `json:"top_workspaces"`
	DayDist        []DayDistribution  `json:"day_distribution"`
}

// GetStats computes analytics for all bookings in the given date range.
// If dateFrom is empty, no date filter is applied.
func (r *Repo) GetStats(dateFrom, dateTo string) (StatsResult, error) {
	var res StatsResult

	dateFilter := ""
	args := []any{}
	if dateFrom != "" && dateTo != "" {
		dateFilter = " AND b.start_time >= $1 AND b.start_time < ($2::date + interval '1 day')"
		args = append(args, dateFrom, dateTo)
	}

	// Total bookings
	qTotal := fmt.Sprintf(`SELECT COUNT(*) FROM bookings b WHERE 1=1%s`, dateFilter)
	if err := r.db.Get(&res.TotalBookings, qTotal, args...); err != nil {
		return res, fmt.Errorf("stats total: %w", err)
	}

	// Active bookings
	qActive := fmt.Sprintf(`SELECT COUNT(*) FROM bookings b WHERE b.status = 'active'%s`, dateFilter)
	if err := r.db.Get(&res.ActiveBookings, qActive, args...); err != nil {
		return res, fmt.Errorf("stats active: %w", err)
	}

	// Room load
	qRoom := fmt.Sprintf(`
SELECT w.room_id, r.name AS room_name, COUNT(*) AS cnt
FROM bookings b
JOIN workspaces w ON w.id = b.workspace_id
JOIN rooms r ON r.id = w.room_id
WHERE 1=1%s
GROUP BY w.room_id, r.name
ORDER BY cnt DESC`, dateFilter)
	if err := r.db.Select(&res.RoomLoad, qRoom, args...); err != nil {
		return res, fmt.Errorf("stats room load: %w", err)
	}
	if res.RoomLoad == nil {
		res.RoomLoad = []RoomLoad{}
	}

	// Top 5 workspaces
	qTop := fmt.Sprintf(`
SELECT b.workspace_id, w.name AS workspace_name, r.name AS room_name, COUNT(*) AS cnt
FROM bookings b
JOIN workspaces w ON w.id = b.workspace_id
JOIN rooms r ON r.id = w.room_id
WHERE 1=1%s
GROUP BY b.workspace_id, w.name, r.name
ORDER BY cnt DESC
LIMIT 5`, dateFilter)
	if err := r.db.Select(&res.TopWorkspaces, qTop, args...); err != nil {
		return res, fmt.Errorf("stats top ws: %w", err)
	}
	if res.TopWorkspaces == nil {
		res.TopWorkspaces = []PopularWorkspace{}
	}

	// Day of week distribution
	qDay := fmt.Sprintf(`
SELECT EXTRACT(DOW FROM b.start_time)::int AS dow, COUNT(*) AS cnt
FROM bookings b
WHERE 1=1%s
GROUP BY dow
ORDER BY dow`, dateFilter)
	if err := r.db.Select(&res.DayDist, qDay, args...); err != nil {
		return res, fmt.Errorf("stats day dist: %w", err)
	}
	if res.DayDist == nil {
		res.DayDist = []DayDistribution{}
	}

	return res, nil
}
