package handler

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/pomclaw/pomclaw/pkg/agent"
	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/contracts"
	"github.com/pomclaw/pomclaw/pkg/protocol"
)

// WSServer is the main gateway server handling WebSocket connections.
type WSServer struct {
	cfg      *config.Config
	agents   *agent.AgentLoop
	sessions contracts.SessionManagerInterface
	msgBus   *bus.MessageBus

	upgrader   websocket.Upgrader
	clients    map[string]*WSClient
	mu         sync.RWMutex
	router     *WSMethodRouter
	httpServer *http.Server
}

// NewWSServer creates a new gateway server.
func NewWSServer(cfg *config.Config, agentLoop *agent.AgentLoop, sessions contracts.SessionManagerInterface, msgBus *bus.MessageBus) *WSServer {
	s := &WSServer{
		cfg:      cfg,
		agents:   agentLoop,
		sessions: sessions,
		msgBus:   msgBus,
		clients:  make(map[string]*WSClient),
	}

	s.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     s.checkOrigin,
	}

	s.router = NewWSMethodRouter(s)
	return s
}

// checkOrigin validates WebSocket connection origin.
// For Phase 1, accept all origins.
func (s *WSServer) checkOrigin(r *http.Request) bool {
	return true
}

// Start begins listening for WebSocket connections.
// Implements service.Service interface.
func (s *WSServer) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHTTPHealth)

	addr := fmt.Sprintf("%s:%d", s.cfg.Gateway.Host, s.cfg.Gateway.Port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	logx.Info("gateway starting:", map[string]interface{}{"addr": addr})

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logx.Error("gateway server error:", map[string]interface{}{"error": err.Error()})
		}
	}()
}

// Stop stops the gateway server.
// Implements service.Service interface.
func (s *WSServer) Stop() {
	if s.httpServer != nil {
		logx.Info("gateway stopping...")
		s.httpServer.Close()
	}
}

// handleWebSocket upgrades HTTP to WebSocket and manages the connection.
func (s *WSServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logx.Error("websocket upgrade failed:", map[string]interface{}{"error": err.Error()})
		return
	}

	client := NewWSClient(conn, s, clientIP(r))
	s.registerClient(client)

	defer func() {
		s.unregisterClient(client)
		client.Close()
	}()

	client.Run(r.Context())
}

// handleHTTPHealth returns a simple HTTP health check response.
func (s *WSServer) handleHTTPHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","protocol":%d}`, protocol.ProtocolVersion)
}

// clientIP extracts the real client IP from the request.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		if i := strings.IndexByte(fwd, ','); i > 0 {
			return strings.TrimSpace(fwd[:i])
		}
		return strings.TrimSpace(fwd)
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

// Router returns the method router for registering additional handlers.
func (s *WSServer) Router() *WSMethodRouter { return s.router }

// ClientList returns a snapshot of all connected clients.
func (s *WSServer) ClientList() []*WSClient {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*WSClient, 0, len(s.clients))
	for _, c := range s.clients {
		list = append(list, c)
	}
	return list
}

// BroadcastEvent sends an event to all connected clients.
func (s *WSServer) BroadcastEvent(event protocol.EventFrame) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, client := range s.clients {
		client.SendEvent(event)
	}
}

// FindClientByUserID finds a client by user ID.
func (s *WSServer) FindClientByUserID(userID string) *WSClient {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.clients {
		if c.userID == userID {
			return c
		}
	}
	return nil
}

func (s *WSServer) registerClient(c *WSClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[c.id] = c

	logx.Info("client connected:", map[string]interface{}{"id": c.id})
}

func (s *WSServer) unregisterClient(c *WSClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, c.id)
	logx.Info("client disconnected:", map[string]interface{}{"id": c.id})
}
