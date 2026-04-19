package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/logger"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: 生产环境需要严格验证Origin
		return true
	},
}

// Server Gateway服务器，实现OpenClaw协议
type Server struct {
	bus      *bus.MessageBus
	clients  sync.Map // map[clientID]*ClientConn
	sessions sync.Map // map[sessionID]*SessionInfo
	mux      *http.ServeMux
	port     int
	uiPath   string
}

// ClientConn WebSocket客户端连接
type ClientConn struct {
	ID     string
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
	SeqNum int
	mu     sync.Mutex
}

// NewServer 创建Gateway服务器
func NewServer(bus *bus.MessageBus, port int, uiPath string) *Server {
	return &Server{
		bus:    bus,
		mux:    http.NewServeMux(),
		port:   port,
		uiPath: uiPath,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	// 注册路由
	s.mux.HandleFunc("/ws", s.handleWebSocket)
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/", s.serveUI)

	logger.InfoC("gateway", "Gateway.Start() called, about to start outbound handler")

	// 启动outbound消息监听
	go s.handleOutboundMessages()

	addr := fmt.Sprintf(":%d", s.port)
	logger.InfoCF("gateway", "Starting Gateway HTTP server", map[string]interface{}{
		"port":    s.port,
		"ui_path": s.uiPath,
		"url":     fmt.Sprintf("http://localhost:%d", s.port),
	})

	return http.ListenAndServe(addr, s.mux)
}

// serveUI 提供静态UI文件
func (s *Server) serveUI(w http.ResponseWriter, r *http.Request) {
	// 如果ui_path为空，返回简单的欢迎页
	if s.uiPath == "" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>PomClaw Gateway</title>
</head>
<body>
    <h1>PomClaw Gateway</h1>
    <p>WebSocket endpoint: ws://localhost:%d/ws</p>
    <p>Health check: <a href="/health">/health</a></p>
    <p>UI not built yet. Run: cd ui && npm install && npm run build</p>
</body>
</html>`, s.port)
		return
	}

	// 提供静态文件
	http.FileServer(http.Dir(s.uiPath)).ServeHTTP(w, r)
}

// handleHealth 健康检查
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	clientCount := 0
	s.clients.Range(func(key, value interface{}) bool {
		clientCount++
		return true
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "ok",
		"clients":      clientCount,
		"websocket_url": fmt.Sprintf("ws://localhost:%d/ws", s.port),
	})
}

// handleWebSocket 处理WebSocket连接
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.ErrorCF("gateway", "WebSocket upgrade failed", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	client := &ClientConn{
		ID:     generateClientID(),
		UserID: "user_" + generateClientID(), // TODO: 从认证获取真实userID
		Conn:   conn,
		Send:   make(chan []byte, 256),
		SeqNum: 0,
	}

	s.clients.Store(client.ID, client)
	defer s.clients.Delete(client.ID)

	logger.InfoCF("gateway", "WebSocket connected", map[string]interface{}{
		"client_id": client.ID,
		"user_id":   client.UserID,
	})

	// 发送hello消息
	s.sendHello(client)

	// 启动读写协程
	go s.writePump(client)
	s.readPump(client)
}

// sendHello 发送连接成功消息
func (s *Server) sendHello(client *ClientConn) {
	hello := HelloOkFrame{
		Type:     "hello-ok",
		Protocol: 1,
		Server: &ServerInfo{
			Version: "pomclaw-1.0.0",
			ConnID:  client.ID,
		},
		Features: &Features{
			Methods: []string{
				"chat.send",
				"sessions.list",
				"sessions.get",
				"sessions.create",
				"sessions.delete",
			},
			Events: []string{
				"message",
				"session.created",
				"session.updated",
			},
		},
		Auth: &AuthInfo{
			Role:       "user",
			IssuedAtMs: time.Now().UnixMilli(),
		},
	}

	data, _ := json.Marshal(hello)
	select {
	case client.Send <- data:
	default:
		logger.WarnCF("gateway", "Failed to send hello", map[string]interface{}{
			"client_id": client.ID,
		})
	}
}

// readPump 从WebSocket读取消息
func (s *Server) readPump(client *ClientConn) {
	defer client.Conn.Close()

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.ErrorCF("gateway", "WebSocket error", map[string]interface{}{
					"error":     err.Error(),
					"client_id": client.ID,
				})
			}
			break
		}

		// 解析消息
		var req RequestFrame
		if err := json.Unmarshal(message, &req); err != nil {
			logger.WarnCF("gateway", "Invalid message format", map[string]interface{}{
				"error":     err.Error(),
				"client_id": client.ID,
			})
			s.sendError(client, "", "invalid_json", "Invalid JSON format")
			continue
		}

		logger.DebugCF("gateway", "Received request", map[string]interface{}{
			"client_id": client.ID,
			"method":    req.Method,
			"req_id":    req.ID,
		})

		// 处理请求
		s.handleRequest(client, &req)
	}
}

// writePump 向WebSocket写入消息
func (s *Server) writePump(client *ClientConn) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.WarnCF("gateway", "Write error", map[string]interface{}{
					"error":     err.Error(),
					"client_id": client.ID,
				})
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleRequest 处理客户端请求
func (s *Server) handleRequest(client *ClientConn, req *RequestFrame) {
	switch req.Method {
	case "chat.send":
		s.handleChatSend(client, req)
	case "sessions.list":
		s.handleSessionsList(client, req)
	case "sessions.get":
		s.handleSessionsGet(client, req)
	case "sessions.create":
		s.handleSessionsCreate(client, req)
	case "sessions.delete":
		s.handleSessionsDelete(client, req)
	default:
		s.sendError(client, req.ID, "method_not_found", "Unknown method: "+req.Method)
	}
}

// sendResponse 发送成功响应
func (s *Server) sendResponse(client *ClientConn, reqID string, payload interface{}) {
	resp := ResponseFrame{
		Type:    "res",
		ID:      reqID,
		OK:      true,
		Payload: payload,
	}

	data, _ := json.Marshal(resp)
	select {
	case client.Send <- data:
	default:
		logger.WarnCF("gateway", "Send buffer full", map[string]interface{}{
			"client_id": client.ID,
		})
	}
}

// sendError 发送错误响应
func (s *Server) sendError(client *ClientConn, reqID, code, message string) {
	resp := ResponseFrame{
		Type: "res",
		ID:   reqID,
		OK:   false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}

	data, _ := json.Marshal(resp)
	select {
	case client.Send <- data:
	default:
	}
}

// sendEvent 发送事件
func (s *Server) sendEvent(client *ClientConn, event string, payload interface{}) {
	client.mu.Lock()
	client.SeqNum++
	seq := client.SeqNum
	client.mu.Unlock()

	evt := EventFrame{
		Type:    "event",
		Event:   event,
		Payload: payload,
		Seq:     seq,
	}

	data, _ := json.Marshal(evt)
	select {
	case client.Send <- data:
	default:
	}
}

// handleOutboundMessages 处理从MessageBus来的响应消息
func (s *Server) handleOutboundMessages() {
	ctx := context.Background()
	logger.InfoC("gateway", "=== Starting outbound message handler ===")

	// 测试日志：确保这个goroutine在运行
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			clientCount := 0
			s.clients.Range(func(key, value interface{}) bool {
				clientCount++
				return true
			})
			logger.InfoCF("gateway", "Heartbeat: handler alive", map[string]interface{}{
				"clients": clientCount,
			})
		}
	}()

	for {
		logger.DebugC("gateway", "Waiting for outbound message...")
		msg, ok := s.bus.SubscribeOutbound(ctx)
		if !ok {
			logger.InfoC("gateway", "Outbound subscription closed")
			return
		}

		logger.InfoCF("gateway", "!!! Received ANY outbound message from bus !!!", map[string]interface{}{
			"channel": msg.Channel,
			"chat_id": msg.ChatID,
		})

		// 只处理gateway channel的消息
		if msg.Channel != "gateway" {
			logger.InfoCF("gateway", "Skipping non-gateway message", map[string]interface{}{
				"channel": msg.Channel,
			})
			continue
		}

		contentPreview := msg.Content
		if len(contentPreview) > 100 {
			contentPreview = contentPreview[:100] + "..."
		}
		logger.InfoCF("gateway", ">>> Processing gateway outbound message <<<", map[string]interface{}{
			"chat_id": msg.ChatID,
			"content": contentPreview,
		})

		// 根据ChatID找到对应的客户端
		val, ok := s.clients.Load(msg.ChatID)
		if !ok {
			logger.WarnCF("gateway", "Client not found for outbound message", map[string]interface{}{
				"chat_id": msg.ChatID,
			})
			continue
		}

		client := val.(*ClientConn)

		// 发送message事件给客户端
		s.sendEvent(client, "message", map[string]interface{}{
			"role":    "assistant",
			"content": msg.Content,
		})
	}
}

// generateClientID 生成客户端ID
func generateClientID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
