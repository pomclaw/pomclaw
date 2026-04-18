# 多 Agent 动态配置改造计划

## 核心思路 ✅

**将 Agent 配置从 config.json 移到数据库，支持运行时动态创建和管理**

```
当前架构：
  config.json (固定配置) → 启动时创建 AgentLoop → 处理所有请求

目标架构：
  用户创建 Agent → 配置存入数据库 POM_AGENT_CONFIGS
                    ↓
  请求到达 → 从数据库读取 Agent 配置 → 动态创建/获取 AgentLoop → 处理请求
```

---

## Phase 1: 数据库表设计（已完成部分）

### 需要的表

```sql
-- 1. Agent 配置表（核心）
CREATE TABLE POM_AGENT_CONFIGS (
    config_id      VARCHAR(64) PRIMARY KEY,
    user_id        VARCHAR(64) NOT NULL,           -- 所属用户
    agent_name     VARCHAR(255) NOT NULL,          -- Agent 名称
    agent_id       VARCHAR(64) UNIQUE NOT NULL,    -- agent_id（唯一标识）
    
    -- Agent 配置
    model          VARCHAR(255) DEFAULT 'glm-5',
    provider       VARCHAR(64) DEFAULT 'openai',
    max_tokens     INTEGER DEFAULT 8192,
    temperature    NUMERIC(3,2) DEFAULT 0.7,
    max_iterations INTEGER DEFAULT 20,
    
    -- 系统提示词
    system_prompt  TEXT,
    
    -- 工作空间配置
    workspace      VARCHAR(512),
    restrict_workspace BOOLEAN DEFAULT true,
    
    -- 状态
    is_active      BOOLEAN DEFAULT true,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES POM_USERS(user_id)
);

CREATE INDEX IDX_POM_AGENT_CONFIGS_USER ON POM_AGENT_CONFIGS(user_id);
CREATE INDEX IDX_POM_AGENT_CONFIGS_AGENT_ID ON POM_AGENT_CONFIGS(agent_id);

-- 2. 用户表（简化版，后续完善）
CREATE TABLE POM_USERS (
    user_id        VARCHAR(64) PRIMARY KEY,
    username       VARCHAR(255) UNIQUE NOT NULL,
    password_hash  VARCHAR(255) NOT NULL,
    email          VARCHAR(255),
    role           VARCHAR(32) DEFAULT 'user',  -- admin/user/viewer
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. 已有的表关联 agent_id
-- POM_SESSIONS (session_key, agent_id, ...)  ✅ 已有
-- POM_MEMORIES (memory_id, agent_id, ...)   ✅ 已有
-- POM_STATE (state_key, agent_id, ...)      ✅ 已有
```

---

## Phase 2: 核心代码改造

### 2.1 新增：AgentConfigStore (pkg/storage/agent_config_store.go)

