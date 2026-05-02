package router

import (
	"net/http"

	"github.com/example/coworking/internal/repository"
	"github.com/gorilla/mux"
)

type Router struct {
	mux *mux.Router
}

func NewRouter(_ *repository.Repo) *Router {
	r := &Router{mux: mux.NewRouter()}
	r.mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
