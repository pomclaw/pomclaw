# 企业级 AI Agent 平台架构设计

本文档描述了将 PomClaw 升级为企业级多租户 AI Agent 平台的完整架构设计。

## 目录

- [1. 核心业务模型](#1-核心业务模型)
- [2. 系统架构](#2-系统架构)
- [3. 认证与授权](#3-认证与授权)
- [4. API 设计](#4-api-设计)
- [5. 前端设计](#5-前端设计)
- [6. 后端实现](#6-后端实现)
- [7. 安全性考虑](#7-安全性考虑)
- [8. 可扩展性](#8-可扩展性)
- [9. 监控与日志](#9-监控与日志)
- [10. 实施路线图](#10-实施路线图)

---

## 1. 核心业务模型

### 1.1 数据模型设计

```
Organization (企业/组织)
  ├── Users (用户)
  │     ├── email, password_hash
  │     ├── role (admin, user, viewer)
  │     └── created_at, last_login
  │
  └── Agents (AI Agent)
        ├── agent_id (UUID)
        ├── owner_id (user_id)
        ├── name, description
        ├── model, provider
        ├── system_prompt
        ├── tools_enabled []
        ├── status (active, paused, deleted)
        └── quota (tokens, requests)
```

### 1.2 数据库表设计

#### 用户表
```sql
CREATE TABLE pom_users (
    user_id VARCHAR(64) PRIMARY KEY,
    org_id VARCHAR(64) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    username VARCHAR(100),
    role VARCHAR(20) DEFAULT 'user',  -- admin, user, viewer
    status VARCHAR(20) DEFAULT 'active',  -- active, suspended
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP
);

CREATE INDEX idx_users_org ON pom_users(org_id);
CREATE INDEX idx_users_email ON pom_users(email);
```

#### 组织表
```sql
CREATE TABLE pom_organizations (
    org_id VARCHAR(64) PRIMARY KEY,
    org_name VARCHAR(255) NOT NULL,
    plan_type VARCHAR(50) DEFAULT 'free',  -- free, pro, enterprise
    quota_limit INTEGER DEFAULT 10000,
    quota_used INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### Agent 配置表
```sql
CREATE TABLE pom_agent_configs (
    agent_id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    org_id VARCHAR(64) NOT NULL,
    agent_name VARCHAR(255) NOT NULL,
    description TEXT,
    provider VARCHAR(50) DEFAULT 'openai',
    model VARCHAR(100) DEFAULT 'gpt-4',
    system_prompt TEXT,
    temperature NUMERIC(3,2) DEFAULT 0.7,
    max_tokens INTEGER DEFAULT 4096,
    tools_enabled JSONB DEFAULT '[]',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES pom_users(user_id),
    FOREIGN KEY (org_id) REFERENCES pom_organizations(org_id)
);

CREATE INDEX idx_agents_user ON pom_agent_configs(user_id);
CREATE INDEX idx_agents_org ON pom_agent_configs(org_id);
CREATE INDEX idx_agents_status ON pom_agent_configs(status);
```

#### Agent 使用统计表
```sql
CREATE TABLE pom_agent_usage (
    usage_id VARCHAR(64) PRIMARY KEY,
    agent_id VARCHAR(64) NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    org_id VARCHAR(64) NOT NULL,
    tokens_input INTEGER DEFAULT 0,
    tokens_output INTEGER DEFAULT 0,
    tokens_total INTEGER DEFAULT 0,
    request_count INTEGER DEFAULT 1,
    cost DECIMAL(10,4) DEFAULT 0,
    date DATE DEFAULT CURRENT_DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (agent_id) REFERENCES pom_agent_configs(agent_id),
    FOREIGN KEY (user_id) REFERENCES pom_users(user_id),
    FOREIGN KEY (org_id) REFERENCES pom_organizations(org_id)
);

CREATE INDEX idx_usage_agent_date ON pom_agent_usage(agent_id, date);
CREATE INDEX idx_usage_user_date ON pom_agent_usage(user_id, date);
CREATE INDEX idx_usage_org_date ON pom_agent_usage(org_id, date);
```

#### 修改现有表，添加租户隔离
```sql
-- 为现有表添加多租户支持
ALTER TABLE pom_memories ADD COLUMN user_id VARCHAR(64);
ALTER TABLE pom_memories ADD COLUMN org_id VARCHAR(64);
CREATE INDEX idx_memories_user ON pom_memories(user_id);
CREATE INDEX idx_memories_org ON pom_memories(org_id);

ALTER TABLE pom_sessions ADD COLUMN user_id VARCHAR(64);
ALTER TABLE pom_sessions ADD COLUMN org_id VARCHAR(64);
CREATE INDEX idx_sessions_user ON pom_sessions(user_id);

ALTER TABLE pom_transcripts ADD COLUMN user_id VARCHAR(64);
ALTER TABLE pom_transcripts ADD COLUMN org_id VARCHAR(64);
CREATE INDEX idx_transcripts_user ON pom_transcripts(user_id);

ALTER TABLE pom_state ADD COLUMN org_id VARCHAR(64);
CREATE INDEX idx_state_org ON pom_state(org_id);

ALTER TABLE pom_daily_notes ADD COLUMN user_id VARCHAR(64);
ALTER TABLE pom_daily_notes ADD COLUMN org_id VARCHAR(64);
CREATE INDEX idx_notes_user ON pom_daily_notes(user_id);
```

---

## 2. 系统架构

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      前端 (Web UI)                           │
│  React/Vue.js + TypeScript + Tailwind CSS                   │
│  ├── 登录/注册页                                              │
│  ├── Agent 管理面板 (列表/创建/编辑)                          │
│  ├── 对话界面 (Chat UI)                                       │
│  └── 用户设置 & 统计面板                                      │
└─────────────────────┬───────────────────────────────────────┘
                      │ REST API + WebSocket
┌─────────────────────▼───────────────────────────────────────┐
│                  API Gateway (Go)                            │
│  ├── JWT 认证中间件                                           │
│  ├── 权限验证                                                 │
│  ├── 请求限流                                                 │
│  └── 路由分发                                                 │
└─────────────┬───────────────────┬───────────────────────────┘
              │                   │
    ┌─────────▼─────────┐  ┌─────▼──────────────┐
    │  User Service     │  │  Agent Service     │
    │  - 注册/登录       │  │  - CRUD            │
    │  - 用户管理        │  │  - 对话处理         │
    │  - Token 签发      │  │  - 会话管理         │
    └─────────┬─────────┘  └─────┬──────────────┘
              │                   │
              └─────────┬─────────┘
                        │
              ┌─────────▼─────────┐
              │  PostgreSQL       │
              │  - 多租户数据隔离   │
              │  - pgvector 向量   │
              └───────────────────┘
```

### 2.2 组件说明

| 组件 | 职责 | 技术栈 |
|------|------|--------|
| 前端 UI | 用户交互界面 | React + TypeScript + Tailwind |
| API Gateway | API 路由、认证、限流 | Go + Chi Router |
| User Service | 用户管理、认证 | Go |
| Agent Service | Agent 生命周期管理 | Go + 现有 agent 包 |
| PostgreSQL | 数据持久化、向量搜索 | PostgreSQL + pgvector |
| Redis (可选) | 会话缓存、限流、配额 | Redis |

---

## 3. 认证与授权

### 3.1 JWT Token 方案

```go
// JWT Claims 结构
type Claims struct {
    UserID string `json:"user_id"`
    OrgID  string `json:"org_id"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    jwt.StandardClaims
}

// 生成 Token
func GenerateToken(user *User) (string, error) {
    claims := &Claims{
        UserID: user.ID,
        OrgID:  user.OrgID,
        Email:  user.Email,
        Role:   user.Role,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
            IssuedAt:  time.Now().Unix(),
            Issuer:    "pomclaw",
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(jwtSecret))
}

// 验证 Token
func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(
        tokenString,
        &Claims{},
        func(token *jwt.Token) (interface{}, error) {
            return []byte(jwtSecret), nil
        },
    )
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, errors.New("invalid token")
}
```

### 3.2 认证中间件

```go
// AuthMiddleware JWT 认证中间件
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. 从 Header 提取 token
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization header", 401)
            return
        }
        
        // 2. 验证 Bearer token
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            http.Error(w, "Invalid authorization header", 401)
            return
        }
        
        // 3. 解析和验证 JWT
        claims, err := ValidateToken(parts[1])
        if err != nil {
            http.Error(w, "Invalid token", 401)
            return
        }
        
        // 4. 将用户信息注入 context
        ctx := context.WithValue(r.Context(), "claims", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// GetClaims 从 context 中获取用户信息
func GetClaims(ctx context.Context) *Claims {
    if claims, ok := ctx.Value("claims").(*Claims); ok {
        return claims
    }
    return nil
}
```

### 3.3 RBAC 权限模型

```go
type Permission string

const (
    // Agent 权限
    PermAgentCreate  Permission = "agent:create"
    PermAgentRead    Permission = "agent:read"
    PermAgentUpdate  Permission = "agent:update"
    PermAgentDelete  Permission = "agent:delete"
    
    // 用户管理权限
    PermUserManage   Permission = "user:manage"
    PermUserInvite   Permission = "user:invite"
    
    // 组织权限
    PermOrgSettings  Permission = "org:settings"
    PermOrgBilling   Permission = "org:billing"
)

// 角色权限映射
var RolePermissions = map[string][]Permission{
    "admin": {
        PermAgentCreate, PermAgentRead, PermAgentUpdate, PermAgentDelete,
        PermUserManage, PermUserInvite,
        PermOrgSettings, PermOrgBilling,
    },
    "user": {
        PermAgentCreate, PermAgentRead, PermAgentUpdate, PermAgentDelete,
    },
    "viewer": {
        PermAgentRead,
    },
}

// CheckPermission 检查权限
func CheckPermission(role string, perm Permission) bool {
    perms, ok := RolePermissions[role]
    if !ok {
        return false
    }
    
    for _, p := range perms {
        if p == perm {
            return true
        }
    }
    return false
}

// RequirePermission 权限检查中间件
func RequirePermission(perm Permission) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := GetClaims(r.Context())
            if claims == nil {
                http.Error(w, "Unauthorized", 401)
                return
            }
            
            if !CheckPermission(claims.Role, perm) {
                http.Error(w, "Permission denied", 403)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## 4. API 设计

### 4.1 RESTful API 端点

#### 认证相关
```
POST   /api/v1/auth/register          - 注册
POST   /api/v1/auth/login             - 登录
POST   /api/v1/auth/logout            - 登出
POST   /api/v1/auth/refresh           - 刷新 token
GET    /api/v1/auth/me                - 获取当前用户信息
POST   /api/v1/auth/forgot-password   - 忘记密码
POST   /api/v1/auth/reset-password    - 重置密码
```

#### 用户管理
```
GET    /api/v1/users                  - 获取用户列表 (admin only)
GET    /api/v1/users/:id              - 获取用户详情
PUT    /api/v1/users/:id              - 更新用户信息
DELETE /api/v1/users/:id              - 删除用户 (admin only)
POST   /api/v1/users/:id/suspend      - 暂停用户 (admin only)
POST   /api/v1/users/:id/activate     - 激活用户 (admin only)
```

#### 组织管理
```
GET    /api/v1/orgs/:id               - 获取组织信息
PUT    /api/v1/orgs/:id               - 更新组织信息 (admin only)
GET    /api/v1/orgs/:id/members       - 获取组织成员
POST   /api/v1/orgs/:id/invite        - 邀请成员 (admin only)
DELETE /api/v1/orgs/:id/members/:uid  - 移除成员 (admin only)
```

#### Agent 管理
```
GET    /api/v1/agents                 - 获取我的 agents 列表
POST   /api/v1/agents                 - 创建新 agent
GET    /api/v1/agents/:id             - 获取 agent 详情
PUT    /api/v1/agents/:id             - 更新 agent 配置
DELETE /api/v1/agents/:id             - 删除 agent
POST   /api/v1/agents/:id/start       - 启动 agent
POST   /api/v1/agents/:id/stop        - 停止 agent
POST   /api/v1/agents/:id/duplicate   - 复制 agent
```

#### 对话相关
```
POST   /api/v1/agents/:id/chat        - 发送消息 (REST)
WS     /api/v1/agents/:id/ws          - WebSocket 实时对话
GET    /api/v1/agents/:id/sessions    - 获取对话历史列表
GET    /api/v1/agents/:id/sessions/:sid  - 获取特定会话详情
DELETE /api/v1/agents/:id/sessions/:sid  - 删除会话
POST   /api/v1/agents/:id/sessions/:sid/clear - 清空会话
```

#### 统计分析
```
GET    /api/v1/agents/:id/usage       - Agent 使用统计
GET    /api/v1/users/:id/usage        - 用户使用统计
GET    /api/v1/orgs/:id/usage         - 组织使用统计
GET    /api/v1/analytics/dashboard    - 控制台概览数据
```

### 4.2 API 请求/响应示例

#### 注册
```json
// POST /api/v1/auth/register
Request:
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "username": "张三",
  "org_name": "我的公司"
}

Response: 201 Created
{
  "user_id": "user_abc123",
  "email": "user@example.com",
  "username": "张三",
  "org_id": "org_xyz789",
  "role": "admin",
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

#### 登录
```json
// POST /api/v1/auth/login
Request:
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}

Response: 200 OK
{
  "user_id": "user_abc123",
  "email": "user@example.com",
  "username": "张三",
  "org_id": "org_xyz789",
  "role": "admin",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2026-04-19T10:00:00Z"
}
```

#### 创建 Agent
```json
// POST /api/v1/agents
Request:
{
  "name": "我的客服助手",
  "description": "处理客户咨询的智能助手",
  "provider": "openai",
  "model": "gpt-4",
  "system_prompt": "你是一个专业的客服助手，负责解答客户问题...",
  "temperature": 0.7,
  "max_tokens": 4096,
  "tools_enabled": ["web_search", "calculator"]
}

Response: 201 Created
{
  "agent_id": "agent_def456",
  "name": "我的客服助手",
  "description": "处理客户咨询的智能助手",
  "provider": "openai",
  "model": "gpt-4",
  "status": "active",
  "created_at": "2026-04-18T10:00:00Z",
  "endpoint": {
    "rest": "/api/v1/agents/agent_def456/chat",
    "websocket": "ws://api.example.com/api/v1/agents/agent_def456/ws"
  }
}
```

#### 获取 Agents 列表
```json
// GET /api/v1/agents?page=1&limit=20&status=active
Response: 200 OK
{
  "agents": [
    {
      "agent_id": "agent_def456",
      "name": "我的客服助手",
      "description": "处理客户咨询",
      "model": "gpt-4",
      "status": "active",
      "created_at": "2026-04-18T10:00:00Z",
      "last_used": "2026-04-18T12:30:00Z",
      "total_requests": 156,
      "total_tokens": 45600
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 5,
    "total_pages": 1
  }
}
```

#### 发送消息
```json
// POST /api/v1/agents/agent_def456/chat
Request:
{
  "message": "你好，请帮我查询订单状态",
  "session_id": "session_ghi789",  // optional，不传则创建新会话
  "stream": false
}

Response: 200 OK
{
  "agent_id": "agent_def456",
  "session_id": "session_ghi789",
  "message_id": "msg_jkl012",
  "content": "您好！请提供您的订单号，我来帮您查询...",
  "tokens_used": {
    "input": 12,
    "output": 25,
    "total": 37
  },
  "timestamp": "2026-04-18T10:05:00Z"
}
```

#### WebSocket 消息格式
```json
// Client -> Server
{
  "type": "chat",
  "message": "你好，请帮我查询订单状态",
  "session_id": "session_ghi789"
}

// Server -> Client (streaming)
{
  "type": "stream",
  "content": "您好",
  "done": false
}
{
  "type": "stream",
  "content": "！",
  "done": false
}
{
  "type": "done",
  "message_id": "msg_jkl012",
  "tokens_used": {
    "input": 12,
    "output": 25,
    "total": 37
  }
}
```

#### 使用统计
```json
// GET /api/v1/agents/agent_def456/usage?start_date=2026-04-01&end_date=2026-04-18
Response: 200 OK
{
  "agent_id": "agent_def456",
  "period": {
    "start": "2026-04-01",
    "end": "2026-04-18"
  },
  "summary": {
    "total_requests": 1250,
    "total_tokens": 456789,
    "total_cost": 23.45,
    "avg_response_time_ms": 1234
  },
  "daily_usage": [
    {
      "date": "2026-04-18",
      "requests": 89,
      "tokens": 25678,
      "cost": 1.32
    }
  ]
}
```

---

## 5. 前端设计

### 5.1 技术栈

```
前端框架: React 18 + TypeScript
UI 组件库: Ant Design / Shadcn UI
状态管理: Zustand / Redux Toolkit
API 客户端: Axios + React Query (TanStack Query)
实时通信: WebSocket / Socket.io-client
路由: React Router v6
构建工具: Vite
样式方案: Tailwind CSS
图表库: Recharts / Chart.js
```

### 5.2 目录结构

```
frontend/
├── public/
│   └── favicon.ico
├── src/
│   ├── main.tsx              - 入口文件
│   ├── App.tsx               - 根组件
│   ├── routes.tsx            - 路由配置
│   │
│   ├── pages/                - 页面组件
│   │   ├── auth/
│   │   │   ├── Login.tsx
│   │   │   ├── Register.tsx
│   │   │   └── ForgotPassword.tsx
│   │   ├── dashboard/
│   │   │   └── Dashboard.tsx
│   │   ├── agents/
│   │   │   ├── AgentList.tsx
│   │   │   ├── AgentCreate.tsx
│   │   │   ├── AgentEdit.tsx
│   │   │   └── AgentChat.tsx
│   │   ├── users/
│   │   │   └── UserManagement.tsx
│   │   ├── settings/
│   │   │   └── Settings.tsx
│   │   └── analytics/
│   │       └── Analytics.tsx
│   │
│   ├── components/           - 通用组件
│   │   ├── layout/
│   │   │   ├── Header.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   └── Layout.tsx
│   │   ├── chat/
│   │   │   ├── ChatWindow.tsx
│   │   │   ├── MessageList.tsx
│   │   │   ├── MessageBubble.tsx
│   │   │   └── InputArea.tsx
│   │   ├── agent/
│   │   │   ├── AgentCard.tsx
│   │   │   └── AgentForm.tsx
│   │   └── common/
│   │       ├── Button.tsx
│   │       ├── Modal.tsx
│   │       └── Loading.tsx
│   │
│   ├── hooks/                - 自定义 Hooks
│   │   ├── useAuth.ts
│   │   ├── useAgent.ts
│   │   ├── useChat.ts
│   │   └── useWebSocket.ts
│   │
│   ├── services/             - API 服务
│   │   ├── api.ts           - Axios 配置
│   │   ├── authService.ts
│   │   ├── agentService.ts
│   │   ├── chatService.ts
│   │   └── userService.ts
│   │
│   ├── stores/               - 状态管理
│   │   ├── authStore.ts
│   │   ├── agentStore.ts
│   │   └── chatStore.ts
│   │
│   ├── types/                - TypeScript 类型
│   │   ├── auth.ts
│   │   ├── agent.ts
│   │   └── chat.ts
│   │
│   ├── utils/                - 工具函数
│   │   ├── token.ts
│   │   ├── format.ts
│   │   └── websocket.ts
│   │
│   └── styles/               - 全局样式
│       └── globals.css
│
├── package.json
├── tsconfig.json
├── vite.config.ts
└── tailwind.config.js
```

### 5.3 关键组件示例

#### AgentChat - 对话界面
```typescript
// src/pages/agents/AgentChat.tsx
import { useState, useEffect, useRef } from 'react';
import { useParams } from 'react-router-dom';
import { useChat } from '@/hooks/useChat';
import { MessageList } from '@/components/chat/MessageList';
import { InputArea } from '@/components/chat/InputArea';

export default function AgentChat() {
  const { agentId } = useParams<{ agentId: string }>();
  const { 
    messages, 
    sendMessage, 
    isLoading,
    sessionId 
  } = useChat(agentId);
  
  const [input, setInput] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // 自动滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = async () => {
    if (!input.trim() || isLoading) return;
    
    await sendMessage(input);
    setInput('');
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="flex flex-col h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b px-6 py-4 shadow-sm">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-semibold">
            {agent?.name || 'Agent Chat'}
          </h1>
          <div className="flex gap-2">
            <button className="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded">
              清空会话
            </button>
            <button className="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded">
              设置
            </button>
          </div>
        </div>
      </header>

      {/* Messages Area */}
      <div className="flex-1 overflow-y-auto px-6 py-4">
        <MessageList messages={messages} />
        <div ref={messagesEndRef} />
      </div>

      {/* Input Area */}
      <div className="bg-white border-t px-6 py-4">
        <InputArea
          value={input}
          onChange={setInput}
          onSend={handleSend}
          onKeyPress={handleKeyPress}
          disabled={isLoading}
          placeholder="输入消息... (Enter 发送, Shift+Enter 换行)"
        />
      </div>
    </div>
  );
}
```

#### useChat - 对话 Hook
```typescript
// src/hooks/useChat.ts
import { useState, useEffect, useCallback } from 'react';
import { useWebSocket } from './useWebSocket';
import { chatService } from '@/services/chatService';
import type { Message } from '@/types/chat';

export function useChat(agentId: string) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [sessionId, setSessionId] = useState<string>();

  const { ws, connect, disconnect, send } = useWebSocket(
    `/api/v1/agents/${agentId}/ws`
  );

  // 加载历史消息
  useEffect(() => {
    if (!agentId) return;
    
    chatService.getSessions(agentId).then(sessions => {
      if (sessions.length > 0) {
        setSessionId(sessions[0].session_id);
        setMessages(sessions[0].messages || []);
      }
    });
  }, [agentId]);

  // WebSocket 消息处理
  useEffect(() => {
    if (!ws) return;

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      if (data.type === 'stream') {
        // 流式响应：追加内容
        setMessages(prev => {
          const lastMsg = prev[prev.length - 1];
          if (lastMsg?.role === 'assistant' && !lastMsg.done) {
            return [
              ...prev.slice(0, -1),
              { ...lastMsg, content: lastMsg.content + data.content }
            ];
          } else {
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
        setMessages(prev => {
          const lastMsg = prev[prev.length - 1];
          return [
            ...prev.slice(0, -1),
            { ...lastMsg, done: true, message_id: data.message_id }
          ];
        });
        setIsLoading(false);
      } else if (data.type === 'error') {
        console.error('WebSocket error:', data.message);
        setIsLoading(false);
      }
    };
  }, [ws]);

  // 发送消息
  const sendMessage = useCallback(async (content: string) => {
    if (!agentId || !content.trim()) return;

    setIsLoading(true);

    // 添加用户消息
    const userMessage: Message = {
      role: 'user',
      content,
      timestamp: new Date()
    };
    setMessages(prev => [...prev, userMessage]);

    try {
      if (ws && ws.readyState === WebSocket.OPEN) {
        // WebSocket 模式
        send({
          type: 'chat',
          message: content,
          session_id: sessionId
        });
      } else {
        // REST 模式
        const response = await chatService.sendMessage(agentId, {
          message: content,
          session_id: sessionId,
          stream: false
        });

        setSessionId(response.session_id);
        setMessages(prev => [...prev, {
          role: 'assistant',
          content: response.content,
          message_id: response.message_id,
          done: true,
          timestamp: new Date()
        }]);
        setIsLoading(false);
      }
    } catch (error) {
      console.error('Failed to send message:', error);
      setIsLoading(false);
    }
  }, [agentId, sessionId, ws, send]);

  return {
    messages,
    sendMessage,
    isLoading,
    sessionId
  };
}
```

#### useWebSocket - WebSocket Hook
```typescript
// src/hooks/useWebSocket.ts
import { useState, useEffect, useCallback, useRef } from 'react';
import { getToken } from '@/utils/token';

export function useWebSocket(url: string) {
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>();

  const connect = useCallback(() => {
    const token = getToken();
    if (!token) return;

    const wsUrl = `${url}?token=${token}`;
    const websocket = new WebSocket(wsUrl);

    websocket.onopen = () => {
      console.log('WebSocket connected');
      setIsConnected(true);
    };

    websocket.onclose = () => {
      console.log('WebSocket disconnected');
      setIsConnected(false);
      
      // 自动重连
      reconnectTimeoutRef.current = setTimeout(() => {
        console.log('Reconnecting WebSocket...');
        connect();
      }, 3000);
    };

    websocket.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    setWs(websocket);
  }, [url]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    ws?.close();
    setWs(null);
    setIsConnected(false);
  }, [ws]);

  const send = useCallback((data: any) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(data));
    }
  }, [ws]);

  useEffect(() => {
    connect();
    return () => disconnect();
  }, []);

  return {
    ws,
    isConnected,
    connect,
    disconnect,
    send
  };
}
```

---

## 6. 后端实现

### 6.1 目录结构

```
cmd/pomclaw/
├── main.go
└── api/
    └── server.go           - HTTP 服务器入口

pkg/
├── api/
│   ├── router.go           - 路由定义
│   ├── middleware/
│   │   ├── auth.go         - JWT 认证中间件
│   │   ├── cors.go         - CORS 中间件
│   │   ├── logger.go       - 日志中间件
│   │   ├── ratelimit.go    - 限流中间件
│   │   └── recovery.go     - 错误恢复中间件
│   │
│   └── handlers/
│       ├── auth.go         - 认证处理器
│       ├── user.go         - 用户管理处理器
│       ├── org.go          - 组织管理处理器
│       ├── agent.go        - Agent CRUD 处理器
│       ├── chat.go         - 对话处理器
│       └── analytics.go    - 统计分析处理器
│
├── models/
│   ├── user.go
│   ├── organization.go
│   ├── agent_config.go
│   └── usage.go
│
├── services/
│   ├── auth_service.go     - 认证服务
│   ├── user_service.go     - 用户服务
│   ├── org_service.go      - 组织服务
│   ├── agent_service.go    - Agent 服务
│   └── quota_service.go    - 配额服务
│
├── repositories/
│   ├── user_repo.go
│   ├── org_repo.go
│   ├── agent_repo.go
│   └── usage_repo.go
│
└── utils/
    ├── jwt.go              - JWT 工具
    ├── password.go         - 密码加密
    └── validator.go        - 输入验证
```

### 6.2 核心代码示例

#### Router - 路由配置
```go
// pkg/api/router.go
package api

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/pomclaw/pomclaw/pkg/api/handlers"
    mw "github.com/pomclaw/pomclaw/pkg/api/middleware"
)

func NewRouter(
    authHandler *handlers.AuthHandler,
    userHandler *handlers.UserHandler,
    agentHandler *handlers.AgentHandler,
    chatHandler *handlers.ChatHandler,
) *chi.Mux {
    r := chi.NewRouter()

    // 全局中间件
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(mw.Logger)
    r.Use(middleware.Recoverer)
    r.Use(mw.CORS)

    // API v1
    r.Route("/api/v1", func(r chi.Router) {
        // 公开端点
        r.Route("/auth", func(r chi.Router) {
            r.Post("/register", authHandler.Register)
            r.Post("/login", authHandler.Login)
            r.Post("/forgot-password", authHandler.ForgotPassword)
            r.Post("/reset-password", authHandler.ResetPassword)
        })

        // 需要认证的端点
        r.Group(func(r chi.Router) {
            r.Use(mw.AuthMiddleware)

            // 认证相关
            r.Get("/auth/me", authHandler.Me)
            r.Post("/auth/logout", authHandler.Logout)
            r.Post("/auth/refresh", authHandler.Refresh)

            // 用户管理
            r.Route("/users", func(r chi.Router) {
                r.Get("/", userHandler.List)
                r.Get("/{userId}", userHandler.Get)
                r.Put("/{userId}", userHandler.Update)
                
                r.With(mw.RequirePermission(PermUserManage)).
                    Delete("/{userId}", userHandler.Delete)
            })

            // Agent 管理
            r.Route("/agents", func(r chi.Router) {
                r.Get("/", agentHandler.List)
                r.Post("/", agentHandler.Create)
                r.Get("/{agentId}", agentHandler.Get)
                r.Put("/{agentId}", agentHandler.Update)
                r.Delete("/{agentId}", agentHandler.Delete)
                r.Post("/{agentId}/start", agentHandler.Start)
                r.Post("/{agentId}/stop", agentHandler.Stop)

                // 对话相关
                r.Post("/{agentId}/chat", chatHandler.Chat)
                r.HandleFunc("/{agentId}/ws", chatHandler.WebSocket)
                
                // 会话管理
                r.Get("/{agentId}/sessions", chatHandler.ListSessions)
                r.Get("/{agentId}/sessions/{sessionId}", chatHandler.GetSession)
                r.Delete("/{agentId}/sessions/{sessionId}", chatHandler.DeleteSession)

                // 使用统计
                r.Get("/{agentId}/usage", agentHandler.GetUsage)
            })
        })
    })

    return r
}
```

#### Agent Handler - Agent 处理器
```go
// pkg/api/handlers/agent.go
package handlers

import (
    "encoding/json"
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/pomclaw/pomclaw/pkg/services"
    "github.com/pomclaw/pomclaw/pkg/models"
)

type AgentHandler struct {
    agentService *services.AgentService
    quotaService *services.QuotaService
}

func NewAgentHandler(
    agentService *services.AgentService,
    quotaService *services.QuotaService,
) *AgentHandler {
    return &AgentHandler{
        agentService: agentService,
        quotaService: quotaService,
    }
}

// Create 创建新 Agent
func (h *AgentHandler) Create(w http.ResponseWriter, r *http.Request) {
    // 1. 从 JWT 获取用户信息
    claims := GetClaims(r.Context())
    if claims == nil {
        RespondError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    // 2. 解析请求
    var req struct {
        Name         string   `json:"name" validate:"required,min=1,max=100"`
        Description  string   `json:"description" validate:"max=500"`
        Provider     string   `json:"provider" validate:"required,oneof=openai anthropic"`
        Model        string   `json:"model" validate:"required"`
        SystemPrompt string   `json:"system_prompt"`
        Temperature  float64  `json:"temperature" validate:"gte=0,lte=2"`
        MaxTokens    int      `json:"max_tokens" validate:"gte=1,lte=32000"`
        ToolsEnabled []string `json:"tools_enabled"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        RespondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // 3. 验证输入
    if err := Validate.Struct(req); err != nil {
        RespondError(w, http.StatusBadRequest, err.Error())
        return
    }

    // 4. 检查配额
    canCreate, err := h.quotaService.CanCreateAgent(claims.UserID, claims.OrgID)
    if err != nil {
        RespondError(w, http.StatusInternalServerError, "Failed to check quota")
        return
    }
    if !canCreate {
        RespondError(w, http.StatusForbidden, "Agent quota exceeded")
        return
    }

    // 5. 创建 Agent
    agent := &models.AgentConfig{
        UserID:       claims.UserID,
        OrgID:        claims.OrgID,
        Name:         req.Name,
        Description:  req.Description,
        Provider:     req.Provider,
        Model:        req.Model,
        SystemPrompt: req.SystemPrompt,
        Temperature:  req.Temperature,
        MaxTokens:    req.MaxTokens,
        ToolsEnabled: req.ToolsEnabled,
        Status:       "active",
    }

    if err := h.agentService.Create(agent); err != nil {
        RespondError(w, http.StatusInternalServerError, "Failed to create agent")
        return
    }

    // 6. 返回响应
    RespondJSON(w, http.StatusCreated, agent)
}

// List 获取用户的 Agents
func (h *AgentHandler) List(w http.ResponseWriter, r *http.Request) {
    claims := GetClaims(r.Context())
    if claims == nil {
        RespondError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    // 解析查询参数
    page := GetQueryInt(r, "page", 1)
    limit := GetQueryInt(r, "limit", 20)
    status := r.URL.Query().Get("status")

    agents, total, err := h.agentService.ListByUser(
        claims.UserID,
        page,
        limit,
        status,
    )
    if err != nil {
        RespondError(w, http.StatusInternalServerError, "Failed to list agents")
        return
    }

    RespondJSON(w, http.StatusOK, map[string]interface{}{
        "agents": agents,
        "pagination": map[string]interface{}{
            "page":        page,
            "limit":       limit,
            "total":       total,
            "total_pages": (total + limit - 1) / limit,
        },
    })
}

// Get 获取单个 Agent
func (h *AgentHandler) Get(w http.ResponseWriter, r *http.Request) {
    claims := GetClaims(r.Context())
    agentID := chi.URLParam(r, "agentId")

    agent, err := h.agentService.Get(agentID)
    if err != nil {
        RespondError(w, http.StatusNotFound, "Agent not found")
        return
    }

    // 验证所有权
    if agent.UserID != claims.UserID {
        // 检查是否是同组织的管理员
        if claims.Role != "admin" || agent.OrgID != claims.OrgID {
            RespondError(w, http.StatusForbidden, "Permission denied")
            return
        }
    }

    RespondJSON(w, http.StatusOK, agent)
}

// Update 更新 Agent 配置
func (h *AgentHandler) Update(w http.ResponseWriter, r *http.Request) {
    claims := GetClaims(r.Context())
    agentID := chi.URLParam(r, "agentId")

    // 验证所有权
    if err := h.agentService.VerifyOwnership(agentID, claims.UserID); err != nil {
        RespondError(w, http.StatusForbidden, "Permission denied")
        return
    }

    var req struct {
        Name         *string   `json:"name,omitempty"`
        Description  *string   `json:"description,omitempty"`
        SystemPrompt *string   `json:"system_prompt,omitempty"`
        Temperature  *float64  `json:"temperature,omitempty"`
        MaxTokens    *int      `json:"max_tokens,omitempty"`
        ToolsEnabled *[]string `json:"tools_enabled,omitempty"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        RespondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    agent, err := h.agentService.Update(agentID, req)
    if err != nil {
        RespondError(w, http.StatusInternalServerError, "Failed to update agent")
        return
    }

    RespondJSON(w, http.StatusOK, agent)
}

// Delete 删除 Agent
func (h *AgentHandler) Delete(w http.ResponseWriter, r *http.Request) {
    claims := GetClaims(r.Context())
    agentID := chi.URLParam(r, "agentId")

    // 验证所有权
    if err := h.agentService.VerifyOwnership(agentID, claims.UserID); err != nil {
        RespondError(w, http.StatusForbidden, "Permission denied")
        return
    }

    if err := h.agentService.Delete(agentID); err != nil {
        RespondError(w, http.StatusInternalServerError, "Failed to delete agent")
        return
    }

    RespondJSON(w, http.StatusOK, map[string]string{
        "message": "Agent deleted successfully",
    })
}
```

#### Auth Handler - 认证处理器
```go
// pkg/api/handlers/auth.go
package handlers

import (
    "encoding/json"
    "net/http"
    "github.com/pomclaw/pomclaw/pkg/services"
    "github.com/pomclaw/pomclaw/pkg/utils"
)

type AuthHandler struct {
    authService *services.AuthService
    userService *services.UserService
}

func NewAuthHandler(
    authService *services.AuthService,
    userService *services.UserService,
) *AuthHandler {
    return &AuthHandler{
        authService: authService,
        userService: userService,
    }
}

// Register 用户注册
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required,min=8"`
        Username string `json:"username" validate:"required,min=2,max=50"`
        OrgName  string `json:"org_name" validate:"required,min=2,max=100"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        RespondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    if err := Validate.Struct(req); err != nil {
        RespondError(w, http.StatusBadRequest, err.Error())
        return
    }

    // 检查邮箱是否已存在
    exists, _ := h.userService.EmailExists(req.Email)
    if exists {
        RespondError(w, http.StatusConflict, "Email already registered")
        return
    }

    // 创建组织和用户
    user, token, err := h.authService.Register(
        req.Email,
        req.Password,
        req.Username,
        req.OrgName,
    )
    if err != nil {
        RespondError(w, http.StatusInternalServerError, "Registration failed")
        return
    }

    RespondJSON(w, http.StatusCreated, map[string]interface{}{
        "user_id":  user.ID,
        "email":    user.Email,
        "username": user.Username,
        "org_id":   user.OrgID,
        "role":     user.Role,
        "token":    token,
    })
}

// Login 用户登录
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        RespondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    user, token, err := h.authService.Login(req.Email, req.Password)
    if err != nil {
        RespondError(w, http.StatusUnauthorized, "Invalid credentials")
        return
    }

    RespondJSON(w, http.StatusOK, map[string]interface{}{
        "user_id":    user.ID,
        "email":      user.Email,
        "username":   user.Username,
        "org_id":     user.OrgID,
        "role":       user.Role,
        "token":      token,
        "expires_at": time.Now().Add(24 * time.Hour),
    })
}

// Me 获取当前用户信息
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
    claims := GetClaims(r.Context())
    if claims == nil {
        RespondError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    user, err := h.userService.GetByID(claims.UserID)
    if err != nil {
        RespondError(w, http.StatusNotFound, "User not found")
        return
    }

    RespondJSON(w, http.StatusOK, user)
}
```

---

## 7. 安全性考虑

### 7.1 数据隔离

所有数据库查询必须加上租户过滤：

```go
// 错误示例 - 缺少租户过滤
func (r *AgentRepository) Get(agentID string) (*AgentConfig, error) {
    var agent AgentConfig
    err := r.db.Where("agent_id = ?", agentID).First(&agent).Error
    return &agent, err
}

// 正确示例 - 包含用户/组织验证
func (r *AgentRepository) GetByUser(agentID, userID string) (*AgentConfig, error) {
    var agent AgentConfig
    err := r.db.Where("agent_id = ? AND user_id = ?", agentID, userID).
        First(&agent).Error
    return &agent, err
}

// 列表查询也必须过滤
func (r *AgentRepository) ListByUser(userID string) ([]*AgentConfig, error) {
    var agents []*AgentConfig
    err := r.db.Where("user_id = ? AND status != ?", userID, "deleted").
        Find(&agents).Error
    return agents, err
}
```

### 7.2 所有权验证

```go
// 服务层验证
func (s *AgentService) VerifyOwnership(agentID, userID string) error {
    agent, err := s.repo.Get(agentID)
    if err != nil {
        return err
    }
    if agent.UserID != userID {
        return errors.New("permission denied")
    }
    return nil
}

// 在所有修改操作前验证
func (h *AgentHandler) Update(w http.ResponseWriter, r *http.Request) {
    claims := GetClaims(r.Context())
    agentID := chi.URLParam(r, "agentId")
    
    // 先验证所有权
    if err := h.agentService.VerifyOwnership(agentID, claims.UserID); err != nil {
        RespondError(w, http.StatusForbidden, "Permission denied")
        return
    }
    
    // 继续更新逻辑...
}
```

### 7.3 密码安全

```go
// pkg/utils/password.go
package utils

import "golang.org/x/crypto/bcrypt"

const bcryptCost = 12

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
    return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// 强密码策略验证
func ValidatePassword(password string) error {
    if len(password) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    
    hasUpper := false
    hasLower := false
    hasDigit := false
    
    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsDigit(char):
            hasDigit = true
        }
    }
    
    if !hasUpper || !hasLower || !hasDigit {
        return errors.New("password must contain uppercase, lowercase, and digit")
    }
    
    return nil
}
```

### 7.4 SQL 注入防护

```go
// 始终使用参数化查询
// ✅ 正确
db.Where("user_id = ? AND status = ?", userID, status).Find(&agents)

// ❌ 错误 - 容易 SQL 注入
db.Where(fmt.Sprintf("user_id = '%s' AND status = '%s'", userID, status)).Find(&agents)
```

### 7.5 XSS 防护

```go
import "html"

// 输出到前端前转义
func SanitizeHTML(input string) string {
    return html.EscapeString(input)
}

// 或使用白名单 Markdown
import "github.com/microcosm-cc/bluemonday"

func SanitizeMarkdown(input string) string {
    p := bluemonday.UGCPolicy()
    return p.Sanitize(input)
}
```

### 7.6 CORS 配置

```go
// pkg/api/middleware/cors.go
package middleware

import (
    "net/http"
    "github.com/rs/cors"
)

func CORS(next http.Handler) http.Handler {
    c := cors.New(cors.Options{
        AllowedOrigins: []string{
            "http://localhost:3000",  // 开发环境
            "https://app.example.com", // 生产环境
        },
        AllowedMethods: []string{
            http.MethodGet,
            http.MethodPost,
            http.MethodPut,
            http.MethodPatch,
            http.MethodDelete,
            http.MethodOptions,
        },
        AllowedHeaders: []string{
            "Authorization",
            "Content-Type",
            "X-Request-ID",
        },
        ExposedHeaders: []string{
            "X-Request-ID",
        },
        AllowCredentials: true,
        MaxAge:           300,
    })
    
    return c.Handler(next)
}
```

### 7.7 限流

```go
// pkg/api/middleware/ratelimit.go
package middleware

import (
    "net/http"
    "golang.org/x/time/rate"
    "sync"
)

type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
    return &RateLimiter{
        limiters: make(map[string]*rate.Limiter),
        rate:     r,
        burst:    b,
    }
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    limiter, exists := rl.limiters[key]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.limiters[key] = limiter
    }
    
    return limiter
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 从 context 获取用户信息作为限流 key
        claims := GetClaims(r.Context())
        key := r.RemoteAddr // 默认使用 IP
        if claims != nil {
            key = claims.UserID // 认证用户使用 user_id
        }
        
        limiter := rl.getLimiter(key)
        if !limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

---

## 8. 可扩展性

### 8.1 水平扩展架构

```
                    ┌─────────────┐
                    │  Nginx LB   │
                    └──────┬──────┘
                           │
       ┌───────────────────┼───────────────────┐
       │                   │                   │
   ┌───▼───┐          ┌───▼───┐          ┌───▼───┐
   │API-1  │          │API-2  │          │API-3  │
   └───┬───┘          └───┬───┘          └───┬───┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
              ┌────────────┴────────────┐
              │                         │
       ┌──────▼──────┐         ┌───────▼──────┐
       │ PostgreSQL  │         │    Redis     │
       │  (Primary)  │         │  (Cache &    │
       │      +      │         │   Session)   │
       │  (Replicas) │         └──────────────┘
       └─────────────┘
```

### 8.2 数据库优化

```sql
-- 添加索引
CREATE INDEX CONCURRENTLY idx_agents_user_status 
ON pom_agent_configs(user_id, status);

CREATE INDEX CONCURRENTLY idx_sessions_agent_updated 
ON pom_sessions(agent_id, updated_at DESC);

CREATE INDEX CONCURRENTLY idx_usage_org_date 
ON pom_agent_usage(org_id, date DESC);

-- 分区表（针对大数据量）
CREATE TABLE pom_agent_usage_2026_04 PARTITION OF pom_agent_usage
FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

-- 物化视图（统计查询优化）
CREATE MATERIALIZED VIEW mv_org_daily_stats AS
SELECT 
    org_id,
    date,
    SUM(tokens_total) as total_tokens,
    SUM(request_count) as total_requests,
    SUM(cost) as total_cost
FROM pom_agent_usage
GROUP BY org_id, date;

CREATE UNIQUE INDEX ON mv_org_daily_stats(org_id, date);
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_org_daily_stats;
```

### 8.3 缓存策略

```go
// Redis 缓存 Agent 配置
func (s *AgentService) Get(agentID string) (*AgentConfig, error) {
    // 1. 尝试从缓存读取
    cacheKey := fmt.Sprintf("agent:%s", agentID)
    cached, err := s.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var agent AgentConfig
        json.Unmarshal([]byte(cached), &agent)
        return &agent, nil
    }
    
    // 2. 缓存未命中，从数据库读取
    agent, err := s.repo.Get(agentID)
    if err != nil {
        return nil, err
    }
    
    // 3. 写入缓存
    data, _ := json.Marshal(agent)
    s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
    
    return agent, nil
}

// 更新时清除缓存
func (s *AgentService) Update(agentID string, updates map[string]interface{}) error {
    if err := s.repo.Update(agentID, updates); err != nil {
        return err
    }
    
    // 清除缓存
    cacheKey := fmt.Sprintf("agent:%s", agentID)
    s.redis.Del(ctx, cacheKey)
    
    return nil
}
```

### 8.4 配额管理

```go
// pkg/services/quota_service.go
package services

import (
    "context"
    "fmt"
    "time"
    "github.com/go-redis/redis/v8"
)

type QuotaService struct {
    redis *redis.Client
    db    *gorm.DB
}

// CheckAndDecrement 检查并扣减配额
func (s *QuotaService) CheckAndDecrement(
    ctx context.Context,
    orgID string,
    tokens int,
) error {
    // 月度配额 key
    key := fmt.Sprintf("quota:%s:%s", orgID, time.Now().Format("2006-01"))
    
    // 使用 Lua 脚本保证原子性
    script := `
    local used = redis.call('GET', KEYS[1])
    if not used then
        used = 0
    end
    
    local limit = tonumber(ARGV[1])
    local increment = tonumber(ARGV[2])
    local new_used = tonumber(used) + increment
    
    if new_used > limit then
        return -1
    end
    
    redis.call('INCRBY', KEYS[1], increment)
    redis.call('EXPIRE', KEYS[1], 2592000)  -- 30天过期
    return new_used
    `
    
    limit := s.getOrgLimit(orgID)
    result, err := s.redis.Eval(ctx, script, []string{key}, limit, tokens).Result()
    if err != nil {
        return err
    }
    
    if result.(int64) == -1 {
        return errors.New("quota exceeded")
    }
    
    return nil
}

// GetUsage 获取当前使用量
func (s *QuotaService) GetUsage(ctx context.Context, orgID string) (int, int, error) {
    key := fmt.Sprintf("quota:%s:%s", orgID, time.Now().Format("2006-01"))
    
    used, err := s.redis.Get(ctx, key).Int()
    if err == redis.Nil {
        used = 0
    } else if err != nil {
        return 0, 0, err
    }
    
    limit := s.getOrgLimit(orgID)
    return used, limit, nil
}
```

---

## 9. 监控与日志

### 9.1 结构化日志

```go
// pkg/api/middleware/logger.go
package middleware

import (
    "net/http"
    "time"
    "go.uber.org/zap"
)

func Logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Wrap ResponseWriter to capture status code
        ww := &responseWriter{ResponseWriter: w, statusCode: 200}
        
        // Process request
        next.ServeHTTP(ww, r)
        
        // Log request
        duration := time.Since(start)
        
        logger.Info("http_request",
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
            zap.Int("status", ww.statusCode),
            zap.Duration("duration", duration),
            zap.String("user_id", getUserID(r)),
            zap.String("ip", r.RemoteAddr),
            zap.String("user_agent", r.UserAgent()),
        )
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}
```

### 9.2 Prometheus Metrics

```go
// pkg/metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP 请求指标
    RequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "api_request_duration_seconds",
            Help:    "HTTP request latencies in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path", "status"},
    )
    
    RequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    // Agent 指标
    AgentsTotal = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "agents_total",
            Help: "Total number of agents",
        },
        []string{"status"},
    )
    
    TokensUsed = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "tokens_used_total",
            Help: "Total tokens consumed",
        },
        []string{"user_id", "agent_id", "model"},
    )
    
    // 数据库指标
    DBQueriesTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "db_queries_total",
            Help: "Total number of database queries",
        },
        []string{"operation", "table"},
    )
    
    DBQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "db_query_duration_seconds",
            Help:    "Database query latencies in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"operation", "table"},
    )
)

