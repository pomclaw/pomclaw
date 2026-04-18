# 多租户改造：最简方案

## 核心思路（用户方案）✅

**直接在 AgentLoop 入参中传递 agentID，不要从 session 反查！**

```
API 层（JWT 解析）→ 获取 userID → 直接传给 AgentLoop
                                       ↓
                            AgentLoop 根据 agentID 动态获取 Store
```

---

## 方案对比

### ❌ 之前的复杂方案（不推荐）
```go
// 从 session 反查 agent_id（多一次数据库查询）
agentID := queryAgentIDFromSession(sessionKey)
sessionStore := storeManager.GetStore(agentID)
```

### ✅ 简单方案（推荐）
```go
// API 层直接传 agentID
agentLoop.ProcessDirect(ctx, message, sessionKey, agentID)

// AgentLoop 内部根据 agentID 获取 Store
sessionStore := al.storeManager.GetStore(agentID)
```

---

## 实现方案

### 方案 A：直接加参数（最直接）

#### 1. 修改 AgentLoop 接口

```go
// pkg/agent/loop.go

// 修改前
func (al *AgentLoop) ProcessDirect(ctx context.Context, content, sessionKey string) (string, error)

// 修改后：增加 agentID 参数
func (al *AgentLoop) ProcessDirect(ctx context.Context, content, sessionKey, agentID string) (string, error) {
    return al.ProcessDirectWithChannel(ctx, content, sessionKey, agentID, "cli", "direct")
}

func (al *AgentLoop) ProcessDirectWithChannel(ctx context.Context, content, sessionKey, agentID, channel, chatID string) (string, error) {
    msg := bus.InboundMessage{
        Channel:    channel,
        SenderID:   "cron",
        ChatID:     chatID,
        Content:    content,
        SessionKey: sessionKey,
        AgentID:    agentID,  // ← 新增字段
    }
    
    return al.processMessage(ctx, msg)
}
```

#### 2. 修改 InboundMessage 结构

```go
// pkg/bus/bus.go

type InboundMessage struct {
    Channel    string
    SenderID   string
    ChatID     string
    Content    string
    SessionKey string
    AgentID    string  // ← 新增字段
}
```

#### 3. 修改 AgentLoop 内部处理

```go
// pkg/agent/loop.go

type AgentLoop struct {
    storeManager *storage.StoreManager  // ← 新增
    // 移除单个 store 字段
    // sessions  SessionManagerInterface  // ← 删除
    // state     StateManagerInterface    // ← 删除
    // ...
}

func (al *AgentLoop) processMessage(ctx context.Context, msg bus.InboundMessage) (string, error) {
    // 从消息中获取 agentID
    agentID := msg.AgentID
    if agentID == "" {
        agentID = "default"  // 向后兼容
    }
    
    // 动态获取该 agent 的 Store
    sessionStore := al.storeManager.GetSessionStore(agentID)
    stateStore := al.storeManager.GetStateStore(agentID)
    memoryStore := al.storeManager.GetMemoryStore(agentID)
    
    // 后续处理使用这些 Store
    history := sessionStore.GetHistory(msg.SessionKey)
    // ...
}
```

#### 4. StoreManager 实现

```go
// pkg/storage/store_manager.go (新建)

type StoreManager struct {
    db         *sql.DB
    cfg        *config.Config
    embSvc     EmbeddingService
    
    // Store 实例缓存
    sessionStores sync.Map  // agentID -> *SessionStore
    stateStores   sync.Map
    memoryStores  sync.Map
}

func NewStoreManager(cfg *config.Config, db *sql.DB, embSvc EmbeddingService) *StoreManager {
    return &StoreManager{
        db:     db,
        cfg:    cfg,
        embSvc: embSvc,
    }
}

// 懒加载 + 缓存
func (sm *StoreManager) GetSessionStore(agentID string) *postgres.SessionStore {
    // 尝试从缓存获取
    if store, ok := sm.sessionStores.Load(agentID); ok {
        return store.(*postgres.SessionStore)
    }
    
    // 创建新实例
    store := postgres.NewSessionStore(sm.db, agentID)
    sm.sessionStores.Store(agentID, store)
    
    return store
}

func (sm *StoreManager) GetStateStore(agentID string) *postgres.StateStore {
    if store, ok := sm.stateStores.Load(agentID); ok {
        return store.(*postgres.StateStore)
    }
    
    store := postgres.NewStateStore(sm.db, agentID)
    sm.stateStores.Store(agentID, store)
    
    return store
}

func (sm *StoreManager) GetMemoryStore(agentID string) *postgres.MemoryStore {
    if store, ok := sm.memoryStores.Load(agentID); ok {
        return store.(*postgres.MemoryStore)
    }
    
    store := postgres.NewMemoryStore(sm.db, agentID, sm.embSvc)
    sm.memoryStores.Store(agentID, store)
    
    return store
}
```

#### 5. API 层调用

