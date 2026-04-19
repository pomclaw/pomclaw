# OpenClaw UI 集成方案

## 发现

OpenClaw有一个完整的Web UI控制面板，使用以下技术栈：
- **前端**: Lit (Web Components) + TypeScript + Vite
- **后端**: Node.js + TypeScript + WebSocket
- **协议**: 自定义Gateway协议（基于WebSocket）

## OpenClaw架构分析

### 前端 (ui/)
```
ui/
├── src/
│   ├── ui/
│   │   ├── app.ts                    # 主应用组件
│   │   ├── gateway.ts               # WebSocket客户端
│   │   ├── app-chat.ts              # 聊天界面
│   │   ├── app-channels.ts          # 频道管理
│   │   └── app-render.ts            # 渲染引擎
│   └── main.ts
├── package.json
└── vite.config.ts
```

### 后端 (src/gateway/)
```
src/gateway/
├── server.impl.ts          # Gateway服务器主实现
├── server-http.ts          # HTTP/WebSocket服务器
├── server-chat.ts          # 聊天处理
├── control-ui.ts           # UI路由
├── client.ts               # 客户端管理
└── protocol/               # 协议定义
```

### 通信协议

**WebSocket消息格式**:
```typescript
// 客户端请求
{
  type: "req",
  id: "uuid",
  method: "chat.send",
  params: { message: "你好" }
}

// 服务器响应
{
  type: "res",
  id: "uuid",
  ok: true,
  payload: { response: "你好！" }
}

// 服务器事件
{
  type: "event",
  event: "message",
  payload: { content: "..." },
  seq: 123
}
```

---

## 集成方案

### 方案A: 直接复用OpenClaw UI (推荐)

**思路**: 将OpenClaw的UI和Gateway协议直接移植到Pomclaw

#### 优势
- ✅ 完整的Web界面（聊天、频道、设置、统计）
- ✅ 成熟的WebSocket协议
- ✅ 已实现认证、会话管理
- ✅ 支持多语言 (en, zh-CN, zh-TW, de, es, pt-BR)
- ✅ 响应式设计

#### 实施步骤

**Step 1: 复制UI到Pomclaw** (1天)

```bash
# 复制前端代码
cp -r /d/go/openclaw-feat-mattermost-block-streaming-rebased/openclaw-feat-mattermost-block-streaming-rebased/ui \
      /d/go/pomclaw/pomclaw/ui

# 修改package.json中的名称
sed -i 's/openclaw-control-ui/pomclaw-control-ui/g' /d/go/pomclaw/pomclaw/ui/package.json
```

**Step 2: 实现Go版本的Gateway服务器** (3-4天)

创建 `pkg/gateway/server.go`:

