# Mattermost 多实例连接池设计文档

## 1. 背景

Pomclaw 原有 Mattermost 渠道仅支持单实例（单 bot token）连接。业务场景中每个用户有独立的 `im_bot_token`，需同时维护大量 WebSocket 连接到同一 Mattermost Server。用户信息存储在外部 MySQL 数据库的 `user_binding_relations` 表中。

## 2. 架构概览

```
Config (MySQLDBConfig)          Config (MattermostPoolConfig)
        |                               |
        v                               v
  pkg/mysql/                     pkg/channels/mattermost/
  +-- ConnectionManager          +-- MattermostPool (Channel 接口)
  +-- BindingStore               |   +-- instance#1 {token_1, ws_1}
        |  定时同步               |   +-- instance#2 {token_2, ws_2}
        +---------------------> |   +-- instance#N
                                 |
                          Channel Manager (一个 "mattermost" 条目)
                                 |
                          MessageBus
```

核心思想：
- MySQL 作为独立组件（`pkg/mysql/`），与 PostgreSQL/Oracle 同级
- `MattermostPool` 实现 `base.Channel` 接口，对 Channel Manager 完全透明
- 连接池通过依赖注入获得 `AccountStore`，不自行管理数据库连接

## 3. 文件结构

### 新增

| 文件 | 包 | 职责 |
|------|-----|------|
| `pkg/mysql/connection.go` | `mysql` | MySQL 连接管理：DSN 构建、连接池、Ping/Close |
| `pkg/mysql/binding_store.go` | `mysql` | 数据访问：BindingRecord 结构体、AccountStore 接口、BindingStore 实现 |
| `pkg/channels/mattermost/mattermost_pool.go` | `mattermost` | 连接池核心：Start/Stop/Send、syncLoop、chatIndex 路由 |

### 修改

| 文件 | 改动 |
|------|------|
| `pkg/channels/mattermost/mattermost.go` | 移除 `config.MattermostConfig` 依赖，新增 `NewMattermostChannelFromBinding()`，添加 `onInbound` 回调、`bindingID`/`agentID`/`familyID` 字段 |
| `pkg/config/config.go` | 新增顶层 `MySQLDBConfig`；`MattermostPoolConfig` 保留 mattermost 特有配置 |
| `pkg/channels/manager.go` | `initChannels()` 中创建 MySQL 连接 -> BindingStore -> 注入到 MattermostPool；出站消息分发使用 ants 协程池 |
| `go.mod` | 新增 `github.com/go-sql-driver/mysql`、`github.com/panjf2000/ants/v2` |

## 4. 核心机制

### 4.1 定时同步 (syncLoop)

每 `sync_interval_sec` 秒（默认 60）：

1. 从 MySQL 查询所有活跃 binding（`is_deleted=0 AND im_bot_token IS NOT NULL`）
2. **新增实例**：binding 不在 `instances` 中 -> 创建 `MattermostChannel` -> 启动 WebSocket
3. **更新实例**：binding 存在但 token/botUserID 变更 -> 停止旧实例 -> 创建新实例
4. **移除实例**：`instances` 中存在但不在活跃列表中 -> 停止并删除，清理 chatIndex

### 4.2 消息路由

**入站**：每个 `MattermostChannel` 收到 WebSocket 事件 -> `handlePosted()` -> 发布到 MessageBus（channel="mattermost"）。消息 Metadata 包含 `binding_id`、`agent_id`、`family_id`。收到消息时调用 `onInbound(chatID)` 注册路由映射。

**出站**：`MattermostPool.Send(msg)`
1. `chatIndex.Load(msg.ChatID)` -> 命中 -> 路由到对应子实例
2. 未命中 -> 遍历所有运行中实例尝试发送
3. 全部失败 -> 返回错误

## 5. 配置

