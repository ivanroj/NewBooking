package router

import (
	"net/http"

	"github.com/example/coworking/internal/handlers"
	"github.com/example/coworking/internal/middleware"
	"github.com/gorilla/mux"
)

type Router struct {
	mux *mux.Router
}

func NewRouter(h *handlers.Handlers) *Router {
	main := mux.NewRouter()

	main.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	api := main.PathPrefix("/api").Subrouter()

	// ── Public auth endpoints ──────────────────────────────────────────────
	api.HandleFunc("/auth/admin/login", h.AdminLogin).Methods(http.MethodPost)
	api.HandleFunc("/auth/student/telegram", h.StudentTelegramAuth).Methods(http.MethodPost)

	// ── Protected (any valid JWT) ──────────────────────────────────────────
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.BearerAuth(h.Auth))

	// Rooms – read
	protected.HandleFunc("/rooms", h.ListRooms).Methods(http.MethodGet)

	// Workspaces – read (student + admin)
	protected.HandleFunc("/rooms/{room_id}/workspaces", h.ListWorkspaces).Methods(http.MethodGet)

	// Bookings – student / self
	protected.HandleFunc("/bookings", h.CreateBooking).Methods(http.MethodPost)
	protected.HandleFunc("/bookings/my", h.ListMyBookings).Methods(http.MethodGet)
	protected.HandleFunc("/bookings/{id}/cancel", h.CancelMyBooking).Methods(http.MethodPatch)

	// ── Admin-only subrouter ───────────────────────────────────────────────
	admin := protected.PathPrefix("").Subrouter()
	admin.Use(middleware.RequireAdmin)

	// Rooms CRUD
	admin.HandleFunc("/rooms", h.AdminCreateRoom).Methods(http.MethodPost)
	admin.HandleFunc("/rooms/{id}", h.AdminUpdateRoom).Methods(http.MethodPatch)
	admin.HandleFunc("/rooms/{id}", h.AdminDeleteRoom).Methods(http.MethodDelete)

	// Workspaces CRUD
	admin.HandleFunc("/rooms/{room_id}/workspaces", h.AdminCreateWorkspace).Methods(http.MethodPost)
	admin.HandleFunc("/rooms/{room_id}/workspaces/{ws_id}", h.AdminUpdateWorkspace).Methods(http.MethodPatch)
	admin.HandleFunc("/rooms/{room_id}/workspaces/{ws_id}", h.AdminDeleteWorkspace).Methods(http.MethodDelete)

	// Admin bookings & settings
	admin.HandleFunc("/admin/bookings", h.AdminListBookings).Methods(http.MethodGet)
	admin.HandleFunc("/admin/bookings/{id}", h.AdminUpdateBooking).Methods(http.MethodPatch)
	admin.HandleFunc("/admin/settings/booking-limit", h.AdminGetBookingLimit).Methods(http.MethodGet)
	admin.HandleFunc("/admin/settings/booking-limit", h.AdminSetBookingLimit).Methods(http.MethodPatch)

	// Admin stats
	admin.HandleFunc("/admin/stats", h.AdminGetStats).Methods(http.MethodGet)

	return &Router{mux: main}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
