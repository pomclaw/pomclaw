package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/pomclaw/pomclaw/pkg/channels/gateway/store"
)

type agentReq struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	SystemPrompt string          `json:"system_prompt"`
	Model        string          `json:"model"`
	Tools        json.RawMessage `json:"tools"`
}

// ListAgents returns all agents for the authenticated user.
func (h *Handler) ListAgents(w http.ResponseWriter, r *http.Request) {
	userID := userIDFrom(r.Context())
	agents, err := store.ListAgents(h.DB, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	if agents == nil {
		agents = []*store.Agent{}
	}
	writeJSON(w, http.StatusOK, agents)
}

// CreateAgent creates a new agent for the authenticated user.
func (h *Handler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	userID := userIDFrom(r.Context())

	var req agentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}
	if req.Name == "" || req.Model == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "name and model are required")
		return
	}
	if req.Tools == nil {
		req.Tools = json.RawMessage("[]")
	}

	agent, err := store.CreateAgent(h.DB, userID, req.Name, req.Description, req.SystemPrompt, req.Model, req.Tools)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, agent)
}

// GetAgent returns a single agent owned by the authenticated user.
func (h *Handler) GetAgent(w http.ResponseWriter, r *http.Request) {
	userID := userIDFrom(r.Context())
	agentID := r.PathValue("agent_id")

	agent, err := store.GetAgent(h.DB, agentID, userID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not_found", "agent not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, agent)
}

// UpdateAgent updates a single agent owned by the authenticated user.
func (h *Handler) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	userID := userIDFrom(r.Context())
	agentID := r.PathValue("agent_id")

	var req agentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}
	if req.Name == "" || req.Model == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "name and model are required")
		return
	}
	if req.Tools == nil {
		req.Tools = json.RawMessage("[]")
	}

	agent, err := store.UpdateAgent(h.DB, agentID, userID, req.Name, req.Description, req.SystemPrompt, req.Model, req.Tools)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not_found", "agent not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, agent)
}

// DeleteAgent removes an agent owned by the authenticated user.
func (h *Handler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	userID := userIDFrom(r.Context())
	agentID := r.PathValue("agent_id")

	err := store.DeleteAgent(h.DB, agentID, userID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not_found", "agent not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