```json
{
  "mysql": {
    "enabled": true,
    "host": "bj-cdb-xxxx.sql.tencentcdb.com",
    "port": 26517,
    "database": "kidclaw_family_agent",
    "user": "root",
    "password": "***",
    "pool_max_open": 10,
    "pool_max_idle": 2
  },
  "channels": {
    "mattermost": {
      "enabled": true,
      "server_url": "https://mm.example.com",
      "bindings_table": "user_binding_relations",
      "sync_interval_sec": 60,
      "max_connections": 500
    }
  }
}
```

### MySQL 配置（顶层）

| 字段 | 环境变量 | 说明 | 默认值 |
|------|---------|------|--------|
| `host` | `POM_MYSQL_HOST` | MySQL 主机 | `localhost` |
| `port` | `POM_MYSQL_PORT` | 端口 | `3306` |
| `database` | `POM_MYSQL_DATABASE` | 数据库名 | 空 |
| `user` | `POM_MYSQL_USER` | 用户名 | 空 |
| `password` | `POM_MYSQL_PASSWORD` | 密码 | 空 |
| `pool_max_open` | `POM_MYSQL_POOL_MAX_OPEN` | 最大连接数 | `10` |
| `pool_max_idle` | `POM_MYSQL_POOL_MAX_IDLE` | 最大空闲连接 | `2` |

### Mattermost Pool 配置

| 字段 | 环境变量 | 说明 | 默认值 |
|------|---------|------|--------|
| `server_url` | `POMCLAW_CHANNELS_MATTERMOST_SERVER_URL` | Mattermost Server 地址 | 必填 |
| `bindings_table` | `POMCLAW_CHANNELS_MATTERMOST_BINDINGS_TABLE` | 用户绑定关系表名 | `user_binding_relations` |
| `sync_interval_sec` | `POMCLAW_CHANNELS_MATTERMOST_SYNC_INTERVAL_SEC` | 同步间隔（秒） | `60` |
| `max_connections` | `POMCLAW_CHANNELS_MATTERMOST_MAX_CONNECTIONS` | 最大 WebSocket 连接数 | `500` |

## 6. 外部 MySQL 数据源

```sql
-- user_binding_relations 表（只读，已由外部服务维护）
SELECT id, user_id, user_name, family_id, agent_id,
       im_bot_token, im_user_token, im_channel_id,
       mattermost_bot_user_id, mattermost_user_id
FROM user_binding_relations
WHERE is_deleted = 0 AND im_bot_token IS NOT NULL AND im_bot_token != ''
```

## 7. 依赖注入流程

```
manager.go initChannels():
  1. mysqlpkg.NewBindingStore(db, bindingsTable)         -> AccountStore
  2. mattermostpkg.NewMattermostPool(cfg, store, bus)     -> MattermostPool（内部创建 ants 协程池）
```

`MattermostPool` 不感知 MySQL 连接细节，只依赖 `AccountStore` 接口。未来如果数据源变更（如改为 API 调用），只需替换 `AccountStore` 实现。

## 8. WebSocket 心跳保活

### 问题

Mattermost Server 通过反向代理（nginx）暴露，代理 `proxy_read_timeout` 为 60 秒。Mattermost Server 每 60 秒发送一次 WebSocket PING，但代理在 PING 到达前就因空闲超时切断了连接。

### 方案

pomclaw 采用心跳方案：每 30 秒通过 `gorilla/websocket` 的 `WriteControl(PingMessage)` 发送客户端 PING。此方法线程安全，开销极低，且保证零消息丢失。

### 时序

```
t=0s    WebSocket 建立，收到 hello + auth OK
t=30s   客户端 PING -> 代理转发 -> 服务端返回 PONG -> 代理超时计时器重置
t=60s   服务端 PING 到达 -> 客户端自动回复 PONG
t=60s   客户端第二次 PING -> 再次重置代理超时
        ...连接稳定保持...
```

## 9. 连接稳定性与重连

### 连续重连保护

