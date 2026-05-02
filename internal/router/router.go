package router

import (
    "net/http"

    "github.com/example/coworking/internal/repository"
    "github.com/gorilla/mux"
)

type Router struct {
    mux *mux.Router
}

func NewRouter(repo *repository.Repo) *Router {
    r := &Router{mux: mux.NewRouter()}
    // Health check
    r.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    }).Methods(http.MethodGet)
    // TODO: add more routes (rooms, auth, bookings)
    return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    r.mux.ServeHTTP(w, req)
}
