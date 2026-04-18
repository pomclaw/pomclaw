# Web Slack Channel 设计方案

## 核心思路

**复用 Slack 协议,为 Web 前端提供标准的 Slack-Compatible API**

```
前端 (Slack SDK) → HTTP/WebSocket → WebSlackChannel → MessageBus → AgentLoop
                                   ↑
                                   复用 slack.go 的消息处理逻辑
```

---

## 1. 架构设计

### 1.1 为什么选择 Slack 协议?

| 优势 | 说明 |
|------|------|
| **前端SDK成熟** | `@slack/web-api`, `@slack/socket-mode` 开箱即用 |
| **后端逻辑复用** | 现有 `slack.go` 代码可直接复用 |
| **消息格式标准** | Slack 的消息格式已经是业界标准 |
| **UI组件丰富** | Slack 有大量开源 React 组件 |
| **易于调试** | 可以用 Slack 官方工具测试 |

---

## 2. 实现方案

### 2.1 创建 WebSlackChannel

```go
// pkg/channels/web_slack.go
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

// WebSlackChannel 实现 Slack-Compatible 的 Web 接口
type WebSlackChannel struct {
    *BaseChannel
    slackChannel *SlackChannel // 复用现有 Slack 逻辑
    
    // WebSocket 连接管理
    clients   sync.Map // map[userID]*WebClient
    upgrader  websocket.Upgrader
    
    // Session 管理
    sessions  sync.Map // map[sessionID]*SessionInfo
}

type WebClient struct {
    UserID    string
    Conn      *websocket.Conn
    Send      chan []byte
    SessionID string
}

type SessionInfo struct {
    SessionID string
    AgentID   string
    ChannelID string // 模拟 Slack 的 channel
    ThreadTS  string // 模拟 Slack 的 thread
    UserID    string
    CreatedAt time.Time
}

// SlackMessage 兼容 Slack 的消息格式
type SlackMessage struct {
    Type      string `json:"type"`       // "message"
    Channel   string `json:"channel"`    // session_id 映射
    User      string `json:"user"`       // user_id
    Text      string `json:"text"`       // 消息内容
    TS        string `json:"ts"`         // timestamp
    ThreadTS  string `json:"thread_ts"`  // 线程(可选)
}

// SlackEvent 兼容 Slack Events API 格式
type SlackEvent struct {
    Type      string      `json:"type"`
    EventType string      `json:"event_type"`
    Event     interface{} `json:"event"`
}

func NewWebSlackChannel(cfg config.WebConfig, messageBus *bus.MessageBus) (*WebSlackChannel, error) {
    base := NewBaseChannel("web_slack", cfg, messageBus, []string{"*"})
    
    // 创建内嵌的 SlackChannel 来复用逻辑
    // 注意: 这里不需要真实的 Slack token,只用其消息处理逻辑
    slackChannel := &SlackChannel{
        BaseChannel: base,
    }
    
    return &WebSlackChannel{
        BaseChannel:  base,
        slackChannel: slackChannel,
        upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                // TODO: 生产环境需要验证 Origin
                return true
            },
        },
    }, nil
}

// Start 启动 HTTP/WebSocket 服务器
func (c *WebSlackChannel) Start(ctx context.Context) error {
    logger.InfoC("web_slack", "Starting Web Slack Channel")
    
    // 注册路由
    http.HandleFunc("/api/slack/rtm.connect", c.handleRTMConnect)
    http.HandleFunc("/api/slack/chat.postMessage", c.handlePostMessage)
    http.HandleFunc("/api/slack/websocket", c.handleWebSocket)
    
    c.setRunning(true)
    logger.InfoC("web_slack", "Web Slack Channel started")
    return nil
}

// handleRTMConnect 模拟 Slack 的 RTM 连接 API
func (c *WebSlackChannel) handleRTMConnect(w http.ResponseWriter, r *http.Request) {
    // 验证 token (简化版)
    token := r.Header.Get("Authorization")
    userID := c.validateToken(token)
    if userID == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // 返回 WebSocket URL
    response := map[string]interface{}{
        "ok":  true,
        "url": fmt.Sprintf("ws://%s/api/slack/websocket?token=%s", r.Host, token),
        "self": map[string]string{
            "id":   userID,
            "name": "user_" + userID,
        },
    }
    
    json.NewEncoder(w).Encode(response)
}

// handlePostMessage 模拟 Slack 的发送消息 API
func (c *WebSlackChannel) handlePostMessage(w http.ResponseWriter, r *http.Request) {
    var msg SlackMessage
    if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // 验证 token
    token := r.Header.Get("Authorization")
    userID := c.validateToken(token)
    if userID == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // 处理消息
    sessionID := msg.Channel // Slack 的 channel 对应我们的 session
    agentID := c.getOrCreateSession(sessionID, userID)
    
    // 构造 InboundMessage
    inboundMsg := bus.InboundMessage{
        Channel:    "web_slack",
        SenderID:   userID,
        ChatID:     c.formatChatID(sessionID, msg.ThreadTS),
        Content:    msg.Text,
        SessionKey: sessionID,
        AgentID:    agentID,
    }
    
    // 发送到 MessageBus
    ctx := context.Background()
    go func() {
        response, err := c.bus.SendAndWait(ctx, inboundMsg, 30*time.Second)
        if err != nil {
            logger.ErrorCF("web_slack", "Failed to process message", map[string]interface{}{
                "error": err.Error(),
            })
            return
        }
        
        // 通过 WebSocket 推送响应
        c.broadcastMessage(userID, SlackMessage{
            Type:     "message",
            Channel:  sessionID,
            User:     "bot",
            Text:     response,
            TS:       fmt.Sprintf("%d", time.Now().Unix()),
            ThreadTS: msg.ThreadTS,
        })
    }()
    
    // 立即返回成功
    json.NewEncoder(w).Encode(map[string]interface{}{
        "ok":      true,
        "channel": sessionID,
        "ts":      fmt.Sprintf("%d", time.Now().Unix()),
    })
}

// handleWebSocket 处理 WebSocket 连接
func (c *WebSlackChannel) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    // 验证 token
    token := r.URL.Query().Get("token")
    userID := c.validateToken(token)
    if userID == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // 升级为 WebSocket
    conn, err := c.upgrader.Upgrade(w, r, nil)
    if err != nil {
        logger.ErrorCF("web_slack", "WebSocket upgrade failed", map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    // 创建客户端
    client := &WebClient{
        UserID: userID,
        Conn:   conn,
        Send:   make(chan []byte, 256),
    }
    
    c.clients.Store(userID, client)
    defer c.clients.Delete(userID)
    
    logger.InfoCF("web_slack", "WebSocket connected", map[string]interface{}{
        "user_id": userID,
    })
    
    // 发送连接成功消息
    helloMsg := map[string]interface{}{
        "type": "hello",
    }
    c.sendToClient(client, helloMsg)
    
    // 启动读写协程
    go c.writePump(client)
    c.readPump(client)
}

// readPump 读取客户端消息
func (c *WebSlackChannel) readPump(client *WebClient) {
    defer client.Conn.Close()
    
    client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    client.Conn.SetPongHandler(func(string) error {
        client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })
    
    for {
        _, message, err := client.Conn.ReadMessage()
        if err != nil {
            break
        }
        
        var msg SlackMessage
        if err := json.Unmarshal(message, &msg); err != nil {
            continue
        }
        
        // 处理消息 (复用 handlePostMessage 的逻辑)
        if msg.Type == "message" {
            c.processClientMessage(client, &msg)
        }
    }
}

// writePump 向客户端写入消息
func (c *WebSlackChannel) writePump(client *WebClient) {
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

// processClientMessage 处理客户端消息
func (c *WebSlackChannel) processClientMessage(client *WebClient, msg *SlackMessage) {
    sessionID := msg.Channel
    if sessionID == "" {
        // 创建新 session
        sessionID = c.createSession(client.UserID)
        client.SessionID = sessionID
    }
    
    agentID := c.getOrCreateSession(sessionID, client.UserID)
    
    // 构造 InboundMessage
    inboundMsg := bus.InboundMessage{
        Channel:    "web_slack",
        SenderID:   client.UserID,
        ChatID:     c.formatChatID(sessionID, msg.ThreadTS),
        Content:    msg.Text,
        SessionKey: sessionID,
        AgentID:    agentID,
    }
    
    // 发送到 MessageBus
    ctx := context.Background()
    go func() {
        response, err := c.bus.SendAndWait(ctx, inboundMsg, 30*time.Second)
        if err != nil {
            c.sendError(client, err.Error())
            return
        }
        
        // 发送响应
        c.sendToClient(client, SlackMessage{
            Type:     "message",
            Channel:  sessionID,
            User:     "bot",
            Text:     response,
            TS:       fmt.Sprintf("%d", time.Now().Unix()),
            ThreadTS: msg.ThreadTS,
        })
    }()
}

// Helper functions
func (c *WebSlackChannel) sendToClient(client *WebClient, msg interface{}) {
    data, _ := json.Marshal(msg)
    select {
    case client.Send <- data:
    default:
        close(client.Send)
        c.clients.Delete(client.UserID)
    }
}

func (c *WebSlackChannel) sendError(client *WebClient, errMsg string) {
    c.sendToClient(client, map[string]interface{}{
        "type":  "error",
        "error": errMsg,
    })
}

func (c *WebSlackChannel) broadcastMessage(userID string, msg SlackMessage) {
    if client, ok := c.clients.Load(userID); ok {
        c.sendToClient(client.(*WebClient), msg)
    }
}

func (c *WebSlackChannel) validateToken(token string) string {
    // TODO: 实现真正的 JWT 验证
    // 临时实现: token 就是 userID
    if token == "" {
        return ""
    }
    return token
}

func (c *WebSlackChannel) getOrCreateSession(sessionID, userID string) string {
    if val, ok := c.sessions.Load(sessionID); ok {
        return val.(*SessionInfo).AgentID
    }
    
    // 创建新 session,默认使用 default agent
    agentID := "default"
    c.sessions.Store(sessionID, &SessionInfo{
        SessionID: sessionID,
        AgentID:   agentID,
        ChannelID: sessionID,
        UserID:    userID,
        CreatedAt: time.Now(),
    })
    
    return agentID
}

func (c *WebSlackChannel) createSession(userID string) string {
    return fmt.Sprintf("session_%s_%d", userID, time.Now().UnixNano())
}

func (c *WebSlackChannel) formatChatID(channel, threadTS string) string {
    if threadTS != "" {
        return fmt.Sprintf("%s:%s", channel, threadTS)
    }
    return channel
}

func (c *WebSlackChannel) Stop(ctx context.Context) error {
    logger.InfoC("web_slack", "Stopping Web Slack Channel")
    c.setRunning(false)
    return nil
}
```

