# Session-Agent 绑定方案分析

## 用户的洞察 💡

**关键观点：session_key 在数据库中已经和 agent_id 绑定了！**

```sql
-- POM_SESSIONS 表结构
CREATE TABLE POM_SESSIONS (
    session_key VARCHAR(255) PRIMARY KEY,  -- ← session 唯一标识
    agent_id    VARCHAR(64) NOT NULL,      -- ← 已经绑定了 agent_id
    messages    TEXT,
    ...
);

-- 数据示例
session_key     | agent_id
----------------|----------
"chat_001"      | "user_A"    ← session 绑定到 user_A
"chat_002"      | "user_B"    ← session 绑定到 user_B
```

**思路：既然有这个绑定关系，能不能从 session_key 反向查询 agent_id？**

---

## 方案实现

### 核心思路

```go
// pkg/storage/store_manager.go
type StoreManager struct {
    db         *sql.DB
    
    // 缓存 session_key -> agent_id 的映射
    sessionAgentMap map[string]string
    mu sync.RWMutex
}

// 从 session_key 查询对应的 agent_id
func (sm *StoreManager) GetAgentIDBySession(sessionKey string) (string, error) {
    // 1. 先查缓存
    sm.mu.RLock()
    if agentID, exists := sm.sessionAgentMap[sessionKey]; exists {
        sm.mu.RUnlock()
        return agentID, nil
    }
    sm.mu.RUnlock()
    
    // 2. 查数据库
    var agentID string
    err := sm.db.QueryRow(`
        SELECT agent_id 
        FROM POM_SESSIONS 
        WHERE session_key = $1
    `, sessionKey).Scan(&agentID)
    
    if err == sql.ErrNoRows {
        // session 不存在，返回空（需要创建新 session）
        return "", nil
    }
    if err != nil {
        return "", err
    }
    
    // 3. 缓存结果
    sm.mu.Lock()
    sm.sessionAgentMap[sessionKey] = agentID
    sm.mu.Unlock()
    
    return agentID, nil
}

// 获取 SessionStore（自动确定 agent_id）
func (sm *StoreManager) GetSessionStore(sessionKey string) (*SessionStore, error) {
    // 从 session_key 查询 agent_id
    agentID, err := sm.GetAgentIDBySession(sessionKey)
    if err != nil {
        return nil, err
    }
    
    if agentID == "" {
        // 新 session，需要外部指定 agent_id
        return nil, fmt.Errorf("session not found, agent_id required")
    }
    
    // 根据 agent_id 获取对应的 Store
    return sm.getOrCreateStore(agentID), nil
}
```

### AgentLoop 使用

```go
// pkg/agent/loop.go
func (al *AgentLoop) runAgentLoop(ctx context.Context, opts processOptions) (string, error) {
    // 尝试从 session_key 自动查询 agent_id
    agentID, err := al.storeManager.GetAgentIDBySession(opts.SessionKey)
    if err != nil {
        return "", err
    }
    
    if agentID == "" {
        // 新 session，从 context 获取 agent_id
        agentID = al.extractAgentIDFromContext(ctx)
    }
    
    // 获取对应的 Store
    sessionStore := al.storeManager.GetSessionStoreByAgentID(agentID)
    stateStore := al.storeManager.GetStateStoreByAgentID(agentID)
    memoryStore := al.storeManager.GetMemoryStoreByAgentID(agentID)
    
    // 后续处理...
    history := sessionStore.GetHistory(opts.SessionKey)
    // ...
}
```

---

## 优势 ✅

### 1. **自动推断 agent_id**
```go
// API 层调用变得简单
func Chat(w http.ResponseWriter, r *http.Request) {
    var req ChatRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 如果是已存在的 session，不需要传 agent_id
    ctx := context.Background()
    response, _ := agentLoop.ProcessDirect(ctx, req.Message, req.SessionKey)
    // AgentLoop 会自动从 session_key 查到 agent_id
}
```

### 2. **减少参数传递**
- 不需要在整个调用链传递 agent_id
- session_key 已经包含了足够的信息

### 3. **符合直觉**
- session 本来就属于某个 agent
- 从 session_key 查 agent_id 很自然

---

## 挑战 ⚠️

### 1. **新 Session 问题**（关键）

**场景：用户首次创建会话**

```typescript
// 前端：创建新会话
const response = await fetch('/api/chat', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` },
  body: JSON.stringify({
    message: "你好",
    session_id: "new_chat_123",  // ← 新 session，数据库中不存在
  }),
});
```

**问题：数据库中查不到这个 session_key，无法获取 agent_id**

**解决方案：**

#### 方案 A：分两步处理
```go
func (al *AgentLoop) runAgentLoop(ctx context.Context, opts processOptions) (string, error) {
    // 1. 尝试从 session_key 查询 agent_id
    agentID, _ := al.storeManager.GetAgentIDBySession(opts.SessionKey)
    
    if agentID == "" {
        // 2. 新 session，从 context 获取 agent_id（JWT token）
        agentID = ctx.Value("agent_id").(string)
        
        // 3. 建立绑定关系（首次保存时会写入数据库）
        al.storeManager.BindSessionToAgent(opts.SessionKey, agentID)
    }
    
    // 使用 agentID 处理...
}
```

#### 方案 B：前端显式传 agent_id（仅首次）
```typescript
// 前端：新会话需要指定 agent_id
const response = await fetch('/api/chat', {
  body: JSON.stringify({
    message: "你好",
    session_id: "new_chat_123",
    agent_id: userId,  // ← 新会话时必须传
  }),
});