```go
package gateway

import (
    "encoding/json"
    "net/http"
    "sync"
    
    "github.com/gorilla/websocket"
    "github.com/pomclaw/pomclaw/pkg/bus"
    "github.com/pomclaw/pomclaw/pkg/logger"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

// GatewayServer 兼容OpenClaw协议的Gateway服务器
type GatewayServer struct {
    bus      *bus.MessageBus
    clients  sync.Map
    mux      *http.ServeMux
}

// OpenClaw协议消息格式
type RequestFrame struct {
    Type   string                 `json:"type"`   // "req"
    ID     string                 `json:"id"`     // UUID
    Method string                 `json:"method"` // "chat.send"
    Params map[string]interface{} `json:"params"`
}

type ResponseFrame struct {
    Type    string      `json:"type"`    // "res"
    ID      string      `json:"id"`
    OK      bool        `json:"ok"`
    Payload interface{} `json:"payload,omitempty"`
    Error   *ErrorInfo  `json:"error,omitempty"`
}

type EventFrame struct {
    Type    string      `json:"type"`    // "event"
    Event   string      `json:"event"`
    Payload interface{} `json:"payload"`
    Seq     int         `json:"seq"`
}

type ErrorInfo struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

type Client struct {
    ID       string
    UserID   string
    Conn     *websocket.Conn
    Send     chan []byte
    SeqNum   int
}

func NewGatewayServer(bus *bus.MessageBus) *GatewayServer {
    return &GatewayServer{
        bus: bus,
        mux: http.NewServeMux(),
    }
}

func (s *GatewayServer) Start(port int) error {
    // 注册路由
    s.mux.HandleFunc("/", s.serveUI)
    s.mux.HandleFunc("/ws", s.handleWebSocket)
    s.mux.HandleFunc("/health", s.handleHealth)
    
    addr := fmt.Sprintf(":%d", port)
    logger.InfoCF("gateway", "Starting Gateway server", map[string]interface{}{
        "port": port,
        "url":  fmt.Sprintf("http://localhost:%d", port),
    })
    
    return http.ListenAndServe(addr, s.mux)
}

// serveUI 提供静态UI文件
func (s *GatewayServer) serveUI(w http.ResponseWriter, r *http.Request) {
    // 提供ui/dist目录下的文件
    http.FileServer(http.Dir("./ui/dist")).ServeHTTP(w, r)
}

// handleWebSocket 处理WebSocket连接
func (s *GatewayServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        logger.ErrorCF("gateway", "WebSocket upgrade failed", map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    client := &Client{
        ID:     generateClientID(),
        UserID: "user_" + generateClientID(), // TODO: 从认证获取
        Conn:   conn,
        Send:   make(chan []byte, 256),
        SeqNum: 0,
    }
    
    s.clients.Store(client.ID, client)
    defer s.clients.Delete(client.ID)
    
    logger.InfoCF("gateway", "WebSocket connected", map[string]interface{}{
        "client_id": client.ID,
    })
    
    // 发送hello消息
    s.sendHello(client)
    
    // 启动读写协程
    go s.writePump(client)
    s.readPump(client)
}

// sendHello 发送初始连接消息
func (s *GatewayServer) sendHello(client *Client) {
    hello := map[string]interface{}{
        "type":     "hello-ok",
        "protocol": 1,
        "server": map[string]string{
            "version": "pomclaw-1.0",
            "connId":  client.ID,
        },
        "features": map[string]interface{}{
            "methods": []string{"chat.send", "sessions.list"},
            "events":  []string{"message", "session.created"},
        },
    }
    
    data, _ := json.Marshal(hello)
    select {
    case client.Send <- data:
    default:
    }
}

// readPump 读取客户端消息
func (s *GatewayServer) readPump(client *Client) {
    defer client.Conn.Close()
    
    for {
        _, message, err := client.Conn.ReadMessage()
        if err != nil {
            break
        }
        
        var req RequestFrame
        if err := json.Unmarshal(message, &req); err != nil {
            logger.WarnCF("gateway", "Invalid message", map[string]interface{}{
                "error": err.Error(),
            })
            continue
        }
        
        // 处理请求
        s.handleRequest(client, &req)
    }
}

// writePump 写入客户端消息
func (s *GatewayServer) writePump(client *Client) {
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

// handleRequest 处理客户端请求
func (s *GatewayServer) handleRequest(client *Client, req *RequestFrame) {
    switch req.Method {
    case "chat.send":
        s.handleChatSend(client, req)
    case "sessions.list":
        s.handleSessionsList(client, req)
    default:
        s.sendError(client, req.ID, "method_not_found", "Unknown method: "+req.Method)
    }
}

// handleChatSend 处理发送消息
func (s *GatewayServer) handleChatSend(client *Client, req *RequestFrame) {
    message, _ := req.Params["message"].(string)
    sessionID, _ := req.Params["sessionId"].(string)
    
    if sessionID == "" {
        sessionID = "session_" + client.UserID + "_" + generateSessionID()
    }
    
    // 构造InboundMessage
    inboundMsg := bus.InboundMessage{
        Channel:    "gateway",
        SenderID:   client.UserID,
        ChatID:     client.ID,
        Content:    message,
        SessionKey: sessionID,
        AgentID:    "default",
    }
    
    // 发送到MessageBus
    ctx := context.Background()
    go func() {
        response, err := s.bus.SendAndWait(ctx, inboundMsg, 30*time.Second)
        if err != nil {
            s.sendError(client, req.ID, "chat_error", err.Error())
            return
        }
        
        // 发送响应
        s.sendResponse(client, req.ID, map[string]interface{}{
            "sessionId": sessionID,
            "response":  response,
        })
        
        // 发送事件（流式显示）
        s.sendEvent(client, "message", map[string]interface{}{
            "sessionId": sessionID,
            "content":   response,
            "role":      "assistant",
        })
    }()
}

// handleSessionsList 处理查询会话列表
func (s *GatewayServer) handleSessionsList(client *Client, req *RequestFrame) {
    // TODO: 从数据库查询
    sessions := []map[string]interface{}{
        {
            "sessionId": "session_123",
            "agentId":   "default",
            "createdAt": time.Now().Unix(),
        },
    }
    
    s.sendResponse(client, req.ID, map[string]interface{}{
        "sessions": sessions,
    })
}

// sendResponse 发送响应
func (s *GatewayServer) sendResponse(client *Client, reqID string, payload interface{}) {
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
    }
}

// sendError 发送错误
func (s *GatewayServer) sendError(client *Client, reqID, code, message string) {
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
func (s *GatewayServer) sendEvent(client *Client, event string, payload interface{}) {
    client.SeqNum++
    evt := EventFrame{
        Type:    "event",
        Event:   event,
        Payload: payload,
        Seq:     client.SeqNum,
    }
    
    data, _ := json.Marshal(evt)
    select {
    case client.Send <- data:
    default:
    }
}

// handleHealth 健康检查
func (s *GatewayServer) handleHealth(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{
        "status": "ok",
    })
}

func generateClientID() string {
    return fmt.Sprintf("client_%d", time.Now().UnixNano())
}

func generateSessionID() string {
    return fmt.Sprintf("%d", time.Now().UnixNano())
}
```