`eventLoop` 维护 `consecutiveReconnects` 计数器：
- 每次成功接收事件 -> 计数器归零
- 每次 EventChannel 关闭或 PingTimeout -> 计数器 +1
- 达到 5 次连续重连 -> 停止该实例，避免无限重连风暴

### 重连退避

`reconnect()` 方法内部循环重试，延迟递增：1s -> 2s -> 3s -> ... -> 30s 封顶。每次重试前检查 `ctx` 是否已取消。

### 竞态安全

- `Stop()` 调用 `cancel()` 后加锁置空 `wsClient`
- `reconnect()` 在持锁期间检查 `ctx.Err()`，防止 Stop 和 reconnect 并发时泄漏新连接
- `closeWS()` 先设置 `ReadDeadline` 为 `time.Now()` 使阻塞的 `ReadMessage` 立即返回，再 `Close()`

## 10. 并发模型

### 连接建立/销毁 — ants 协程池

`syncOnce()` 分三阶段执行：

- **Phase 1（计算 diff）**：持写锁比较 instances 与 MySQL binding 列表，计算 toStop/toStart
- **Phase 2（停止旧实例）**：所有待停止实例并发执行 `ch.Stop()`，`sync.WaitGroup` 等待全部完成
- **Phase 3（启动新实例）**：通过 `ants.Pool`（预分配 20 worker）提交启动任务，`sync.WaitGroup` 等待本轮全部完成

ants 协程池在 `NewMattermostPool()` 时创建，`Stop()` 时释放。采用预分配模式（`WithPreAlloc`），worker goroutine 常驻复用，避免高频 sync 周期下的 goroutine 创建/销毁开销和 GC 压力。

```
syncOnce Phase 3:

  ants.Pool (size=20, pre-allocated workers)
  +---+---+---+- ... -+----+
  | W1| W2| W3|       | W20|  <-- 20 个常驻 worker goroutine
  +-+-+-+-+-+-+       +-+--+
    |   |   |           |
    v   v   v           v
  Start Start Start   Start  <-- 每个执行: NewChannel -> ch.Start (HTTP auth + WS dial)
    |   |   |           |
    v   v   v           v
  p.mu.Lock -> p.instances[id] = ch -> p.mu.Unlock
```

Pool 满时 `Submit()` 阻塞，与旧信号量语义一致；Submit 失败时正确处理 WaitGroup 计数平衡。

### 消息接收 — 每实例独立协程

每个 `MattermostChannel.Start()` 启动一个 `go eventLoop()` 协程，负责：
- 从 WebSocket `EventChannel` 读取事件
- 发送 30s 心跳 PING
- 处理 `handlePosted()` -> `PublishInbound()` 写入 MessageBus

### 消息发送 — ants 协程池 dispatcher

`manager.go` 的 `dispatchOutbound()` 从 MessageBus 读取出站消息，通过 `ants.Pool`（预分配 50 worker）分发：

```go
m.sendPool.Submit(func() {
    sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    ch.Send(sendCtx, outMsg)
})
```

ants 协程池在 `NewManager()` 时创建，`StopAll()` 时释放。worker 常驻复用，一个 bot 的 `Send` 慢不会阻塞其他 bot 或其他 channel 的消息发送。

### Bot 间回声防护

`MattermostChannel.handlePosted()` 中：
1. `post.UserId == c.botUserID` -> 跳过自己的消息
2. `c.allBotIDs.Load(post.UserId)` -> 跳过 pool 内所有其他 bot 的消息

`allBotIDs` 是 `*sync.Map`，由 `MattermostPool` 在启动实例时注入，所有实例共享同一个 map。

## 11. LLM 流式输出 (Streaming)

### 11.1 背景

LLM 生成长回复时，用户需等待整个生成完成才能看到消息。流式输出在 LLM 生成过程中实时更新 Mattermost 帖子，实现"打字机"效果，显著改善用户体验。

