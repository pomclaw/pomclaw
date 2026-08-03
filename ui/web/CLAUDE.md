# PomClaw UI - 开发指南

## 项目概述

React 19 + TypeScript + Vite 前端，连接 go-zero 后端。

**核心功能**：智能体管理、WebSocket 实时通信、内存系统、Provider 管理。

## 快速启动

```bash
npm install
npm run dev          # http://localhost:5173/pomclaw
npm run build        # 生产构建
npm run lint:fix     # 格式修复
```

## 配置

`.env.local`:
```env
VITE_BACKEND_HOST=localhost
VITE_BACKEND_PORT=18790
VITE_API_PREFIX=/pomclaw
```

## 项目结构

```
src/
├── api/                    # 核心层（手写）
│   ├── http-client.ts      # HTTP 客户端（鉴权 + apiPrefix）
│   ├── ws-client.ts        # WebSocket
│   ├── protocol.ts         # Protocol v3
│   ├── errors.ts           # 错误定义
│   ├── voices.ts           # 声音管理（无生成 API）
│   └── tts-capabilities.ts # TTS 能力（无生成 API）
├── client/                 # ⚠️ AUTO-GENERATED (勿改)
│   ├── pomclaw.ts          # 44+ API 方法
│   ├── pomclawComponents.ts# 类型定义
│   └── gocliRequest.ts     # HTTP 基础客户端
├── hooks/
│   ├── use-api-client.ts   # API 包装层（鉴权）✅
│   ├── use-system-health.ts
│   ├── use-agents.ts
│   └── use-*.ts            # React Query hooks ✅
├── components/             # React 组件 ✅
├── pages/                  # 页面 ✅
├── stores/                 # Zustand 状态 ✅
└── types/                  # 类型定义 ✅

✅ = 手写代码（安全修改）
⚠️ = 生成代码（勿改，会被覆盖）
```

## 核心开发模式

### 黄金法则：协议优先

所有 HTTP API 的**类型和定义都由后端自动生成**，不手写。

```
修改 API 定义 (docs/api/pomclaw.api)
    ↓
make generate
  ↙        ↘
后端代码   前端代码
(handler)  (useApiClient)
          ↓
        hooks + 组件
```

### 新增 HTTP 接口（5 步）

**1️⃣ 修改 API 定义** (`docs/api/pomclaw.api`)
```api
service pomclaw-api {
    @doc "创建智能体"
    @handler CreateAgent
    post /v1/agents (CreateAgentReq) returns(CreateAgentResp)
}

type CreateAgentReq { name string; model string }
type CreateAgentResp { id string; name string }
```

**2️⃣ 生成代码**
```bash
make generate
```

**3️⃣ 后端实现** (`internal/logic/createagentlogic.go`)
```go
func (l *CreateAgentLogic) CreateAgent(req *types.CreateAgentReq) (*types.CreateAgentResp, error) {
    // 业务逻辑
    return &types.CreateAgentResp{...}, nil
}
```

**4️⃣ 前端 Hook** (`src/hooks/use-agents.ts`)
```typescript
export function useCreateAgent() {
  const api = useApiClient();
  return useMutation({
    mutationFn: (req: CreateAgentReq) => api.createAgent(req),
  });
}
```

**5️⃣ 在组件中用**
```typescript
const { mutate } = useCreateAgent();
mutate({ name: "Agent", model: "gpt-4" });
```

### 鉴权原理

生成的代码**不处理鉴权**（无法访问 React 上下文）。

- ✗ 直接用生成代码：无认证头
- ✓ 用 `useApiClient()`：自动添加认证头

```typescript
// ✓ 这样用 - useApiClient() 负责鉴权
const api = useApiClient();
await api.createAgent({...});

// ✗ 不要这样 - 没有认证
await webapi.post("/v1/agents", {...});
```

### 必读规则

**不要改的** ❌
- `src/client/*.ts` - 下次 `make generate` 被覆盖

**要手写的** ✅
- `src/hooks/use-*.ts` - React Query hooks
- `src/components/**` - React 组件
- 后端 `internal/logic/` - 业务逻辑

**问题代码怎么办？**
- 改 `gen_models.sh` 或生成工具
- 重新运行 `make generate`
- 不改生成的文件

## WebSocket（手写）

无法自动生成，手写实现：

```typescript
const ws = useWs();
const response = await ws.call(Methods.AGENT_LIST, params);
```

## State Management

- **Zustand** - 认证、UI 状态
- **React Query** - 服务器状态（agents、sessions 等）

## 常用命令

```bash
npm run dev        # 开发
npm run build      # 构建
npm run lint:fix   # 修复格式
npm test           # 测试
npm test:watch     # 测试监听

# 后端（项目根目录）
make generate      # 生成代码
make build         # 构建
make run           # 运行
```

## 相关文档

- 后端架构：`../CLAUDE.md`
- API 定义：`../../docs/api/pomclaw.api`
- WebSocket 协议：`../../docs/WEBSOCKET_GUIDE.md`

## 环境变量

**开发** (`.env.local`)
```env
VITE_BACKEND_HOST=localhost
VITE_BACKEND_PORT=18790
VITE_API_PREFIX=/pomclaw
```

**生产**：通过 CI/CD 注入

## 类型安全

所有类型来自自动生成，无需手写：

```typescript
// 类型自动从生成的 pomclawComponents.ts 导入
import type { CreateAgentReq, CreateAgentResp } from "@/client";
```

## 常见问题

**Q: 为什么不能改 `src/client/` 中的代码？**
A: 下次运行 `make generate` 就被覆盖，白改。修改要在包装层（useApiClient、hooks）做。

**Q: WebSocket 怎么调用？**
A: 手写，通过 `useWs()` hook 调用。HTTP API 用自动生成的。

**Q: 怎么添加新的 API？**
A: 改 `docs/api/pomclaw.api` → `make generate` → 实现业务逻辑 → 前端包装 hook。
