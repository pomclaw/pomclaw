# HTTP Server Channel 设计方案

## 核心需求

**创建一个原生的HTTP/WebSocket服务端Channel，作为多租户Agent平台的前端接入层**

```
前端 (React)
    ↓ HTTP/WebSocket
HTTPServerChannel (端口18790)
    ↓ MessageBus
AgentLoop (多个Agent)
    ↓
PostgreSQL
```

---

## 1. 设计思路

### 1.1 与现有Channels的区别

| Channel类型 | 模式 | 依赖 | 适用场景 |
|------------|------|------|---------|
| **Slack/Discord** | 客户端 | 需要外部平台token | 接入现有平台 |
| **HTTPServerChannel** | **服务端** | **无外部依赖** | **原生Web应用** |

### 1.2 HTTPServerChannel特点

✅ **自带HTTP服务器** - 监听端口，接收连接  
✅ **WebSocket支持** - 实时双向通信  
✅ **RESTful API** - 标准HTTP接口  
✅ **Session管理** - 用户session生命周期管理  
✅ **用户认证** - JWT token验证  
✅ **多租户隔离** - 每个用户独立的Agent和Session

---

## 2. 架构设计

### 2.1 HTTPServerChannel结构

```go
// pkg/channels/http_server.go
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
    "github.com/pomclaw/pomclaw/pkg/config"
    "github.com/pomclaw/pomclaw/pkg/logger"
)

type HTTPServerChannel struct {
    *BaseChannel
    
    // HTTP服务器
    server   *http.Server
    port     int
    
    // WebSocket连接管理
    clients  sync.Map // map[userID]*WSClient
    upgrader websocket.Upgrader
    
    // Session管理
    sessions sync.Map // map[sessionID]*SessionInfo
    
    // 认证
    jwtSecret string
}

type WSClient struct {
    UserID    string
    Conn      *websocket.Conn
    Send      chan []byte
    mu        sync.Mutex
}

type SessionInfo struct {
    SessionID string
    AgentID   string
    UserID    string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// 前端发送的消息格式
type ClientMessage struct {
    Type      string `json:"type"`       // "chat", "ping"
    SessionID string `json:"session_id"` // 可选，为空则创建新session
    Message   string `json:"message"`    // 消息内容
}

// 后端返回的消息格式
type ServerMessage struct {
    Type      string `json:"type"`       // "message", "error", "session_created"
    SessionID string `json:"session_id"`
    Content   string `json:"content"`
    Timestamp int64  `json:"timestamp"`
}

func NewHTTPServerChannel(cfg config.HTTPServerConfig, messageBus *bus.MessageBus) (*HTTPServerChannel, error) {
    base := NewBaseChannel("http_server", cfg, messageBus, []string{"*"})
    
    return &HTTPServerChannel{
        BaseChannel: base,
        port:        cfg.Port,
        jwtSecret:   cfg.JWTSecret,
        upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                // TODO: 生产环境验证CORS
                return true
            },
        },
    }, nil
}

// Start 启动HTTP服务器
func (c *HTTPServerChannel) Start(ctx context.Context) error {
    logger.InfoCF("http_server", "Starting HTTP Server Channel", map[string]interface{}{
        "port": c.port,
    })
    
    // 创建路由
    mux := http.NewServeMux()
    
    // WebSocket端点
    mux.HandleFunc("/ws", c.handleWebSocket)
    
    // HTTP API
    mux.HandleFunc("/api/health", c.handleHealth)
    mux.HandleFunc("/api/sessions", c.handleSessions)
    
    // 创建HTTP服务器
    c.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", c.port),
        Handler: c.corsMiddleware(mux),
    }
    
    // 启动服务器
    go func() {
        if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.ErrorCF("http_server", "Server error", map[string]interface{}{
                "error": err.Error(),
            })
        }
    }()
    
    c.setRunning(true)
    logger.InfoCF("http_server", "HTTP Server Channel started", map[string]interface{}{
        "url": fmt.Sprintf("http://localhost:%d", c.port),
    })
    
    return nil
}

// Stop 停止服务器
func (c *HTTPServerChannel) Stop(ctx context.Context) error {
    logger.InfoC("http_server", "Stopping HTTP Server Channel")
    
    if c.server != nil {
        if err := c.server.Shutdown(ctx); err != nil {
            return err
        }
    }
    
    c.setRunning(false)
    logger.InfoC("http_server", "HTTP Server Channel stopped")
    return nil
}

// handleWebSocket 处理WebSocket连接
func (c *HTTPServerChannel) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    // 1. 验证token（可选，前期可以先不验证）
    token := r.URL.Query().Get("token")
    userID := c.validateToken(token)
    if userID == "" {
        // 如果没有token，生成临时userID
        userID = fmt.Sprintf("guest_%d", time.Now().UnixNano())
    }
    
    // 2. 升级为WebSocket
    conn, err := c.upgrader.Upgrade(w, r, nil)
    if err != nil {
        logger.ErrorCF("http_server", "WebSocket upgrade failed", map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    // 3. 创建客户端
    client := &WSClient{
        UserID: userID,
        Conn:   conn,
        Send:   make(chan []byte, 256),
    }
    
    c.clients.Store(userID, client)
    
    logger.InfoCF("http_server", "WebSocket connected", map[string]interface{}{
        "user_id": userID,
    })
    
    // 4. 发送欢迎消息
    c.sendToClient(client, ServerMessage{
        Type:      "connected",
        Content:   "Welcome to PomClaw!",
        Timestamp: time.Now().Unix(),
    })
    
    // 5. 启动读写协程
    go c.writePump(client)
    c.readPump(client)
    
    // 6. 清理
    c.clients.Delete(userID)
    logger.InfoCF("http_server", "WebSocket disconnected", map[string]interface{}{
        "user_id": userID,
    })
}

// readPump 从WebSocket读取消息
func (c *HTTPServerChannel) readPump(client *WSClient) {
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
                logger.ErrorCF("http_server", "WebSocket error", map[string]interface{}{
                    "error": err.Error(),
                })
            }
            break
        }
        
        // 解析消息
        var msg ClientMessage
        if err := json.Unmarshal(message, &msg); err != nil {
            logger.WarnCF("http_server", "Invalid message format", map[string]interface{}{
                "error": err.Error(),
            })
            continue
        }
        
        // 处理消息
        c.handleClientMessage(client, &msg)
    }
}

// writePump 向WebSocket写入消息
func (c *HTTPServerChannel) writePump(client *WSClient) {
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

// handleClientMessage 处理客户端消息
func (c *HTTPServerChannel) handleClientMessage(client *WSClient, msg *ClientMessage) {
    ctx := context.Background()
    
    switch msg.Type {
    case "ping":
        c.sendToClient(client, ServerMessage{
            Type:      "pong",
            Timestamp: time.Now().Unix(),
        })
        return
        
    case "chat":
        // 处理聊天消息
        c.handleChatMessage(ctx, client, msg)
    }
}

// handleChatMessage 处理聊天消息
func (c *HTTPServerChannel) handleChatMessage(ctx context.Context, client *WSClient, msg *ClientMessage) {
    // 1. 确定sessionID
    sessionID := msg.SessionID
    if sessionID == "" {
        // 创建新session
        sessionID = c.createSession(client.UserID)
        c.sendToClient(client, ServerMessage{
            Type:      "session_created",
            SessionID: sessionID,
            Timestamp: time.Now().Unix(),
        })
    }
    
    // 2. 获取或创建session信息
    agentID := c.getOrCreateSession(sessionID, client.UserID)
    
    // 3. 构造InboundMessage
    inboundMsg := bus.InboundMessage{
        Channel:    "http_server",
        SenderID:   client.UserID,
        ChatID:     sessionID,
        Content:    msg.Message,
        SessionKey: sessionID,
        AgentID:    agentID,
    }
    
    logger.DebugCF("http_server", "Processing message", map[string]interface{}{
        "user_id":    client.UserID,
        "session_id": sessionID,
        "agent_id":   agentID,
        "message":    msg.Message,
    })
    
    // 4. 发送到MessageBus
    go func() {
        response, err := c.bus.SendAndWait(ctx, inboundMsg, 30*time.Second)
        if err != nil {
            logger.ErrorCF("http_server", "Failed to process message", map[string]interface{}{
                "error": err.Error(),
            })
            c.sendToClient(client, ServerMessage{
                Type:      "error",
                Content:   fmt.Sprintf("Error: %v", err),
                SessionID: sessionID,
                Timestamp: time.Now().Unix(),
            })
            return
        }
        
        // 5. 返回响应
        c.sendToClient(client, ServerMessage{
            Type:      "message",
            SessionID: sessionID,
            Content:   response,
            Timestamp: time.Now().Unix(),
        })
    }()
}

// sendToClient 向客户端发送消息
func (c *HTTPServerChannel) sendToClient(client *WSClient, msg ServerMessage) {
    data, err := json.Marshal(msg)
    if err != nil {
        logger.ErrorCF("http_server", "Failed to marshal message", map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    select {
    case client.Send <- data:
    default:
        logger.WarnCF("http_server", "Client send buffer full", map[string]interface{}{
            "user_id": client.UserID,
        })
        close(client.Send)
        c.clients.Delete(client.UserID)
    }
}

// getOrCreateSession 获取或创建session
func (c *HTTPServerChannel) getOrCreateSession(sessionID, userID string) string {
    // 从内存查询
    if val, ok := c.sessions.Load(sessionID); ok {
        session := val.(*SessionInfo)
        session.UpdatedAt = time.Now()
        return session.AgentID
    }
    
    // 创建新session，默认使用default agent
    agentID := "default"
    session := &SessionInfo{
        SessionID: sessionID,
        AgentID:   agentID,
        UserID:    userID,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    c.sessions.Store(sessionID, session)
    
    logger.InfoCF("http_server", "Session created", map[string]interface{}{
        "session_id": sessionID,
        "user_id":    userID,
        "agent_id":   agentID,
    })
    
    return agentID
}

// createSession 创建新session ID
func (c *HTTPServerChannel) createSession(userID string) string {
    return fmt.Sprintf("session_%s_%d", userID, time.Now().UnixNano())
}

// validateToken 验证JWT token
func (c *HTTPServerChannel) validateToken(token string) string {
    // TODO: 实现JWT验证
    // 临时实现：直接返回token作为userID
    if token == "" {
        return ""
    }
    return token
}

// handleHealth 健康检查
func (c *HTTPServerChannel) handleHealth(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":  "ok",
        "channel": "http_server",
        "clients": c.countClients(),
    })
}

// handleSessions 查询sessions
func (c *HTTPServerChannel) handleSessions(w http.ResponseWriter, r *http.Request) {
    // TODO: 实现session列表查询
    sessions := []map[string]interface{}{}
    
    c.sessions.Range(func(key, value interface{}) bool {
        session := value.(*SessionInfo)
        sessions = append(sessions, map[string]interface{}{
            "session_id": session.SessionID,
            "user_id":    session.UserID,
            "agent_id":   session.AgentID,
            "created_at": session.CreatedAt,
        })
        return true
    })
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "sessions": sessions,
    })
}

// countClients 统计连接数
func (c *HTTPServerChannel) countClients() int {
    count := 0
    c.clients.Range(func(key, value interface{}) bool {
        count++
        return true
    })
    return count
}

// corsMiddleware CORS中间件
func (c *HTTPServerChannel) corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// Send 实现BaseChannel接口
func (c *HTTPServerChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
    // 根据ChatID找到对应的client
    if client, ok := c.clients.Load(msg.ChatID); ok {
        c.sendToClient(client.(*WSClient), ServerMessage{
            Type:      "message",
            Content:   msg.Content,
            Timestamp: time.Now().Unix(),
        })
        return nil
    }
    
    return fmt.Errorf("client not found: %s", msg.ChatID)
}
```