```go
package storage

import (
    "database/sql"
    "fmt"
    "time"
)

// AgentConfig 表示数据库中的 Agent 配置
type AgentConfig struct {
    ConfigID          string
    UserID            string
    AgentName         string
    AgentID           string
    
    // LLM 配置
    Model             string
    Provider          string
    MaxTokens         int
    Temperature       float64
    MaxIterations     int
    
    // 提示词
    SystemPrompt      string
    
    // 工作空间
    Workspace         string
    RestrictWorkspace bool
    
    IsActive          bool
    CreatedAt         time.Time
    UpdatedAt         time.Time
}

type AgentConfigStore struct {
    db *sql.DB
}

func NewAgentConfigStore(db *sql.DB) *AgentConfigStore {
    return &AgentConfigStore{db: db}
}

// GetByAgentID 根据 agent_id 查询配置
func (s *AgentConfigStore) GetByAgentID(agentID string) (*AgentConfig, error) {
    var cfg AgentConfig
    err := s.db.QueryRow(`
        SELECT config_id, user_id, agent_name, agent_id,
               model, provider, max_tokens, temperature, max_iterations,
               system_prompt, workspace, restrict_workspace,
               is_active, created_at, updated_at
        FROM POM_AGENT_CONFIGS
        WHERE agent_id = $1 AND is_active = true
    `, agentID).Scan(
        &cfg.ConfigID, &cfg.UserID, &cfg.AgentName, &cfg.AgentID,
        &cfg.Model, &cfg.Provider, &cfg.MaxTokens, &cfg.Temperature, &cfg.MaxIterations,
        &cfg.SystemPrompt, &cfg.Workspace, &cfg.RestrictWorkspace,
        &cfg.IsActive, &cfg.CreatedAt, &cfg.UpdatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("agent config not found: %s", agentID)
    }
    if err != nil {
        return nil, err
    }
    
    return &cfg, nil
}

// GetByUserID 查询用户的所有 Agent
func (s *AgentConfigStore) GetByUserID(userID string) ([]*AgentConfig, error) {
    rows, err := s.db.Query(`
        SELECT config_id, user_id, agent_name, agent_id,
               model, provider, max_tokens, temperature, max_iterations,
               system_prompt, workspace, restrict_workspace,
               is_active, created_at, updated_at
        FROM POM_AGENT_CONFIGS
        WHERE user_id = $1
        ORDER BY created_at DESC
    `, userID)
    
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var configs []*AgentConfig
    for rows.Next() {
        var cfg AgentConfig
        if err := rows.Scan(
            &cfg.ConfigID, &cfg.UserID, &cfg.AgentName, &cfg.AgentID,
            &cfg.Model, &cfg.Provider, &cfg.MaxTokens, &cfg.Temperature, &cfg.MaxIterations,
            &cfg.SystemPrompt, &cfg.Workspace, &cfg.RestrictWorkspace,
            &cfg.IsActive, &cfg.CreatedAt, &cfg.UpdatedAt,
        ); err != nil {
            return nil, err
        }
        configs = append(configs, &cfg)
    }
    
    return configs, nil
}

// Create 创建新的 Agent 配置
func (s *AgentConfigStore) Create(cfg *AgentConfig) error {
    _, err := s.db.Exec(`
        INSERT INTO POM_AGENT_CONFIGS (
            config_id, user_id, agent_name, agent_id,
            model, provider, max_tokens, temperature, max_iterations,
            system_prompt, workspace, restrict_workspace,
            is_active, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
        )
    `, cfg.ConfigID, cfg.UserID, cfg.AgentName, cfg.AgentID,
       cfg.Model, cfg.Provider, cfg.MaxTokens, cfg.Temperature, cfg.MaxIterations,
       cfg.SystemPrompt, cfg.Workspace, cfg.RestrictWorkspace,
       cfg.IsActive, cfg.CreatedAt, cfg.UpdatedAt)
    
    return err
}

// Update 更新 Agent 配置
func (s *AgentConfigStore) Update(cfg *AgentConfig) error {
    _, err := s.db.Exec(`
        UPDATE POM_AGENT_CONFIGS
        SET agent_name = $1, model = $2, provider = $3,
            max_tokens = $4, temperature = $5, max_iterations = $6,
            system_prompt = $7, workspace = $8, restrict_workspace = $9,
            updated_at = CURRENT_TIMESTAMP
        WHERE agent_id = $10
    `, cfg.AgentName, cfg.Model, cfg.Provider,
       cfg.MaxTokens, cfg.Temperature, cfg.MaxIterations,
       cfg.SystemPrompt, cfg.Workspace, cfg.RestrictWorkspace,
       cfg.AgentID)
    
    return err
}

// Delete 删除（软删除）
func (s *AgentConfigStore) Delete(agentID string) error {
    _, err := s.db.Exec(`
        UPDATE POM_AGENT_CONFIGS
        SET is_active = false, updated_at = CURRENT_TIMESTAMP
        WHERE agent_id = $1
    `, agentID)
    
    return err
}
```

---

### 2.2 改造：StoreManager (pkg/storage/store_manager.go)

