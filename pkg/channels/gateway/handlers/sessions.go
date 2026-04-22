package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/pomclaw/pomclaw/pkg/channels/gateway/store"
)

type createSessionReq struct {
	AgentID string `json:"agent_id"`
	Title   string `json:"title"`
}

// ListSessions returns all sessions for the authenticated user.
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID := userIDFrom(r.Context())
	sessions, err := store.ListSessions(h.DB, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	if sessions == nil {
		sessions = []*store.GatewaySession{}
	}
	writeJSON(w, http.StatusOK, sessions)
}

// CreateSession creates a new session for the authenticated user.
func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	userID := userIDFrom(r.Context())

	var req createSessionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}
	if req.AgentID == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "agent_id is required")
		return
	}

	session, err := store.CreateSession(h.DB, userID, req.AgentID, req.Title)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

// GetSession returns a single session owned by the authenticated user.
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	userID := userIDFrom(r.Context())
	sessionID := r.PathValue("session_id")

	session, err := store.GetSession(h.DB, sessionID, userID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not_found", "session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, session)
}