---

## 3. 配置文件

```yaml
# config/config.yaml
channels:
  http_server:
    enabled: true
    port: 18790
    jwt_secret: "your-secret-key-here"
    cors_origins:
      - "http://localhost:3000"
      - "https://app.example.com"
    max_connections: 1000
```

---

## 4. 前端实现

### 4.1 WebSocket Hook

```typescript
// src/hooks/useWebSocket.ts
import { useState, useEffect, useCallback, useRef } from 'react';

interface Message {
  role: 'user' | 'assistant';
  content: string;
  timestamp: number;
}

interface ServerMessage {
  type: string;
  session_id?: string;
  content?: string;
  timestamp: number;
}

export function useWebSocket(url: string, token?: string) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [sessionId, setSessionId] = useState<string>('');
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    // 构造WebSocket URL
    const wsUrl = token ? `${url}?token=${token}` : url;
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('WebSocket connected');
      setIsConnected(true);
    };

    ws.onmessage = (event) => {
      const data: ServerMessage = JSON.parse(event.data);
      
      switch (data.type) {
        case 'connected':
          console.log('Server:', data.content);
          break;
          
        case 'session_created':
          console.log('New session:', data.session_id);
          setSessionId(data.session_id!);
          break;
          
        case 'message':
          // 添加助手消息
          setMessages((prev) => [
            ...prev,
            {
              role: 'assistant',
              content: data.content!,
              timestamp: data.timestamp,
            },
          ]);
          break;
          
        case 'error':
          console.error('Error:', data.content);
          break;
      }
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
      setIsConnected(false);
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    return () => {
      ws.close();
    };
  }, [url, token]);

  const sendMessage = useCallback((content: string) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      console.error('WebSocket not connected');
      return;
    }

    // 添加用户消息
    setMessages((prev) => [
      ...prev,
      {
        role: 'user',
        content,
        timestamp: Date.now(),
      },
    ]);

    // 发送到服务器
    wsRef.current.send(JSON.stringify({
      type: 'chat',
      session_id: sessionId,
      message: content,
    }));
  }, [sessionId]);

  return {
    messages,
    isConnected,
    sessionId,
    sendMessage,
  };
}
```