```go
package storage

import (
    "database/sql"
    "sync"
    
    "github.com/pomclaw/pomclaw/pkg/config"
    postgres "github.com/pomclaw/pomclaw/pkg/postgres"
)

type StoreManager struct {
    db               *sql.DB
    cfg              *config.Config
    embSvc           EmbeddingService
    agentConfigStore *AgentConfigStore
    
    // Store 实例缓存（按 agentID）
    sessionStores sync.Map  // agentID -> *SessionStore
    stateStores   sync.Map
    memoryStores  sync.Map
}

func NewStoreManager(cfg *config.Config, db *sql.DB, embSvc EmbeddingService) *StoreManager {
    return &StoreManager{
        db:               db,
        cfg:              cfg,
        embSvc:           embSvc,
        agentConfigStore: NewAgentConfigStore(db),
    }
}

// GetSessionStore 根据 agentID 获取 SessionStore（懒加载 + 缓存）
func (sm *StoreManager) GetSessionStore(agentID string) (*postgres.SessionStore, error) {
    // 1. 从缓存获取
    if store, ok := sm.sessionStores.Load(agentID); ok {
        return store.(*postgres.SessionStore), nil
    }
    
    // 2. 创建新实例
    store := postgres.NewSessionStore(sm.db, agentID)
    sm.sessionStores.Store(agentID, store)
    
    return store, nil
}

func (sm *StoreManager) GetStateStore(agentID string) (*postgres.StateStore, error) {
    if store, ok := sm.stateStores.Load(agentID); ok {
        return store.(*postgres.StateStore), nil
    }
    
    store := postgres.NewStateStore(sm.db, agentID)
    sm.stateStores.Store(agentID, store)
    
    return store, nil
}

func (sm *StoreManager) GetMemoryStore(agentID string) (*postgres.MemoryStore, error) {
    if store, ok := sm.memoryStores.Load(agentID); ok {
        return store.(*postgres.MemoryStore), nil
    }
    
    store := postgres.NewMemoryStore(sm.db, agentID, sm.embSvc)
    sm.memoryStores.Store(agentID, store)
    
    return store, nil
}

// GetAgentConfig 获取 Agent 配置
func (sm *StoreManager) GetAgentConfig(agentID string) (*AgentConfig, error) {
    return sm.agentConfigStore.GetByAgentID(agentID)
}
```

---

### 2.3 改造：AgentLoop 支持动态配置