```go
// pkg/api/handlers/chat.go

func Chat(w http.ResponseWriter, r *http.Request) {
    // 1. JWT 认证（中间件已完成）
    claims := r.Context().Value("user").(*auth.Claims)
    userID := claims.UserID  // ← 从 JWT 获取 user_id
    
    // 2. 解析请求
    var req ChatRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 3. 直接调用 AgentLoop，传入 agentID
    ctx := context.Background()
    response, err := agentLoop.ProcessDirect(
        ctx,
        req.Message,
        req.SessionKey,
        userID,  // ← 直接传 agentID！
    )
    
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    
    // 4. 返回结果
    json.NewEncoder(w).Encode(ChatResponse{
        Response: response,
    })
}
```

#### 6. 初始化改造

```go
// cmd/pomclaw/main.go

func initDatabaseAgent(cfg *config.Config, msgBus *bus.MessageBus, provider providers.LLMProvider) (*agent.AgentLoop, storage.ConnectionManager, error) {
    // ...连接数据库...
    
    // 创建 StoreManager（替代单个 Store）
    storeManager := storage.NewStoreManager(cfg, db, embSvc)
    
    // 创建 AgentLoop（传入 StoreManager）
    agentLoop := agent.NewAgentLoopWithStoreManager(
        cfg,
        msgBus,
        provider,
        storeManager,  // ← 传 StoreManager 而不是单个 Store
    )
    
    return agentLoop, conn, nil
}
```

---

## 优势

### ✅ 简单直接
- API 层从 JWT 获取 userID
- 直接作为参数传给 AgentLoop
- 无需反查数据库

### ✅ 调用清晰
```go
// 调用者明确知道要用哪个 agent
agentLoop.ProcessDirect(ctx, message, sessionKey, "user_123")
```

### ✅ 无性能损耗
- 不需要从 session 反查 agent_id
- Store 实例懒加载 + 缓存

### ✅ 向后兼容
```go
// 如果 agentID 为空，降级到 "default"
if agentID == "" {
    agentID = "default"
}
```

---

## 调用流程

```
┌─────────────┐
│  前端请求   │
│  JWT Token  │
└──────┬──────┘
       │
       ▼
┌─────────────────────┐
│  JWT 中间件         │
│  解析出 userID      │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│  API Handler                            │
│  agentLoop.ProcessDirect(              │
│      ctx,                               │
│      "你好",                            │
│      "chat_123",                        │
│      "user_abc"  ← 直接传 agentID      │
│  )                                      │
└──────┬──────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│  AgentLoop.processMessage               │
│  agentID = msg.AgentID  ← 从消息获取    │
│                                         │
│  sessionStore = storeManager.Get(agentID)│
│  stateStore = storeManager.Get(agentID)  │
│  memoryStore = storeManager.Get(agentID) │
└──────┬──────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│  StoreManager                           │
│  GetSessionStore(agentID) {             │
│      if cached { return cached }        │
│      store = NewSessionStore(db, agentID)│
│      cache.Store(agentID, store)        │
│      return store                       │
│  }                                      │
└──────┬──────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│  SessionStore (agentID 绑定)            │
│  db.Query(...                           │
│      WHERE agent_id = $1, agentID)      │
└─────────────────────────────────────────┘
```

---

## 改动清单

### 核心改动（必须）

1. **pkg/bus/bus.go**
   - `InboundMessage` 增加 `AgentID string` 字段

2. **pkg/agent/loop.go**
   - `ProcessDirect` 增加 `agentID` 参数
   - `ProcessDirectWithChannel` 增加 `agentID` 参数
   - `AgentLoop` 结构体：移除单个 Store，增加 `StoreManager`
   - `processMessage` 方法：从 `msg.AgentID` 获取 agentID，动态获取 Store

3. **pkg/storage/store_manager.go**（新建）
   - 实现 `StoreManager` 管理多个 agentID 的 Store 实例

4. **cmd/pomclaw/main.go**
   - `initDatabaseAgent` 创建 `StoreManager` 替代单个 Store

5. **所有调用 ProcessDirect 的地方**
   - 增加 `agentID` 参数
   - CLI: 默认传 `"default"`
   - API: 从 JWT 传 `userID`

### 向后兼容

```go
// CLI 调用（向后兼容）
agentLoop.ProcessDirect(ctx, message, sessionKey, "default")

// API 调用（多租户）
agentLoop.ProcessDirect(ctx, message, sessionKey, userID)
```

---

## 工作量估算

| 任务 | 工作量 |
|------|--------|
| 修改 AgentLoop 接口 | 0.5天 |
| 实现 StoreManager | 1天 |
| 改造 AgentLoop 内部逻辑 | 1天 |
| 修改所有调用点（CLI/Gateway/API） | 0.5天 |
| 测试 | 0.5天 |
| **总计** | **3.5天** |

---

## 总结

### 用户方案的优势 ⭐

- **最简单**：直接传参，无需反查
- **最清晰**：调用者明确知道用哪个 agent
- **最高效**：无额外数据库查询
- **最灵活**：支持 CLI、API、Gateway 等所有场景

### 核心改动

```go
// 1. 接口增加参数
ProcessDirect(ctx, message, sessionKey, agentID string)

// 2. StoreManager 管理多 agent 的 Store
storeManager.GetSessionStore(agentID)

// 3. API 层直接传 userID
agentLoop.ProcessDirect(ctx, msg, sessionKey, claims.UserID)
```

**这是最佳方案！改造工作量：3.5天**