### 11.2 架构

```
                          Provider 层                 Bus 层               Channel 层
                    ┌─────────────────┐       ┌──────────────┐     ┌───────────────────┐
                    │ StreamingProvider│       │  MessageBus  │     │  Manager          │
                    │  .ChatStream()  │       │  .streamer   │     │  (StreamDelegate) │
                    └───────┬─────────┘       └──────┬───────┘     └───────┬───────────┘
                            │                        │                     │
   Agent Loop               │  onChunk callback      │  GetStreamer()      │  GetStreamer()
   callProviderWithStreaming │◄──────────────────     │◄────────────       │◄────────────
          │                 │                        │                     │
          │  accumulated    │                        │                     ▼
          │  text chunks    │                        │             ┌───────────────────┐
          ▼                 │                        │             │ MattermostPool    │
   streamer.Update()────────┼────────────────────────┼────────────►│  .NewStreamer()   │
   streamer.Finalize()      │                        │             │       │           │
   streamer.Cancel()        │                        │             │       ▼           │
                            │                        │             │ MattermostChannel │
                            │                        │             │  .NewStreamer()   │
                            │                        │             └───────┬───────────┘
                            │                        │                     │
                            │                        │                     ▼
                            │                        │             mattermostStreamer
                            │                        │              CreatePost / UpdatePost
```

### 11.3 接口定义

#### StreamingProvider (`pkg/providers/types.go`)

```go
type StreamingProvider interface {
    LLMProvider
    ChatStream(ctx context.Context, messages []Message, tools []ToolDefinition,
        model string, options map[string]interface{},
        onChunk func(accumulated string)) (*LLMResponse, error)
}
```

`onChunk` 回调在每次收到增量文本时触发，参数为**已累积的完整文本**（非 delta）。已实现：`HTTPProvider`、`CodexProvider`。

#### Streamer / StreamDelegate (`pkg/bus/stream.go`)

```go
type Streamer interface {
    Update(ctx context.Context, content string) error   // 增量更新
    Finalize(ctx context.Context, content string) error  // 最终确认
    Cancel(ctx context.Context)                          // 取消（删除中间态帖子）
    HasPosted() bool                                     // 是否已创建过帖子
}

type StreamDelegate interface {
    GetStreamer(ctx context.Context, channel, chatID string) (Streamer, bool)
}
```

### 11.4 文件清单

| 文件 | 职责 |
|------|------|
| `pkg/bus/stream.go` | `Streamer` 和 `StreamDelegate` 接口定义 |
| `pkg/providers/types.go` | `StreamingProvider` 接口定义 |
| `pkg/providers/http_provider.go` | `HTTPProvider.ChatStream()` — OpenAI 兼容 SSE 流解析 |
| `pkg/providers/codex_provider.go` | `CodexProvider.ChatStream()` — OpenAI Responses API 流解析 |
| `pkg/agent/loop.go` | `callProviderWithStreaming()` — 协调 Provider 与 Streamer |
| `pkg/channels/mattermost/mattermost.go` | `mattermostStreamer` — 创建/更新/删除 Mattermost 帖子 |
| `pkg/channels/mattermost/mattermost_pool.go` | `MattermostPool.NewStreamer()` — 路由到正确的子实例 |
| `pkg/channels/manager.go` | `Manager.GetStreamer()` — 实现 `StreamDelegate`，桥接 Bus 与 Channel |
| `pkg/tools/message.go` | `MarkSentInRound()` — 防止流式完成后重复发送 |

### 11.5 完整链路

#### 初始化

```
main.go
  → channels.NewManager(cfg, bus, opts...)
    → messageBus.SetStreamDelegate(manager)   // Manager 注册为 StreamDelegate
```

#### 运行时流程

