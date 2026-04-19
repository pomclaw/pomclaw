package channels

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

// GatewayChannel 实现 Channel 接口的 Gateway
type GatewayChannel struct {
	*BaseChannel
	clients  sync.Map // map[clientID]*ClientConn
	sessions sync.Map // map[sessionID]*SessionInfo
	server   *http.Server
	port     int
	uiPath   string
	ctx      context.Context
	cancel   context.CancelFunc
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

// NewGatewayChannel 创建Gateway Channel
func NewGatewayChannel(messageBus *bus.MessageBus, port int, uiPath string) *GatewayChannel {
	base := NewBaseChannel("gateway", nil, messageBus, []string{"*"}) // 允许所有来源

	return &GatewayChannel{
		BaseChannel: base,
		port:        port,
		uiPath:      uiPath,
	}
}

// Start 启动Gateway服务器
func (g *GatewayChannel) Start(ctx context.Context) error {
	g.ctx, g.cancel = context.WithCancel(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", g.handleWebSocket)
	mux.HandleFunc("/health", g.handleHealth)
	mux.HandleFunc("/", g.serveUI)

	g.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", g.port),
		Handler: mux,
	}

	logger.InfoCF("gateway", "Starting Gateway Channel", map[string]interface{}{
		"port":    g.port,
		"ui_path": g.uiPath,
		"url":     fmt.Sprintf("http://localhost:%d", g.port),
	})

	go func() {
		if err := g.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.ErrorCF("gateway", "Gateway server error", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	return nil
}

// Stop 停止Gateway服务器
func (g *GatewayChannel) Stop(ctx context.Context) error {
	logger.InfoC("gateway", "Stopping Gateway channel")

	if g.cancel != nil {
		g.cancel()
	}

	if g.server != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := g.server.Shutdown(shutdownCtx); err != nil {
			logger.ErrorCF("gateway", "Error shutting down gateway server", map[string]interface{}{
				"error": err.Error(),
			})
			return err
		}
	}

	// 关闭所有客户端连接
	g.clients.Range(func(key, value interface{}) bool {
		if client, ok := value.(*ClientConn); ok {
			client.Conn.Close()
		}
		return true
	})

	logger.InfoC("gateway", "Gateway channel stopped")
	return nil
}

// Send 实现 Channel 接口：发送消息到客户端
func (g *GatewayChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	logger.InfoCF("gateway", "=== Gateway.Send() called ===", map[string]interface{}{
		"chat_id": msg.ChatID,
		"content": fmt.Sprintf("%.100s", msg.Content),
	})

	// 根据 ChatID 找到对应的客户端
	val, ok := g.clients.Load(msg.ChatID)
	if !ok {
		logger.ErrorCF("gateway", "Client not found for ChatID", map[string]interface{}{
			"chat_id": msg.ChatID,
		})
		return fmt.Errorf("client not found: %s", msg.ChatID)
	}

	client := val.(*ClientConn)
	logger.InfoCF("gateway", "Found client, sending message event", map[string]interface{}{
		"client_id": client.ID,
	})

	// 发送 message 事件给客户端
	g.sendEvent(client, "message", map[string]interface{}{
		"role":    "assistant",
		"content": msg.Content,
	})

	logger.InfoCF("gateway", "=== Gateway.Send() completed ===", map[string]interface{}{
		"client_id": client.ID,
	})

	return nil
}

// serveUI 提供静态UI文件
func (g *GatewayChannel) serveUI(w http.ResponseWriter, r *http.Request) {
	if g.uiPath == "" {
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
</html>`, g.port)
		return
	}

	http.FileServer(http.Dir(g.uiPath)).ServeHTTP(w, r)
}

// handleHealth 健康检查
func (g *GatewayChannel) handleHealth(w http.ResponseWriter, r *http.Request) {
	clientCount := 0
	g.clients.Range(func(key, value interface{}) bool {
		clientCount++
		return true
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "ok",
		"clients":       clientCount,
		"websocket_url": fmt.Sprintf("ws://localhost:%d/ws", g.port),
	})
}

// handleWebSocket 处理WebSocket连接
func (g *GatewayChannel) handleWebSocket(w http.ResponseWriter, r *http.Request) {
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

	g.clients.Store(client.ID, client)
	defer g.clients.Delete(client.ID)

	logger.InfoCF("gateway", "WebSocket connected", map[string]interface{}{
		"client_id": client.ID,
		"user_id":   client.UserID,
	})

	// 发送hello消息
	g.sendHello(client)

	// 启动读写协程
	go g.writePump(client)
	g.readPump(client)
}

// sendHello 发送连接成功消息
func (g *GatewayChannel) sendHello(client *ClientConn) {
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
func (g *GatewayChannel) readPump(client *ClientConn) {
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
			g.sendError(client, "", "invalid_json", "Invalid JSON format")
			continue
		}

		logger.DebugCF("gateway", "Received request", map[string]interface{}{
			"client_id": client.ID,
			"method":    req.Method,
			"req_id":    req.ID,
		})

		// 处理请求
		g.handleRequest(client, &req)
	}
}

// writePump 向WebSocket写入消息
func (g *GatewayChannel) writePump(client *ClientConn) {
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
func (g *GatewayChannel) handleRequest(client *ClientConn, req *RequestFrame) {
	switch req.Method {
	case "chat.send":
		g.handleChatSend(client, req)
	case "sessions.list":
		g.handleSessionsList(client, req)
	case "sessions.get":
		g.handleSessionsGet(client, req)
	case "sessions.create":
		g.handleSessionsCreate(client, req)
	case "sessions.delete":
		g.handleSessionsDelete(client, req)
	default:
		g.sendError(client, req.ID, "method_not_found", "Unknown method: "+req.Method)
	}
}

// sendResponse 发送成功响应
func (g *GatewayChannel) sendResponse(client *ClientConn, reqID string, payload interface{}) {
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
func (g *GatewayChannel) sendError(client *ClientConn, reqID, code, message string) {
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
func (g *GatewayChannel) sendEvent(client *ClientConn, event string, payload interface{}) {
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
		logger.WarnCF("gateway", "Failed to send event, buffer full", map[string]interface{}{
			"client_id": client.ID,
			"event":     event,
		})
	}
}

// generateClientID 生成客户端ID
func generateClientID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
