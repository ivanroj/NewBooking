package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/example/coworking/internal/repository"
	"github.com/gorilla/mux"
)

type createWSReq struct {
	Name string `json:"name"`
}

type updateWSReq struct {
	Name string `json:"name"`
}

// AdminCreateWorkspace POST /api/rooms/{room_id}/workspaces — admin only.
func (h *Handlers) AdminCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	roomID, err := parseID(mux.Vars(r)["room_id"])
	if err != nil {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid_room_id"})
		return
	}
	if _, err = h.Repo.GetRoomByID(roomID); errors.Is(err, repository.ErrNotFound) {
		_ = writeJSON(w, http.StatusNotFound, errResp{Error: "room_not_found"})
		return
	} else if err != nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
		return
	}

	var req createWSReq
	if err = readJSONBody(r, &req); err != nil || req.Name == "" {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "name_required"})
		return
	}

	ws, err := h.Repo.CreateWorkspace(roomID, req.Name)
	if err != nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
		return
	}
	_ = writeJSON(w, http.StatusCreated, ws)
}

// ListWorkspaces GET /api/rooms/{room_id}/workspaces — protected.
// Optional query: ?start=2006-01-02T15:04:05Z&end=...
func (h *Handlers) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	roomID, err := parseID(mux.Vars(r)["room_id"])
	if err != nil {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid_room_id"})
		return
	}

	var start, end time.Time
	if s := r.URL.Query().Get("start"); s != "" {
		start, _ = time.Parse(time.RFC3339, s)
	}
	if e := r.URL.Query().Get("end"); e != "" {
		end, _ = time.Parse(time.RFC3339, e)
	}

	list, err := h.Repo.ListWorkspaces(roomID, start, end)
	if err != nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
		return
	}
	if list == nil {
		list = []repository.WorkspaceAvailability{}
	}
	_ = writeJSON(w, http.StatusOK, list)
}

// AdminUpdateWorkspace PATCH /api/rooms/{room_id}/workspaces/{ws_id} — admin only.
func (h *Handlers) AdminUpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := parseID(vars["room_id"])
	if err != nil {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid_room_id"})
		return
	}
	wsID, err := parseID(vars["ws_id"])
	if err != nil {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid_ws_id"})
		return
	}

	var req updateWSReq
	if err = readJSONBody(r, &req); err != nil || req.Name == "" {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "name_required"})
		return
	}

	ws, err := h.Repo.UpdateWorkspace(roomID, wsID, req.Name)
	if errors.Is(err, repository.ErrNotFound) {
		_ = writeJSON(w, http.StatusNotFound, errResp{Error: "not_found"})
		return
	}
	if err != nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
		return
	}
	_ = writeJSON(w, http.StatusOK, ws)
}

// AdminDeleteWorkspace DELETE /api/rooms/{room_id}/workspaces/{ws_id} — admin only.
func (h *Handlers) AdminDeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := parseID(vars["room_id"])
	if err != nil {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid_room_id"})
		return
	}
	wsID, err := parseID(vars["ws_id"])
	if err != nil {
		_ = writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid_ws_id"})
		return
	}
	if err = h.Repo.DeleteWorkspace(roomID, wsID); errors.Is(err, repository.ErrNotFound) {
		_ = writeJSON(w, http.StatusNotFound, errResp{Error: "not_found"})
		return
	} else if err != nil {
		_ = writeJSON(w, http.StatusInternalServerError, errResp{Error: "database_error"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
