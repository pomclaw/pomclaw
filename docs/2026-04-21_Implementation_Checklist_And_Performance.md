# 多租户改造：实施检查清单 & 性能优化

**时间**：2026-04-21
**目的**：提供详细的实施步骤和性能优化建议

---

## 📋 实施检查清单

### Phase 1: 消息层改造

#### Step 1.1: InboundMessage 改造
- [ ] 修改 `pkg/bus/types.go`
  ```go
  type InboundMessage struct {
      AgentID    string              // 新增
      Channel    string
      // ... 其他字段
  }
  ```
- [ ] 搜索所有 InboundMessage 创建处，添加默认值或从 context 提取
  ```bash
  grep -r "InboundMessage{" --include="*.go" cmd/ pkg/
  ```
- [ ] 验证编译通过
  ```bash
  go build ./cmd/pomclaw
  ```

#### Step 1.2: 消息发送方适配
- [ ] **CLI 通道**：添加默认 agent_id
  ```go
  msg := bus.InboundMessage{
      AgentID: "default",  // 或从 config 读取
      Channel: "cli",
      // ...
  }
  ```

- [ ] **Cron 通道**：记录 agent_id
  ```go
  msg := bus.InboundMessage{
      AgentID: opts.AgentID,  // 从 cron job 中获取
      // ...
  }
  ```

- [ ] **其他通道**：根据实际情况实现

---

### Phase 2: Store 管理器实现

#### Step 2.1: 创建 StoreManager
- [ ] 新建 `pkg/storage/store_manager.go`
- [ ] 实现缓存机制（sync.Map）
- [ ] 实现 LRU 清理逻辑
- [ ] 单元测试
  ```bash
  go test ./pkg/storage/... -v
  ```

#### Step 2.2: 验证 Store 实现
- [ ] 检查 `pkg/oracle/session_store.go` - 验证 agent_id 隔离
  ```go
  // 应该有类似的查询
  WHERE agent_id = :1 AND session_key = :2
  ```

- [ ] 检查 `pkg/oracle/state_store.go` - 验证状态隔离
- [ ] 检查 `pkg/oracle/memory_store.go` - 验证记忆隔离

#### Step 2.3: 测试多 agent 隔离
- [ ] 写入 Test 验证不同 agent 的数据不会相互干扰
  ```go
  func TestStoreManager_MultiAgentIsolation(t *testing.T) {
      sm := NewStoreManager(cfg, db)

      // Agent 1 保存数据
      store1, _ := sm.GetSessionStore("agent-001")
      store1.AddMessage("session-1", "user", "Hello from A")

      // Agent 2 保存数据
      store2, _ := sm.GetSessionStore("agent-002")
      store2.AddMessage("session-1", "user", "Hello from B")

      // 验证隔离
      hist1 := store1.GetHistory("session-1")
      hist2 := store2.GetHistory("session-1")

      assert.NotEqual(t, hist1, hist2)  // 不应该相等
  }
  ```

---

### Phase 3: AgentLoop 改造

#### Step 3.1: 修改 processMessage
- [ ] 添加 agent_id 提取逻辑
- [ ] 添加日志记录
- [ ] 更新函数签名（如需要）
- [ ] 编译验证

#### Step 3.2: 改造 runAgentLoop
- [ ] 添加 agent_id 参数
- [ ] 从 StoreManager 获取隔离的 Store
  ```go
  sessionStore, _ := al.storeManager.GetSessionStore(opts.AgentID)
  stateStore, _ := al.storeManager.GetStateStore(opts.AgentID)
  memoryStore, _ := al.storeManager.GetMemoryStore(opts.AgentID)
  ```
- [ ] 所有数据操作改为使用隔离的 Store
- [ ] 单元测试验证

#### Step 3.3: 改造 runLLMIteration
- [ ] 工具执行时传递 agent_id 到 context
  ```go
  tools.WithToolContext(ctx, opts.Channel, opts.ChatID, opts.AgentID)
  ```
- [ ] 验证工具能正确获取 agent_id
- [ ] 并发测试（多 goroutine）

#### Step 3.4: 改造 maybeSummarize
- [ ] 传递隔离的 sessionStore
- [ ] 验证摘要也是隔离的

---

### Phase 4: 工具系统适配

