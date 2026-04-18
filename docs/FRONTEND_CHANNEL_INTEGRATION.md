# 前端与Channel集成方案

## 项目目标

构建企业级多Agent租赁平台的前后端交互系统:
- **前端**: 提供登录、创建session、实时聊天的Web界面
- **后端Channel**: 监听用户连接,处理消息,路由到对应Agent

---

## 1. 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      前端 (Web UI)                           │
│  React + TypeScript + WebSocket Client                      │
│  ┌───────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  登录页   │  │  Session列表  │  │   Chat界面   │         │
│  └───────────┘  └──────────────┘  └──────────────┘         │
└───────────────────────────┬─────────────────────────────────┘
                            │ WebSocket / HTTP REST
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                  后端 Channel 层                              │
│  pkg/channels/web_channel.go                                │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐       │
│  │ WebSocket    │  │  HTTP API    │  │  Session    │       │
│  │  Handler     │  │   Handler    │  │  Manager    │       │
│  └──────┬───────┘  └──────┬───────┘  └──────┬──────┘       │
│         │                  │                  │              │
│         └──────────────────┴──────────────────┘              │
│                            │                                 │
│                   ┌────────▼────────┐                        │
│                   │  Message Bus    │                        │
│                   │  (pkg/bus)      │                        │
│                   └────────┬────────┘                        │
└────────────────────────────┼──────────────────────────────────┘
                             │
┌────────────────────────────▼──────────────────────────────────┐
│                     Agent Loop 层                              │
│  pkg/agent/loop.go                                            │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────┐         │
│  │  Default    │  │    Agent-1   │  │   Agent-2   │         │
│  │  Agent      │  │  (用户自定义) │  │  (用户自定义) │         │
│  └─────────────┘  └──────────────┘  └─────────────┘         │
└────────────────────────────────────────────────────────────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │   PostgreSQL    │
                    │  - Sessions     │
                    │  - Agents       │
                    │  - Memories     │
                    └─────────────────┘
```

---

## 2. 核心流程设计

### 2.1 用户首次访问流程

```
用户打开网页
    ↓
前端: 显示登录界面
    ↓
用户输入用户名/密码 (简化版: 暂不实现注册,用默认账号)
    ↓
前端: POST /api/auth/login
    ↓
后端: 验证成功,返回 JWT token
    ↓
前端: 保存 token,跳转到聊天页面
    ↓
前端: 建立 WebSocket 连接 ws://localhost:18790/ws?token={JWT}
    ↓
Channel: 验证 token,绑定 user_id
    ↓
前端: 发送第一条消息 {"type": "chat", "message": "你好", "session_id": ""}
    ↓
Channel: session_id 为空,创建新 session → 绑定到 default agent
    ↓
AgentLoop: 处理消息,调用 LLM
    ↓
Channel: 流式返回响应给前端
    ↓
前端: 实时显示 Agent 回复
```

---

### 2.2 后续对话流程 (已有session)

```
用户发送消息 "帮我查询天气"
    ↓
前端: 通过 WebSocket 发送 {"type": "chat", "message": "帮我查询天气", "session_id": "session_123"}
    ↓
Channel: 从 session_id 查询对应的 agent_id
    ↓
Channel: 构造 InboundMessage {
    Channel:    "web",
    SessionKey: "session_123",
    AgentID:    "agent_abc", (从 session 查询得到)
    Content:    "帮我查询天气",
}
    ↓
Channel: 发送到 MessageBus
    ↓
AgentLoop: 接收消息,调用 LLM
    ↓
AgentLoop: 返回响应到 MessageBus
    ↓
Channel: 监听 MessageBus,收到响应
    ↓
Channel: 通过 WebSocket 推送给前端 {"type": "message", "content": "今天晴天..."}
    ↓
