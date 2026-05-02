package handler

import (
	"net/http"

	"github.com/pomclaw/pomclaw/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodGet, Path: "/api/sessions", Handler: HandleListSessionsHandler(serverCtx)},         // List sessions
			{Method: http.MethodPost, Path: "/api/sessions", Handler: HandleCreateSessionHandler(serverCtx)},       // Create session
			{Method: http.MethodGet, Path: "/api/sessions/:id", Handler: HandleGetSessionHandler(serverCtx)},       // Get session details
			{Method: http.MethodDelete, Path: "/api/sessions/:id", Handler: HandleDeleteSessionHandler(serverCtx)}, // Delete session

			{Method: http.MethodGet, Path: "/api/v1/agents", Handler: ListAgentsHandler(serverCtx)},               // List agents
			{Method: http.MethodPost, Path: "/api/v1/agents", Handler: CreateAgentHandler(serverCtx)},             // Create agent
			{Method: http.MethodGet, Path: "/api/v1/agents/:agent_id", Handler: GetAgentHandler(serverCtx)},       // Get agent details
			{Method: http.MethodPut, Path: "/api/v1/agents/:agent_id", Handler: UpdateAgentHandler(serverCtx)},    // Update agent
			{Method: http.MethodDelete, Path: "/api/v1/agents/:agent_id", Handler: DeleteAgentHandler(serverCtx)}, // Delete agent

			{Method: http.MethodPost, Path: "/api/v1/auth/login", Handler: LoginHandler(serverCtx)},       // User login
			{Method: http.MethodPost, Path: "/api/v1/auth/logout", Handler: LogoutHandler(serverCtx)},     // User logout
			{Method: http.MethodGet, Path: "/api/v1/auth/me", Handler: GetMeHandler(serverCtx)},           // Get current user info
			{Method: http.MethodPost, Path: "/api/v1/auth/refresh", Handler: RefreshHandler(serverCtx)},   // Refresh authentication token
			{Method: http.MethodPost, Path: "/api/v1/auth/register", Handler: RegisterHandler(serverCtx)}, // User registration
		},
	)
}
