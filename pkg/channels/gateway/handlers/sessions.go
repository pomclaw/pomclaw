package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/pomclaw/pomclaw/pkg/channels/gateway/store"
)

type createSessionReq struct {
	AgentID string `json:"agent_id"`
	Title   string `json:"title"`
}

type sessionListItem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Preview      string `json:"preview"`
	MessageCount int    `json:"message_count"`
	Created      string `json:"created"`
	Updated      string `json:"updated"`
}

type sessionChatMessage struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Media   []string `json:"media,omitempty"`
}

// HandleListSessions returns a list of session summaries with pagination.
//
//	GET /api/sessions?offset=0&limit=20
func (h *Handler) HandleListSessions(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")

	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	offset := 0
	limit := 20

	if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
		offset = val
	}
	if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
		limit = val
	}

	items, err := store.ListSessionsWithPagination(h.DB, agentID, offset, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	if items == nil {
		items = []map[string]interface{}{}
	}
	writeJSON(w, http.StatusOK, items)
}

// HandleCreateSession creates a new session for the authenticated user.
//
//	POST /api/sessions
func (h *Handler) HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req createSessionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}
	if req.AgentID == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "agent_id is required")
		return
	}

	session, err := store.CreateSession(h.DB, "", req.AgentID, req.Title)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":       session.ID,
		"messages": []interface{}{},
		"summary":  "",
		"created":  session.CreatedAt.Format(time.RFC3339),
		"updated":  session.CreatedAt.Format(time.RFC3339),
	})
}

// HandleGetSession returns the full message history for a specific session.
//
//	GET /api/sessions/{id}
func (h *Handler) HandleGetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "missing session id")
		return
	}

	session, err := store.GetSessionWithMessages(h.DB, sessionID)
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

// HandleDeleteSession deletes a specific session.
//
//	DELETE /api/sessions/{id}
func (h *Handler) HandleDeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "missing session id")
		return
	}

	err := store.DeleteSession(h.DB, sessionID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not_found", "session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