```
1. AgentLoop.runLLMIteration()
   → callProviderWithStreaming(ctx, messages, tools, llmOpts, opts, iteration)

2. 前置检查（任一不满足则 fallback 到非流式 Chat）：
   ├── provider 实现 StreamingProvider？
   ├── bus 非 nil？
   ├── opts.Channel 和 opts.ChatID 非空？
   ├── 非内部渠道 (constants.IsInternalChannel)？
   └── bus.GetStreamer() 返回可用 Streamer？

3. 调用 streamingProvider.ChatStream(ctx, ..., onChunk)
   │
   │  HTTPProvider 内部：
   │  ├── 发送 POST /chat/completions, stream=true
   │  ├── 逐行解析 SSE data: {...}
   │  ├── 每收到 delta → contentBuilder.WriteString(delta)
   │  └── 调用 onChunk(contentBuilder.String())
   │
   │  onChunk 回调：
   │  ├── 跳过空内容和重复内容
   │  ├── streamer.Update(ctx, accumulated)
   │  │     首次 → Mattermost CreatePost（创建帖子）
   │  │     后续 → Mattermost UpdatePost（原地更新）
   │  └── Update 失败 → streamer.Cancel() + streamer 置 nil

4. ChatStream 返回 LLMResponse
   ├── 有 ToolCalls → streamer.Cancel()（删除中间帖子，后续走工具调用流程）
   ├── Finalize 成功 → MarkSentInRound()（防止 outbound 重复发送）
   └── Finalize 失败且已创建帖子 → MarkSentInRound()（防止重复）
```

### 11.6 重试策略

Agent Loop 的 LLM 调用支持最多 3 次重试。**仅首次尝试使用流式输出**，重试时 fallback 到非流式 `Chat()`，避免用户看到多个"正在输入"的帖子。

```go
for retry := 0; retry <= maxRetries; retry++ {
    if retry == 0 {
        response, err = al.callProviderWithStreaming(...)
    } else {
        response, err = al.provider.Chat(...)  // 重试不走流式
    }
}
```

### 11.7 mattermostStreamer 实现

`mattermostStreamer` 管理单次流式回复的 Mattermost 帖子生命周期：

| 方法 | 行为 |
|------|------|
| `Update(content)` | 首次调用 `CreatePost`，后续调用 `UpdatePost` 原地更新 |
| `Finalize(content)` | 持锁执行最终 Update + 设置 finalized + ackPending（添加 ✅ 反应） |
| `Cancel()` | 设置 finalized + 删除已创建的帖子 |
| `HasPosted()` | 返回 postID 是否非空 |

**线程安全**：所有方法通过 `sync.Mutex` 保护 `postID` 和 `finalized` 状态。`Finalize` 使用内部 `doUpdateLocked()` 方法在持锁状态下完成更新和标记，避免与 `Cancel` 的竞态窗口。

### 11.8 MattermostPool 的 Streamer 路由

`MattermostPool.NewStreamer()` 通过 `chatIndex` 精确路由到拥有该 chatID 的子实例：

```go
func (p *MattermostPool) NewStreamer(ctx, chatID) (Streamer, bool) {
    val, ok := p.chatIndex.Load(chatID)  // O(1) 查找
    if !ok { return nil, false }         // 未命中直接返回，不盲目 fallback
    ch := p.instances[val.(int)]
    return ch.NewStreamer(ctx, chatID)
}
```

不使用 fallback 遍历，确保不会用错误的 Bot Token 发送消息。

### 11.9 防重复发送

流式输出完成后，`MarkSentInRound()` 设置 `MessageTool.sentInRound = true`。Agent Loop 的 outbound 逻辑检查此标志，跳过已通过流式发送的回复，避免用户收到两条相同消息。

### 11.10 超时保护

`HTTPProvider.ChatStream()` 在调用方 context 无 deadline 时自动添加 5 分钟超时，防止 SSE 流 hang 住导致永久阻塞：

```go
if _, hasDeadline := ctx.Deadline(); !hasDeadline {
    ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()
}
```
