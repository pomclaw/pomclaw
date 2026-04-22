# Gateway API 设计文档

**Date**: 2026-04-22
**Version**: v1.0
**Status**: Draft
**Scope**: `pkg/channels/gateway/` HTTP REST + WebSocket API

---

## 一、背景与目标

Pomclaw 是多租户 AI Agent 平台。每个用户可以：

1. **注册 / 登录**，获得独立身份
2. **创建 Agent**，定制 AI 角色（Prompt、模型、工具）
3. **创建 Session**，把 User ↔ Agent 绑定为一次对话上下文
4. **实时对话**，通过 WebSocket（Pico 协议）与 Agent 互动

Gateway 同时承担两个职责：
- **REST API Server**（JSON over HTTP）：用户、Agent、Session 的 CRUD
- **WebSocket Server**（Pico 协议）：实时消息收发

---

## 二、Domain 模型

```
User (1) ──has─many──> Agent (N)
Agent (1) ──has─many──> Session (N)
User  (1) ──has─many──> Session (N)
Session (1) ──has─many──> Message (N)
```

### 2.1 User

```json
{
  "id":         "uuid",
  "username":   "alice",
  "email":      "alice@example.com",
  "status":     "active | suspended",
  "created_at": "2026-04-22T08:00:00Z",
  "updated_at": "2026-04-22T08:00:00Z"
}
```

### 2.2 Agent

```json
{
  "id":            "uuid",
  "user_id":       "uuid",
  "name":          "My Assistant",
  "description":   "用于日常问答的助手",
  "system_prompt": "You are a helpful assistant...",
  "model":         "claude-sonnet-4-6",
  "tools":         ["web_search", "code_exec"],
  "status":        "active | inactive",
  "created_at":    "2026-04-22T08:00:00Z",
  "updated_at":    "2026-04-22T08:00:00Z"
}
```

### 2.3 Session

```json
{
  "id":            "uuid",
  "user_id":       "uuid",
  "agent_id":      "uuid",
  "title":         "关于 Go 并发的讨论",
  "status":        "active | archived",
  "message_count": 42,
  "created_at":    "2026-04-22T08:00:00Z",
  "updated_at":    "2026-04-22T08:10:00Z"
}
```

### 2.4 Message

```json
{
  "id":         "uuid",
  "session_id": "uuid",
  "role":       "user | assistant",
  "content":    "string",
  "created_at": "2026-04-22T08:05:00Z"
}
```

---

## 三、API 总览

> 共 **20 个接口**（含 1 个 WebSocket）

| 序号 | 分组     | 方法      | Path                                   | 说明                   | Auth  |
|------|----------|-----------|----------------------------------------|------------------------|-------|
| 1    | Auth     | POST      | `/api/v1/auth/register`                | 用户注册               | 否    |
| 2    | Auth     | POST      | `/api/v1/auth/login`                   | 用户登录，返回 JWT     | 否    |
| 3    | Auth     | POST      | `/api/v1/auth/logout`                  | 注销（服务端失效 token）| 是    |
| 4    | Auth     | POST      | `/api/v1/auth/refresh`                 | 刷新 JWT               | 是*   |
| 5    | Auth     | GET       | `/api/v1/auth/me`                      | 获取当前用户信息       | 是    |
| 6    | Agents   | GET       | `/api/v1/agents`                       | 列出当前用户的 Agent   | 是    |
| 7    | Agents   | POST      | `/api/v1/agents`                       | 创建 Agent             | 是    |
| 8    | Agents   | GET       | `/api/v1/agents/{agent_id}`            | 获取 Agent 详情        | 是    |
| 9    | Agents   | PUT       | `/api/v1/agents/{agent_id}`            | 更新 Agent 配置        | 是    |
| 10   | Agents   | DELETE    | `/api/v1/agents/{agent_id}`            | 删除 Agent             | 是    |
| 11   | Sessions | GET       | `/api/v1/sessions`                     | 列出当前用户的 Session | 是    |
| 12   | Sessions | POST      | `/api/v1/sessions`                     | 创建 Session（指定 Agent）| 是  |
| 13   | Sessions | GET       | `/api/v1/sessions/{session_id}`        | 获取 Session 详情      | 是    |
| 14   | Sessions | PATCH     | `/api/v1/sessions/{session_id}`        | 更新 Session（改标题等）| 是   |
| 15   | Sessions | DELETE    | `/api/v1/sessions/{session_id}`        | 删除 Session 及历史    | 是    |
| 16   | Messages | GET       | `/api/v1/sessions/{session_id}/messages`| 获取消息历史          | 是    |
| 17   | Messages | DELETE    | `/api/v1/sessions/{session_id}/messages`| 清空消息历史          | 是    |
| 18   | Chat     | WebSocket | `/ws/{session_id}`                     | Pico 协议实时对话      | 是†   |
| 19   | System   | GET       | `/api/v1/system/health`                | 健康检查               | 否    |
| 20   | System   | GET       | `/api/v1/system/models`                | 可用模型列表           | 是    |

