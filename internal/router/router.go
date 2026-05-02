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
	api.HandleFunc("/auth/admin/login", h.AdminLogin).Methods(http.MethodPost)
	api.HandleFunc("/auth/student/telegram", h.StudentTelegramAuth).Methods(http.MethodPost)

	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.BearerAuth(h.Auth))
	protected.HandleFunc("/rooms", h.ListRooms).Methods(http.MethodGet)

	return &Router{mux: main}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
