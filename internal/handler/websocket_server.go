package handler

import (
	"github.com/pomclaw/pomclaw/internal/svc"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/pomclaw/pomclaw/pkg/protocol"
)

// WSServer is the main gateway server handling WebSocket connections.
// Adapted from PomClaw's Protocol v3 implementation.
type WSServer struct {
	serverCtx *svc.ServiceContext

	router      *WSMethodRouter
	upgrader    websocket.Upgrader
	rateLimiter *RateLimiter
	clients     map[string]*WSClient
	mu          sync.RWMutex
	startedAt   time.Time
}

// NewWSServer creates a new Protocol v3 WebSocket gateway server.
// Phase 1: Simplified auth, no HTTP handlers, direct integration with Eino agent loop.
func NewWSServer(svc *svc.ServiceContext) *WSServer {
	s := &WSServer{
		serverCtx: svc,
		clients:   make(map[string]*WSClient),
		startedAt: time.Now(),
	}

	s.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     s.checkOrigin,
	}

	// Initialize rate limiter (0 = disabled by default)
	rateLimitRPM := 0 // TODO: Get from config when rate limiting is configured
	s.rateLimiter = NewRateLimiter(rateLimitRPM, 5)

	s.router = NewWSMethodRouter(s)
	return s
}

// checkOrigin validates WebSocket connection origin against allowed origins whitelist.
// If no origins are configured, all origins are allowed (backward compatibility / dev mode).
// Empty Origin header (non-browser clients like CLI/SDK) is always allowed.
func (s *WSServer) checkOrigin(r *http.Request) bool {
	// TODO: Get allowed origins from config when CORS is configured
	return true // Phase 1: accept all origins
}

// Start is a no-op for service.Service interface compatibility.
// WebSocket routes are registered directly with the main REST server.
func (s *WSServer) Start() {
	logx.Info("WebSocket gateway initialized")
}

// Stop is a no-op for service.Service interface compatibility.
// Client cleanup happens automatically when connections close.
func (s *WSServer) Stop() {
	logx.Info("WebSocket gateway stopping...")
	// Close all active client connections
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, client := range s.clients {
		client.Close()
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

// RateLimiter returns the server's rate limiter for use by method handlers.
func (s *WSServer) RateLimiter() *RateLimiter { return s.rateLimiter }

// Router returns the method router for registering additional handlers.
func (s *WSServer) Router() *WSMethodRouter { return s.router }

// StartedAt returns the server start time.
func (s *WSServer) StartedAt() time.Time { return s.startedAt }

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
		if c.UserID() == userID {
			return c
		}
	}
	return nil
}

// FindClientsBySessionKey finds all clients with the specified active session key.
// Used by WSStreamer to route events to correct clients.
func (s *WSServer) FindClientsBySessionKey(sessionKey string) []*WSClient {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var clients []*WSClient
	for _, c := range s.clients {
		if c.activeSessionKey == sessionKey {
			clients = append(clients, c)
		}
	}
	return clients
}

func (s *WSServer) registerClient(c *WSClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[c.id] = c

	// Note: Event subscription is handled by WSStreamer, not here.
	// WSStreamer subscribes to msgBus.OutboundMessages and routes to clients by sessionKey.

	logx.Infof("client connected: %s (remote: %s)", c.id, c.remoteAddr)
}

func (s *WSServer) unregisterClient(c *WSClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, c.id)

	logx.Infof("client disconnected: %s", c.id)
}