```go
// pkg/agent/loop.go

type AgentLoop struct {
    bus           *bus.MessageBus
    provider      providers.LLMProvider
    storeManager  *storage.StoreManager  // ← 新增
    
    // 移除固定的 Store
    // sessions  SessionManagerInterface  // ← 删除
    // state     StateManagerInterface    // ← 删除
    
    // 保留全局配置（作为默认值）
    defaultCfg    *config.Config
    
    // 其他字段...
}

// processMessage 处理消息（核心改动）
func (al *AgentLoop) processMessage(ctx context.Context, msg bus.InboundMessage) (string, error) {
    // 1. 从消息中获取 agentID
    agentID := msg.AgentID
    if agentID == "" {
        agentID = "default"  // 向后兼容
    }
    
    // 2. 从数据库加载 Agent 配置
    agentConfig, err := al.storeManager.GetAgentConfig(agentID)
    if err != nil {
        // 如果找不到配置，使用全局默认配置
        logger.WarnCF("agent", "Agent config not found, using default", 
                     map[string]interface{}{"agent_id": agentID})
        agentConfig = al.getDefaultAgentConfig(agentID)
    }
    
    // 3. 动态获取该 agent 的 Store
    sessionStore, _ := al.storeManager.GetSessionStore(agentID)
    stateStore, _ := al.storeManager.GetStateStore(agentID)
    memoryStore, _ := al.storeManager.GetMemoryStore(agentID)
    
    // 4. 使用 Agent 配置处理消息
    return al.runAgentLoopWithConfig(ctx, processOptions{
        SessionKey:      msg.SessionKey,
        Channel:         msg.Channel,
        ChatID:          msg.ChatID,
        UserMessage:     msg.Content,
        DefaultResponse: "I've completed processing but have no response to give.",
        EnableSummary:   true,
        SendResponse:    false,
    }, agentConfig, sessionStore, stateStore, memoryStore)
}

// runAgentLoopWithConfig 使用动态配置运行
func (al *AgentLoop) runAgentLoopWithConfig(
    ctx context.Context,
    opts processOptions,
    agentConfig *storage.AgentConfig,
    sessionStore agent.SessionManagerInterface,
    stateStore agent.StateManagerInterface,
    memoryStore agent.MemoryStoreInterface,
) (string, error) {
    // 使用 agentConfig 的配置
    model := agentConfig.Model
    maxTokens := agentConfig.MaxTokens
    temperature := agentConfig.Temperature
    
    // 构建消息（包含 system_prompt）
    history := sessionStore.GetHistory(opts.SessionKey)
    summary := sessionStore.GetSummary(opts.SessionKey)
    
    messages := al.contextBuilder.BuildMessagesWithSystemPrompt(
        history,
        summary,
        opts.UserMessage,
        agentConfig.SystemPrompt,  // ← 使用自定义 system_prompt
        nil,
        opts.Channel,
        opts.ChatID,
    )
    
    // 调用 LLM（使用动态配置）
    response, err := al.provider.Chat(ctx, messages, providerToolDefs, model, map[string]interface{}{
        "max_tokens":  maxTokens,
        "temperature": temperature,
    })
    
    // ... 后续处理
}

// getDefaultAgentConfig 获取默认配置（向后兼容）
func (al *AgentLoop) getDefaultAgentConfig(agentID string) *storage.AgentConfig {
    return &storage.AgentConfig{
        AgentID:           agentID,
        Model:             al.defaultCfg.Agents.Defaults.Model,
        Provider:          al.defaultCfg.Agents.Defaults.Provider,
        MaxTokens:         al.defaultCfg.Agents.Defaults.MaxTokens,
        Temperature:       al.defaultCfg.Agents.Defaults.Temperature,
        MaxIterations:     al.defaultCfg.Agents.Defaults.MaxToolIterations,
        Workspace:         al.defaultCfg.WorkspacePath(),
        RestrictWorkspace: al.defaultCfg.Agents.Defaults.RestrictToWorkspace,
    }
}
```

---

### 2.4 修改：ProcessDirect 增加 agentID 参数

```go
// pkg/agent/loop.go

// 修改前
func (al *AgentLoop) ProcessDirect(ctx context.Context, content, sessionKey string) (string, error)

// 修改后
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
        AgentID:    agentID,  // ← 新增
    }
    
    return al.processMessage(ctx, msg)
}
```

---

### 2.5 修改：InboundMessage 结构

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

---

## Phase 3: RESTful API（多租户支持）

### 3.1 Agent 管理 API

