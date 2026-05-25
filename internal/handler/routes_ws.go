package handler

import (
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/zeromicro/go-zero/rest"
	"net/http"
)

func RegisterWsHandlers(server *rest.Server, serverCtx *svc.ServiceContext, wsServer *WSServer) {
	// WebSocket route - no authentication required at connection time
	// Authentication is handled during the WebSocket handshake
	server.AddRoutes(
		[]rest.Route{
			rest.Route{Method: http.MethodGet, Path: "/ws", Handler: wsServer.handleWebSocket},
		},
	)
}
