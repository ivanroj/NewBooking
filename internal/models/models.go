package models

type Room struct {
	ID          int64  `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
}

type Workspace struct {
	ID     int64  `db:"id" json:"id"`
	RoomID int64  `db:"room_id" json:"room_id"`
	Name   string `db:"name" json:"name"`
}

type User struct {
	ID         int64  `db:"id" json:"id"`
	TelegramID string `db:"telegram_id" json:"telegram_id"`
	Role       string `db:"role" json:"role"`
	Email      string `db:"email" json:"email"`
}

type Booking struct {
	ID          int64  `db:"id" json:"id"`
	UserID      int64  `db:"user_id" json:"user_id"`
	WorkspaceID int64  `db:"workspace_id" json:"workspace_id"`
	StartTime   string `db:"start_time" json:"start_time"`
	EndTime     string `db:"end_time" json:"end_time"`
	Status      string `db:"status" json:"status"`
}

type BookingSettings struct {
	ID        int64 `db:"id" json:"id"`
	UserID    int64 `db:"user_id" json:"user_id"`
	MaxActive int   `db:"max_active" json:"max_active"`
}
