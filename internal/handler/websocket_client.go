package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/pomclaw/pomclaw/pkg/protocol"
)

// WSClient represents a single WebSocket connection.
// Adapted from PomClaw's Protocol v3 implementation.
// Phase 1: Simplified auth (no roles, tenants, pairing).
type WSClient struct {
	id            string
	conn          *websocket.Conn
	server        *WSServer
	authenticated bool
	userID        string // external user ID (set during connect)
	locale        string // user's preferred locale (e.g. "en", "zh", "vi")
	send          chan []byte

	// Session tracking for event routing
	activeSessionKey string // current active session key (set during chat.send)

	connectedAt time.Time
	remoteAddr  string
}

// NewWSClient creates a new WebSocket client.
func NewWSClient(conn *websocket.Conn, server *WSServer, remoteIP string) *WSClient {
	return &WSClient{
		id:          uuid.NewString(),
		conn:        conn,
		server:      server,
		send:        make(chan []byte, 256),
		connectedAt: time.Now(),
		remoteAddr:  remoteIP,
	}
}

// Run starts the read and write pumps for this client.
func (c *WSClient) Run(ctx context.Context) {
	go c.writePump()
	c.readPump(ctx)
}

const maxWSMessageSize = 512 * 1024 // 512KB max message size

// readPump reads frames from the WebSocket connection.
func (c *WSClient) readPump(ctx context.Context) {
	defer c.conn.Close()

	c.conn.SetReadLimit(maxWSMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logx.Error("websocket read error:", map[string]interface{}{
					"client": c.id,
					"error":  err.Error(),
				})
			}
			return
		}

		// Reset read deadline on activity
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		c.handleFrame(ctx, data)
	}
}

// writePump writes frames and pings to the WebSocket connection.
func (c *WSClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleFrame parses and dispatches a single frame.
func (c *WSClient) handleFrame(ctx context.Context, data []byte) {
	frameType, err := protocol.ParseFrameType(data)
	if err != nil {
		c.sendError("", protocol.ErrInvalidRequest, "invalid frame: "+err.Error())
		return
	}

	switch frameType {
	case protocol.FrameTypeRequest:
		var req protocol.RequestFrame
		if err := json.Unmarshal(data, &req); err != nil {
			c.sendError("", protocol.ErrInvalidRequest, "malformed request: "+err.Error())
			return
		}

		// First request must be "connect" (Protocol v3 requirement)
		if !c.authenticated && req.Method != protocol.MethodConnect {
			c.sendError(req.ID, protocol.ErrUnauthorized, "first request must be 'connect'")
			return
		}

		// Dispatch to method router
		c.server.router.Handle(ctx, c, &req)

	default:
		c.sendError("", protocol.ErrInvalidRequest, "unexpected frame type: "+frameType)
	}
}

// SendResponse sends a response frame to this client.
func (c *WSClient) SendResponse(resp *protocol.ResponseFrame) {
	data, err := json.Marshal(resp)
	if err != nil {
		logx.Error("marshal response failed:", map[string]interface{}{"error": err.Error()})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			logx.Debug("client gone, dropping response:", map[string]interface{}{"client": c.id})
		}
	}()
	select {
	case c.send <- data:
	default:
		logx.Info("client send buffer full, dropping message:", map[string]interface{}{"client": c.id})
	}
}

// SendEvent sends an event frame to this client.
func (c *WSClient) SendEvent(event protocol.EventFrame) {
	data, err := json.Marshal(event)
	if err != nil {
		logx.Error("marshal event failed:", map[string]interface{}{"error": err.Error()})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			logx.Debug("client gone, dropping event:", map[string]interface{}{"client": c.id})
		}
	}()
	select {
	case c.send <- data:
	default:
		logx.Info("client send buffer full, dropping event:", map[string]interface{}{"client": c.id})
	}
}

func (c *WSClient) sendError(id, code, message string) {
	c.SendResponse(protocol.NewErrorResponse(id, code, message))
}

// ID returns the client's unique identifier.
func (c *WSClient) ID() string { return c.id }

// UserID returns the external user ID set during connect.
func (c *WSClient) UserID() string { return c.userID }

// Locale returns the user's preferred locale.
func (c *WSClient) Locale() string { return c.locale }

// ConnectedAt returns when the client connected.
func (c *WSClient) ConnectedAt() time.Time { return c.connectedAt }

// RemoteAddr returns the peer IP:port.
func (c *WSClient) RemoteAddr() string { return c.remoteAddr }

// SetActiveSessionKey sets the current active session key for event routing.
func (c *WSClient) SetActiveSessionKey(sessionKey string) {
	c.activeSessionKey = sessionKey
}

// ActiveSessionKey returns the current active session key.
func (c *WSClient) ActiveSessionKey() string { return c.activeSessionKey }

// Close shuts down the client connection.
func (c *WSClient) Close() {
	close(c.send)
}