> `*`：`/auth/refresh` 使用 Refresh Token（不用 Access Token）
> `†`：WebSocket 连接时通过 query param `?token=<jwt>` 传递认证

---

## 四、接口详细设计

### 4.1 Auth 模块

#### POST `/api/v1/auth/register`

注册新用户。

**Request Body**:
```json
{
  "username": "alice",
  "email":    "alice@example.com",
  "password": "P@ssw0rd!"
}
```

**Response 201**:
```json
{
  "user": {
    "id":       "550e8400-e29b-41d4-a716-446655440000",
    "username": "alice",
    "email":    "alice@example.com"
  },
  "access_token":  "<jwt>",
  "refresh_token": "<refresh_jwt>",
  "expires_in":    3600
}
```

**Errors**: `400` 参数校验失败 | `409` 用户名/邮箱已存在

---

#### POST `/api/v1/auth/login`

**Request Body**:
```json
{
  "username": "alice",
  "password": "P@ssw0rd!"
}
```

**Response 200**:
```json
{
  "user": { "id": "...", "username": "alice" },
  "access_token":  "<jwt>",
  "refresh_token": "<refresh_jwt>",
  "expires_in":    3600
}
```

**Errors**: `401` 用户名或密码错误

---

#### POST `/api/v1/auth/logout`

服务端将当前 access_token 加入黑名单（或清除 refresh token）。

**Request Headers**: `Authorization: Bearer <token>`

**Response 204**: No Content

---

#### POST `/api/v1/auth/refresh`

用 Refresh Token 换新的 Access Token。

**Request Body**:
```json
{ "refresh_token": "<refresh_jwt>" }
```

**Response 200**:
```json
{
  "access_token": "<new_jwt>",
  "expires_in":   3600
}
```

**Errors**: `401` Refresh Token 无效或已过期

---

#### GET `/api/v1/auth/me`

**Response 200**:
```json
{
  "id":         "550e...",
  "username":   "alice",
  "email":      "alice@example.com",
  "status":     "active",
  "created_at": "2026-04-22T08:00:00Z"
}
```

---

### 4.2 Agents 模块

> 所有 Agent 接口均对当前用户作用域隔离，不能访问他人的 Agent。

#### GET `/api/v1/agents`

**Query Params**:
- `page` (default: 1)
- `page_size` (default: 20, max: 100)
- `status` (optional: `active | inactive`)

**Response 200**:
```json
{
  "total": 3,
  "page":  1,
  "items": [
    {
      "id":          "...",
      "name":        "My Assistant",
      "description": "用于日常问答",
      "model":       "claude-sonnet-4-6",
      "status":      "active",
      "created_at":  "2026-04-22T08:00:00Z"
    }
  ]
}
```

---

#### POST `/api/v1/agents`

**Request Body**:
```json
{
  "name":          "My Assistant",
  "description":   "用于日常问答的助手",
  "system_prompt": "You are a helpful assistant...",
  "model":         "claude-sonnet-4-6",
  "tools":         ["web_search"]
}
```

**Response 201**:
```json
{
  "id":            "...",
  "user_id":       "...",
  "name":          "My Assistant",
  "description":   "用于日常问答的助手",
  "system_prompt": "You are a helpful assistant...",
  "model":         "claude-sonnet-4-6",
  "tools":         ["web_search"],
  "status":        "active",
  "created_at":    "2026-04-22T08:00:00Z",
  "updated_at":    "2026-04-22T08:00:00Z"
}
```

**Errors**: `400` 参数校验失败 | `422` 模型不可用

---

#### GET `/api/v1/agents/{agent_id}`

**Response 200**: 返回 Agent 完整结构（含 system_prompt）

**Errors**: `404` Agent 不存在 | `403` 无权限

---

#### PUT `/api/v1/agents/{agent_id}`

全量更新（与 POST body 结构相同，所有字段均可修改）。

**Response 200**: 返回更新后的 Agent

---

#### DELETE `/api/v1/agents/{agent_id}`

同时会使该 Agent 下的所有 active Session 变为 `archived`。

**Response 204**: No Content

**Errors**: `404` | `403`

---

### 4.3 Sessions 模块

#### GET `/api/v1/sessions`

