package handler

import (
	"net/http"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext, wsServer *WSServer) {
	// WebSocket route - no authentication required at connection time
	// Authentication is handled during the WebSocket handshake
	server.AddRoutes(
		[]rest.Route{
			rest.Route{Method: http.MethodGet, Path: "/ws", Handler: wsServer.handleWebSocket},
		},
	)

	// Public routes - no authentication required
	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodPost, Path: "/v1/auth/register", Handler: RegisterHandler(serverCtx)}, // User registration
			{Method: http.MethodPost, Path: "/v1/auth/login", Handler: LoginHandler(serverCtx)},       // User login
		},
	)

	// Protected routes - JWT authentication required
	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodGet, Path: "/v1/auth/me", Handler: GetMeHandler(serverCtx)},         // Get current user info
			{Method: http.MethodPost, Path: "/v1/auth/refresh", Handler: RefreshHandler(serverCtx)}, // Refresh authentication token
			{Method: http.MethodPost, Path: "/v1/auth/logout", Handler: LogoutHandler(serverCtx)},   // User logout

			{Method: http.MethodGet, Path: "/v1/agents", Handler: ListAgentsHandler(serverCtx)},                       // List agents
			{Method: http.MethodPost, Path: "/v1/agents", Handler: CreateAgentHandler(serverCtx)},                     // Create agent
			{Method: http.MethodGet, Path: "/v1/agents/:agent_id", Handler: GetAgentHandler(serverCtx)},               // Get agent details
			{Method: http.MethodPut, Path: "/v1/agents/:agent_id", Handler: UpdateAgentHandler(serverCtx)},            // Update agent
			{Method: http.MethodDelete, Path: "/v1/agents/:agent_id", Handler: DeleteAgentHandler(serverCtx)},         // Delete agent
			{Method: http.MethodGet, Path: "/v1/agents/:agent_id/skills", Handler: ListAgentSkillsHandler(serverCtx)}, // List agent skills

			{Method: http.MethodGet, Path: "/v1/sessions", Handler: HandleListSessionsHandler(serverCtx)},         // List sessions
			{Method: http.MethodPost, Path: "/v1/sessions", Handler: HandleCreateSessionHandler(serverCtx)},       // Create session
			{Method: http.MethodGet, Path: "/v1/sessions/:id", Handler: HandleGetSessionHandler(serverCtx)},       // Get session details
			{Method: http.MethodDelete, Path: "/v1/sessions/:id", Handler: HandleDeleteSessionHandler(serverCtx)}, // Delete session

			// Providers endpoints
			{Method: http.MethodGet, Path: "/v1/providers", Handler: ListProvidersHandler(serverCtx)},                 // List providers
			{Method: http.MethodPost, Path: "/v1/providers", Handler: CreateProviderHandler(serverCtx)},               // Create provider
			{Method: http.MethodGet, Path: "/v1/providers/:id", Handler: GetProviderHandler(serverCtx)},               // Get provider
			{Method: http.MethodPut, Path: "/v1/providers/:id", Handler: UpdateProviderHandler(serverCtx)},            // Update provider
			{Method: http.MethodDelete, Path: "/v1/providers/:id", Handler: DeleteProviderHandler(serverCtx)},         // Delete provider
			{Method: http.MethodGet, Path: "/v1/providers/:id/models", Handler: ListProviderModelsHandler(serverCtx)}, // List provider models
			{Method: http.MethodPost, Path: "/v1/providers/:id/verify", Handler: VerifyProviderHandler(serverCtx)},    // Verify provider

			// Skills endpoints
			{Method: http.MethodGet, Path: "/v1/skills", Handler: ListSkillsHandler(serverCtx)},                          // List skills
			{Method: http.MethodPost, Path: "/v1/skills", Handler: CreateSkillHandler(serverCtx)},                        // Create skill
			{Method: http.MethodGet, Path: "/v1/skills/:id", Handler: GetSkillHandler(serverCtx)},                        // Get skill
			{Method: http.MethodPost, Path: "/v1/skills/:id/grant", Handler: GrantSkillHandler(serverCtx)},               // Grant skill to agent
			{Method: http.MethodDelete, Path: "/v1/skills/:id/revoke/:agent_id", Handler: RevokeSkillHandler(serverCtx)}, // Revoke skill from agent

			// Usage analytics endpoints
			{Method: http.MethodGet, Path: "/v1/usage/timeseries", Handler: GetUsageTimeSeriesHandler(serverCtx)}, // Get usage time series
			{Method: http.MethodGet, Path: "/v1/usage/summary", Handler: GetUsageSummaryHandler(serverCtx)},       // Get usage summary

			// System health endpoint
			{Method: http.MethodGet, Path: "/v1/system/health", Handler: GetSystemHealthHandler(serverCtx)}, // Get system health
		},
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
	)
}