---

## 3. 前端实现

### 3.1 使用 Slack SDK

```typescript
// src/services/slackClient.ts
import { WebClient } from '@slack/web-api';

class SlackClientService {
  private client: WebClient;
  private ws: WebSocket | null = null;
  private token: string;

  constructor(baseURL: string, token: string) {
    this.token = token;
    
    // 配置 Slack 客户端指向我们的后端
    this.client = new WebClient(token, {
      slackApiUrl: baseURL, // 指向我们的服务器
    });
  }

  // 建立 WebSocket 连接
  async connect(): Promise<void> {
    // 调用 rtm.connect 获取 WebSocket URL
    const response = await fetch('http://localhost:18790/api/slack/rtm.connect', {
      headers: {
        'Authorization': this.token,
      },
    });
    
    const data = await response.json();
    
    // 连接 WebSocket
    this.ws = new WebSocket(data.url);
    
    this.ws.onopen = () => {
      console.log('WebSocket connected');
    };
    
    this.ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      this.handleMessage(msg);
    };
  }

  // 发送消息 (使用 Slack 标准 API)
  async sendMessage(channel: string, text: string): Promise<void> {
    await this.client.chat.postMessage({
      channel: channel,
      text: text,
    });
  }

  // 监听消息
  private handleMessage(msg: any) {
    if (msg.type === 'message') {
      console.log('Received:', msg);
      // 触发回调
      this.onMessage?.(msg);
    }
  }

  onMessage?: (msg: any) => void;
}

export default SlackClientService;
```

