package gateway

import (
	"database/sql"
	"net/http"

	"github.com/pomclaw/pomclaw/pkg/channels/gateway/handlers"
)

// setupAPIRoutes registers all REST API routes on mux.
func setupAPIRoutes(mux *http.ServeMux, db *sql.DB, secret string) {
	h := &handlers.Handler{DB: db, Secret: secret}

	// Auth (no JWT required)
	mux.HandleFunc("POST /api/v1/auth/register", h.Register)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)

	// Auth (JWT required)
	mux.Handle("GET /api/v1/auth/me", jwtMiddleware(secret, h.GetMe))
	mux.Handle("POST /api/v1/auth/refresh", jwtMiddleware(secret, h.Refresh))
	mux.Handle("POST /api/v1/auth/logout", jwtMiddleware(secret, h.Logout))

	// Agents (JWT required)
	mux.Handle("GET /api/v1/agents", jwtMiddleware(secret, h.ListAgents))
	mux.Handle("POST /api/v1/agents", jwtMiddleware(secret, h.CreateAgent))
	mux.Handle("GET /api/v1/agents/{agent_id}", jwtMiddleware(secret, h.GetAgent))
	mux.Handle("PUT /api/v1/agents/{agent_id}", jwtMiddleware(secret, h.UpdateAgent))
	mux.Handle("DELETE /api/v1/agents/{agent_id}", jwtMiddleware(secret, h.DeleteAgent))

	// Sessions (JWT required)
	mux.Handle("GET /api/v1/sessions", jwtMiddleware(secret, h.ListSessions))
	mux.Handle("POST /api/v1/sessions", jwtMiddleware(secret, h.CreateSession))
	mux.Handle("GET /api/v1/sessions/{session_id}", jwtMiddleware(secret, h.GetSession))
}