#### Step 4.1: 扩展工具上下文
- [ ] 修改 `pkg/tools/base.go`
  ```go
  var ctxKeyAgentID = &toolCtxKey{"agentID"}

  func ToolAgentID(ctx context.Context) string {
      v, _ := ctx.Value(ctxKeyAgentID).(string)
      return v
  }
  ```

#### Step 4.2: 适配关键工具
- [ ] WriteDailyNoteTool
  ```go
  agentID := tools.ToolAgentID(ctx)
  // 使用 agentID 隔离日记
  ```

- [ ] CronTool - 可选（如需支持多 agent 调度）
- [ ] 其他工具 - 根据需要

---

### Phase 5: HTTP 通道实现

#### Step 5.1: 创建 HTTP 通道
- [ ] 新建 `pkg/channels/http_channel.go`
- [ ] 实现请求处理
- [ ] 实现响应转发
- [ ] 错误处理

#### Step 5.2: 测试 HTTP 通道
- [ ] 单一请求测试
  ```bash
  curl -X POST http://localhost:18790/api/v1/chat \
    -H "Content-Type: application/json" \
    -d '{
      "agent_id": "test-001",
      "session_key": "session_1",
      "message": "Hello"
    }'
  ```

- [ ] 并发请求测试（查看响应是否正确）
- [ ] 超时处理测试
- [ ] 错误处理测试

---

### Phase 6: 集成与启动

#### Step 6.1: 修改 main.go
- [ ] 创建 StoreManager
- [ ] 修改 AgentLoop 创建
- [ ] 启动 HTTP 通道
- [ ] 消息路由（outbound 消息转发）

#### Step 6.2: 编译测试
- [ ] 编译检查
  ```bash
  go build ./cmd/pomclaw
  ```
- [ ] 启动服务
  ```bash
  ./pomclaw gateway
  ```
- [ ] 日志检查

---

### Phase 7: 测试验证

#### Step 7.1: 单元测试
- [ ] 所有新代码覆盖率 > 80%
  ```bash
  go test ./pkg/... -cover -v
  ```

#### Step 7.2: 集成测试
- [ ] 端到端流程测试
  ```bash
  # 1. 启动服务
  ./pomclaw gateway &

  # 2. 发送测试请求
  curl -X POST http://localhost:18790/api/v1/chat \
    -d '{"agent_id": "test-001", "session_key": "s1", "message": "test"}'

  # 3. 验证数据隔离
  # 从数据库查看数据是否按 agent_id 正确隔离
  ```

#### Step 7.3: 压力测试
- [ ] 并发 1000+ 请求
  ```bash
  # 使用 wrk 或 Apache Bench
  ab -n 10000 -c 1000 \
    -p payload.json \
    -T application/json \
    http://localhost:18790/api/v1/chat
  ```

- [ ] 监控内存使用
- [ ] 检查缓存大小

#### Step 7.4: 数据隔离验证
- [ ] 查询数据库验证隔离
  ```sql
  -- 验证 session 隔离
  SELECT agent_id, COUNT(*) as session_count
  FROM POM_SESSIONS
  GROUP BY agent_id
  ORDER BY agent_id;

  -- 验证 state 隔离
  SELECT agent_id, COUNT(*) as state_count
  FROM POM_STATE
  GROUP BY agent_id
  ORDER BY agent_id;
  ```

---

## 🚀 性能优化建议

### 缓存优化

#### 1. LRU 缓存替代 sync.Map
**当前**：无限制缓存，可能 OOM

**改进**：使用 LRU 库
```go
import "github.com/hashicorp/golang-lru"

type StoreManager struct {
    sessionStoreLRU *lru.Cache  // 最多 5000 个
}

func (sm *StoreManager) GetSessionStore(agentID string) SessionStore {
    if store, ok := sm.sessionStoreLRU.Get(agentID); ok {
        return store.(SessionStore)
    }
    // ... 创建新实例 ...
}
```

#### 2. 二级缓存（Redis）
**适用场景**：多进程部署

```go
type StoreManager struct {
    localCache *lru.Cache      // 内存缓存
    redis      *redis.Client   // 分布式缓存
}

func (sm *StoreManager) GetSessionStore(agentID string) SessionStore {
    // 1. 检查本地缓存
    if store, ok := sm.localCache.Get(agentID); ok {
        return store.(SessionStore)
    }

    // 2. 检查 Redis
    data, _ := sm.redis.Get(ctx, "agent:"+agentID).Bytes()
    if len(data) > 0 {
        // 反序列化并缓存
        sm.localCache.Add(agentID, store)
        return store
    }

    // 3. 创建新实例
    // ...
}
```

