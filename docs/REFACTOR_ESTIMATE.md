# 多租户架构改造难度评估

## 总体评估

**难度等级：** 中等偏高 (7/10)  
**预计工作量：** 3-4 周（1个全职开发者）  
**风险等级：** 中等

---

## 分模块评估

### 1. 核心架构改造 ⭐⭐⭐⭐ (4/5 难度)

#### 当前状态
```go
// cmd/pomclaw/main.go
var agentLoop *agent.AgentLoop  // 全局单例
agentLoop = agent.NewAgentLoop(cfg, msgBus, provider)
go agentLoop.Run(ctx)  // 单个后台循环处理所有请求
```

#### 目标架构
```go
// pkg/agent/pool.go (新建)
type AgentPool struct {
    agents   map[string]*agent.AgentLoop  // userID -> AgentLoop
    configs  map[string]*UserAgentConfig  // 用户自定义配置
    mu       sync.RWMutex
    provider providers.LLMProvider
    cfg      *config.Config
}

func (p *AgentPool) GetOrCreate(userID string) (*agent.AgentLoop, error) {
    p.mu.RLock()
    if a, exists := p.agents[userID]; exists {
        p.mu.RUnlock()
        return a, nil
    }
    p.mu.RUnlock()

    // 创建用户专属的 AgentLoop
    p.mu.Lock()
    defer p.mu.Unlock()
    
    userCfg := p.getUserConfig(userID)  // 从数据库加载
    agentID := fmt.Sprintf("user_%s", userID)
    
    // 使用用户的 agent_id 创建数据库 stores
    sessionStore := storage.NewSessionStore(p.cfg, db)
    stateStore := storage.NewStateStore(p.cfg, db)
    memoryStore := storage.NewMemoryStore(p.cfg, db, embSvc)
    
    loop := agent.NewAgentLoopWithStores(userCfg, msgBus, provider, 
                                         sessionStore, stateStore, memoryStore)
    p.agents[userID] = loop
    return loop, nil
}
```

#### 优势
✅ **AgentLoop 本身无需改动** - 已经是实例化设计  
✅ **数据库天然支持多租户** - 所有表都有 `agent_id` 字段  
✅ **MessageBus 可复用** - 每个 AgentLoop 可以共享或独立  

#### 改动点
- 新增 `pkg/agent/pool.go` - Agent 实例池管理器
- 修改 `cmd/pomclaw/main.go` - gateway 命令使用 AgentPool
- 修改存储工厂函数 - 支持动态传入 `agent_id`

**工作量：** 2-3 天

---

### 2. 认证授权系统 ⭐⭐⭐ (3/5 难度)

#### 需要新增

##### 2.1 数据库表 (已在架构文档中设计)
```sql
-- 已设计好的表结构
CREATE TABLE POM_USERS (
    user_id VARCHAR(64) PRIMARY KEY,
    organization_id VARCHAR(64),
    username VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255),
    email VARCHAR(255),
    role VARCHAR(32),  -- admin/user/viewer
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE POM_ORGANIZATIONS (
    org_id VARCHAR(64) PRIMARY KEY,
    org_name VARCHAR(255),
    plan VARCHAR(32),  -- free/pro/enterprise
    created_at TIMESTAMP
);

CREATE TABLE POM_AGENT_CONFIGS (
    config_id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64),
    agent_name VARCHAR(255),
    model VARCHAR(255),
    max_tokens INTEGER,
    temperature NUMERIC,
    system_prompt TEXT,
    created_at TIMESTAMP
);

CREATE TABLE POM_AGENT_USAGE (
    usage_id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64),
    config_id VARCHAR(64),
    tokens_used INTEGER,
    cost NUMERIC,
    timestamp TIMESTAMP
);
```

##### 2.2 JWT 认证中间件
```go
// pkg/auth/jwt.go (新建)
type Claims struct {
    UserID         string `json:"user_id"`
    OrganizationID string `json:"organization_id"`
    Role           string `json:"role"`
    jwt.StandardClaims
}

func GenerateToken(userID, orgID, role string) (string, error)
func ValidateToken(tokenString string) (*Claims, error)

// pkg/auth/middleware.go (新建)
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        claims, err := ValidateToken(token)
        if err != nil {
            http.Error(w, "Unauthorized", 401)
            return
        }
        // 将 claims 存入 context
        ctx := context.WithValue(r.Context(), "user", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

##### 2.3 RBAC 权限检查
```go
// pkg/auth/rbac.go (新建)
type Permission string