**Step 3: 注册到main.go** (0.5天)

```go
// cmd/pomclaw/main.go
func gatewayCmd() {
    // ... 现有初始化代码
    
    // 创建Gateway服务器
    gatewayServer := gateway.NewGatewayServer(msgBus)
    
    // 启动Gateway（包含UI）
    go func() {
        if err := gatewayServer.Start(cfg.Gateway.Port); err != nil {
            logger.ErrorCF("gateway", "Gateway server error", map[string]interface{}{
                "error": err.Error(),
            })
        }
    }()
    
    fmt.Printf("✓ Gateway UI available at http://%s:%d\n", cfg.Gateway.Host, cfg.Gateway.Port)
    
    // ... 其他启动代码
}
```

**Step 4: 构建UI** (0.5天)

```bash
cd /d/go/pomclaw/pomclaw/ui
npm install
npm run build  # 输出到 ui/dist
```

**Step 5: 修改UI配置**

修改 `ui/src/ui/gateway.ts` 中的连接地址:

```typescript
// 默认连接到本地Gateway
const DEFAULT_GATEWAY_URL = "ws://localhost:8080/ws";
```

---

## 方案B: 创建简化版UI

如果OpenClaw UI太复杂，可以创建简化版：

**使用技术栈**:
- React + TypeScript
- Tailwind CSS
- 直接复用OpenClaw的WebSocket协议

**优势**:
- 更轻量
- 更容易定制
- 学习成本低

---

## 推荐实施

**选择方案A: 直接复用OpenClaw UI**

### 理由
1. **完整功能** - 包含聊天、会话管理、设置、统计
2. **成熟稳定** - OpenClaw已经在生产环境使用
3. **节省时间** - 无需重新开发UI
4. **兼容性好** - 协议设计合理，易于实现

### 工作量估算

| 任务 | 时间 |
|------|------|
| 复制UI代码 | 0.5天 |
| 实现Go Gateway服务器 | 3-4天 |
| 协议对接 | 1天 |
| 测试调试 | 1-2天 |
| **总计** | **5.5-7.5天** |

---

## 配置示例

```yaml
# config/config.yaml
gateway:
  enabled: true
  host: "0.0.0.0"
  port: 8080
  ui_path: "./ui/dist"
  auth:
    enabled: false  # 简化版暂不启用认证
```

---

## 下一步

1. ✅ 确认方案
2. 复制OpenClaw UI到Pomclaw
3. 实现Go版本的Gateway服务器
4. 对接MessageBus
5. 测试前后端通信
6. 部署上线

---

**这个方案完美解决了你的需求：通用、成熟、节省开发时间！**