---

### 3.2 React 组件

```typescript
// src/pages/Chat.tsx
import React, { useState, useEffect } from 'react';
import SlackClientService from '../services/slackClient';

export default function Chat() {
  const [messages, setMessages] = useState<any[]>([]);
  const [input, setInput] = useState('');
  const [client, setClient] = useState<SlackClientService | null>(null);
  const [sessionId, setSessionId] = useState('session_default');

  useEffect(() => {
    // 初始化 Slack 客户端
    const token = localStorage.getItem('auth_token') || 'test_token';
    const slackClient = new SlackClientService(
      'http://localhost:18790',
      token
    );

    // 监听消息
    slackClient.onMessage = (msg) => {
      setMessages((prev) => [...prev, msg]);
    };

    // 连接 WebSocket
    slackClient.connect();
    setClient(slackClient);

    return () => {
      // 清理连接
    };
  }, []);

  const handleSend = async () => {
    if (!client || !input.trim()) return;

    // 使用 Slack API 发送消息
    await client.sendMessage(sessionId, input);
    
    // 添加到消息列表
    setMessages((prev) => [
      ...prev,
      {
        type: 'message',
        user: 'me',
        text: input,
        ts: Date.now().toString(),
      },
    ]);
    
    setInput('');
  };

  return (
    <div className="chat-container">
      <div className="messages">
        {messages.map((msg, i) => (
          <div key={i} className={`message ${msg.user === 'me' ? 'sent' : 'received'}`}>
            <p>{msg.text}</p>
          </div>
        ))}
      </div>
      
      <div className="input-area">
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyPress={(e) => e.key === 'Enter' && handleSend()}
          placeholder="输入消息..."
        />
        <button onClick={handleSend}>发送</button>
      </div>
    </div>
  );
}
```

