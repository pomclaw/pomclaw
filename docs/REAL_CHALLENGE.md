# 多租户改造：真正的难点分析

## 用户洞察 ✅

**关键观点：不需要每个用户一个 AgentLoop！**

- AgentLoop 本身就是消息处理循环，可以处理多个用户的请求
- 数据库层已经通过 `agent_id` 字段做了隔离
- **真正的难点：Store 层的 agent_id 是构造时固定的，无法动态切换**

---

## 问题根源

### 当前架构

```go
// pkg/postgres/session_store.go:24-32
type SessionStore struct {
    db       *sql.DB
    agentID  string  // ← 构造时固定！
    sessions map[string]*PostgresSession
    mu       sync.RWMutex
}

func NewSessionStore(db *sql.DB, agentID string) *SessionStore {
    return &SessionStore{
        db:      db,
        agentID: agentID,  // ← 从配置读取，启动时固定
    }
}
```

### 数据库操作使用固定 agentID

```go
// pkg/postgres/session_store.go:158-163
func (ss *SessionStore) Save(key string) error {
    _, err = ss.db.Exec(`
        INSERT INTO POM_SESSIONS (session_key, agent_id, messages, summary, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (session_key) DO UPDATE
        SET messages = $3, summary = $4, updated_at = $6
    `, key, ss.agentID, ...)  // ← 使用构造时的固定值
}
```

### 初始化流程

```go
// cmd/pomclaw/main.go:1687-1693
agentID := storage.GetAgentID(cfg)  // ← 从配置读："default"

sessionStore := storage.NewSessionStore(cfg, db)   // agentID = "default"
stateStore := storage.NewStateStore(cfg, db)       // agentID = "default"
memoryStore := storage.NewMemoryStore(cfg, db, embSvc)  // agentID = "default"

agentLoop := agent.NewAgentLoopWithStores(cfg, msgBus, provider, 
                                         sessionStore, stateStore, memoryStore)
```

**结果：所有用户的数据都写入到 `agent_id = 'default'`**

---

## 改造方案对比

### 方案 1：修改 Store 接口（最彻底但影响最大）

**改动点：所有 Store 方法增加 agentID 参数**

```go
// 现在
type SessionManagerInterface interface {
    AddMessage(key, role, content string)
    GetHistory(key string) []providers.Message
    Save(key string) error
}

// 改为
type SessionManagerInterface interface {
    AddMessage(agentID, key, role, content string)
    GetHistory(agentID, key string) []providers.Message
    Save(agentID, key string) error
}
```

**优点：**
- ✅ 最彻底，完全支持动态 agent_id
- ✅ 线程安全，无状态 Store

**缺点：**
- ❌ 影响面巨大：所有调用 Store 的地方都要改
- ❌ AgentLoop 内部几十处调用都要修改
- ❌ 需要在整个调用链传递 agentID

**工作量：** 7-10 天

---

### 方案 2：Store 工厂模式（推荐）⭐

**改动点：Store 不直接使用，通过工厂获取**

```go
// pkg/storage/store_manager.go (新建)
type StoreManager struct {
    db         *sql.DB
    cfg        *config.Config
    embSvc     EmbeddingService
    
    // 实例缓存（带过期淘汰）
    sessionStores map[string]*SessionStore  // agentID -> Store
    stateStores   map[string]*StateStore
    memoryStores  map[string]*MemoryStore
    
    mu sync.RWMutex
}

func NewStoreManager(cfg *config.Config, db *sql.DB, embSvc EmbeddingService) *StoreManager {
    return &StoreManager{
        db:            db,
        cfg:           cfg,
        embSvc:        embSvc,
        sessionStores: make(map[string]*SessionStore),
        stateStores:   make(map[string]*StateStore),
        memoryStores:  make(map[string]*MemoryStore),
    }
}

// 懒加载：首次请求时创建，后续复用
func (sm *StoreManager) GetSessionStore(agentID string) *SessionStore {
    sm.mu.RLock()
    store, exists := sm.sessionStores[agentID]
    sm.mu.RUnlock()
    
    if exists {
        return store
    }
    
    // 创建新实例
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    // Double check
    if store, exists := sm.sessionStores[agentID]; exists {
        return store
    }
    
    store = postgres.NewSessionStore(sm.db, agentID)
    sm.sessionStores[agentID] = store
    return store
}

func (sm *StoreManager) GetStateStore(agentID string) *StateStore { ... }
func (sm *StoreManager) GetMemoryStore(agentID string) *MemoryStore { ... }
```

**AgentLoop 改造：接收 StoreManager 而不是具体 Store**

```go
// pkg/agent/loop.go
type AgentLoop struct {
    storeManager  *storage.StoreManager  // ← 新增
    // 移除单个 store 字段
    // sessions  SessionManagerInterface  // ← 删除
    // state     StateManagerInterface    // ← 删除
    // ...
}

// 处理消息时动态获取 Store
func (al *AgentLoop) processMessage(ctx context.Context, msg bus.InboundMessage) (string, error) {
    // 从消息中提取 userID/agentID（通过 JWT token 或其他方式）
    agentID := al.extractAgentID(msg)
    
    // 动态获取该用户的 Store
    sessionStore := al.storeManager.GetSessionStore(agentID)
    stateStore := al.storeManager.GetStateStore(agentID)
    memoryStore := al.storeManager.GetMemoryStore(agentID)
    
    // 后续处理使用该 Store
    history := sessionStore.GetHistory(msg.SessionKey)
    // ...
}
```

**优点：**
- ✅ AgentLoop 单例即可，无需多实例
- ✅ Store 实例按需创建，自动隔离
- ✅ 线程安全，每个 agentID 有独立 Store
- ✅ 可以加入 LRU 淘汰机制，控制内存

**缺点：**
- ⚠️ 需要改造 AgentLoop 内部逻辑（但范围可控）
- ⚠️ 需要在消息中携带 agentID 信息

**工作量：** 3-4 天

---

### 方案 3：Session Key 编码 agentID（最小改动）

**思路：在 session_key 中编码 agent_id**

```go
// 现在
sessionKey = "chat_123"

// 改为
sessionKey = "agent:user_abc123:chat_123"
//            ^^^^^ ^^^^^^^^^^^ ^^^^^^^^
//            前缀   agentID     原始key
```

**Store 层改造：从 key 解析 agentID**

```go
// pkg/postgres/session_store.go
type SessionStore struct {
    db       *sql.DB
    // agentID  string  // ← 删除固定字段
    sessions map[string]*PostgresSession
    mu       sync.RWMutex
}

func (ss *SessionStore) parseAgentID(key string) string {
    // "agent:user_123:chat_456" -> "user_123"
    parts := strings.Split(key, ":")
    if len(parts) >= 3 && parts[0] == "agent" {
        return parts[1]
    }
    return "default"  // 向后兼容
}

func (ss *SessionStore) Save(key string) error {
    agentID := ss.parseAgentID(key)  // ← 动态解析
    
    _, err = ss.db.Exec(`
        INSERT INTO POM_SESSIONS (session_key, agent_id, messages, ...)
        VALUES ($1, $2, $3, ...)
    `, key, agentID, ...)  // ← 使用解析出的 agentID
}
```

**API 层构造 session_key**

```go
// pkg/api/handlers/chat.go
func Chat(w http.ResponseWriter, r *http.Request) {
    claims := r.Context().Value("user").(*auth.Claims)
    userID := claims.UserID
    
    // 构造编码了 agentID 的 session_key
    sessionKey := fmt.Sprintf("agent:%s:%s", userID, req.SessionID)
    
    // 调用 AgentLoop 处理（AgentLoop 无需改动）
    response, _ := agentLoop.ProcessDirect(ctx, req.Message, sessionKey)
}
```

**优点：**
- ✅ 改动最小，AgentLoop 无需改动
- ✅ 无需多实例，无内存开销
- ✅ Store 接口不需要改

**缺点：**
- ❌ 不优雅，session_key 语义混淆
- ❌ 需要修改所有 Store 的数据库操作逻辑
- ❌ 需要处理向后兼容（旧的 key 格式）

**工作量：** 2-3 天

---

### 方案 4：Context 传递 agentID（中等改动）

**思路：在 context 中传递 agentID**

```go
// pkg/agent/loop.go
func (al *AgentLoop) ProcessDirect(ctx context.Context, content, sessionKey string) (string, error) {
    // 从 context 获取 agentID
    agentID, ok := ctx.Value("agent_id").(string)
    if !ok {
        agentID = "default"
    }
    
    // 动态获取 Store
    sessionStore := al.storeManager.GetSessionStore(agentID)
    // ...
}
```

**API 层注入 agentID**

```go
// pkg/api/handlers/chat.go
func Chat(w http.ResponseWriter, r *http.Request) {
    claims := r.Context().Value("user").(*auth.Claims)
    
    // 注入 agent_id 到 context
    ctx := context.WithValue(r.Context(), "agent_id", claims.UserID)
    
    response, _ := agentLoop.ProcessDirect(ctx, req.Message, sessionKey)
}
```

**优点：**
- ✅ 符合 Go 惯例（context 传递请求级数据）
- ✅ AgentLoop 改动可控
- ✅ 类型安全（可以用自定义 context key）

**缺点：**
- ⚠️ 需要在整个调用链传递 context
- ⚠️ Store 层仍需改造（方案2）

**工作量：** 3-4 天

---

## 推荐方案

### 最佳实践：方案 2 (StoreManager) + 方案 4 (Context 传递)

**理由：**
1. ✅ 架构清晰：StoreManager 管理多租户 Store
2. ✅ 符合 Go 惯例：通过 context 传递请求级数据
3. ✅ 性能可控：Store 实例复用 + LRU 淘汰
4. ✅ 单 AgentLoop：无需多实例，用户说的对

**核心改动点：**

1. **新增 StoreManager** (pkg/storage/store_manager.go)
   - 管理多个 agentID 对应的 Store 实例
   - 懒加载 + 实例缓存

2. **改造 AgentLoop** (pkg/agent/loop.go)
   - 接收 StoreManager 替代单个 Store
   - 从 context 获取 agentID
   - 动态获取对应的 Store

3. **API 层** (pkg/api/handlers/*.go)
   - JWT 认证获取 userID
   - 注入 agent_id 到 context
   - 调用 AgentLoop

**工作量：** 3-4 天（后端）

---

## 与前端的配合

**前端无需关心 agent_id 细节**

```typescript
// src/services/api.ts
export async function sendMessage(message: string, sessionId: string) {
  const response = await fetch('/api/chat', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,  // ← JWT token 包含 userID
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      message,
      session_id: sessionId,  // ← 只需传普通的 sessionId
    }),
  });
  return response.json();
}
```

**后端自动从 JWT 提取 agentID：**

```go
// pkg/api/handlers/chat.go
func Chat(w http.ResponseWriter, r *http.Request) {
    // 1. 中间件已解析 JWT，存入 context
    claims := r.Context().Value("user").(*auth.Claims)
    
    // 2. 注入 agent_id
    ctx := context.WithValue(r.Context(), "agent_id", claims.UserID)
    
    // 3. AgentLoop 自动从 context 获取
    response, _ := agentLoop.ProcessDirect(ctx, req.Message, req.SessionID)
}
```

---

## 总结

### 用户的洞察完全正确 ✅

- **不需要多 AgentLoop 实例**
- **真正的难点是 Store 层的动态 agent_id**

### 推荐方案

**StoreManager (工厂模式) + Context 传递**

- 改动可控：3-4 天
- 架构清晰：符合 Go 最佳实践
- 性能优秀：实例复用 + 懒加载

### 核心改造

```
1. 新增 StoreManager               - 1天
2. 改造 AgentLoop                   - 1-2天
3. 添加 JWT + API 层                - 2-3天（如果要前端再加 7天）
```

**总工作量（仅后端）：4-6 天**