const (
    PermCreateAgent Permission = "agent:create"
    PermDeleteAgent Permission = "agent:delete"
    PermViewUsage   Permission = "usage:view"
    // ...
)

var RolePermissions = map[string][]Permission{
    "admin": {PermCreateAgent, PermDeleteAgent, PermViewUsage, ...},
    "user":  {PermCreateAgent, PermViewUsage, ...},
    "viewer": {PermViewUsage},
}

func CheckPermission(role string, perm Permission) bool
```

**工作量：** 2 天

---

### 3. RESTful API 层 ⭐⭐⭐⭐ (4/5 难度)

#### 需要新增的 API

```go
// pkg/api/router.go (新建)
func NewRouter(agentPool *agent.AgentPool, db *sql.DB) *chi.Mux {
    r := chi.NewRouter()
    
    // 公开接口
    r.Post("/api/auth/login", handlers.Login)
    r.Post("/api/auth/register", handlers.Register)
    
    // 需要认证的接口
    r.Group(func(r chi.Router) {
        r.Use(auth.AuthMiddleware)
        
        // Agent 管理
        r.Get("/api/agents", handlers.ListAgents)
        r.Post("/api/agents", handlers.CreateAgent)
        r.Put("/api/agents/{id}", handlers.UpdateAgent)
        r.Delete("/api/agents/{id}", handlers.DeleteAgent)
        
        // 对话接口
        r.Post("/api/chat", handlers.Chat)              // 同步
        r.Get("/api/chat/stream", handlers.ChatStream)  // WebSocket 流式
        
        // 用户管理 (admin only)
        r.Get("/api/users", handlers.ListUsers)
        r.Post("/api/users", handlers.CreateUser)
        
        // 使用量统计
        r.Get("/api/usage", handlers.GetUsage)
    })
    
    return r
}
```

#### 核心处理器示例

```go
// pkg/api/handlers/chat.go (新建)
func Chat(w http.ResponseWriter, r *http.Request) {
    // 1. 从 context 获取用户信息
    claims := r.Context().Value("user").(*auth.Claims)
    
    // 2. 解析请求
    var req ChatRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 3. 从 AgentPool 获取用户的 AgentLoop
    agentLoop, err := agentPool.GetOrCreate(claims.UserID)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    
    // 4. 处理对话
    ctx := context.Background()
    response, err := agentLoop.ProcessDirect(ctx, req.Message, req.SessionKey)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    
    // 5. 记录使用量
    recordUsage(claims.UserID, req.AgentConfigID, tokensUsed)
    
    // 6. 返回结果
    json.NewEncoder(w).Encode(ChatResponse{
        Response: response,
        TokensUsed: tokensUsed,
    })
}

// WebSocket 流式接口
func ChatStream(w http.ResponseWriter, r *http.Request) {
    conn, _ := upgrader.Upgrade(w, r, nil)
    defer conn.Close()
    
    // 获取用户的 AgentLoop
    claims := r.Context().Value("user").(*auth.Claims)
    agentLoop, _ := agentPool.GetOrCreate(claims.UserID)
    
    // 通过 MessageBus 实时推送响应
    // ...
}
```

**工作量：** 4-5 天

---

### 4. 前端开发 ⭐⭐⭐⭐⭐ (5/5 难度)

#### 技术栈
- React 18 + TypeScript
- Zustand (状态管理)
- React Query (API 客户端)
- Tailwind CSS (UI)

#### 核心页面

```
src/
├── pages/
│   ├── Login.tsx           # 登录页
│   ├── Dashboard.tsx       # 控制台首页
│   ├── Agents.tsx          # Agent 列表和配置
│   ├── Chat.tsx            # 对话界面
│   ├── Usage.tsx           # 使用量统计
│   └── Users.tsx           # 用户管理 (admin only)
├── components/
│   ├── AgentCard.tsx       # Agent 卡片
│   ├── ChatWindow.tsx      # 对话窗口
│   ├── MessageBubble.tsx   # 消息气泡
│   └── UsageChart.tsx      # 使用量图表
├── hooks/
│   ├── useAuth.ts          # 认证 hook
│   ├── useAgents.ts        # Agent 管理 hook
│   └── useChat.ts          # 对话 hook (WebSocket)
├── services/
│   └── api.ts              # API 客户端
└── stores/
    ├── authStore.ts        # 认证状态
    └── chatStore.ts        # 对话状态