---

## 4. 优势总结

| 方面 | 优势 |
|------|------|
| **开发速度** | 前端直接用 Slack SDK,后端复用现有代码 |
| **代码复用** | 最大化复用 `slack.go` 的逻辑 |
| **调试方便** | 可以用 Slack 官方工具测试 |
| **易于扩展** | 支持真实 Slack 和 Web 双模式 |
| **生态丰富** | 大量 Slack UI 组件可用 |

---

## 5. 实施步骤

1. **Phase 1: 后端** (2-3天)
   - [ ] 创建 `web_slack.go`
   - [ ] 实现 HTTP API (`rtm.connect`, `chat.postMessage`)
   - [ ] 实现 WebSocket 处理
   - [ ] Session 管理

2. **Phase 2: 前端** (2-3天)
   - [ ] 配置 Slack SDK
   - [ ] 创建聊天界面
   - [ ] 连接 WebSocket
   - [ ] 消息发送/接收

3. **Phase 3: 联调** (1-2天)
   - [ ] 端到端测试
   - [ ] 修复 Bug
   - [ ] 性能优化

---

## 6. 配置示例

```yaml
# config/config.yaml
channels:
  web_slack:
    enabled: true
    port: 18790
    allow_from: ["*"]
    cors_origins:
      - "http://localhost:3000"
      - "https://app.example.com"
```

---

**总结: 这个方案完美结合了 Slack 的成熟生态和你现有的架构,最小化开发成本!**
