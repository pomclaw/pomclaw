package handler

import (
	"context"

	"github.com/pomclaw/pomclaw/pkg/protocol"
)

// ClientInterface abstracts a WebSocket client connection.
// This interface allows for easy testing and extension.
type ClientInterface interface {
	// SendResponse sends a response frame to the client
	SendResponse(resp *protocol.ResponseFrame)

	// SendEvent sends an event frame to the client
	SendEvent(event protocol.EventFrame)

	// ID returns the unique client identifier
	ID() string

	// UserID returns the external user ID set during connect
	UserID() string

	// Role returns the client's permission role (for future auth)
	Role() string
}

// MethodHandler processes a single RPC method request.
// ctx contains request context with user info, locale, etc.
// client is the WebSocket client making the request.
// req is the parsed request frame from the client.
type MethodHandler func(ctx context.Context, client ClientInterface, req *protocol.RequestFrame)