```go
// pkg/api/handlers/agent.go (新建)

// ListAgents 查询用户的所有 Agent
func ListAgents(w http.ResponseWriter, r *http.Request) {
    claims := r.Context().Value("user").(*auth.Claims)
    
    configs, err := agentConfigStore.GetByUserID(claims.UserID)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "agents": configs,
    })
}

// CreateAgent 创建新 Agent
func CreateAgent(w http.ResponseWriter, r *http.Request) {
    claims := r.Context().Value("user").(*auth.Claims)
    
    var req struct {
        AgentName    string  `json:"agent_name"`
        Model        string  `json:"model"`
        Temperature  float64 `json:"temperature"`
        SystemPrompt string  `json:"system_prompt"`
    }
    json.NewDecoder(r.Body).Decode(&req)
    
    // 生成唯一的 agent_id
    agentID := fmt.Sprintf("agent_%s_%s", claims.UserID, uuid.New().String()[:8])
    
    cfg := &storage.AgentConfig{
        ConfigID:     uuid.New().String(),
        UserID:       claims.UserID,
        AgentName:    req.AgentName,
        AgentID:      agentID,
        Model:        req.Model,
        Temperature:  req.Temperature,
        SystemPrompt: req.SystemPrompt,
        MaxTokens:    8192,
        IsActive:     true,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }
    
    if err := agentConfigStore.Create(cfg); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "agent_id": agentID,
        "message":  "Agent created successfully",
    })
}

// UpdateAgent 更新 Agent 配置
func UpdateAgent(w http.ResponseWriter, r *http.Request) {
    claims := r.Context().Value("user").(*auth.Claims)
    agentID := chi.URLParam(r, "id")
    
    // 验证权限：只能修改自己的 Agent
    cfg, err := agentConfigStore.GetByAgentID(agentID)
    if err != nil || cfg.UserID != claims.UserID {
        http.Error(w, "Agent not found or access denied", 404)
        return
    }
    
    var req struct {
        AgentName    string  `json:"agent_name"`
        Model        string  `json:"model"`
        Temperature  float64 `json:"temperature"`
        SystemPrompt string  `json:"system_prompt"`
    }
    json.NewDecoder(r.Body).Decode(&req)
    
    cfg.AgentName = req.AgentName
    cfg.Model = req.Model
    cfg.Temperature = req.Temperature
    cfg.SystemPrompt = req.SystemPrompt
    
    if err := agentConfigStore.Update(cfg); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Agent updated successfully",
    })
}

// DeleteAgent 删除 Agent
func DeleteAgent(w http.ResponseWriter, r *http.Request) {
    claims := r.Context().Value("user").(*auth.Claims)
    agentID := chi.URLParam(r, "id")
    
    // 验证权限
    cfg, err := agentConfigStore.GetByAgentID(agentID)
    if err != nil || cfg.UserID != claims.UserID {
        http.Error(w, "Agent not found or access denied", 404)
        return
    }
    
    if err := agentConfigStore.Delete(agentID); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Agent deleted successfully",
    })
}
```

### 3.2 Chat API（使用指定 Agent）

```go
// pkg/api/handlers/chat.go

func Chat(w http.ResponseWriter, r *http.Request) {
    claims := r.Context().Value("user").(*auth.Claims)
    
    var req struct {
        AgentID    string `json:"agent_id"`     // ← 用户指定使用哪个 Agent
        Message    string `json:"message"`
        SessionKey string `json:"session_key"`
    }
    json.NewDecoder(r.Body).Decode(&req)
    
    // 验证 Agent 属于该用户
    cfg, err := agentConfigStore.GetByAgentID(req.AgentID)
    if err != nil || cfg.UserID != claims.UserID {
        http.Error(w, "Agent not found or access denied", 404)
        return
    }
    
    // 调用 AgentLoop（传入 agentID）
    ctx := context.Background()
    response, err := agentLoop.ProcessDirect(
        ctx,
        req.Message,
        req.SessionKey,
        req.AgentID,  // ← 使用用户指定的 Agent
    )
    
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "response": response,
    })
}
```

---

## Phase 4: 支持的功能

### ✅ 核心功能

1. **动态多 Agent**
   - 每个用户可以创建多个 Agent
   - 每个 Agent 有独立的配置（model, temperature, system_prompt）
   
2. **Agent 配置管理**
   - 创建 Agent：`POST /api/agents`
   - 查询 Agent 列表：`GET /api/agents`
   - 更新 Agent：`PUT /api/agents/{id}`
   - 删除 Agent：`DELETE /api/agents/{id}`

3. **独立的对话历史**
   - 每个 Agent 有独立的 session/memory/state
   - 数据通过 agent_id 隔离

4. **自定义 System Prompt**
   - 每个 Agent 可以有不同的角色定位
   - 例如：代码助手、写作助手、客服机器人

5. **灵活的 LLM 配置**
   - 不同 Agent 可以使用不同的模型
   - 独立调整 temperature、max_tokens 等参数

