package handler

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/pomclaw/pomclaw/pkg/protocol"
)

// WSMethodRouter maps method names to handlers.
type WSMethodRouter struct {
	handlers map[string]MethodHandler
	server   *WSServer
}

// NewWSMethodRouter creates a new method router.
func NewWSMethodRouter(server *WSServer) *WSMethodRouter {
	r := &WSMethodRouter{
		handlers: make(map[string]MethodHandler),
		server:   server,
	}
	r.registerDefaults()
	return r
}

// Register adds a method handler.
func (r *WSMethodRouter) Register(method string, handler MethodHandler) {
	r.handlers[method] = handler
}

// Handle dispatches a request to the appropriate handler.
func (r *WSMethodRouter) Handle(ctx context.Context, client ClientInterface, req *protocol.RequestFrame) {
	handler, ok := r.handlers[req.Method]
	if !ok {
		logx.Info("unknown method:", map[string]interface{}{
			"method": req.Method,
			"client": client.ID(),
		})
		client.SendResponse(protocol.NewErrorResponse(
			req.ID,
			protocol.ErrInvalidRequest,
			"unknown method: "+req.Method,
		))
		return
	}

	logx.Debug("handling method:", map[string]interface{}{
		"method": req.Method,
		"client": client.ID(),
		"req_id": req.ID,
	})
	handler(ctx, client, req)
}

// registerDefaults registers built-in method handlers.
func (r *WSMethodRouter) registerDefaults() {
	r.Register(protocol.MethodConnect, r.handleConnect)
	r.Register(protocol.MethodHealth, r.handleHealth)
}

// handleConnect processes the connect handshake.
func (r *WSMethodRouter) handleConnect(ctx context.Context, client ClientInterface, req *protocol.RequestFrame) {
	var params struct {
		UserID string `json:"user_id"`
	}
	if req.Params != nil {
		json.Unmarshal(req.Params, &params)
	}

	// For Phase 1, simplified auth: accept any connection
	// Set user ID on client if provided
	if c, ok := client.(*WSClient); ok {
		c.userID = params.UserID
		if c.userID == "" {
			c.userID = "anonymous"
		}
		c.role = "user" // Default role
	}

	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]any{
		"protocol": protocol.ProtocolVersion,
		"user_id":  params.UserID,
		"server": map[string]any{
			"name":    "pomclaw",
			"version": "0.1.0",
		},
	}))

	logx.Info("client connected:", map[string]interface{}{
		"client":  client.ID(),
		"user_id": params.UserID,
	})
}

// handleHealth returns server health status.
func (r *WSMethodRouter) handleHealth(ctx context.Context, client ClientInterface, req *protocol.RequestFrame) {
	clientList := r.server.ClientList()

	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]any{
		"status":    "ok",
		"version":   "0.1.0",
		"clients":   len(clientList),
		"currentId": client.ID(),
	}))
}
