# PomClaw

**企业级分布式 AI Agent 平台**

<p>
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Database-PostgreSQL%2FOracle-336791?style=for-the-badge&logo=postgresql&logoColor=white" alt="Database">
  <img src="https://img.shields.io/badge/Execution-SSH%20Sandbox-FF6600?style=for-the-badge" alt="SSH Sandbox">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Version-1.0.0-blue?style=for-the-badge" alt="Version">
</p>

[English](README.en.md) | [中文](#-项目概述)

---

## 📋 目录

- [项目概述](#-项目概述)
- [核心开发理念：API 优先](#-核心开发理念api-优先)
- [技术栈](#-技术栈)
- [核心功能](#-核心功能)
- [快速开始](#-快速开始)
- [架构设计](#-架构设计)
- [配置](#-配置)
- [应用场景](#-应用场景)
- [性能与扩展](#-性能与扩展)
- [安全](#-安全)
- [贡献](#-贡献)
- [许可证](#-许可证)

---

## 🎯 项目概述

PomClaw 是一个企业级平台，用最少的基础设施成本大规模部署 AI Agent。与个人版本需要为每个 Agent 配置一个独立 VM 不同，PomClaw 通过以下核心创新实现**无限 Agent 共享基础设施**：

- **分布式记忆存储**：所有 Agent 的记忆、对话和状态统一存储在数据库中
- **SSH 沙盒执行**：无需独立 VM，通过 SSH 沙盒安全隔离执行环境
- **多租户隔离**：支持数千个 Agent 的精细权限管理
- **成本降低 90%**：用 M 个计算节点（M ≈ N/10）服务 N 个 Agent

### 快速对比

| 方面 | 传统方案 | PomClaw |
|------|---------|---------|
| **架构** | 1 个 Agent = 1 个 VM | 共享基础设施 |
| **100个 Agent 成本** | 100 × $10/月 = $1000 | 10 × $10/月 = $100 |
| **存储** | 本地文件 | 分布式数据库 |
| **执行** | 本地计算 | SSH 沙盒池 |
| **可扩展性** | 随 Agent 线性增长 | 随数据集线性增长 |
| **管理** | 独立管理每个 VM | 统一中央平台 |

---

## 🎨 核心开发理念：API 优先

PomClaw 采用 **API 优先（API-First）** 的开发模式，核心思想是：

> **一份 API 定义，同时生成前后端代码，保证协议绝对一致。**

### 为什么需要 API 优先？

在传统前后端分离开发中，最消耗 token 的环节是**沟通和修复协议不一致**：

```
传统模式：需求讨论 → 后端写 API → 前端写类型 → 联调发现字段名不一致 → 改后端 → 改前端 → 重新联调...
消耗大量 token 在「对齐字段名」和「排查类型错误」
```

PomClaw 的 API 优先模式：

```
API 优先：修改 docs/api/pomclaw.api → make generate → 前后端代码自动生成 → 零误差联调
```

**核心优势**：

| 对比项 | 传统模式 | API 优先模式 |
|--------|---------|-------------|
| **类型定义** | 前后端各写一遍 | 从 API 定义自动生成 |
| **协议一致性** | 人工维护，容易漏改 | 自动生成，100% 一致 |
| **联调时间** | 占开发周期 30%+ | 几乎为零 |
| **Token 消耗** | 反复沟通字段名、类型、格式 | 一次性定义，零消耗在对齐上 |
| **修改成本** | 改字段需改前后端 + 文档 | 改 API 定义 → 重新生成 |
| **错误率** | 手写类型容易出错（拼写、类型、null 处理） | 生成代码经过 goctl 验证 |

### 开发工作流

```
┌─────────────────────────────────────────────────────────────────┐
│                   1. 定义 API (单⼀真相源)                         │
│              docs/api/pomclaw.api                                │
│   type CreateAgentReq { display_name string; model string }      │
│   post /v1/agents (CreateAgentReq) returns (CreateAgentResp)     │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                   2. make generate (自动生成)                     │
│                                                                   │
│   ┌──────────────────────┐    ┌──────────────────────────────┐   │
│   │ 后端生成 (goctl)      │    │ 前端生成 (goctl)             │   │
│   │                      │    │                              │   │
│   │ internal/handler/    │    │ ui/web/src/client/           │   │
│   │   createagenthandler.go│  │   pomclaw.ts (API 方法)      │   │
│   │ internal/types/      │    │   pomclawComponents.ts (类型) │   │
│   │   types.go           │    │                              │   │
│   └──────────────────────┘    └──────────────────────────────┘   │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                   3. 实现业务逻辑 (手写)                           │
│                                                                   │
│   ┌──────────────────────┐    ┌──────────────────────────────┐   │
│   │ 后端:                 │    │ 前端:                        │   │
│   │ internal/logic/      │    │ src/hooks/use-*.ts          │   │
│   │   createagentlogic.go│    │   useCreateAgent()           │   │
│   │                      │    │ src/components/              │   │
│   │   // 只需要写业务逻辑     │    │   agent-create-dialog.tsx   │   │
│   └──────────────────────┘    └──────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### 具体步骤

#### 1️⃣ 定义或修改 API

在 `docs/api/pomclaw.api` 中定义接口：

```api
// 定义请求/响应类型
type CreateAgentReq {
    DisplayName string `json:"display_name"`
    Model       string `json:"model"`
    ProviderID  int64  `json:"provider_id"`
}

type CreateAgentResp {
    Agent Agent `json:"agent"`
}

// 注册路由
service pomclaw {
    @doc "Create a new agent"
    @handler CreateAgent
    post /v1/agents (CreateAgentReq) returns (CreateAgentResp)
}
```

#### 2️⃣ 运行代码生成

```bash
make generate
```

这一步自动生成：

| 生成产物 | 文件 | 说明 |
|---------|------|------|
| **后端 Handler** | `internal/handler/createagenthandler.go` | HTTP 路由处理（含参数解析、鉴权、响应包装） |
| **后端 Types** | `internal/types/types.go` | `CreateAgentReq` / `CreateAgentResp` 结构体 |
| **前端 API 方法** | `ui/web/src/client/pomclaw.ts` | `createAgent()` 函数 |
| **前端类型定义** | `ui/web/src/client/pomclawComponents.ts` | TypeScript 接口定义 |

#### 3️⃣ 实现业务逻辑

**后端** — 在 `internal/logic/createagentlogic.go` 中写业务逻辑：

```go
func (l *CreateAgentLogic) CreateAgent(req *types.CreateAgentReq) (*types.CreateAgentResp, error) {
    // 只需要写业务逻辑，参数已经解析好，响应会自动序列化
    agent, err := l.svcCtx.AgentModel.Insert(l.ctx, req.DisplayName, req.Model)
    if err != nil {
        return nil, err
    }
    return &types.CreateAgentResp{Agent: *agent}, nil
}
```

**前端** — 通过 `useApiClient()` 调用，自动携带鉴权：

```typescript
import { useApiClient } from "@/hooks/use-api-client";
import { useMutation } from "@tanstack/react-query";

function useCreateAgent() {
    const api = useApiClient();
    return useMutation({
        mutationFn: (req: CreateAgentReq) => api.createAgent(req),
    });
}
```

> `useApiClient()` 自动注入 JWT token、租户 ID、用户 ID，无需手动处理鉴权。

### ⚠️ 重要规则

| 文件 | 规则 |
|------|------|
| `internal/handler/*.go` | **不要手动编辑** — 下次 `make generate` 会被覆盖 |
| `internal/model/*_gen.go` | **不要手动编辑** — 从数据库表结构自动生成 |
| `internal/types/*.go` | **不要手动编辑** — 从 API 定义自动生成 |
| `ui/web/src/client/*.ts` | **不要手动编辑** — 从 API 定义自动生成 |
| `internal/logic/*.go` | 手写业务逻辑 ✅ |
| `ui/web/src/hooks/*.ts` | 手写 React Query 包装 ✅ |
| `ui/web/src/components/*.tsx` | 手写 UI 组件 ✅ |

---

## 🏗️ 技术栈

### 后端框架体系
- **[go-zero](https://github.com/zeromicro/go-zero)** — 企业级微服务框架
  - `goctl` 代码生成：从 API 定义自动生成 HTTP handler 和 types
  - 高性能 RPC 和 HTTP 服务
  - 内置熔断、限流、超时控制
  - 分布式追踪和可观测性

- **[eino](https://github.com/cloudwego/eino)** — AI Agent 工程框架
  - 模块化 Agent 架构设计
  - 灵活的工具链和插件系统
  - 内置记忆、规划和推理能力
  - 完整的 LLM 集成支持

### 前端技术栈
- **React 19** + TypeScript — 现代化前端框架
- **Vite** — 超高速构建工具
- **Jotai** — 原子化状态管理
- **TanStack Router** — 类型安全的路由方案
- **Tailwind CSS** — 实用优先的样式框架
- **shadcn/ui** — 无障碍 UI 组件库

### 数据持久化
- **PostgreSQL / Oracle** — 企业级关系数据库
- **pgvector** — 向量检索和语义搜索
- 完整的多租户数据隔离

---

## ✨ 核心功能

### 🗄️ 分布式记忆存储
- **统一后端**：支持 PostgreSQL、Oracle 或任何 SQL 数据库
- **向量检索**：内置 pgvector 支持语义搜索
- **多租户隔离**：自动隔离不同组织/Agent 的数据
- **完整持久化**：保留所有对话历史、状态和元数据

### 🏗️ SSH 沙盒执行
- **安全隔离**：在隔离环境中执行代码，无需 VM 开销
- **灵活部署**：将任何 Linux/Unix 服务器连接为执行节点
- **负载均衡**：自动跨多个沙盒节点分配任务
- **资源控制**：内置超时和资源限制机制

### 💰 企业经济学
- **基础设施整合**：在同一硬件上运行数百个 Agent
- **按需扩展**：添加 SSH 节点而不是 Agent 节点
- **运维简化**：统一的日志、监控和升级管理
- **遗留系统集成**：与现有本地基础设施兼容

### 🔒 安全与合规
- **多租户 RBAC**：组织级和 Agent 级的访问控制
- **审计日志**：完整的操作审计跟踪
- **网络隔离**：支持 VPC、SSH 密钥管理、堡垒机
- **数据加密**：传输层和存储层加密

### 📊 可观测性
- **统一仪表板**：从一个地方监控所有 Agent
- **实时日志**：流式输出 Agent 执行日志和错误
- **性能指标**：CPU、内存、执行时间追踪
- **分布式追踪**：完整的系统端到端追踪

---

## 🚀 快速开始（10 分钟）

### 前置要求
- **Go 1.24+**
- **Node.js 18+**（用于前端构建）
- **PostgreSQL 13+**（或 Oracle 数据库）
- **SSH 访问沙盒节点**

### 1. 克隆和编译

```bash
git clone https://github.com/pomclaw/pomclaw.git
cd pomclaw
make build  # 自动构建后端和前端 UI
```

> **说明**：`make build` 会自动：
> - 编译前端 UI（使用 `npm run build`）
> - 编译后端二进制文件
> - 将前端打包到 `dist/control-ui/` 目录

### 2. 初始化数据库

```bash
# 创建数据库
createdb pomclaw

# 导入数据库表结构
psql pomclaw < docs/sql/pom_meta.sql
psql pomclaw < docs/sql/pom_users.sql
psql pomclaw < docs/sql/pom_agents_v2.sql
psql pomclaw < docs/sql/pom_config.sql
psql pomclaw < docs/sql/pom_memories.sql
psql pomclaw < docs/sql/pom_prompts.sql
psql pomclaw < docs/sql/pom_sessions.sql
psql pomclaw < docs/sql/pom_transcripts.sql
psql pomclaw < docs/sql/pom_daily_notes.sql
psql pomclaw < docs/sql/pom_state.sql
```

### 3. 启动 Gateway

```bash
./build/pomclaw

# Gateway 运行在 http://localhost:18790
# 自动提供前端 UI：http://localhost:18790（使用 dist/control-ui）
```

**Gateway Web UI 界面：**

![PomClaw Gateway Chat UI](docs/screenshots/pomclaw_chat.jpg)

---

## 📋 架构设计

```
┌──────────────────────────────────────────────────────────┐
│          分布式数据库（PostgreSQL/Oracle）                  │
│  - 记忆、对话、状态（多租户）                               │
│  - pgvector 向量嵌入                                      │
└──────────────────────────────────────────────────────────┘
                          ↑
                ┌─────────┼─────────┐
                ↓         ↓         ↓
         ┌──────────┐┌──────────┐┌──────────┐
         │SSH Node1 ││SSH Node2 ││SSH Node3 │
         │（沙盒）  ││（沙盒）  ││（沙盒）  │
         └──────────┘└──────────┘└──────────┘
                ↑         ↑         ↑
                └─────────┼─────────┘
                          │
    ┌─────────────────────┴─────────────────────┐
    │    PomClaw Gateway API + WebSocket         │
    │  （单一控制平面服务所有 Agent）            │
    └─────────────────────┬─────────────────────┘
         ↑                 ↑                ↑
    ┌────────────┐   ┌────────────┐   ┌────────────┐
    │  Agent-1   │   │  Agent-2   │   │  Agent-N   │
    └────────────┘   └────────────┘   └────────────┘
```

### 代码架构

```
pomclaw/
├── docs/api/pomclaw.api          # 🎯 API 定义（单一真相源）
├── cmd/pomclaw/                   # 入口
├── internal/
│   ├── handler/                   # ⚠️ 自动生成（HTTP 路由）
│   ├── types/                     # ⚠️ 自动生成（请求/响应类型）
│   ├── logic/                     # ✅ 手写业务逻辑
│   ├── model/                     # ⚠️ 自动生成（数据库 CRUD）
│   ├── storage/                   # ✅ 数据访问层
│   ├── agent/                     # ✅ Agent 循环
│   ├── svc/                       # ✅ 依赖注入
│   ├── tools/                     # ✅ 工具实现
│   └── contracts/                 # ✅ 接口定义
├── ui/web/
│   ├── src/
│   │   ├── client/                # ⚠️ 自动生成（API 方法 + 类型）
│   │   ├── hooks/                 # ✅ React Query 包装
│   │   ├── components/            # ✅ UI 组件
│   │   ├── pages/                 # ✅ 页面
│   │   └── stores/                # ✅ 状态管理
│   └── ...
├── etc/                           # YAML 配置
└── Makefile                       # 构建 + 代码生成
```

---

## 🔧 配置

### 数据库配置

```json
{
  "storage_type": "postgres",
  "postgres": {
    "enabled": true,
    "host": "db.example.com",
    "port": 5432,
    "database": "pomclaw",
    "user": "pomclaw",
    "password": "${POSTGRES_PASSWORD}",
    "ssl_mode": "require",
    "pool_max_open": 25,
    "pool_max_idle": 5
  }
}
```

---

## 📚 应用场景

### 🏢 企业 AI 客服
从 10 个扩展到 1000+ 个支持 Agent，成本增长无关

### 🤖 工作流自动化平台
用于 RPA、数据处理和业务逻辑自动化的分布式任务执行引擎

### 📊 大规模数据分析
为每个用户/组织提供隔离、安全工作区的多租户分析平台

### 🔬 科研计算
用于科学模拟和数据处理的高可用计算集群

### 🎓 教育平台
为数千名学生管理 AI 助手，拥有隔离且安全的工作区

---

## 📊 性能与扩展

### 容量规划

| 配置 | Agent 数 | 内存/Agent | CPU | 数据库 |
|------|---------|-----------|-----|--------|
| 小型 | 100 | 256MB | 2-4 核 | PostgreSQL 13 |
| 中型 | 1,000 | 256MB | 8-16 核 | PostgreSQL 14 |
| 大型 | 10,000 | 256MB | 32+ 核 | PostgreSQL 14+ 或 Oracle 21c |
| 企业 | 100,000+ | 256MB | 多节点 | 分布式数据库 |

### 存储需求

- **每个 Agent**：~1MB 元数据 + 10MB 对话（因使用情况而异）
- **向量存储**：~1,500 字节/条记忆（384 维嵌入）

---

## 🔒 安全

### 身份认证与授权
- JWT 令牌认证
- 组织级和 Agent 级的 RBAC
- API 密钥管理和轮换

### 网络安全
- SSH 密钥认证（无密码）
- 所有通信都使用 TLS 1.3
- VPC/网络隔离支持
- 堡垒机兼容

### 数据保护
- 存储层加密（数据库级）
- 传输层加密（TLS）
- 所有操作的审计日志
- 数据保留和合规策略

---

## 🤝 贡献

欢迎贡献代码！请：

1. Fork 仓库
2. 创建功能分支（`git checkout -b feature/amazing-feature`）
3. 提交更改（`git commit -m 'Add amazing feature'`）
4. Push 到分支（`git push origin feature/amazing-feature`）
5. 开启 Pull Request

### 开发原则

- **API 优先**：修改 `docs/api/pomclaw.api` → `make generate` → 实现业务逻辑
- **不要编辑生成文件**：handler、types、client 目录下的文件会被覆盖
- **遵循现有代码风格**：使用 `gofmt` 和项目 ESLint 配置

---

## 📜 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

---

## 📞 支持

- **问题反馈**: [GitHub Issues](https://github.com/pomclaw/pomclaw/issues)
- **讨论**: [GitHub Discussions](https://github.com/pomclaw/pomclaw/discussions)
- **企业支持**: contact@pomclaw.com

---

## 🎉 致谢

PomClaw 基于以下优秀开源项目：
- go-zero 和 eino 社区
- 开源数据库和 SSH 社区
- Go 生态系统贡献者