#### 3. 缓存预热（可选）
```go
// 启动时加载常用 agent 的 store
func (sm *StoreManager) WarmupCache(agentIDs []string) {
    for _, agentID := range agentIDs {
        sm.GetSessionStore(agentID)
        sm.GetStateStore(agentID)
        sm.GetMemoryStore(agentID)
    }
}
```

---

### 数据库优化

#### 1. 索引优化
```sql
-- 确保已有这些索引
CREATE INDEX idx_sessions_agent_key
ON POM_SESSIONS(agent_id, session_key);

CREATE INDEX idx_state_agent_key
ON POM_STATE(agent_id, state_key);

CREATE INDEX idx_memories_agent_id
ON POM_MEMORIES(agent_id);
```

#### 2. 查询优化
```go
// 避免 N+1 查询
// 不要这样做：
for _, agentID := range agentIDs {
    store := sm.GetSessionStore(agentID)  // 每次都查 DB
}

// 改为：
stores := make(map[string]SessionStore)
for _, agentID := range agentIDs {
    stores[agentID], _ = sm.GetSessionStore(agentID)  // 使用缓存
}
```

#### 3. 批量操作
```go
// 如果需要保存多个 session，使用批量 INSERT
// 而非逐个 INSERT
```

---

### 并发优化

#### 1. 使用 sync.Pool 减少 GC
```go
var messagePool = sync.Pool{
    New: func() interface{} {
        return &providers.Message{}
    },
}

func getMessage() *providers.Message {
    return messagePool.Get().(*providers.Message)
}

func putMessage(m *providers.Message) {
    messagePool.Put(m)
}
```

#### 2. Goroutine 池控制
```go
// 限制并发工具执行数量
type toolExecutor struct {
    semaphore chan struct{}  // 容量为 100
}

func (te *toolExecutor) Execute(ctx context.Context, tool string, args map[string]interface{}) {
    te.semaphore <- struct{}{}        // 获取令牌
    defer func() { <-te.semaphore }() // 释放令牌

    // 执行工具
}
```

#### 3. Context 超时管理
```go
// 为每个请求添加超时
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()

response, err := agentLoop.ProcessDirect(ctx, ...)
```

---

### 监控和诊断

#### 1. 缓存命中率
```go
type StoreManager struct {
    // ...
    cacheMissCount int64
    cacheHitCount  int64
}

func (sm *StoreManager) GetCacheStats() map[string]interface{} {
    total := sm.cacheMissCount + sm.cacheHitCount
    hitRate := 0.0
    if total > 0 {
        hitRate = float64(sm.cacheHitCount) / float64(total) * 100
    }

    return map[string]interface{}{
        "hit_count":   sm.cacheHitCount,
        "miss_count":  sm.cacheMissCount,
        "hit_rate":    hitRate,
    }
}
```

#### 2. 性能指标
```go
// 添加到 main.go
func startMetricsServer() {
    http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        stats := map[string]interface{}{
            "cache_stats":    storeManager.GetCacheStats(),
            "goroutine_count": runtime.NumGoroutine(),
            "memory":         getMemStats(),
        }
        json.NewEncoder(w).Encode(stats)
    })

    go http.ListenAndServe(":9090", nil)
}

func getMemStats() map[string]interface{} {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    return map[string]interface{}{
        "allocated": m.Alloc,
        "total_allocated": m.TotalAlloc,
        "system": m.Sys,
        "gc_runs": m.NumGC,
    }
}
```

#### 3. 日志聚合
```go
// 在关键路径添加日志
logger.DebugCF("store_manager", "GetSessionStore",
    map[string]interface{}{
        "agent_id": agentID,
        "cache_hit": cacheHit,
        "duration_ms": duration.Milliseconds(),
    })
```

---

## 📊 性能基准（预期）

### 单机性能（8 核，16GB 内存）

| 指标 | 值 |
|------|-----|
| **吞吐量** | 5000+ req/sec |
| **P50 延迟** | < 50ms |
| **P99 延迟** | < 200ms |
| **缓存命中率** | > 90% |
| **内存占用** | < 500MB |
| **并发 Agent 数** | 5000+ |

### 压力测试命令