// 后续请求可以省略 agent_id
const response2 = await fetch('/api/chat', {
  body: JSON.stringify({
    message: "继续聊",
    session_id: "new_chat_123",  // ← 已存在，自动查到 agent_id
    // agent_id 可以不传
  }),
});
```

---

### 2. **性能问题**

**每次请求都要查数据库？**

```go
// 每次都查询
agentID, _ := db.Query("SELECT agent_id FROM POM_SESSIONS WHERE session_key = $1")
```

**解决方案：多级缓存**

```go
type StoreManager struct {
    // 内存缓存
    sessionAgentCache *lru.Cache  // session_key -> agent_id
    
    // Redis 缓存（可选）
    redis *redis.Client
}

func (sm *StoreManager) GetAgentIDBySession(sessionKey string) string {
    // 1. 内存缓存
    if agentID, ok := sm.sessionAgentCache.Get(sessionKey); ok {
        return agentID.(string)
    }
    
    // 2. Redis 缓存
    if sm.redis != nil {
        agentID, _ := sm.redis.Get(ctx, "session:"+sessionKey).Result()
        if agentID != "" {
            sm.sessionAgentCache.Add(sessionKey, agentID)
            return agentID
        }
    }
    
    // 3. 数据库查询
    var agentID string
    sm.db.QueryRow("SELECT agent_id FROM POM_SESSIONS WHERE session_key = $1", 
                   sessionKey).Scan(&agentID)
    
    // 4. 更新缓存
    sm.sessionAgentCache.Add(sessionKey, agentID)
    if sm.redis != nil {
        sm.redis.Set(ctx, "session:"+sessionKey, agentID, 24*time.Hour)
    }
    
    return agentID
}
```

---

### 3. **缓存一致性**

**问题：如果 session 的 agent_id 被修改了（虽然很少见）**

```go
// 管理员操作：将 session 转移给另一个 agent
UPDATE POM_SESSIONS 
SET agent_id = 'new_agent' 
WHERE session_key = 'chat_123';

// 但缓存中还是旧的 agent_id
```

**解决方案：**
- 通常不需要支持 session 转移
- 如果需要，提供缓存刷新机制

---

## 推荐方案：混合模式 ⭐

```go
// pkg/agent/loop.go
func (al *AgentLoop) runAgentLoop(ctx context.Context, opts processOptions) (string, error) {
    var agentID string
    
    // 策略 1: 优先从 context 获取（API 请求）
    if ctxAgentID, ok := ctx.Value("agent_id").(string); ok && ctxAgentID != "" {
        agentID = ctxAgentID
    } else {
        // 策略 2: 从 session_key 查询（已存在的 session）
        agentID, _ = al.storeManager.GetAgentIDBySession(opts.SessionKey)
    }
    
    // 策略 3: 降级到默认值
    if agentID == "" {
        agentID = "default"
    }
    
    // 获取对应的 Store
    sessionStore := al.storeManager.GetSessionStoreByAgentID(agentID)
    // ...
}
```

**优先级：**
1. **Context 传入的 agent_id**（最明确，来自 JWT）
2. **Session_key 查询的 agent_id**（缓存查询）
3. **默认值 "default"**（向后兼容）

---

## API 使用示例

### 前端代码（用户无感知）

```typescript
// src/services/api.ts
export async function sendMessage(sessionId: string, message: string) {
  const response = await fetch('/api/chat', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,  // ← JWT 包含 user_id
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      session_id: sessionId,  // ← 只需传 session_id
      message: message,
      // 不需要传 agent_id！后端自动处理
    }),
  });
  return response.json();
}
```

### 后端处理

```go
// pkg/api/handlers/chat.go
func Chat(w http.ResponseWriter, r *http.Request) {
    // 1. JWT 认证（中间件已完成）
    claims := r.Context().Value("user").(*auth.Claims)
    
    // 2. 解析请求
    var req ChatRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 3. 注入 agent_id 到 context（优先级最高）
    ctx := context.WithValue(r.Context(), "agent_id", claims.UserID)
    
    // 4. 调用 AgentLoop（自动处理 agent_id）
    response, _ := agentLoop.ProcessDirect(ctx, req.Message, req.SessionID)
    
    // 5. 返回结果
    json.NewEncoder(w).Encode(ChatResponse{Response: response})
}
```

**处理流程：**
```
新 Session:
  Context 有 agent_id → 使用它 → 保存时建立 session-agent 绑定

已存在 Session:
  Context 有 agent_id → 使用它（优先）
  没有？→ 从 session_key 查询缓存 → 使用查到的 agent_id
```

---

## 总结

### 你的方案可行！✅

**核心优势：**
- session_key 确实能锁定 agent_id
- 减少参数传递
- 符合直觉

**需要解决：**
1. **新 session 首次创建**：从 JWT/Context 获取 agent_id
2. **性能优化**：使用缓存（内存 + Redis）
3. **混合策略**：Context > Session 查询 > 默认值

**推荐实现：**
```go
// 混合模式：优先 Context，降级到 Session 查询
agentID := getAgentID(ctx, sessionKey)

func getAgentID(ctx context.Context, sessionKey string) string {
    // 1. 从 Context（JWT）
    if id := ctx.Value("agent_id"); id != nil {
        return id.(string)
    }
    // 2. 从 Session（缓存查询）
    if id := storeManager.GetAgentIDBySession(sessionKey); id != "" {
        return id
    }
    // 3. 默认值
    return "default"
}
```

**改造工作量：** 3-4 天（加上缓存优化）