**Query Params**:
- `agent_id` (optional, 按 Agent 过滤)
- `status` (optional: `active | archived`)
- `page`, `page_size`

**Response 200**:
```json
{
  "total": 10,
  "page":  1,
  "items": [
    {
      "id":            "...",
      "agent_id":      "...",
      "agent_name":    "My Assistant",
      "title":         "关于 Go 并发的讨论",
      "status":        "active",
      "message_count": 42,
      "updated_at":    "2026-04-22T08:10:00Z"
    }
  ]
}
```

---

#### POST `/api/v1/sessions`

创建新 Session，将用户与指定 Agent 绑定。这是开始对话的必要步骤。

**Request Body**:
```json
{
  "agent_id": "550e8400-...",
  "title":    "新对话"
}
```

> `title` 可选，不填则由服务器生成默认标题（如"新对话 #1"）。

**Response 201**:
```json
{
  "id":            "...",
  "user_id":       "...",
  "agent_id":      "...",
  "agent_name":    "My Assistant",
  "title":         "新对话",
  "status":        "active",
  "message_count": 0,
  "created_at":    "2026-04-22T09:00:00Z",
  "updated_at":    "2026-04-22T09:00:00Z"
}
```

**Errors**: `404` Agent 不存在 | `403` Agent 不属于当前用户

---

#### GET `/api/v1/sessions/{session_id}`

**Response 200**: 返回 Session 完整信息

---

#### PATCH `/api/v1/sessions/{session_id}`

部分更新（目前支持修改 title 和 status）。

**Request Body**:
```json
{
  "title":  "重命名后的标题",
  "status": "archived"
}
```

**Response 200**: 返回更新后的 Session

---

#### DELETE `/api/v1/sessions/{session_id}`

删除 Session 及其所有消息历史。

**Response 204**: No Content

---

### 4.4 Messages 模块

#### GET `/api/v1/sessions/{session_id}/messages`

获取该 Session 的消息历史，支持分页（从新到旧排序）。

**Query Params**:
- `page` (default: 1)
- `page_size` (default: 50, max: 200)
- `before` (optional: cursor，消息 ID，获取此 ID 之前的消息)

**Response 200**:
```json
{
  "total": 42,
  "page":  1,
  "items": [
    {
      "id":         "...",
      "session_id": "...",
      "role":       "user",
      "content":    "Go 中如何优雅地处理并发？",
      "created_at": "2026-04-22T08:05:00Z"
    },
    {
      "id":         "...",
      "session_id": "...",
      "role":       "assistant",
      "content":    "在 Go 中处理并发主要有以下几种方式...",
      "created_at": "2026-04-22T08:05:02Z"
    }
  ]
}
```

---

#### DELETE `/api/v1/sessions/{session_id}/messages`

清空当前 Session 的所有消息历史（保留 Session 本身）。

**Response 204**: No Content

---

### 4.5 Chat 模块（WebSocket）

#### WebSocket `/ws/{session_id}`

实时双向对话，基于 Pico 协议（已有实现）。

**连接认证**:
```
GET /ws/550e8400-...?token=<jwt_access_token>
Upgrade: websocket
```

**Gateway 在 Upgrade 前验证**：
1. JWT 有效性
2. session_id 对应的 Session 存在且属于该用户
3. Session 关联的 Agent 处于 active 状态

**消息格式**（沿用现有 Pico 协议，扩展字段）:

客户端发送消息：
```json
{
  "type":       "message.send",
  "id":         "msg-uuid",
  "session_id": "session-uuid",
  "timestamp":  1745321234567,
  "payload": {
    "content": "你好，帮我写一个 Go 并发示例"
  }
}
```

服务端推送消息（流式）：
```json
{
  "type":       "message.create",
  "id":         "msg-uuid",
  "session_id": "session-uuid",
  "timestamp":  1745321234600,
  "payload": {
    "role":    "assistant",
    "content": "好的，以下是一个 Go 并发示例...",
    "done":    false
  }
}
```

流结束：
```json
{
  "type":       "message.create",
  "id":         "msg-uuid",
  "session_id": "session-uuid",
  "timestamp":  1745321235000,
  "payload": {
    "role":    "assistant",
    "content": "",
    "done":    true
  }
}
```

**错误推送**:
```json
{
  "type":    "error",
  "id":      "msg-uuid",
  "payload": {
    "code":    "agent_not_running",
    "message": "Agent is not active"
  }
}
```

---

### 4.6 System 模块

#### GET `/api/v1/system/health`

```json
{
  "status":   "ok",
  "version":  "0.1.0",
  "uptime_s": 3600
}
```

---

#### GET `/api/v1/system/models`

返回系统支持的模型列表，供 Agent 创建时使用。

