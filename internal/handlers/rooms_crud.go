package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/example/coworking/internal/repository"
	"github.com/gorilla/mux"
)

type createRoomReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateRoomReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AdminCreateRoom POST /api/rooms — admin only.
func (h *Handlers) AdminCreateRoom(w http.ResponseWriter, r *http.Request) {
	var req createRoomReq
	if err := readJSONBody(r, &req); err != nil {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "bad_json"})
		return
	}
	if req.Name == "" {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "name_required"})
		return
	}
	room, err := h.Repo.CreateRoom(req.Name, req.Description)
	if err != nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
		return
	}
	_ = writeJSON(w, http.StatusCreated, room)
}

// AdminUpdateRoom PATCH /api/rooms/{id} — admin only.
func (h *Handlers) AdminUpdateRoom(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(mux.Vars(r)["id"])
	if err != nil {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid_id"})
		return
	}
	// Fetch existing to allow partial update.
	existing, err := h.Repo.GetRoomByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		_ = writeJSON(w, http.StatusNotFound, errResp{Error: "not_found"})
		return
	}
	if err != nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
		return
	}

	var req updateRoomReq
	_ = readJSONBody(r, &req)
	if req.Name == "" {
		req.Name = existing.Name
	}
	if req.Description == "" {
		req.Description = existing.Description
	}

	room, err := h.Repo.UpdateRoom(id, req.Name, req.Description)
	if errors.Is(err, repository.ErrNotFound) {
		_ = writeJSON(w, http.StatusNotFound, errResp{Error: "not_found"})
		return
	}
	if err != nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
		return
	}
	_ = writeJSON(w, http.StatusOK, room)
}

// AdminDeleteRoom DELETE /api/rooms/{id} — admin only.
func (h *Handlers) AdminDeleteRoom(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(mux.Vars(r)["id"])
	if err != nil {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid_id"})
		return
	}
	if err := h.Repo.DeleteRoom(id); errors.Is(err, repository.ErrNotFound) {
		_ = writeJSON(w, http.StatusNotFound, errResp{Error: "not_found"})
		return
	} else if err != nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