```

#### 示例实现

```typescript
// src/hooks/useChat.ts
export function useChat(agentId: string) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const ws = useRef<WebSocket>();

  const sendMessage = async (content: string) => {
    setIsLoading(true);
    
    // WebSocket 流式响应
    ws.current = new WebSocket(`ws://api/chat/stream`);
    ws.current.onmessage = (event) => {
      const chunk = JSON.parse(event.data);
      setMessages(prev => [...prev, chunk]);
    };
    
    ws.current.send(JSON.stringify({
      agent_id: agentId,
      message: content,
      session_key: sessionKey,
    }));
    
    setIsLoading(false);
  };

  return { messages, sendMessage, isLoading };
}
```

**工作量：** 7-10 天（如果有前端经验）

---

### 5. 数据迁移 ⭐⭐ (2/5 难度)

#### 现有数据迁移脚本

```sql
-- 1. 创建新表
CREATE TABLE POM_USERS (...);
CREATE TABLE POM_ORGANIZATIONS (...);
CREATE TABLE POM_AGENT_CONFIGS (...);

-- 2. 迁移现有数据（给默认用户）
INSERT INTO POM_USERS (user_id, username, role, organization_id)
VALUES ('default_user', 'admin', 'admin', 'default_org');

INSERT INTO POM_ORGANIZATIONS (org_id, org_name, plan)
VALUES ('default_org', 'Default Organization', 'enterprise');

-- 3. 现有 agent_id='default' 的数据关联到默认用户
-- 不需要修改现有数据，只需在 POM_AGENT_CONFIGS 中创建关联
INSERT INTO POM_AGENT_CONFIGS (config_id, user_id, agent_name, model)
VALUES ('default_config', 'default_user', 'Default Agent', 'glm-5');
```

**工作量：** 0.5 天

---

## 实施路线图

### Phase 1: 后端基础 (5-6 天)
1. ✅ 数据库表设计（已完成 - 见 ENTERPRISE_ARCHITECTURE.md）
2. Agent Pool 实现
3. JWT 认证系统
4. RBAC 权限系统

### Phase 2: API 层 (4-5 天)
1. RESTful API 路由
2. 核心业务处理器（Chat, Agent CRUD）
3. WebSocket 流式接口
4. 使用量统计

### Phase 3: 前端开发 (7-10 天)
1. 登录/注册页面
2. Dashboard 和 Agent 管理
3. 对话界面（含流式显示）
4. 用户管理（admin）

### Phase 4: 测试和部署 (2-3 天)
1. 单元测试
2. 集成测试
3. 性能测试
4. Docker 部署配置

---

## 风险点

### 高风险
1. **并发 Agent 管理** - 多个用户同时对话时的资源管理
   - 缓解：设置 Agent 超时自动回收机制
   
2. **WebSocket 连接管理** - 大量并发 WebSocket 连接
   - 缓解：使用连接池 + 心跳检测

### 中风险
3. **Token 消耗追踪** - 准确计算每个请求的 token 使用量
   - 缓解：在 provider 层统一拦截计量

4. **数据隔离** - 确保用户只能访问自己的数据
   - 缓解：所有查询强制加 `WHERE agent_id = user_agent_id`

---

## 性能优化建议

1. **Agent 实例池化**
   ```go
   // 限制最大 Agent 数量，LRU 淘汰
   type AgentPool struct {
       maxAgents int  // 例如 100
       cache *lru.Cache
   }
   ```

2. **连接复用**
   - 所有 AgentLoop 共享同一个数据库连接池
   - 不需要为每个用户创建新连接

3. **缓存策略**
   - Redis 缓存用户配置、Agent 配置
   - 减少数据库查询

---

## 总结

### 改造可行性：✅ 高

**理由：**
- 现有架构设计良好（AgentLoop 实例化、数据库 agent_id 隔离）
- 改动点清晰，主要是"包装层"而非"核心重构"
- 数据库表结构已设计完成

### 推荐方案

**如果你的目标是快速上线 MVP：**

1. **简化版（2 周）**
   - 只做后端 API + Agent Pool
   - 使用现有的 CLI 或 Postman 测试
   - 前端暂时不做，或用简单的 HTML + jQuery

2. **完整版（3-4 周）**
   - 按上述 4 个 Phase 完整实施
   - 包含完整的 React 前端

### 需要的技能

- ✅ Go 后端开发（你已经具备）
- ✅ PostgreSQL（你已在用）
- ⚠️ React/TypeScript（如果没经验，前端会花更多时间）
- ✅ JWT/RBAC（通用知识，学习成本低）

**结论：改造难度在可控范围内，关键是前端开发时间。如果只做后端 API，2 周可完成。**
