package handler

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/pomclaw/pomclaw/pkg/protocol"
)

// MethodHandler processes a single RPC method request.
// Adapted from GoClaw's Protocol v3 implementation.
type MethodHandler func(ctx context.Context, client *WSClient, req *protocol.RequestFrame)

// WSMethodRouter maps method names to handlers.
// Phase 1: Simplified auth with no permission checks.
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
func (r *WSMethodRouter) Handle(ctx context.Context, client *WSClient, req *protocol.RequestFrame) {
	handler, ok := r.handlers[req.Method]
	if !ok {
		logx.Infof("unknown method: %s (client: %s)", req.Method, client.id)
		client.SendResponse(protocol.NewErrorResponse(
			req.ID,
			protocol.ErrInvalidRequest,
			"unknown method: "+req.Method,
		))
		return
	}

	// Phase 1: No permission checks (simplified auth)
	// Future: Add permission engine check here

	logx.Debugf("handling method: %s (client: %s, req_id: %s)", req.Method, client.id, req.ID)
	handler(ctx, client, req)
}

// registerDefaults registers built-in method handlers.
func (r *WSMethodRouter) registerDefaults() {
	r.Register(protocol.MethodConnect, r.handleConnect)
	r.Register(protocol.MethodHealth, r.handleHealth)

	// Register chat methods (Phase 1)
	chatHandler := NewChatHandlerV3(r.server.serverCtx, r.server.rateLimiter)
	chatHandler.Register(r)

	// Register sessions methods
	sessionsHandler := NewSessionsMethods(r.server.serverCtx)
	sessionsHandler.Register(r)
}

// handleConnect processes the connect handshake.
// Phase 1: Simplified auth - accept all connections with userID.
func (r *WSMethodRouter) handleConnect(ctx context.Context, client *WSClient, req *protocol.RequestFrame) {
	var params struct {
		UserID   string `json:"user_id"`
		SenderID string `json:"sender_id"`        // optional: client identifier for reconnection
		Locale   string `json:"locale"`           // user's preferred locale (e.g. "en", "zh", "vi")
		Protocol int    `json:"protocol_version"` // protocol version (should be 3)
	}
	if req.Params != nil {
		json.Unmarshal(req.Params, &params)
	}

	// Phase 1: Accept all connections (no token check)
	client.authenticated = true
	client.userID = params.UserID
	if client.userID == "" {
		client.userID = "anonymous"
	}
	if params.Locale != "" {
		client.locale = params.Locale
	} else {
		client.locale = "en"
	}

	logx.Infof("client authenticated: %s (user: %s, locale: %s)", client.id, client.userID, client.locale)

	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]any{
		"protocol": protocol.ProtocolVersion,
		"user_id":  client.userID,
		"role":     "user", // Phase 1: All users have "user" role
		"server": map[string]any{
			"name":    "pomclaw",
			"version": "0.1.0",
		},
	}))
}

// handleHealth returns server health status.
func (r *WSMethodRouter) handleHealth(ctx context.Context, client *WSClient, req *protocol.RequestFrame) {
	clientList := r.server.ClientList()
	var clients []map[string]any
	for _, c := range clientList {
		clients = append(clients, map[string]any{
			"id":           c.ID(),
			"user_id":      c.UserID(),
			"connected_at": c.ConnectedAt().Unix(),
			"remote_addr":  c.RemoteAddr(),
		})
	}

	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]any{
		"status":      "ok",
		"protocol":    protocol.ProtocolVersion,
		"version":     "0.1.0",
		"clients":     len(clientList),
		"current_id":  client.ID(),
		"client_list": clients,
	}))
}
