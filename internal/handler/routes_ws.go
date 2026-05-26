package handler

import (
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/zeromicro/go-zero/rest"
	"net/http"
)

func RegisterWsHandlers(server *rest.Server, serverCtx *svc.ServiceContext, wsServer *WSServer) {

	server.AddRoutes(
		[]rest.Route{
			{
				// User login
				Method:  http.MethodPost,
				Path:    "/v1/auth/login",
				Handler: LoginHandler(serverCtx),
			},
			{
				// User logout
				Method:  http.MethodPost,
				Path:    "/v1/auth/logout",
				Handler: LogoutHandler(serverCtx),
			},
			{
				// Get current user info
				Method:  http.MethodGet,
				Path:    "/v1/auth/me",
				Handler: GetMeHandler(serverCtx),
			},
			{
				// Refresh authentication token
				Method:  http.MethodPost,
				Path:    "/v1/auth/refresh",
				Handler: RefreshHandler(serverCtx),
			},
			{
				// User registration
				Method:  http.MethodPost,
				Path:    "/v1/auth/register",
				Handler: RegisterHandler(serverCtx),
			},
		},
	)

	// WebSocket route - no authentication required at connection time
	// Authentication is handled during the WebSocket handshake
	server.AddRoutes(
		[]rest.Route{
			rest.Route{Method: http.MethodGet, Path: "/ws", Handler: wsServer.handleWebSocket},
		},
	)
}
