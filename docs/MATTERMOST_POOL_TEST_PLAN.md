# Mattermost 连接池测试方案

## 1. 单元测试

### 1.1 现有测试（`mattermost_test.go`）

| 测试 | 说明 |
|------|------|
| `TestParseMattermostChatID` | chatID 解析 |
| `TestMattermostWSURL` | HTTP->WS URL 转换 |
| `TestNewMattermostChannelFromBinding` | 参数校验和构造 |
| `TestMattermostSendNotRunning` | 非运行状态发送 |
| `TestMattermostSendInvalidChatID` | 空 chatID 发送 |

### 1.2 待补充

#### MattermostPool 核心逻辑

```
TestNewMattermostPool_MissingServerURL       -- 缺少 server_url 返回错误
TestNewMattermostPool_DefaultValues          -- 验证 SyncIntervalSec、MaxConnections 默认值
TestMattermostPool_SendNotRunning            -- 未启动时发送消息返回错误
TestMattermostPool_ChatIndexRouting          -- chatIndex 命中时路由到正确实例
TestMattermostPool_ChatIndexMiss             -- chatIndex 未命中时遍历所有实例
TestMattermostPool_CleanChatIndex            -- 清理指定 bindingID 的 chatIndex 条目
```

#### syncOnce 逻辑（需 mock AccountStore）

```
TestSyncOnce_NewBinding                     -- 新 binding -> 创建实例
TestSyncOnce_RemovedBinding                 -- binding 消失 -> 停止实例、清理 chatIndex
TestSyncOnce_ChangedBinding                 -- token 变更 -> 重建实例
TestSyncOnce_UnchangedBinding               -- 无变化 -> 不重建
TestSyncOnce_MaxConnectionsLimit            -- 超过 max_connections -> 停止创建
```

#### MySQL 组件（`pkg/mysql/`）

```
TestBuildDSN                                -- 验证 DSN 格式正确
TestBindingStore_ListActive                 -- 查询活跃记录
TestBindingStore_FilterDeleted              -- is_deleted=1 被过滤
TestBindingStore_FilterEmptyToken           -- 空 token 被过滤
```

### 1.3 Mock 实现

```go
type MockAccountStore struct {
    bindings []mysql.BindingRecord
    err      error
}
func (m *MockAccountStore) ListActive(_ context.Context) ([]mysql.BindingRecord, error) {
    return m.bindings, m.err
}
func (m *MockAccountStore) Close() error { return nil }
```

## 2. 集成测试

### 环境变量

```
MM_SERVER_URL=https://mm.example.com
MM_BOT_TOKEN=xoxb-...
MM_CHANNEL_ID=f6c1msw84...
POM_MYSQL_HOST=...
POM_MYSQL_PORT=26517
POM_MYSQL_DATABASE=kidclaw_family_agent
POM_MYSQL_USER=root
POM_MYSQL_PASSWORD=***
```

### 测试场景

| 编号 | 场景 | 验证点 |
|------|------|--------|
| IT-1 | 单实例连接 | bot token 认证、WebSocket 连接、消息收发 |
| IT-2 | Pool 启动多个 binding | 全部建立 WebSocket、各自独立收发 |
| IT-3 | 运行中新增 binding | sync 后新实例自动启动 |
| IT-4 | 运行中删除 binding | sync 后旧实例自动停止、chatIndex 清理 |
| IT-5 | Token 变更 | sync 后旧实例停止、新实例启动 |
| IT-6 | WebSocket 断连重连 | 自动重连、消息恢复 |
| IT-7 | Pool Stop | 所有实例停止、ants pool 释放 |
| IT-8 | MySQL 连接 | 连接池建立和查询 |

### 现有集成测试

- `TestMattermostIntegration` -- 单实例认证和发送
- `TestMattermostE2E` -- 用户发消息->bot 接收->验证内容
- `TestMattermostListen` -- 交互式监听（手动测试辅助）

## 3. 性能测试

| 指标 | 目标 |
|------|------|
| 启动 200 个 WebSocket 连接 | < 30s |
| sync 查询 500 条 binding | < 1s |
| 内存占用（200 连接） | < 500MB |
| chatIndex 10000 条目查找 | < 1ms |
| 单实例消息发送延迟 | < 200ms |

## 4. 执行命令

```bash
# 单元测试
go test ./pkg/mysql/... ./pkg/channels/... ./pkg/config/... -short -v -count=1

# 完整测试（需设置环境变量）
MM_SERVER_URL=... MM_BOT_TOKEN=... MM_CHANNEL_ID=... \
  go test ./pkg/channels/mattermost/ -v -count=1

# 静态检查
go vet ./pkg/mysql/... ./pkg/channels/... ./pkg/config/...

# 构建验证
go build ./pkg/mysql/... ./pkg/channels/... ./pkg/config/...
```