// RecordRequest 记录 HTTP 请求
func RecordRequest(method, path string, status int, duration time.Duration) {
    RequestDuration.WithLabelValues(method, path, strconv.Itoa(status)).
        Observe(duration.Seconds())
    RequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
}

// RecordTokenUsage 记录 token 使用
func RecordTokenUsage(userID, agentID, model string, tokens int) {
    TokensUsed.WithLabelValues(userID, agentID, model).Add(float64(tokens))
}
```

### 9.3 健康检查

```go
// pkg/health/health.go
package health

type HealthCheck struct {
    db    *gorm.DB
    redis *redis.Client
}

func NewHealthCheck(db *gorm.DB, redis *redis.Client) *HealthCheck {
    return &HealthCheck{db: db, redis: redis}
}

func (h *HealthCheck) Check(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    result := map[string]interface{}{
        "status": "healthy",
        "checks": map[string]interface{}{},
    }
    
    // 检查数据库
    if err := h.db.Exec("SELECT 1").Error; err != nil {
        result["status"] = "unhealthy"
        result["checks"].(map[string]interface{})["database"] = map[string]interface{}{
            "status": "down",
            "error":  err.Error(),
        }
    } else {
        result["checks"].(map[string]interface{})["database"] = map[string]interface{}{
            "status": "up",
        }
    }
    
    // 检查 Redis
    if err := h.redis.Ping(ctx).Err(); err != nil {
        result["status"] = "unhealthy"
        result["checks"].(map[string]interface{})["redis"] = map[string]interface{}{
            "status": "down",
            "error":  err.Error(),
        }
    } else {
        result["checks"].(map[string]interface{})["redis"] = map[string]interface{}{
            "status": "up",
        }
    }
    
    statusCode := http.StatusOK
    if result["status"] == "unhealthy" {
        statusCode = http.StatusServiceUnavailable
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(result)
}
```

---

## 10. 实施路线图

### Phase 1: MVP 基础 (2-3 周)

**目标**: 实现核心功能，支持基本的用户注册、Agent 创建和对话

**任务清单**:
- [ ] 数据库表设计和迁移脚本
- [ ] 用户注册/登录 API
- [ ] JWT 认证中间件
- [ ] Agent CRUD API
- [ ] 基础前端页面 (登录、注册、Agent 列表)
- [ ] REST 对话 API
- [ ] 简单的 Chat UI

**交付物**:
- 可运行的后端 API
- 基础前端界面
- Docker Compose 部署脚本

---

### Phase 2: 核心功能增强 (3-4 周)

**目标**: 完善用户体验，增加实时对话和会话管理

**任务清单**:
- [ ] WebSocket 实时对话
- [ ] 流式响应支持
- [ ] 会话历史查看
- [ ] Agent 配置页面
- [ ] 用户配额管理
- [ ] 使用统计基础版
- [ ] 完善前端 UI/UX

**交付物**:
- 完整的对话体验
- Agent 管理功能
- 基础配额系统

---

### Phase 3: 企业特性 (4-6 周)

**目标**: 支持多租户、团队协作和高级管理功能

**任务清单**:
- [ ] 组织/团队管理
- [ ] RBAC 权限系统
- [ ] 成员邀请和管理
- [ ] 详细使用统计与分析
- [ ] API Key 管理
- [ ] Webhook 集成
- [ ] 管理员面板

**交付物**:
- 完整的企业级特性
- 多租户隔离
- 权限管理系统

---

### Phase 4: 优化与扩展 (持续)

**目标**: 性能优化、监控告警、生态扩展

**任务清单**:
- [ ] Redis 缓存层
- [ ] 数据库查询优化
- [ ] 监控告警系统 (Prometheus + Grafana)
- [ ] 日志聚合 (ELK Stack)
- [ ] 限流和反爬虫
- [ ] 多语言支持 (i18n)
- [ ] 插件/扩展市场
- [ ] API 文档自动生成

**交付物**:
- 高性能系统
- 完整监控体系
- 可扩展架构

---

## 11. 技术选型总结

| 组件 | 推荐方案 | 理由 |
|------|---------|------|
| **前端框架** | React 18 + TypeScript | 生态成熟，类型安全，社区活跃 |
| **UI 库** | Ant Design / Shadcn UI | 企业级组件，开箱即用 |
| **状态管理** | Zustand | 轻量级，API 简洁 |
| **API 通信** | React Query + Axios | 自动缓存，请求去重 |
| **实时通信** | WebSocket | 原生支持，低延迟 |
| **后端框架** | Go + Chi Router | 高性能，并发模型优秀 |
| **数据库** | PostgreSQL + pgvector | 向量搜索，事务完整 |
| **缓存** | Redis | 高性能 KV 存储 |
| **认证** | JWT | 无状态，易扩展 |
| **监控** | Prometheus + Grafana | 标准化指标采集 |
| **日志** | Zap + ELK | 结构化日志，易查询 |
| **容器化** | Docker + Docker Compose | 环境一致性 |
| **编排 (可选)** | Kubernetes | 自动扩展，高可用 |

---

## 12. 估算成本

### 开发成本 (3 人团队)

| 阶段 | 时长 | 人员配置 |
|------|------|---------|
| Phase 1 (MVP) | 2-3 周 | 1 后端 + 1 前端 + 0.5 架构 |
| Phase 2 (核心功能) | 3-4 周 | 1 后端 + 1 前端 + 0.5 架构 |
| Phase 3 (企业特性) | 4-6 周 | 1.5 后端 + 1 前端 + 0.5 架构 |
| Phase 4 (优化) | 持续 | 1 后端 + 0.5 前端 + 0.5 运维 |

### 运维成本 (月度)

| 资源 | 配置 | 估算成本 |
|------|------|---------|
| API 服务器 (2 实例) | 2vCPU, 4GB RAM | $40 |
| PostgreSQL | 2vCPU, 8GB RAM, 100GB SSD | $60 |
| Redis | 1GB | $15 |
| 负载均衡 | 标准版 | $20 |
| 对象存储 (日志) | 50GB | $5 |
| **合计** | - | **$140/月** |

*注：成本基于云服务商中等配置，实际成本取决于流量和用户规模*

---

## 13. 下一步行动

1. **评审架构设计** - 与团队确认技术栈和架构方案
2. **创建数据库 Schema** - 编写迁移脚本
3. **搭建项目骨架** - 前后端脚手架
4. **实现 Phase 1 MVP** - 2-3 周交付基础版本
5. **用户测试** - 收集反馈，迭代优化

---

## 附录

### A. 数据库完整 Schema
参见本文档第 1 章节

### B. API 完整文档
建议使用 Swagger/OpenAPI 自动生成

### C. 部署脚本示例
```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  api:
    image: pomclaw-api:latest
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgresql://user:pass@postgres:5432/pomclaw
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - postgres
      - redis
  
  postgres:
    image: pgvector/pgvector:pg16
    volumes:
      - postgres-data:/var/lib/postgresql/data
    environment:
      - POSTGRES_PASSWORD=${DB_PASSWORD}
  
  redis:
    image: redis:7-alpine
    volumes:
      - redis-data:/data
  
  frontend:
    image: pomclaw-frontend:latest
    ports:
      - "3000:3000"
    environment:
      - API_URL=http://api:8080

volumes:
  postgres-data:
  redis-data:
```

---

**文档版本**: 1.0  
**最后更新**: 2026-04-18  
**维护者**: PomClaw Team