前端: 追加到消息列表显示
```

---

## 3. 技术实现方案

### 3.1 前端实现

#### 3.1.1 目录结构

```
frontend/
├── public/
│   └── index.html
├── src/
│   ├── main.tsx                 # 入口
│   ├── App.tsx                  # 根组件
│   ├── pages/
│   │   ├── Login.tsx            # 登录页
│   │   └── Chat.tsx             # 聊天页
│   ├── components/
│   │   ├── MessageList.tsx      # 消息列表
│   │   ├── MessageInput.tsx     # 输入框
│   │   └── SessionSidebar.tsx   # Session侧边栏
│   ├── hooks/
│   │   ├── useWebSocket.ts      # WebSocket Hook
│   │   └── useAuth.ts           # 认证 Hook
│   ├── services/
│   │   ├── api.ts               # HTTP API
│   │   └── websocket.ts         # WebSocket 管理
│   └── types/
│       └── chat.ts              # 类型定义
├── package.json
├── tsconfig.json
└── vite.config.ts
```

---

#### 3.1.2 核心代码示例

**WebSocket Hook**

```typescript
// src/hooks/useWebSocket.ts
import { useState, useEffect, useCallback, useRef } from 'react';

interface Message {
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
}

export function useWebSocket(url: string, token: string) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [isTyping, setIsTyping] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    // 建立 WebSocket 连接
    const ws = new WebSocket(`${url}?token=${token}`);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('WebSocket connected');
      setIsConnected(true);
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      if (data.type === 'stream') {
        // 流式响应: 追加内容
        setIsTyping(true);
        setMessages((prev) => {
          const lastMsg = prev[prev.length - 1];
          if (lastMsg?.role === 'assistant' && !lastMsg.done) {
            // 追加到最后一条消息
            return [
              ...prev.slice(0, -1),
              { ...lastMsg, content: lastMsg.content + data.content }
            ];
          } else {
            // 新建助手消息
            return [
              ...prev,
              {
                role: 'assistant',
                content: data.content,
                done: false,
                timestamp: new Date()
              }
            ];
          }
        });
      } else if (data.type === 'done') {
        // 完成标记
        setIsTyping(false);
        setMessages((prev) => {
          const lastMsg = prev[prev.length - 1];
          return [
            ...prev.slice(0, -1),
            { ...lastMsg, done: true }
          ];
        });
      } else if (data.type === 'error') {
        console.error('Error:', data.message);
        setIsTyping(false);
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

  const sendMessage = useCallback((content: string, sessionId?: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      // 添加用户消息到列表
      setMessages((prev) => [
        ...prev,
        {
          role: 'user',
          content,
          timestamp: new Date()
        }
      ]);

      // 发送到后端
      wsRef.current.send(JSON.stringify({
        type: 'chat',
        message: content,
        session_id: sessionId || ''
      }));
    }
  }, []);

  return {
    messages,
    isConnected,
    isTyping,
    sendMessage
  };
}
```

---

**Chat 组件**

```typescript
// src/pages/Chat.tsx
import React, { useState } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';
import MessageList from '../components/MessageList';
import MessageInput from '../components/MessageInput';

const WEBSOCKET_URL = 'ws://localhost:18790/ws';

export default function Chat() {
  const [sessionId, setSessionId] = useState<string>('');
  const token = localStorage.getItem('auth_token') || '';
  
  const { messages, isConnected, isTyping, sendMessage } = useWebSocket(
    WEBSOCKET_URL,
    token
  );

  const handleSend = (content: string) => {
    sendMessage(content, sessionId);
  };

  return (
    <div className="flex h-screen bg-gray-100">
      {/* 侧边栏 (未来扩展: Session 列表) */}
      <aside className="w-64 bg-white border-r">
        <div className="p-4">
          <h2 className="text-lg font-semibold">对话历史</h2>
          <button 
            onClick={() => setSessionId('')}
            className="mt-2 w-full py-2 px-4 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            新建对话
          </button>
        </div>
      </aside>

      {/* 聊天区域 */}
      <main className="flex-1 flex flex-col">
        {/* Header */}
        <header className="bg-white border-b px-6 py-4 shadow-sm">
          <h1 className="text-xl font-semibold">AI Agent 对话</h1>
          <div className="text-sm text-gray-500">
            {isConnected ? '🟢 已连接' : '🔴 未连接'}
            {sessionId && ` | Session: ${sessionId.slice(0, 8)}...`}
          </div>
        </header>

        {/* 消息列表 */}
        <div className="flex-1 overflow-y-auto p-6">
          <MessageList messages={messages} isTyping={isTyping} />
        </div>

        {/* 输入框 */}
        <div className="bg-white border-t p-4">
          <MessageInput 
            onSend={handleSend}
            disabled={!isConnected}
          />
        </div>
      </main>
    </div>
  );
}
```

---

**消息列表组件**

```typescript
// src/components/MessageList.tsx
import React, { useRef, useEffect } from 'react';

