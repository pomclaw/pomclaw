# PomClaw

**企业级分布式 AI Agent 平台** — 用最少的基础设施成本，大规模部署 AI Agent。

<p>
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Database-PostgreSQL%2FOracle-336791?style=for-the-badge&logo=postgresql&logoColor=white" alt="Database">
  <img src="https://img.shields.io/badge/Execution-SSH%20Sandbox-FF6600?style=for-the-badge" alt="SSH Sandbox">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Version-1.0.0-blue?style=for-the-badge" alt="Version">
</p>

[English](README.en.md) | [中文](#-项目概述)

---

## 🎯 项目背景

传统个人版 OpenClaw 的问题很明显，并不适合企业做 toC 项目：

* **每个 Agent 一台机器**，数量越多成本越线性涨

* **记忆、对话都存在本地文件**，分散在各机器上，没法统一管

* **每台机器要独立升级、监控**，运维很痛苦

**云端虾的核心目标：少数几台机器服务大批量 Agent，成本打下来，管理集中起来。**

PomClaw 用 M 个计算节点（M ≈ N/10）服务 N 个 Agent，共享基础设施：

| 方面 | 传统方案 | PomClaw |
|------|---------|---------|
| **架构** | 1 个 Agent = 1 个 VM | 共享基础设施 |
| **100个 Agent 成本** | 100 × $10/月 = $1000 | 10 × $10/月 = $100 |
| **存储** | 本地文件 | 分布式数据库 |
| **执行** | 本地计算 | SSH 沙盒池 |
| **管理** | 独立管理每个 VM | 统一中央平台 |

> 这个想法受到我们在做的小孩子伴学点子宠物 kidclaw 的启发，想要一个能规模化、低成本跑起来的 AI Agent，于是就有了云端虾。

---

## 💡 设计理念

> pom 代表 **pomelo**，是名称的创意部分。"pomclaw 结合了 **pomelo（柚子）** 和 **claw（钳子）**"。

![PomClaw Logo](docs/screenshots/logo_1.png) ![PomClaw Logo](docs/screenshots/logo_2.png)

### API 优先（核心开发理念）

> **一份 API 定义，同时生成前后端代码，保证协议绝对一致。**

基于 go-zero 的 **goctl / zero-api 中间语言**：先定义、再生成、只写业务逻辑，不让 AI 从头写前后端协议代码。这个模式能有效限制 AI 乱发挥，100% 保证代码精确性，也能减少 token 消耗。

```
API 优先：修改 docs/api/pomclaw.api → make generate → 前后端代码自动生成 → 零误差联调
```

**工作流**：

1. 定义 API：在 `docs/api/pomclaw.api` 中定义接口和类型
2. 生成代码：`make generate` 自动生成后端 handler/types + 前端 TS 客户端
3. 实现逻辑：只写 `internal/logic/` 和前端 hooks/组件

> ⚠️ **不要手动编辑生成文件**（handler、types、client），下次 `make generate` 会被覆盖。只手写 `internal/logic/` 业务逻辑。

---

## 🚀 项目核心场景

从产品角度讲，我们想做的事：

1. **让每个人能快速创建自己的虾**
2. **虾能实现自己的功能**
3. **让训练好的虾能流通起来给别人复用**

### 虾的基础功能：创建虾，能干活

用户进来第一件事，就是创建自己的虾。平台支持：

* 起名字、挑模型、选 Provider，写一句描述告诉虾它是干什么的
* 初始化人格：SOUL.md、AGENTS.md 这些启动文件，定义虾的性格、能力和做事边界
* 系统提示词预览：启用前先看看组装出来的 system prompt 长什么样
* 创建完就能对话，基于 WebSocket 流式返回，实时聊天

![主功能菜单模块图](docs/screenshots/main_menu.png)

### 虾的 Agent 市场：直接能复用

这个模块的设计初衷是，大家都可以把自己训练好的虾共享出来，形成一个虾的人才市场。别人不用从零训练，看中哪只直接用就行。

平台靠共享机制支撑：虾可以标记为共享，记录创建者信息，别人在市场上看到就能复用；虾配套的技能也可以一起共享，训练好的虾往往带一套技能，技能跟着虾一起流通。

![虾的agent市场图](docs/screenshots/agent_market.png)

### 虾的对话：核心场景

举个例子：

**kidclaw 闯龙宫** — 小朋友的闯龙宫，把剧情、规则、关卡打包成一个 skill，上传授权给虾，虾就学会了带小朋友闯关。不用改代码，加个技能包就行。

![kidclaw虾的闯龙宫skill效果图](docs/screenshots/kidclaw_skill.png)

---

## 🏗️ 技术全景

整体就一个 Gateway 服务所有 Agent，所有 Agent 共享同一个数据库和计算资源。

* **后端**：Go 1.25 + [go-zero](https://github.com/zeromicro/go-zero)（微服务框架）+ [eino](https://github.com/cloudwego/eino)（AI Agent 框架）
* **存储**：PostgreSQL + pgvector，管数据和向量
* **前端**：React 19 + Vite
* **监控**：OpenTelemetry

![pomclaw框架图](docs/screenshots/framework.png)

底层引擎是使用的 Go 项目中一直在用的 **eino 框架**：

> [开源 GitHub](https://github.com/cloudwego/eino) | [官方文档](https://www.cloudwego.io/zh/docs/eino/overview/)

下面不铺开讲每个模块，挑 **3 条核心链路**，讲讲每条链路上我们实际怎么做的。

---

## 🔗 三条核心链路

### 链路一：一条消息的完整旅程

这一条讲的是，用户在聊天框里发一句话，背后发生了什么。

**消息怎么进来** — 前端通过 WebSocket 实时把消息推给后端，用自研的 Protocol v3 协议定义消息格式，Agent 处理结果再流式推回前端，用户能看到打字机一样的效果。

> 这里现在看，已经有成熟的 AI 前后端交互通信协议，比如 [ag-ui-protocol](https://github.com/ag-ui-protocol/ag-ui)，只不过 pomclaw 是自己实现的 AI 交互协议，同样基于 ws 通信。

**Agent 怎么处理** — Agent 核心用 eino 的 ChatModelAgent 重写，一条消息大致走这几步：

1. 解析出这条消息属于哪个 Agent、哪个工作目录、哪个渠道
2. 组装上下文：系统提示词、SOUL.md / AGENTS.md 这些上下文文件、会话历史
3. 交给模型，模型决定要不要调工具
4. 调工具、拿结果、再交回模型，循环到模型觉得回答完为止
5. 最终答案通过 WebSocket 流式返回

工程结构上用工厂模式加 context builder 接口解耦，以后想换 Agent 实现不用动上层，多个渠道接入也方便。

```go
adkAgent, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
   Name:          "pomclaw",
   MaxIterations: l.svcCtx.Config.Agents.Defaults.MaxToolIterations,
   ToolsConfig: adk.ToolsConfig{
      ToolsNodeConfig: toolsNodeConfig,
   },
   Model: llm,
})
// Discover MCP tools for this agent and append to tool config
mcpTools, mcpClosers := l.discoverMCPTools(l.ctx, agentRecord.AgentId)
toolsNodeConfig.Tools = append(toolsNodeConfig.Tools, mcpTools...)
```

**模型能调哪些工具** — 内置工具集分几类：

| 类别 | 工具 |
|:----|:----|
| 文件系统 | 读、编辑、列表、写入 |
| 执行 | shell 命令 |
| 记忆 | recall 短期回忆、remember 长期写入 |
| 技能 | use_skill、run_skill_script、read_skill_file |

工具不光是 shell 和文件，还接了记忆和技能，这个后面两条链路展开讲。

### 链路二：Agent 的记忆

这一条讲的是 Agent 怎么记住东西、需要的时候怎么想起来。这是 Agent 好不好用的关键。

pomclaw 中是自己实现的记忆工具，也可以展示一些记忆外化：

![agent的长期记忆图](docs/screenshots/memory_overview.png)

![agent的长期记忆内容](docs/screenshots/memory_content.png)

记忆模块放到平台中，好处是可以随时查阅 agent 的记忆内容，方便跟踪调整和修改。

**存储分两层**：

* **文档层**（memory_documents）：存原始文档，日记、笔记、项目背景
* **分块层**（memory_chunks）：文档按行切块，每块单独生成向量和全文索引

**双路检索** — 查记忆的时候走两条路并行：

* **语义检索**：用 pgvector 的 HNSW 索引，按语义相似度找，这是向量那一路
* **关键词检索**：用全文索引，精确匹配关键词，这是关键词那一路

两路结果混合打分再返回。好处是语义和关键词互补，光靠语义有时会漏掉精确词，光靠关键词又抓不住意思相近的说法。

> 为什么使用 PostgreSQL 作为底层存储？因为开源，支持很多插件。pgvector 就是本次记忆所使用的向量扩展插件，支持向量检索。其实也可以用其他的比如 Milvus。

**实现方式** — 增加两个记忆工具：

* **remember**：记忆写入，Agent 记住了就存到文档层并分块向量化
* **recall**：记忆提取，对话过程中按需要检索相关记忆喂回上下文

实际用起来的效果就是：Agent 能记住用户聊过的偏好，下次会话能想起来；也能基于项目背景文档回答问题，而不是每次从零开始。

> 这里的记忆是我们自己实现的。除此之外，agent 长记忆除了比较知名的 mem0，现在也有其他比较成熟的 memory 方案，graphiti 实时知识图谱构建、letta 记忆分级等，可以尝试使用效果，看看能否应用，提升记忆效果。

### 链路三：Agent 的可观测

这一条讲的是，Agent 跑起来以后，怎么知道它干了什么、花多少钱、出了错怎么查。

**调用链还原** — 用 trace 和 span 两张表组成父子调用链。一次 Agent 处理会形成一个 trace，里面按步骤拆成多个 span，span 之间用父子关系串起来，能完整还原一次调用：先调了哪个模型、中间调了哪个工具、耗时多少、最后结果如何。

**指标全记录** — 每次调用会记录这些数据：

* token 用量（输入 / 输出）
* 成本
* LLM 调用次数、工具调用次数
* 耗时、状态、错误信息

上报方式是接 eino 的 OpenTelemetry 回调，自动采集，不用埋点。

**实现方式** — 通过实现 eino 的 callbacks 接口，来捕捉到各个节点开始和结束的时间，然后组装成链路完成指标记录。

```go
tracesModel := model.NewTracesModel(psqlConn)
spansModel := model.NewSpansModel(psqlConn)

traceExporter := callback.NewPGExporter(tracesModel, spansModel)
traceProvider := callback.NewLocalTracerProvider(traceExporter)
meterProvider := metric.NewMeterProvider()
opentelemetry.SetProvider(traceProvider, meterProvider)

traceHandler, shutdown, err := apmplus.NewApmplusHandler(&apmplus.Config{
   Host:        "local",
   AppKey:      "local",
   ServiceName: c.Name,
})
```

---

## 🛠️ 开发原则：AI + zero-api 中间语言

最后说下怎么开发这套东西的，这是我们实践下来最有价值的方法论。

基于 go-zero 的 goctl 工具，核心思路是：**先定义、再生成、只写业务逻辑**，不让 AI 从头写前后端协议代码。

> **zero-api**（[goctl](https://github.com/zeromicro/zero-api)）是一个 RESTful API 描述中间语言。这个概念很早就有了，类似于 gRPC，只不过 zero-api 专注于 RESTful HTTP API。这次在本项目中实践了 AI + goctl 的开发模式，能够有效限制 AI 乱发挥的毛病。

![中间语言产生各端代码](docs/screenshots/codegen_diagram.png)

首先设计最为重要的两个协议：**① API+WS 的接口定义**、**② SQL 表结构定义**。这两块内容由开发者严格把控，最早生成：

```go
@server (
   prefix: /pomclaw-api
   jwt:    Auth
)
service pomclaw {
   @doc "List all agents"
   @handler ListAgents
   get /v1/agents (ListAgentsReq) returns (ListAgentsResp)

   @doc "Create a new agent"
   @handler CreateAgent
   post /v1/agents (CreateAgentReq) returns (CreateAgentResp)
}
```

```sql
-- Pomclaw MCP Servers Table
create table mcp_servers
(
    id          serial primary key,
    user_id     uuid                                   not null,
    name        varchar(255)                           not null,
    description varchar(255),
    transport   varchar(50)                            not null, -- stdio, sse, streamable-http
    command     text,                                            -- stdio: command to spawn
    args        jsonb                    default '[]'::jsonb,    -- stdio: command arguments
    url         text,                                            -- sse/http: server URL
    headers     jsonb                    default '{}'::jsonb,    -- sse/http: HTTP headers
    env         jsonb                    default '{}'::jsonb,    -- stdio: environment variables
    api_key     varchar(512),
    tool_prefix varchar(50),
    timeout_sec integer                  default 60    not null,
    settings    jsonb                    default '{}'::jsonb not null,
    enabled     boolean                  default true  not null,
    is_shared   boolean                  default false not null,
    created_at  timestamp with time zone default now() not null,
    updated_at  timestamp with time zone default now() not null,
    constraint mcp_servers_name_key unique (name)
);
```

在项目中，通过 CLAUDE.md 文件，严格要求 AI 的开发内容，强调 AI 直接修改代码并不会生效，必须通过修改 `docs/api`、`docs/sql` 等文件来间接修改代码：

```markdown
## Goctl 代码生成（核心工具链）

# 1. 从数据库生成模型 CRUD
goctl model pg datasource \
  --url='postgres://user:pass@host:port/db' \
  -t='table_names' \
  -d='internal/model'

# 2. 从 .api 文件生成 后端 GO语言 HTTP handlers + types 代码
goctl api go --api docs/api/pomclaw.api -dir ./

# 3. 从 .api 文件生成 前端 ts语言代码
goctl api ts --api docs/api/pomclaw.api -dir ./ui/src/client

**⚠️ DO NOT EDIT**:
- `internal/handler/*.go` - Auto-generated HTTP handlers
- `internal/model/*_gen.go` - Auto-generated CRUD
- `internal/types/*.go` - Auto-generated request/response types
```

最终生成大量的前后端、数据层等协议文件：

![生成的前后端协议文件 1](docs/screenshots/generated_files_1.png)
![生成的前后端协议文件 2](docs/screenshots/generated_files_2.png)
![生成的前后端协议文件 3](docs/screenshots/generated_files_3.png)
![生成的前后端协议文件 4](docs/screenshots/generated_files_4.png)

对于这些重复可生成的代码，我们不会让 AI 反复确认和修改，可以很好的限制 AI 的自由发挥、100% 的保证了代码的精确性，并且也能减少 token 消耗。

---

## 🚀 快速开始

### 前置要求
- **Go 1.24+**、**Node.js 18+**
- **PostgreSQL 13+**（或 Oracle）
- **SSH 访问沙盒节点**

### 1. 克隆和编译

```bash
git clone https://github.com/pomclaw/pomclaw.git
cd pomclaw
make build  # 自动构建后端和前端 UI
```

> **说明**：`make build` 会自动编译前端 UI（`npm run build`）、编译后端二进制、并把前端打包到 `dist/control-ui/` 目录。

### 2. 初始化数据库

```bash
createdb pomclaw
for f in docs/sql/*.sql; do psql pomclaw < $f; done
```

### 3. 启动 Gateway

```bash
./build/pomclaw  # Gateway 运行在 http://localhost:18790，自动提供前端 UI
```

---

## 🔧 配置

配置为 YAML 格式，位于 `etc/` 目录：

- `etc/config.example.yaml` — 完整示例
- `etc/local.yaml` — 本地开发配置

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