---

### 4.2 Chat组件

```typescript
// src/pages/Chat.tsx
import React, { useState } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';

const WEBSOCKET_URL = 'ws://localhost:18790/ws';

export default function Chat() {
  const [input, setInput] = useState('');
  const token = localStorage.getItem('auth_token') || ''; // 可选
  
  const { messages, isConnected, sessionId, sendMessage } = useWebSocket(
    WEBSOCKET_URL,
    token
  );

  const handleSend = () => {
    if (!input.trim() || !isConnected) return;
    sendMessage(input);
    setInput('');
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="h-screen flex flex-col bg-gray-100">
      {/* Header */}
      <header className="bg-white border-b px-6 py-4 shadow-sm">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-semibold">PomClaw Agent Chat</h1>
          <div className="text-sm">
            <span className={isConnected ? 'text-green-600' : 'text-red-600'}>
              {isConnected ? '● Connected' : '○ Disconnected'}
            </span>
            {sessionId && (
              <span className="ml-4 text-gray-500">
                Session: {sessionId.slice(0, 12)}...
              </span>
            )}
          </div>
        </div>
      </header>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-6">
        <div className="max-w-3xl mx-auto space-y-4">
          {messages.map((msg, i) => (
            <div
              key={i}
              className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
            >
              <div
                className={`max-w-xl px-4 py-2 rounded-lg ${
                  msg.role === 'user'
                    ? 'bg-blue-500 text-white'
                    : 'bg-white text-gray-900 shadow'
                }`}
              >
                <p className="whitespace-pre-wrap">{msg.content}</p>
                <span className="text-xs opacity-70 mt-1 block">
                  {new Date(msg.timestamp).toLocaleTimeString()}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Input */}
      <div className="bg-white border-t p-4">
        <div className="max-w-3xl mx-auto flex gap-2">
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyPress={handleKeyPress}
            disabled={!isConnected}
            placeholder="输入消息... (Enter发送, Shift+Enter换行)"
            className="flex-1 px-4 py-2 border rounded-lg resize-none focus:outline-none focus:border-blue-500 disabled:bg-gray-100"
            rows={3}
          />
          <button
            onClick={handleSend}
            disabled={!input.trim() || !isConnected}
            className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-300 disabled:cursor-not-allowed"
          >
            发送
          </button>
        </div>
      </div>
    </div>
  );
}
```