interface Message {
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
}

interface Props {
  messages: Message[];
  isTyping: boolean;
}

export default function MessageList({ messages, isTyping }: Props) {
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // 自动滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  return (
    <div className="space-y-4">
      {messages.map((msg, index) => (
        <div
          key={index}
          className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
        >
          <div
            className={`max-w-xl px-4 py-2 rounded-lg ${
              msg.role === 'user'
                ? 'bg-blue-500 text-white'
                : 'bg-gray-200 text-gray-900'
            }`}
          >
            <p className="whitespace-pre-wrap">{msg.content}</p>
            <span className="text-xs opacity-70 mt-1 block">
              {msg.timestamp.toLocaleTimeString()}
            </span>
          </div>
        </div>
      ))}
      
      {isTyping && (
        <div className="flex justify-start">
          <div className="bg-gray-200 px-4 py-2 rounded-lg">
            <span className="animate-pulse">正在输入...</span>
          </div>
        </div>
      )}
      
      <div ref={messagesEndRef} />
    </div>
  );
}
```

---

**输入框组件**

```typescript
// src/components/MessageInput.tsx
import React, { useState } from 'react';

interface Props {
  onSend: (message: string) => void;
  disabled?: boolean;
}

export default function MessageInput({ onSend, disabled }: Props) {
  const [input, setInput] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (input.trim() && !disabled) {
      onSend(input);
      setInput('');
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="flex gap-2">
      <textarea
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={handleKeyDown}
        disabled={disabled}
        placeholder="输入消息... (Enter 发送, Shift+Enter 换行)"
        className="flex-1 px-4 py-2 border border-gray-300 rounded-lg resize-none focus:outline-none focus:border-blue-500"
        rows={3}
      />
      <button
        type="submit"
        disabled={!input.trim() || disabled}
        className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-300 disabled:cursor-not-allowed"
      >
        发送
      </button>
    </form>
  );
}
```

---

### 3.2 后端实现

#### 3.2.1 WebSocket Channel

```go
// pkg/channels/web_channel.go
package channels

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    "github.com/pomclaw/pomclaw/pkg/bus"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        // TODO: 生产环境需要严格验证 Origin
        return true
    },
}

type WebChannel struct {
    bus           *bus.MessageBus
    clients       sync.Map // map[clientID]*Client
    sessionStore  SessionStore
}

type Client struct {
    ID        string
    UserID    string
    Conn      *websocket.Conn
    Send      chan []byte
    SessionID string
}

// ClientMessage 前端发送的消息格式
type ClientMessage struct {
    Type      string `json:"type"`       // "chat"
    Message   string `json:"message"`    // 用户消息内容
    SessionID string `json:"session_id"` // 可选,为空则创建新session
}

// ServerMessage 后端返回的消息格式
type ServerMessage struct {
    Type       string `json:"type"`        // "stream", "done", "error"
    Content    string `json:"content"`     // 消息内容
    SessionID  string `json:"session_id"`  // session标识
    Done       bool   `json:"done"`        // 是否完成
}

func NewWebChannel(bus *bus.MessageBus, sessionStore SessionStore) *WebChannel {
    return &WebChannel{
        bus:          bus,
        sessionStore: sessionStore,
    }
}

// HandleWebSocket 处理 WebSocket 连接
func (wc *WebChannel) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    // 1. 验证 JWT token (从查询参数或 Header)
    token := r.URL.Query().Get("token")
    userID, err := wc.validateToken(token)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // 2. 升级为 WebSocket
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade failed: %v", err)
        return
    }

    // 3. 创建客户端
    client := &Client{
        ID:     generateClientID(),
        UserID: userID,
        Conn:   conn,
        Send:   make(chan []byte, 256),
    }

    wc.clients.Store(client.ID, client)
    defer wc.clients.Delete(client.ID)

    log.Printf("WebSocket client connected: %s (user: %s)", client.ID, userID)

    // 4. 启动读写协程
    go wc.writePump(client)
    wc.readPump(client)
}