### ✅ 向后兼容

- CLI 默认使用 `agent_id = "default"`
- 没有配置的 Agent 降级到全局默认配置

---

## Phase 5: 使用示例

### 用户创建多个 Agent

```bash
# 1. 创建代码助手 Agent
curl -X POST http://localhost:18790/api/agents \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "agent_name": "Code Assistant",
    "model": "gpt-4",
    "temperature": 0.2,
    "system_prompt": "You are an expert programmer. Always provide clean, well-documented code."
  }'
# 返回：{ "agent_id": "agent_user123_abc123" }

# 2. 创建写作助手 Agent
curl -X POST http://localhost:18790/api/agents \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "agent_name": "Writing Assistant",
    "model": "claude-3",
    "temperature": 0.8,
    "system_prompt": "You are a creative writing assistant. Help users write engaging content."
  }'
# 返回：{ "agent_id": "agent_user123_def456" }

# 3. 查询我的所有 Agent
curl http://localhost:18790/api/agents \
  -H "Authorization: Bearer $JWT_TOKEN"
# 返回：{ "agents": [{ "agent_id": "...", "agent_name": "..." }, ...] }
```

### 使用不同的 Agent 对话

```bash
# 使用代码助手
curl -X POST http://localhost:18790/api/chat \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "agent_id": "agent_user123_abc123",
    "message": "帮我写一个快速排序",
    "session_key": "code_session_1"
  }'

# 使用写作助手
curl -X POST http://localhost:18790/api/chat \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "agent_id": "agent_user123_def456",
    "message": "帮我写一篇关于AI的文章",
    "session_key": "writing_session_1"
  }'
```

---

## 改动文件清单

### 新增文件
1. `pkg/storage/agent_config_store.go` - Agent 配置管理
2. `pkg/storage/store_manager.go` - Store 管理器
3. `pkg/api/handlers/agent.go` - Agent 管理 API
4. `pkg/api/handlers/chat.go` - 对话 API
5. `pkg/api/middleware/auth.go` - JWT 认证中间件
6. `pkg/api/router.go` - API 路由

### 修改文件
1. `pkg/bus/bus.go` - InboundMessage 增加 AgentID
2. `pkg/agent/loop.go` - 支持动态配置和 StoreManager
3. `pkg/postgres/schema.go` - 增加 POM_AGENT_CONFIGS 表
4. `pkg/oracle/schema.go` - 增加 POM_AGENT_CONFIGS 表
5. `cmd/pomclaw/main.go` - 初始化 StoreManager
6. 所有调用 `ProcessDirect` 的地方 - 增加 agentID 参数

---

## 工作量估算

| 任务 | 工作量 |
|------|--------|
| **Phase 1: 数据库表** | 0.5天 |
| **Phase 2: 核心代码** | 3天 |
| - AgentConfigStore | 0.5天 |
| - StoreManager | 1天 |
| - AgentLoop 改造 | 1.5天 |
| **Phase 3: API 层** | 2天 |
| - Agent 管理 API | 1天 |
| - Chat API | 0.5天 |
| - JWT 认证 | 0.5天 |
| **Phase 4: 测试** | 1天 |
| **总计** | **6.5天** |

---

## 总结

### ✅ 你的方案非常合理！

**核心优势：**
1. **配置存数据库** - 支持动态创建和管理
2. **多租户友好** - 天然支持多用户多 Agent
3. **灵活性高** - 每个 Agent 独立配置
4. **向后兼容** - CLI 仍然能用

**架构清晰：**
```
用户 → 创建多个 Agent → 配置存数据库
                        ↓
请求 → 指定 agent_id → 加载配置 → 动态处理
```

**改造后支持：**
- ✅ 每用户多 Agent
- ✅ 独立配置（model, prompt, temperature）
- ✅ 独立数据（session, memory, state）
- ✅ RESTful API 管理
- ✅ 为多租户 SaaS 打好基础