**Response 200**:
```json
{
  "models": [
    {
      "id":          "claude-sonnet-4-6",
      "name":        "Claude Sonnet 4.6",
      "description": "平衡性能与成本的旗舰模型",
      "context_len": 200000
    },
    {
      "id":          "claude-haiku-4-5",
      "name":        "Claude Haiku 4.5",
      "description": "高速低成本模型",
      "context_len": 200000
    }
  ]
}
```

---

## 五、认证方案

### JWT 双 Token 机制

| Token         | 有效期   | 存储位置            |
|---------------|----------|---------------------|
| Access Token  | 1 小时   | Memory（前端）      |
| Refresh Token | 7 天     | HttpOnly Cookie     |

**Header 格式**:
```
Authorization: Bearer <access_token>
```

**Flow**:
```
Register/Login
    ↓
返回 access_token + refresh_token
    ↓
每个 API 请求携带 access_token
    ↓
access_token 过期 → POST /auth/refresh → 新 access_token
    ↓
refresh_token 过期 → 重新登录
```

---

## 六、完整对话生命周期

```
1. POST /api/v1/auth/register          → 用户注册
2. POST /api/v1/auth/login             → 获得 JWT
3. POST /api/v1/agents                 → 创建 Agent（定制 Prompt / Model）
4. POST /api/v1/sessions               → 创建 Session（指定 agent_id）
5. WebSocket /ws/{session_id}?token=.. → 建立连接
6. message.send                        → 发消息
7. message.create (stream)             → 收流式回复
8. ... 多轮对话 ...
9. PATCH /api/v1/sessions/{id}         → 归档 Session
10. DELETE /api/v1/sessions/{id}       → 删除历史
```

---

## 七、错误码规范

所有错误统一格式：

```json
{
  "error": {
    "code":    "session_not_found",
    "message": "Session 550e... not found",
    "details": {}
  }
}
```

| HTTP 状态 | code                   | 含义                    |
|-----------|------------------------|-------------------------|
| 400       | `invalid_request`      | 请求参数不合法          |
| 401       | `unauthorized`         | 未登录 / Token 无效     |
| 403       | `forbidden`            | 无权限访问该资源        |
| 404       | `not_found`            | 资源不存在              |
| 409       | `conflict`             | 资源已存在（注册冲突）  |
| 422       | `model_unavailable`    | 指定模型不可用          |
| 500       | `internal_error`       | 服务器内部错误          |

---

## 八、实现建议

### 8.1 文件结构（在 `pkg/channels/gateway/` 中新增）

```
pkg/channels/gateway/
├── pico.go                  # 已有：WebSocket 服务器（需扩展认证）
├── pico_protocol.go         # 已有：Pico 消息类型
├── router.go                # NEW：HTTP 路由注册（chi 或 net/http）
├── middleware.go            # NEW：JWT 中间件、限流、CORS
├── handlers/
│   ├── auth.go              # NEW：注册、登录、刷新
│   ├── agents.go            # NEW：Agent CRUD
│   ├── sessions.go          # NEW：Session CRUD
│   ├── messages.go          # NEW：消息历史
│   └── system.go            # NEW：健康检查、模型列表
└── models/
    ├── user.go              # NEW：User DB model
    ├── agent.go             # NEW：Agent DB model
    └── session.go           # NEW：Session DB model（扩展现有）
```

### 8.2 Session Key 映射

创建 Session 后，生成 session_key 格式：`gateway:{session_id}`

与现有 BaseChannel 的 `channel:chatID` 约定保持一致，这样 AgentLoop 无需修改即可复用。

### 8.3 Multi-Tenant 隔离

所有查询均加 `WHERE user_id = ?` 条件：
- Agent 只属于创建它的用户
- Session 只属于创建它的用户
- WebSocket 连接时校验 session_id 归属

### 8.4 与现有 AgentLoop 集成

WebSocket 连接建立时，需将 `agent_id` 注入到 InboundMessage（对应 Phase 1 of Multi-Tenant refactoring）：

```go
// pico.go handleMessageSend 中
msg := bus.InboundMessage{
    Channel:    "gateway",
    SenderID:   userID,       // 从 JWT 解析
    ChatID:     sessionID,
    SessionKey: "gateway:" + sessionID,
    AgentID:    agentID,      // 从 Session 记录中读取
    Content:    content,
}
```

---

## 九、接口数量汇总

| 分组     | 数量 |
|----------|------|
| Auth     | 5    |
| Agents   | 5    |
| Sessions | 5    |
| Messages | 2    |
| Chat WS  | 1    |
| System   | 2    |
| **Total**| **20** |

---

*End of Document*
