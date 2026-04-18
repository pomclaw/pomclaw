# HTTP Server Channel - 最小实现示例

## 核心代码（50行）

```go
// pkg/channels/http_server.go
package channels

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
    
    "github.com/gorilla/websocket"
    "github.com/pomclaw/pomclaw/pkg/bus"
)

// WebSocket升级器（允许所有来源）
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

type HTTPServerChannel struct {
    bus *bus.MessageBus
}

// 前端发送的消息
type ClientMessage struct {
    Type      string `json:"type"`
    SessionID string `json:"session_id"`
    Message   string `json:"message"`
}

// 后端返回的消息
type ServerMessage struct {
    Type      string `json:"type"`
    SessionID string `json:"session_id"`
    Content   string `json:"content"`
}

func NewHTTPServerChannel(messageBus *bus.MessageBus) *HTTPServerChannel {
    return &HTTPServerChannel{bus: messageBus}
}

// 启动HTTP服务器（1个函数）
func (c *HTTPServerChannel) Start() error {
    http.HandleFunc("/ws", c.handleWebSocket)
    return http.ListenAndServe(":18790", nil)
}

// 处理WebSocket连接（核心函数）
func (c *HTTPServerChannel) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    // 1. 升级为WebSocket（1行）
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    defer conn.Close()
    
    // 2. 循环读取消息
    for {
        // 读取JSON消息（1行）
        _, message, err := conn.ReadMessage()
        if err != nil {
            break
        }
        
        // 解析消息（2行）
        var msg ClientMessage
        json.Unmarshal(message, &msg)
        
        // 3. 发送到MessageBus
        ctx := context.Background()
        inboundMsg := bus.InboundMessage{
            Channel:    "http_server",
            SessionKey: msg.SessionID,
            Content:    msg.Message,
            AgentID:    "default",
        }
        
        // 4. 等待响应
        response, _ := c.bus.SendAndWait(ctx, inboundMsg, 30*time.Second)
        
        // 5. 返回给前端（3行）
        reply := ServerMessage{
            Type:      "message",
            SessionID: msg.SessionID,
            Content:   response,
        }
        data, _ := json.Marshal(reply)
        conn.WriteMessage(websocket.TextMessage, data)
    }
}
```

---

## 前端代码（30行）

```typescript
// 连接WebSocket
const ws = new WebSocket('ws://localhost:18790/ws');

ws.onopen = () => {
    console.log('Connected');
};

ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    console.log('Received:', msg.content);
};

// 发送消息
function sendMessage(text: string) {
    ws.send(JSON.stringify({
        type: 'chat',
        session_id: 'my_session',
        message: text
    }));
}

// 使用
sendMessage('你好');
```

---

## 就这么简单！

### 库做的事（你不需要写）
- ✅ TCP连接管理
- ✅ WebSocket握手
- ✅ 帧解析/封装
- ✅ 心跳保活

### 你需要写的事
- ⚠️ 调用 `upgrader.Upgrade()` - 升级连接
- ⚠️ 调用 `conn.ReadMessage()` - 读消息
- ⚠️ 调用 `json.Unmarshal()` - 解析JSON
- ⚠️ 调用 `bus.SendAndWait()` - 发到MessageBus
- ⚠️ 调用 `conn.WriteMessage()` - 写消息

**核心代码 < 100行，大部分是if判断和错误处理。**

---

## 完整版本的额外功能

我之前给的 `HTTP_SERVER_CHANNEL_DESIGN.md` 有这些增强：

| 功能 | 代码量 | 必需？ |
|------|--------|--------|
| 基础收发消息 | 50行 | ✅ 必需 |
| 并发管理（多个连接） | +50行 | ✅ 推荐 |
| Session管理 | +30行 | ✅ 推荐 |
| 心跳保活 | +20行 | ✅ 推荐 |
| JWT认证 | +30行 | ⚠️ 可选 |
| CORS中间件 | +10行 | ⚠️ 可选 |
| 日志记录 | +20行 | ⚠️ 可选 |

**MVP版本（基础+并发+Session）约130行代码**

---

## 总结

**你需要写的是"业务逻辑"，不是"协议实现"**

就像写Web后端：
- 你不需要实现HTTP协议 → Go标准库已实现
- 你不需要实现WebSocket协议 → gorilla/websocket已实现
- **你只需要写：收到消息后怎么处理，响应什么内容**

工作量：**1-2天就能写完基础版本**