```bash
# 1. 简单吞吐量测试（1000 并发，10000 请求）
ab -n 10000 -c 1000 \
  -p payload.json \
  -T application/json \
  http://localhost:18790/api/v1/chat

# 2. 长连接测试（模拟真实客户端）
# 使用 wrk lua 脚本

# 3. 内存压力测试（10000+ 并发 agent）
for i in {1..10000}; do
  curl -X POST http://localhost:18790/api/v1/chat \
    -H "Content-Type: application/json" \
    -d "{\"agent_id\": \"agent-$i\", \"session_key\": \"s$i\", \"message\": \"test\"}" &

  if [ $((i % 100)) -eq 0 ]; then
    wait
    echo "Sent $i requests..."
  fi
done
wait

# 4. 内存监控
watch -n 1 'ps aux | grep pomclaw'
```

---

## 🔍 故障排查指南

### 问题 1: 缓存无限增长

**症状**：内存占用持续上升

**排查步骤**：
```go
// 1. 检查缓存大小
curl http://localhost:9090/metrics | jq '.cache_stats'

// 2. 如果缓存大小 > 5000，说明 LRU 清理未执行
// 3. 检查日志中的清理消息
tail -f logs/pomclaw.log | grep "Cache cleanup"
```

**解决方案**：
- 降低 MaxCachedAgents 阈值
- 增加清理频率

### 问题 2: 跨租户数据污染

**症状**：Agent A 的数据中出现 Agent B 的内容

**排查步骤**：
```sql
-- 1. 检查 session 数据
SELECT agent_id, session_key, COUNT(*) as msg_count
FROM POM_SESSIONS
GROUP BY agent_id, session_key
ORDER BY agent_id;

-- 2. 如果发现某个 session_key 有多个 agent_id
-- 这表示 session 被污染了

-- 3. 找出污染来源
SELECT * FROM POM_SESSIONS
WHERE session_key = 'PROBLEMATIC_KEY'
ORDER BY updated_at DESC;
```

**根本原因**：
- StoreManager 缓存使用错误的 agent_id
- SessionKey 冲突（应该由 agent_id 限定）

### 问题 3: 请求超时

**症状**：某些请求返回 "Request timeout"

**排查步骤**：
```bash
# 1. 检查 AgentLoop 是否卡住
curl http://localhost:9090/metrics | jq '.goroutine_count'

# 2. 检查 LLM 响应时间
tail -f logs/pomclaw.log | grep "LLM response"

# 3. 检查数据库连接
ps aux | grep -i postgres  # 或 oracle
```

**解决方案**：
- 增加超时时间（从 60s 到 120s）
- 检查数据库性能
- 检查 LLM API 响应时间

---

## 📝 发布清单

### 版本号
```
当前：v1.0.0（单 agent）
发布：v2.0.0（多租户）
```

### 发布文档
- [ ] API 文档（新增 HTTP 通道）
- [ ] 迁移指南（从单 agent 到多租户）
- [ ] 故障排查指南
- [ ] 性能调优指南

### 发布流程
1. 在测试环境验证
2. 灰度发布（10% 流量）
3. 监控关键指标
4. 全量发布

---

## ✅ 最后检查

在发布前，运行此清单：

```bash
# 1. 编译
go build ./cmd/pomclaw
[ $? -eq 0 ] && echo "✅ Compile OK" || echo "❌ Compile Failed"

# 2. 单元测试
go test ./pkg/... -v
[ $? -eq 0 ] && echo "✅ Unit Tests OK" || echo "❌ Unit Tests Failed"

# 3. 集成测试
./test-integration.sh
[ $? -eq 0 ] && echo "✅ Integration Tests OK" || echo "❌ Integration Tests Failed"

# 4. 性能测试
./test-performance.sh
[ $? -eq 0 ] && echo "✅ Performance OK" || echo "❌ Performance Issues"

# 5. 数据隔离验证
./test-isolation.sh
[ $? -eq 0 ] && echo "✅ Isolation OK" || echo "❌ Isolation Failed"

# 6. 向后兼容性
./test-backward-compat.sh
[ $? -eq 0 ] && echo "✅ Backward Compatibility OK" || echo "❌ Compatibility Issues"

echo "All checks passed! Ready to release."
```

---

## 📞 支持

如有问题，参考：
- 主要方案文档：`2026-04-21_HTTP_MultiTenant_AgentLoop_Architecture.md`
- 现有参考文档：
  - `MULTI_AGENT_PLAN.md`
  - `SESSION_AGENT_BINDING.md`
  - `ENTERPRISE_ARCHITECTURE.md`