---

## 5. 注册到Manager

```go
// pkg/channels/manager.go
func (m *ChannelManager) InitializeChannels() error {
    // ... 其他channels
    
    // HTTP Server Channel
    if m.config.Channels.HTTPServer.Enabled {
        httpChannel, err := NewHTTPServerChannel(m.config.Channels.HTTPServer, m.bus)
        if err != nil {
            return fmt.Errorf("failed to create http_server channel: %w", err)
        }
        m.channels["http_server"] = httpChannel
        logger.InfoC("channel_manager", "HTTP Server channel registered")
    }
    
    return nil
}
```

---

## 6. 测试流程

### 6.1 启动后端

```bash
cd pomclaw
go run cmd/pomclaw/main.go
# HTTP Server Channel listening on :18790
```

### 6.2 测试WebSocket (浏览器控制台)

```javascript
const ws = new WebSocket('ws://localhost:18790/ws');

ws.onopen = () => {
    console.log('Connected');
    
    // 发送消息
    ws.send(JSON.stringify({
        type: 'chat',
        message: '你好',
        session_id: '' // 空表示创建新session
    }));
};

ws.onmessage = (event) => {
    console.log('Received:', JSON.parse(event.data));
};
```

### 6.3 启动前端

```bash
cd frontend
npm run dev
# 访问 http://localhost:3000
```

---

## 7. 优势总结

| 特性 | 说明 |
|------|------|
| **独立运行** | 不依赖任何外部平台 |
| **完全掌控** | 所有逻辑在你的服务器上 |
| **多租户友好** | 每个用户独立的session和agent |
| **实时通信** | WebSocket双向推送 |
| **易于扩展** | 可添加RESTful API |
| **前端简单** | 标准WebSocket协议,无需第三方SDK |

---

## 8. 开发时间估算

| 任务 | 时间 |
|------|------|
| 后端 HTTPServerChannel | 2-3天 |
| 前端 WebSocket Hook | 1-2天 |
| 聊天界面 | 1-2天 |
| 联调测试 | 1天 |
| **总计** | **5-8天** |

---

**这才是真正适合多租户Agent平台的Channel!**