// readPump 读取客户端消息
func (wc *WebChannel) readPump(client *Client) {
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
                log.Printf("WebSocket error: %v", err)
            }
            break
        }

        // 解析消息
        var msg ClientMessage
        if err := json.Unmarshal(message, &msg); err != nil {
            log.Printf("Invalid message format: %v", err)
            continue
        }

        // 处理消息
        wc.handleClientMessage(client, &msg)
    }
}

// writePump 向客户端写入消息
func (wc *WebChannel) writePump(client *Client) {
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
                // Channel 关闭
                client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            w, err := client.Conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            w.Write(message)

            if err := w.Close(); err != nil {
                return
            }

        case <-ticker.C:
            // 发送心跳 Ping
            client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

// handleClientMessage 处理客户端消息
func (wc *WebChannel) handleClientMessage(client *Client, msg *ClientMessage) {
    ctx := context.Background()

    // 1. 确定 session_id
    sessionID := msg.SessionID
    if sessionID == "" {
        // 创建新 session
        sessionID = generateSessionID(client.UserID)
        client.SessionID = sessionID
        log.Printf("Created new session: %s for user: %s", sessionID, client.UserID)
    }

    // 2. 从 session 查询 agent_id (或使用 default)
    agentID := wc.sessionStore.GetAgentID(sessionID)
    if agentID == "" {
        agentID = "default" // 默认 agent
        wc.sessionStore.BindSession(sessionID, agentID)
    }

    // 3. 构造 InboundMessage
    inboundMsg := bus.InboundMessage{
        Channel:    "web",
        SenderID:   client.UserID,
        ChatID:     client.ID,
        Content:    msg.Message,
        SessionKey: sessionID,
        AgentID:    agentID,
    }

    // 4. 发送到 MessageBus
    go func() {
        response, err := wc.bus.SendAndWait(ctx, inboundMsg, 30*time.Second)
        if err != nil {
            // 发送错误消息
            wc.sendToClient(client, ServerMessage{
                Type:    "error",
                Content: err.Error(),
            })
            return
        }

        // 5. 流式返回响应 (简化版: 直接返回完整响应)
        // TODO: 支持真正的流式输出
        wc.sendToClient(client, ServerMessage{
            Type:      "stream",
            Content:   response,
            SessionID: sessionID,
        })

        wc.sendToClient(client, ServerMessage{
            Type:      "done",
            SessionID: sessionID,
            Done:      true,
        })
    }()
}

// sendToClient 向客户端发送消息
func (wc *WebChannel) sendToClient(client *Client, msg ServerMessage) {
    data, _ := json.Marshal(msg)
    select {
    case client.Send <- data:
    default:
        // Channel 满了,关闭连接
        close(client.Send)
        wc.clients.Delete(client.ID)
    }
}

// validateToken 验证 JWT token
func (wc *WebChannel) validateToken(token string) (string, error) {
    // TODO: 实现 JWT 验证
    // 临时实现: 直接返回 token 作为 userID
    if token == "" {
        return "", fmt.Errorf("missing token")
    }
    return token, nil
}

// SessionStore 接口
type SessionStore interface {
    GetAgentID(sessionID string) string
    BindSession(sessionID, agentID string)
}

// Helper functions
func generateClientID() string {
    return fmt.Sprintf("client_%d", time.Now().UnixNano())
}

func generateSessionID(userID string) string {
    return fmt.Sprintf("session_%s_%d", userID, time.Now().UnixNano())
}
```

---

#### 3.2.2 HTTP 路由注册

```go
// cmd/pomclaw/main.go (添加 WebSocket 路由)

func main() {
    // ... 现有初始化代码

    // 创建 WebChannel
    webChannel := channels.NewWebChannel(messageBus, sessionStore)

    // 注册 HTTP 路由
    http.HandleFunc("/ws", webChannel.HandleWebSocket)
    http.HandleFunc("/health", healthCheck)

    // 启动服务器
    addr := ":18790"
    log.Printf("Starting WebSocket server on %s", addr)
    if err := http.ListenAndServe(addr, nil); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}
```

---

#### 3.2.3 Session Store 实现

```go
// pkg/channels/session_store.go
package channels

import (
    "database/sql"
    "sync"
)

type MemorySessionStore struct {
    // 内存缓存: session_id -> agent_id
    sessions sync.Map
    db       *sql.DB
}

func NewMemorySessionStore(db *sql.DB) *MemorySessionStore {
    return &MemorySessionStore{
        db: db,
    }
}

func (s *MemorySessionStore) GetAgentID(sessionID string) string {
    // 1. 从缓存查
    if agentID, ok := s.sessions.Load(sessionID); ok {
        return agentID.(string)
    }

    // 2. 从数据库查
    var agentID string
    err := s.db.QueryRow(`
        SELECT agent_id FROM POM_SESSIONS WHERE session_key = $1
    `, sessionID).Scan(&agentID)

    if err == nil {
        s.sessions.Store(sessionID, agentID)
        return agentID
    }

    return ""
}

func (s *MemorySessionStore) BindSession(sessionID, agentID string) {
    s.sessions.Store(sessionID, agentID)
    
    // 异步写入数据库 (首次创建session时)
    go func() {
        _, err := s.db.Exec(`
            INSERT INTO POM_SESSIONS (session_key, agent_id, messages, created_at)
            VALUES ($1, $2, '[]', NOW())
            ON CONFLICT (session_key) DO NOTHING
        `, sessionID, agentID)
        
        if err != nil {
            log.Printf("Failed to persist session: %v", err)
        }
    }()
}
```

---

## 4. 部署与测试

### 4.1 本地开发环境

**后端启动:**
```bash
cd pomclaw
go run cmd/pomclaw/main.go
# WebSocket 服务运行在 ws://localhost:18790/ws
```

**前端启动:**
```bash
cd frontend
npm install
npm run dev
# 前端运行在 http://localhost:3000
```

---

### 4.2 测试流程

**Step 1: 测试 WebSocket 连接**
```javascript
// 浏览器控制台测试
const ws = new WebSocket('ws://localhost:18790/ws?token=test_user_123');

ws.onopen = () => {
    console.log('Connected');
    ws.send(JSON.stringify({
        type: 'chat',
        message: '你好',
        session_id: ''
    }));
};

ws.onmessage = (event) => {
    console.log('Received:', JSON.parse(event.data));
};
```

**Step 2: 测试前端界面**
1. 打开 http://localhost:3000
2. 输入消息 "你好"
3. 查看是否收到 Agent 回复

---

## 5. 未来扩展

### 5.1 短期 (MVP 阶段)
- [ ] 实现简单的登录认证 (JWT)
- [ ] 支持创建多个 session
- [ ] Session 列表侧边栏
- [ ] 流式输出优化

### 5.2 中期
- [ ] 支持多个自定义 Agent
- [ ] Agent 配置页面
- [ ] 会话历史搜索
- [ ] 导出对话记录

### 5.3 长期
- [ ] 多租户支持
- [ ] 语音输入/输出
- [ ] 移动端适配
- [ ] 插件市场

---

## 6. 关键技术点总结

| 组件 | 技术选型 | 原因 |
|------|---------|------|
| 前端框架 | React + TypeScript | 生态成熟,类型安全 |
| 实时通信 | WebSocket | 低延迟,双向通信 |
| 后端框架 | Go + Gorilla WebSocket | 高并发,原生支持 |
| Session管理 | 内存缓存 + PostgreSQL | 快速查询 + 持久化 |
| 认证 | JWT | 无状态,易扩展 |

---

## 7. 项目文件清单

### 新增文件
```
frontend/
├── src/
│   ├── pages/
│   │   ├── Login.tsx          (新增)
│   │   └── Chat.tsx           (新增)
│   ├── components/
│   │   ├── MessageList.tsx    (新增)
│   │   └── MessageInput.tsx   (新增)
│   └── hooks/
│       └── useWebSocket.ts    (新增)

backend/
├── pkg/channels/
│   ├── web_channel.go         (新增)
│   └── session_store.go       (新增)
└── cmd/pomclaw/
    └── main.go                (修改: 添加WebSocket路由)
```

---

## 8. 开发时间估算

| 任务 | 预计时间 |
|------|---------|
| 后端 WebSocket Channel | 2-3 天 |
| 前端基础界面 | 2-3 天 |
| Session 管理 | 1-2 天 |
| 联调测试 | 1-2 天 |
| **总计** | **6-10 天** |

---

**文档版本**: 1.0  
**创建日期**: 2026-04-19  
**作者**: PomClaw Team
