package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/config"
	"github.com/pomclaw/pomclaw/pkg/logger"
)

// picoConn represents a single WebSocket connection.
type picoConn struct {
	id        string
	conn      *websocket.Conn
	sessionID string
	writeMu   sync.Mutex
	closed    atomic.Bool
}

// PicoChannel handles Pico Protocol WebSocket connections.
type PicoChannel struct {
	*BaseChannel
	config      config.PicoSettings
	upgrader    websocket.Upgrader
	connections map[string]*picoConn            // connID -> *picoConn
	bySession   map[string]map[string]*picoConn // sessionID -> connID -> *picoConn
	connsMu     sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	server      *http.Server
}

// NewPicoChannel creates a new Pico Protocol channel.
func NewPicoChannel(cfg config.PicoSettings, messageBus *bus.MessageBus) (*PicoChannel, error) {
	base := NewBaseChannel("pico", cfg, messageBus, cfg.AllowFrom)

	ch := &PicoChannel{
		BaseChannel: base,
		config:      cfg,
		upgrader:    websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		connections: make(map[string]*picoConn),
		bySession:   make(map[string]map[string]*picoConn),
	}
	return ch, nil
}

// Start starts the Pico channel and its HTTP server.
func (c *PicoChannel) Start(ctx context.Context) error {
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.running = true

	port := c.config.Port
	if port == 0 {
		port = 18792 // default port
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", c.handleWebSocket)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	c.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	logger.InfoCF("pico", "Pico Protocol channel started", map[string]any{
		"port": port,
		"url":  fmt.Sprintf("ws://localhost:%d/ws", port),
	})

	go func() {
		if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.ErrorCF("pico", "Pico server error", map[string]any{
				"error": err.Error(),
			})
		}
	}()

	return nil
}

// Stop stops the Pico channel and closes all connections.
func (c *PicoChannel) Stop(ctx context.Context) error {
	c.running = false

	c.connsMu.Lock()
	for _, pc := range c.connections {
		pc.close()
	}
	clear(c.connections)
	clear(c.bySession)
	c.connsMu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}

	if c.server != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := c.server.Shutdown(shutdownCtx); err != nil {
			logger.ErrorCF("pico", "Error shutting down server", map[string]any{
				"error": err.Error(),
			})
		}
	}

	logger.InfoC("pico", "Pico Protocol channel stopped")
	return nil
}

// handleWebSocket handles WebSocket upgrades for Pico Protocol.
func (c *PicoChannel) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "only GET requests allowed", http.StatusMethodNotAllowed)
		return
	}

	if !c.running {
		http.Error(w, "channel not running", http.StatusServiceUnavailable)
		return
	}

	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.ErrorCF("pico", "WebSocket upgrade failed", map[string]any{"error": err.Error()})
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	pc := c.addConnection(conn, sessionID)
	if pc == nil {
		conn.Close()
		return
	}

	logger.InfoCF("pico", "WebSocket connected", map[string]any{
		"conn_id":    pc.id,
		"session_id": sessionID,
	})

	go c.readLoop(pc)
}

// addConnection adds a new connection to the registry.
func (c *PicoChannel) addConnection(conn *websocket.Conn, sessionID string) *picoConn {
	c.connsMu.Lock()
	defer c.connsMu.Unlock()

	pc := &picoConn{
		id:        uuid.New().String(),
		conn:      conn,
		sessionID: sessionID,
	}

	c.connections[pc.id] = pc
	if _, ok := c.bySession[sessionID]; !ok {
		c.bySession[sessionID] = make(map[string]*picoConn)
	}
	c.bySession[sessionID][pc.id] = pc

	return pc
}

// removeConnection removes a connection from the registry.
func (c *PicoChannel) removeConnection(connID string) {
	c.connsMu.Lock()
	defer c.connsMu.Unlock()

	pc, ok := c.connections[connID]
	if !ok {
		return
	}

	delete(c.connections, connID)
	if bySession, ok := c.bySession[pc.sessionID]; ok {
		delete(bySession, connID)
		if len(bySession) == 0 {
			delete(c.bySession, pc.sessionID)
		}
	}
}

// getSessionConnections returns all connections for a session.
func (c *PicoChannel) getSessionConnections(sessionID string) []*picoConn {
	c.connsMu.RLock()
	defer c.connsMu.RUnlock()

	bySession, ok := c.bySession[sessionID]
	if !ok {
		return nil
	}

	conns := make([]*picoConn, 0, len(bySession))
	for _, pc := range bySession {
		conns = append(conns, pc)
	}
	return conns
}

// Send sends a message to all connections in a session.
func (c *PicoChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	if !c.running {
		return fmt.Errorf("channel not running")
	}

	sessionID := strings.TrimPrefix(msg.ChatID, "pico:")
	outMsg := PicoMessage{
		Type:      TypeMessageCreate,
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Timestamp: time.Now().UnixMilli(),
		Payload: map[string]any{
			"content": msg.Content,
		},
	}

	conns := c.getSessionConnections(sessionID)
	if len(conns) == 0 {
		return fmt.Errorf("no active connections for session %s", sessionID)
	}

	for _, pc := range conns {
		if err := pc.writeJSON(outMsg); err != nil {
			logger.DebugCF("pico", "Write failed", map[string]any{"error": err.Error()})
		}
	}

	return nil
}

// writeJSON helper for picoConn.
func (pc *picoConn) writeJSON(v any) error {
	if pc.closed.Load() {
		return fmt.Errorf("connection closed")
	}
	pc.writeMu.Lock()
	defer pc.writeMu.Unlock()
	return pc.conn.WriteJSON(v)
}

// close closes the connection.
func (pc *picoConn) close() {
	if pc.closed.CompareAndSwap(false, true) {
		pc.conn.Close()
	}
}

// readLoop reads messages from a WebSocket connection.
func (c *PicoChannel) readLoop(pc *picoConn) {
	defer func() {
		pc.close()
		c.removeConnection(pc.id)
		logger.InfoCF("pico", "WebSocket disconnected", map[string]any{"conn_id": pc.id})
	}()

	pc.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	pc.conn.SetPongHandler(func(string) error {
		pc.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		_, rawMsg, err := pc.conn.ReadMessage()
		if err != nil {
			return
		}

		pc.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		var msg PicoMessage
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			pc.writeJSON(newError("invalid_message", "failed to parse message"))
			continue
		}

		c.handleMessage(pc, msg)
	}
}

// handleMessage processes an inbound message.
func (c *PicoChannel) handleMessage(pc *picoConn, msg PicoMessage) {
	switch msg.Type {
	case TypePing:
		pong := PicoMessage{
			Type:      TypePong,
			ID:        msg.ID,
			Timestamp: time.Now().UnixMilli(),
		}
		_ = pc.writeJSON(pong)

	case TypeMessageSend:
		c.handleMessageSend(pc, msg)
	}
}

// handleMessageSend processes an inbound message.send.
func (c *PicoChannel) handleMessageSend(pc *picoConn, msg PicoMessage) {
	content, ok := msg.Payload["content"].(string)
	if !ok || strings.TrimSpace(content) == "" {
		pc.writeJSON(newError("empty_content", "message content is empty"))
		return
	}

	sessionID := msg.SessionID
	if sessionID == "" {
		sessionID = pc.sessionID
	}

	// Use BaseChannel's HandleMessage to process inbound messages
	// This ensures consistent allow-list checking and sessionKey generation
	c.HandleMessage(pc.id, sessionID, content, []string{}, map[string]string{})
}
