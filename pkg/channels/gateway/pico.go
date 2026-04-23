package gateway

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/channels/base"
	"github.com/pomclaw/pomclaw/pkg/config"
	"github.com/pomclaw/pomclaw/pkg/logger"
	postgresdb "github.com/pomclaw/pomclaw/pkg/postgres"
	pomclaui "github.com/pomclaw/pomclaw/ui"
)

// picoConn represents a single WebSocket connection.
type picoConn struct {
	id        string
	conn      *websocket.Conn
	sessionID string
	agentID   string
	writeMu   sync.Mutex
	closed    atomic.Bool
}

// PicoChannel handles Gateway Protocol WebSocket connections.
type PicoChannel struct {
	*base.BaseChannel
	config   config.GatewayConfig
	pgConfig config.PostgresDBConfig // may be nil; owned by caller (config.Config)

	db          *sql.DB // created in Start(), closed in Stop()
	upgrader    websocket.Upgrader
	connections map[string]*picoConn            // connID -> *picoConn
	bySession   map[string]map[string]*picoConn // sessionID -> connID -> *picoConn
	connsMu     sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	server      *http.Server
}

// NewPicoChannel creates a new Gateway Protocol channel.
// pgConfig may be nil; when nil the REST API routes are skipped.
func NewPicoChannel(cfg config.GatewayConfig, pgConfig config.PostgresDBConfig, messageBus *bus.MessageBus) (*PicoChannel, error) {
	baseChannel := base.NewBaseChannel("gateway", cfg, messageBus, cfg.AllowFrom)

	ch := &PicoChannel{
		BaseChannel: baseChannel,
		config:      cfg,
		pgConfig:    pgConfig,
		upgrader:    websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		connections: make(map[string]*picoConn),
		bySession:   make(map[string]map[string]*picoConn),
	}
	return ch, nil
}

// Start starts the Gateway channel and its HTTP server.
func (c *PicoChannel) Start(ctx context.Context) error {
	if c.config.JWTSecret == "" {
		return fmt.Errorf("gateway: jwt_secret is required but not configured")
	}

	c.ctx, c.cancel = context.WithCancel(ctx)
	c.SetRunning(true)

	port := c.config.Port
	if port == 0 {
		port = 18792 // default port
	}

	// Initialize DB connection from Postgres config.
	if c.pgConfig.Enabled {
		pgConn, err := postgresdb.NewConnectionManager(&c.pgConfig)
		if err != nil {
			return fmt.Errorf("gateway: failed to connect to postgres: %w", err)
		}
		c.db = pgConn.DB()

	} else {
		logger.WarnC("pico", "no database configured – REST API routes will not be available")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", c.handleWebSocket)
	setupAPIRoutes(mux, c.db, c.config.JWTSecret)
	logger.InfoC("pico", "REST API routes registered")

	// Mount the embedded UI – must be last so API/WS routes take precedence.
	distFS, err := fs.Sub(pomclaui.DistFS, "dist")
	if err == nil {
		mux.Handle("/", spaHandler(distFS))
		logger.InfoC("pico", "UI static files mounted at /")
	} else {
		logger.WarnCF("pico", "failed to mount UI static files", map[string]any{"error": err.Error()})
	}

	c.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: corsMiddleware(mux),
	}

	logger.InfoCF("pico", "Gateway Protocol channel started", map[string]any{
		"port": port,
		"url":  fmt.Sprintf("ws://localhost:%d/ws", port),
	})

	go func() {
		if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.ErrorCF("pico", "Gateway server error", map[string]any{
				"error": err.Error(),
			})
		}
	}()

	return nil
}

// Stop stops the Gateway channel and closes all connections.
func (c *PicoChannel) Stop(ctx context.Context) error {
	c.SetRunning(false)

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

	if c.db != nil {
		if err := c.db.Close(); err != nil {
			logger.ErrorCF("pico", "Error closing database connection", map[string]any{
				"error": err.Error(),
			})
		}
		c.db = nil
	}

	logger.InfoC("pico", "Gateway Protocol channel stopped")
	return nil
}

// handleWebSocket handles WebSocket upgrades for Gateway Protocol.
func (c *PicoChannel) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "only GET requests allowed", http.StatusMethodNotAllowed)
		return
	}

	if !c.IsRunning() {
		http.Error(w, "channel not running", http.StatusServiceUnavailable)
		return
	}

	// Validate parameters before upgrade
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		http.Error(w, "agent_id query parameter is required", http.StatusBadRequest)
		logger.WarnC("pico", "WebSocket connection rejected: missing agent_id")
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// Upgrade connection after validation
	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.ErrorCF("pico", "WebSocket upgrade failed", map[string]any{"error": err.Error()})
		return
	}

	pc := c.addConnection(conn, sessionID, agentID)
	if pc == nil {
		conn.Close()
		return
	}

	logger.InfoCF("pico", "WebSocket connected", map[string]any{
		"conn_id":    pc.id,
		"session_id": sessionID,
		"agent_id":   agentID,
	})

	go c.readLoop(pc)
}

// addConnection adds a new connection to the registry.
func (c *PicoChannel) addConnection(conn *websocket.Conn, sessionID string, agentID string) *picoConn {
	c.connsMu.Lock()
	defer c.connsMu.Unlock()

	pc := &picoConn{
		id:        uuid.New().String(),
		conn:      conn,
		sessionID: sessionID,
		agentID:   agentID,
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
	if !c.IsRunning() {
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

	// Extract media from payload if present
	var media []string
	if mediaList, ok := msg.Payload["media"].([]interface{}); ok {
		for _, m := range mediaList {
			if s, ok := m.(string); ok {
				media = append(media, s)
			}
		}
	}

	// Pass agentID through metadata for multi-tenant isolation
	metadata := make(map[string]string)
	if pc.agentID != "" {
		metadata["agent_id"] = pc.agentID
	}

	// Use BaseChannel's HandleMessage to process inbound messages
	// This ensures consistent allow-list checking and sessionKey generation
	c.HandleMessage(pc.id, sessionID, content, media, metadata)
}
