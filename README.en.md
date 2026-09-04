# PomClaw

**Enterprise-Grade Distributed AI Agent Platform** — Deploy AI Agents at scale with minimal infrastructure costs.

<p>
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Database-PostgreSQL%2FOracle-336791?style=for-the-badge&logo=postgresql&logoColor=white" alt="Database">
  <img src="https://img.shields.io/badge/Execution-SSH%20Sandbox-FF6600?style=for-the-badge" alt="SSH Sandbox">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Version-1.0.0-blue?style=for-the-badge" alt="Version">
</p>

[English](#-project-background) | [中文](README.md)

---

## 🎯 Project Background

The problems with traditional personal-use OpenClaw are obvious — it's not suited for enterprises building toC products:

* **One machine per Agent** — costs grow linearly as the number of Agents increases

* **Memory and conversations stored in local files** — scattered across machines, impossible to manage centrally

* **Each machine needs independent upgrades and monitoring** — painful operations

**The core goal of Cloud Shrimp (云端虾): serve a large number of Agents with just a few machines, cut costs, and centralize management.**

PomClaw serves N Agents with M compute nodes (M ≈ N/10), sharing infrastructure:

| Aspect | Traditional | PomClaw |
|------|---------|---------|
| **Architecture** | 1 Agent = 1 VM | Shared infrastructure |
| **Cost for 100 Agents** | 100 × $10/mo = $1000 | 10 × $10/mo = $100 |
| **Storage** | Local files | Distributed database |
| **Execution** | Local compute | SSH sandbox pool |
| **Management** | Manage each VM individually | Centralized platform |

> This idea was inspired by our kidclaw companion-pet project for kids — we wanted an AI Agent that could scale and run at low cost, and so Cloud Shrimp was born.

---

## 💡 Design Philosophy

> pom stands for **pomelo**, the creative part of the name. "pomclaw combines **pomelo** and **claw**".

![PomClaw Logo](docs/screenshots/logo_1.png) ![PomClaw Logo](docs/screenshots/logo_2.png)

### API First (Core Development Philosophy)

> **One API definition generates both frontend and backend code, guaranteeing 100% protocol consistency.**

Based on go-zero's **goctl / zero-api intermediate language**: define first, then generate, only write business logic — don't let AI write frontend/backend protocol code from scratch. This pattern effectively limits AI's free rein, guarantees 100% code accuracy, and reduces token consumption.

```
API First: edit docs/api/pomclaw.api → make generate → frontend + backend code auto-generated → zero-error integration
```

**Workflow**:

1. Define the API: define interfaces and types in `docs/api/pomclaw.api`
2. Generate code: `make generate` auto-generates backend handlers/types + frontend TS client
3. Implement logic: only write `internal/logic/` and frontend hooks/components

> ⚠️ **Don't manually edit generated files** (handler, types, client) — they'll be overwritten on the next `make generate`. Only hand-write `internal/logic/` business logic.

---

## 🚀 Core Scenarios

From a product perspective, what we want to build:

1. **Let everyone quickly create their own Shrimp**
2. **Shrimp can implement their own functions**
3. **Let trained Shrimp circulate and be reused by others**

### Basic Shrimp Features: Create a Shrimp, It Can Work

The first thing users do is create their own Shrimp. The platform supports:

* Name it, pick a model, select a Provider, and write a one-line description telling the Shrimp what it does
* Initialize personality: startup files like SOUL.md and AGENTS.md define the Shrimp's character, abilities, and boundaries
* System prompt preview: see the assembled system prompt before enabling it
* Chat right after creation — real-time streaming via WebSocket

![Main menu module diagram](docs/screenshots/main_menu.png)

### Shrimp Agent Market: Directly Reusable

The design intent of this module is that everyone can share their trained Shrimp, forming a talent market for Shrimp. Others don't need to train from scratch — just pick the one they like and use it directly.

The platform is supported by a sharing mechanism: Shrimp can be marked as shared with creator info recorded, so others can reuse them from the market. Accompanying skills can also be shared — trained Shrimp often carry a set of skills, and skills circulate along with the Shrimp.

![Shrimp agent market diagram](docs/screenshots/agent_market.png)

### Shrimp Chat: The Core Scenario

For example:

**kidclaw explores the Dragon Palace** — a kids' adventure where the story, rules, and levels are packaged into a skill, uploaded and authorized to the Shrimp, and the Shrimp learns to guide kids through the game. No code changes needed — just add a skill pack.

![kidclaw Shrimp's Dragon Palace skill demo](docs/screenshots/kidclaw_skill.png)

---

## 🏗️ Technical Overview

A single Gateway serves all Agents, sharing the same database and compute resources.

* **Backend**: Go 1.25 + [go-zero](https://github.com/zeromicro/go-zero) (microservices framework) + [eino](https://github.com/cloudwego/eino) (AI Agent framework)
* **Storage**: PostgreSQL + pgvector, for data and vectors
* **Frontend**: React 19 + Vite
* **Monitoring**: OpenTelemetry

![pomclaw framework diagram](docs/screenshots/framework.png)

The underlying engine is the **eino framework**, widely used in Go projects:

> [Open Source GitHub](https://github.com/cloudwego/eino) | [Official Docs](https://www.cloudwego.io/zh/docs/eino/overview/)

Instead of covering every module, we'll highlight **3 core chains** and how we actually built each one.

---

## 🔗 Three Core Chains

### Chain 1: The Full Journey of a Message

This covers what happens behind the scenes when a user types a message in the chat box.

**How the message comes in** — The frontend pushes messages to the backend in real-time via WebSocket, using our self-developed Protocol v3 to define message formats. Agent results are streamed back to the frontend, giving users a typewriter-like effect.

> There are already mature AI frontend/backend interaction protocols, such as [ag-ui-protocol](https://github.com/ag-ui-protocol/ag-ui), but pomclaw implements its own AI interaction protocol, also based on WebSocket.

**How the Agent processes it** — The Agent core is rewritten with eino's ChatModelAgent. A message roughly goes through these steps:

1. Parse which Agent, working directory, and channel the message belongs to
2. Assemble context: system prompt, context files like SOUL.md / AGENTS.md, conversation history
3. Hand to the model; the model decides whether to call tools
4. Call tools, get results, hand back to the model, loop until the model decides it's done
5. Final answer streamed back via WebSocket

The engineering structure uses a factory pattern plus a context builder interface to decouple — future Agent implementation changes won't touch the upper layer, and multiple channel integrations are easy.

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

**Which tools the model can call** — Built-in tools are grouped into several categories:

| Category | Tools |
|:----|:----|
| File system | read, edit, list, write |
| Execution | shell commands |
| Memory | recall (short-term retrieval), remember (long-term write) |
| Skills | use_skill, run_skill_script, read_skill_file |

Tools aren't just shell and files — they also integrate memory and skills, covered in the next two chains.

### Chain 2: Agent Memory

This covers how the Agent remembers things and recalls them when needed. It's the key to how useful an Agent is.

pomclaw implements its own memory tools, which also demonstrate memory externalization:

![agent long-term memory overview](docs/screenshots/memory_overview.png)

![agent long-term memory content](docs/screenshots/memory_content.png)

Putting the memory module in the platform lets you inspect an Agent's memory content anytime, making it easy to track, adjust, and modify.

**Two-layer storage**:

* **Document layer** (memory_documents): raw documents — diaries, notes, project background
* **Chunk layer** (memory_chunks): documents split into line-based chunks, each chunk gets its own vector and full-text index

**Dual-path retrieval** — memory queries run two paths in parallel:

* **Semantic retrieval**: pgvector's HNSW index, finds by semantic similarity — the vector path
* **Keyword retrieval**: full-text index, exact keyword matching — the keyword path

Results from both paths are mixed, scored, and returned. The benefit is complementarity: semantic alone can miss exact words, keyword alone can't catch similar phrasing.

> Why PostgreSQL as the underlying storage? Because it's open source and supports many plugins. pgvector is the vector extension used for memory, supporting vector retrieval. Alternatives like Milvus also work.

**Implementation** — two memory tools:

* **remember**: memory write; once the Agent remembers something, it's stored in the document layer, chunked and vectorized
* **recall**: memory retrieval; during conversation, relevant memories are retrieved and fed back into context as needed

In practice: the Agent remembers user preferences discussed and recalls them in the next session; it can also answer questions based on project background documents instead of starting from zero each time.

> This memory is our own implementation. Beyond the well-known mem0, there are other mature memory solutions — graphiti (real-time knowledge graph construction), letta (tiered memory), etc. — worth trying to see if they can further improve memory effectiveness.

### Chain 3: Agent Observability

This covers how you know what an Agent did, how much it cost, and how to trace errors once it's running.

**Call chain reconstruction** — trace and span tables form parent-child call chains. One Agent processing forms a trace, split into multiple spans by step, linked by parent-child relationships — fully reconstructing a single call: which model was called first, which tool in between, how long it took, and the final result.

**Full metric recording** — every call records:

* token usage (input / output)
* cost
* LLM call count, tool call count
* duration, status, error messages

Reporting uses eino's OpenTelemetry callbacks, auto-collected without instrumentation.

**Implementation** — by implementing eino's callbacks interface, we capture the start and end times of each node, then assemble them into call chains and complete metric recording.

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

## 🛠️ Development Principles: AI + zero-api Intermediate Language

Finally, how we developed this — the most valuable methodology from our practice.

Based on go-zero's goctl tool, the core idea is: **define first, generate, write only business logic** — don't let AI write frontend/backend protocol code from scratch.

> **zero-api** ([goctl](https://github.com/zeromicro/zero-api)) is a RESTful API description intermediate language. This concept has existed for a long time — similar to gRPC, but zero-api focuses on RESTful HTTP APIs. In this project we practiced the AI + goctl development pattern, which effectively limits AI's tendency to improvise.

![Intermediate language generating per-side code](docs/screenshots/codegen_diagram.png)

First design the two most important protocols: **① API+WS interface definitions**、**② SQL table structure**. These are strictly controlled by developers and generated first:

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

Through the CLAUDE.md file, we strictly require AI development content, emphasizing that AI modifying code directly won't take effect — it must modify `docs/api`, `docs/sql` files to indirectly change code:

```markdown
## Goctl Code Generation (Core Toolchain)

# 1. Generate model CRUD from database
goctl model pg datasource \
  --url='postgres://user:pass@host:port/db' \
  -t='table_names' \
  -d='internal/model'

# 2. Generate backend Go HTTP handlers + types from .api file
goctl api go --api docs/api/pomclaw.api -dir ./

# 3. Generate frontend TS code from .api file
goctl api ts --api docs/api/pomclaw.api -dir ./ui/src/client

**⚠️ DO NOT EDIT**:
- `internal/handler/*.go` - Auto-generated HTTP handlers
- `internal/model/*_gen.go` - Auto-generated CRUD
- `internal/types/*.go` - Auto-generated request/response types
```

This generates a large number of protocol files for frontend/backend/data layers:

![Generated frontend/backend protocol files 1](docs/screenshots/generated_files_1.png)
![Generated frontend/backend protocol files 2](docs/screenshots/generated_files_2.png)
![Generated frontend/backend protocol files 3](docs/screenshots/generated_files_3.png)
![Generated frontend/backend protocol files 4](docs/screenshots/generated_files_4.png)

For these repeatedly-generatable code files, we don't have AI repeatedly confirm and modify them — this effectively limits AI's free rein, guarantees 100% code accuracy, and reduces token consumption.

---

## 🚀 Quick Start

### Prerequisites
- **Go 1.24+**、**Node.js 18+**
- **PostgreSQL 13+** (or Oracle)
- **SSH access to sandbox nodes**

### 1. Clone and Build

```bash
git clone https://github.com/pomclaw/pomclaw.git
cd pomclaw
make build  # automatically builds backend and frontend UI
```

> **Note**: `make build` automatically compiles the frontend UI (`npm run build`), compiles the backend binary, and packages the frontend into `dist/control-ui/`.

### 2. Initialize Database

```bash
createdb pomclaw
for f in docs/sql/*.sql; do psql pomclaw < $f; done
```

### 3. Start Gateway

```bash
./build/pomclaw  # Gateway runs on http://localhost:18790, serves frontend UI automatically
```

---

## 🔧 Configuration

Configuration is YAML format, located in the `etc/` directory:

- `etc/config.example.yaml` — full example
- `etc/local.yaml` — local development config

---

## 🤝 Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Principles

- **API First**: edit `docs/api/pomclaw.api` → `make generate` → implement business logic
- **Don't edit generated files**: handler, types, client directories are overwritten
- **Follow existing code style**: use `gofmt` and the project's ESLint configuration

---

## 📜 License

MIT License - see [LICENSE](LICENSE) file for details

---

## 📞 Support

- **Issue feedback**: [GitHub Issues](https://github.com/pomclaw/pomclaw/issues)
- **Discussions**: [GitHub Discussions](https://github.com/pomclaw/pomclaw/discussions)
- **Enterprise support**: contact@pomclaw.com

---

## 🎉 Acknowledgments

PomClaw builds on the excellent work of:
- go-zero and eino communities
- Open source database and SSH communities
- Go ecosystem contributors
