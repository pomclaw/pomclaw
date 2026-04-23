# 多租户改造：风险管理 & 回滚方案

**时间**：2026-04-21
**目的**：识别风险、制定回滚策略、保证业务连续性

---

## 🚨 风险识别矩阵

### 高风险项 (🔴)

#### R1: 数据隔离失败 - 跨租户污染
**影响**：严重（客户数据混乱）
**概率**：中
**根本原因**：
- SessionKey 冲突（不同 agent 使用相同 key）
- Cache key 使用错误（agent_id 丢失）
- SessionStore 实现缺陷

**检测方法**：
```sql
-- 检查某个 session 是否有多个 agent
SELECT agent_id, COUNT(DISTINCT agent_id) as agent_count
FROM POM_SESSIONS
WHERE session_key = 'SUSPECTED_KEY'
GROUP BY session_key
HAVING COUNT(DISTINCT agent_id) > 1;
```

**缓解方案**：
```go
// 1. 强制 SessionKey 格式为：{agent_id}:{unique_id}
// 这样从 key 本身就能验证隔离
func ValidateSessionKey(sessionKey, agentID string) error {
    parts := strings.Split(sessionKey, ":")
    if len(parts) != 2 || parts[0] != agentID {
        return fmt.Errorf("session key mismatch: key=%s, agent=%s", sessionKey, agentID)
    }
    return nil
}

// 2. 在 Store 层加入验证
func (s *SessionStore) GetHistory(sessionKey string) []Message {
    // 验证 sessionKey 是否属于这个 agent
    if !strings.HasPrefix(sessionKey, s.agentID+":") {
        panic("SecurityError: session key agent mismatch")
    }
    // ... 继续正常流程
}
```

**监控告警**：
```go
// 定时检查数据隔离
func MonitorDataIsolation(db *sql.DB) {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        var violationCount int
        db.QueryRow(`
            SELECT COUNT(*)
            FROM POM_SESSIONS
            WHERE session_key IN (
                SELECT session_key
                FROM POM_SESSIONS
                GROUP BY session_key
                HAVING COUNT(DISTINCT agent_id) > 1
            )
        `).Scan(&violationCount)

        if violationCount > 0 {
            logger.AlertCF("security", "Data isolation violation detected",
                map[string]interface{}{"violation_count": violationCount})
            // 发出告警
        }
    }
}
```

---

#### R2: 内存爆炸 - 缓存无限增长
**影响**：严重（服务 OOM 宕机）
**概率**：中高
**根本原因**：
- 10000+ agent 同时活跃，Store 缓存无限制增长
- LRU 清理失败或未执行

**检测方法**：
```bash
# 监控内存占用
watch -n 5 'ps aux | grep pomclaw | grep -v grep'

# 检查缓存大小
curl http://localhost:9090/metrics | jq '.cache_stats'
```

**缓解方案**：
```go
// 1. 使用真正的 LRU（非简单的 sync.Map）
import "github.com/hashicorp/golang-lru/v2"

type StoreManager struct {
    sessionStoreLRU *lru.Cache[string, SessionStore]  // 最多 5000
    // ...
}

func NewStoreManager(cfg *config.Config, db *sql.DB) *StoreManager {
    sessionLRU, _ := lru.New[string, SessionStore](5000)

    return &StoreManager{
        sessionStoreLRU: sessionLRU,
        // ...
    }
}

// 2. 添加监控和自动告警
func (sm *StoreManager) MonitorCacheSize() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        size := sm.sessionStoreLRU.Len()
        usage := float64(size) / 5000.0 * 100

        logger.DebugCF("store_manager", "Cache size",
            map[string]interface{}{
                "size": size,
                "usage_percent": usage,
            })

        if usage > 80 {
            logger.WarnCF("store_manager", "Cache usage high",
                map[string]interface{}{"usage_percent": usage})
        }
        if usage > 95 {
            logger.AlertCF("store_manager", "Cache nearly full",
                map[string]interface{}{"usage_percent": usage})
        }
    }
}

// 3. 设置内存限制
func (sm *StoreManager) CheckMemoryUsage() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    if m.Alloc > 1000*1024*1024 {  // > 1GB
        logger.AlertCF("store_manager", "Memory usage critical",
            map[string]interface{}{
                "allocated_mb": m.Alloc / 1024 / 1024,
            })
        // 强制清理某些缓存
        sm.sessionStoreLRU.Purge()
    }
}
```

**监控告警**：
- 内存占用 > 500MB：警告
- 内存占用 > 1GB：严重告警，尝试清理
- 内存占用 > 2GB：熔断，返回 503

---

#### R3: 并发竞态条件
**影响**：中（数据不一致）
**概率**：低（但很难调试）
**根本原因**：
- 多 goroutine 并发访问同一个 Store
- 缺少同步机制

**检测方法**：
```bash
# 运行 Go race detector
go test ./pkg/agent/... -race

go build -race ./cmd/pomclaw
```

**缓解方案**：
```go
// 1. Store 实现内部线程安全
type SessionStore struct {
    mu       sync.RWMutex  // 保护 sessions map
    sessions map[string]*Session
}

// 2. 在关键操作加锁
func (s *SessionStore) AddMessage(key, role, content string) {
    s.mu.Lock()
    defer s.mu.Unlock()

    session := s.sessions[key]
    if session == nil {
        session = &Session{}
        s.sessions[key] = session
    }
    session.Messages = append(session.Messages, Message{
        Role:    role,
        Content: content,
    })
}

// 3. 定期运行 race 检测
// 在 CI/CD 中添加：
run_race_tests() {
    go test -race ./pkg/... 2>&1 | tee race-test.log
    if grep -q "WARNING: DATA RACE" race-test.log; then
        echo "Race condition detected!"
        exit 1
    fi
}
```

---

### 中风险项 (🟠)

#### R4: HTTP 通道连接泄漏
**影响**：中（连接用尽）
**缓解**：
```go
// 1. 设置连接超时
const ResponseTimeout = 60 * time.Second

// 2. 定期清理过期连接
func (hc *HTTPChannel) CleanupExpiredConnections() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        now := time.Now()
        hc.clients.Range(func(key, value interface{}) bool {
            conn := value.(*clientConnection)
            if now.Sub(conn.lastActivity) > ResponseTimeout+10*time.Second {
                logger.WarnCF("http_channel", "Cleaning up expired connection",
                    map[string]interface{}{"client_id": key})
                hc.clients.Delete(key)
                close(conn.responseChan)
            }
            return true
        })
    }
}
```

#### R5: 向后兼容性破坏
**影响**：中（现有客户端报错）
**缓解**：
```go
// 1. InboundMessage.AgentID 字段添加默认值
msg := bus.InboundMessage{
    AgentID: "default",  // 如果未指定
    // ...
}

// 2. 在 HTTP 通道验证
if req.AgentID == "" {
    req.AgentID = extractAgentIDFromAuth(r)  // 从 JWT 提取
    if req.AgentID == "" {
        req.AgentID = "default"  // 最后的降级
    }
}
```

---

### 低风险项 (🟡)

#### R6: 性能下降
**影响**：低（用户体验下降）
**缓解**：参考性能优化文档

#### R7: 文档过期
**影响**：低（维护成本增加）
**缓解**：
- 维护版本控制的文档
- 在代码中添加注释说明 agent_id 隔离逻辑

---

## 🔄 阶段化发布策略

### Phase 0: 预发布（3-5 天内）

#### Pre 0.1: 完整测试
- [ ] 所有单元测试通过（go test -race）
- [ ] 集成测试覆盖 80%+
- [ ] 压力测试 1000+ 并发
- [ ] 数据隔离验证通过

#### Pre 0.2: 性能基准
- [ ] 记录单 agent 性能
- [ ] 记录 100 agent 性能
- [ ] 记录 1000 agent 性能
- [ ] 确认无性能退化

#### Pre 0.3: 兼容性检查
- [ ] 旧版本数据能正确读取
- [ ] 旧版本客户端能降级处理
- [ ] 数据迁移无损

---

### Phase 1: 灰度发布（5-10%）

**目标**：识别隐藏的问题

**发布方式**：
```go
// 在 load balancer 或 API gateway 层控制
func ShouldUseNewVersion(userID string) bool {
    hash := crc32.ChecksumIEEE([]byte(userID))
    percentage := (hash % 100) + 1
    return percentage <= 5  // 5% 用户使用新版本
}
```

**监控指标**：
- 错误率（目标：< 0.1%）
- 响应时间（P99 < 200ms）
- 数据污染（0 个）
- 内存占用（< 500MB）

**告警阈值**：
- 错误率 > 0.5%：立即回滚
- P99 > 500ms：立即回滚
- 检测到数据污染：立即回滚
- 内存占用 > 1GB：立即回滚

---

### Phase 2: 扩大发布（50%）

**前置条件**：
- [ ] Phase 1 运行 24 小时无异常
- [ ] 错误率 < 0.1%
- [ ] 无严重问题反馈

**发布方式**：
```yaml
# 在 Kubernetes 中
replicas: 10
strategy:
  canary:
    steps:
      - setWeight: 50
        duration: 2h
```

**监控**：
- 同上，继续严格监控

---

### Phase 3: 全量发布（100%）

**前置条件**：
- [ ] Phase 2 运行 48 小时无异常
- [ ] 用户反馈积极
- [ ] 性能指标达预期

---

## 📋 回滚方案

### 快速回滚（5 分钟内）

#### 场景 1: 严重 Bug（OOM、数据污染）
```bash
#!/bin/bash
# 1. 立即停止新版本
kubectl scale deployment pomclaw-new --replicas=0

# 2. 验证旧版本状态
kubectl get pods -l version=v1

# 3. 确认回滚
kubectl scale deployment pomclaw-v1 --replicas=10

# 4. 监控错误率
watch kubectl top pods
```

#### 场景 2: 性能问题（P99 > 1s）
```bash
# 1. 关闭新功能（StoreManager）
# 在代码中添加 feature flag
if os.Getenv("DISABLE_MULTITENANT") == "true" {
    // 使用旧的单 agent 逻辑
}

# 2. 环境变量禁用新版本
kubectl set env deployment/pomclaw \
    DISABLE_MULTITENANT=true

# 3. 逐个重启 Pod
kubectl rollout restart deployment/pomclaw

# 4. 监控恢复
```

---

### 完整回滚（不兼容改动）

#### 场景：发现无法修复的架构问题

**回滚步骤**：
```bash
# 1. 立即切换回旧版本
git tag rollback-v1.0.0 <old-commit>
docker build -t pomclaw:rollback .
docker push pomclaw:rollback

# 2. 部署旧版本
kubectl set image deployment/pomclaw \
    pomclaw=pomclaw:rollback

# 3. 清理新版本资源
kubectl delete deployment pomclaw-new
kubectl delete pvc pomclaw-cache  # 如有的话

# 4. 数据恢复（如需要）
# 备份新版本产生的数据
mysqldump pomclaw > backup-new-$(date +%s).sql

# 从旧版本备份恢复
mysql pomclaw < backup-old-version.sql

# 5. 验证
curl http://api.pomclaw.com/health
```

---

### 数据恢复

#### 场景：新版本产生了坏数据

**恢复步骤**：
```sql
-- 1. 备份坏数据
CREATE TABLE POM_SESSIONS_BAD AS
SELECT * FROM POM_SESSIONS
WHERE agent_id NOT IN (SELECT agent_id FROM POM_AGENT_CONFIGS);

-- 2. 清理坏数据
DELETE FROM POM_SESSIONS
WHERE agent_id NOT IN (SELECT agent_id FROM POM_AGENT_CONFIGS);

-- 3. 从备份恢复好数据
INSERT INTO POM_SESSIONS
SELECT * FROM POM_SESSIONS_BACKUP
WHERE created_at > '2026-04-21 10:00:00'
  AND created_at < '2026-04-21 12:00:00';

-- 4. 验证
SELECT COUNT(*) FROM POM_SESSIONS;
SELECT agent_id, COUNT(*) FROM POM_SESSIONS GROUP BY agent_id;
```

---

### 部分回滚（granular）

#### 场景：某个 Agent 受影响

```sql
-- 1. 识别受影响的 agent
SELECT DISTINCT agent_id
FROM POM_SESSIONS
WHERE updated_at > '2026-04-21 10:00:00'
  AND (
    -- 数据异常条件
    character_length(content) > 100000 OR
    updated_at - created_at > interval '1 hour'
  );

-- 2. 隔离其数据
BEGIN TRANSACTION;
  DELETE FROM POM_SESSIONS WHERE agent_id = 'bad-agent-id';
  INSERT INTO POM_SESSIONS
    SELECT * FROM POM_SESSIONS_BACKUP
    WHERE agent_id = 'bad-agent-id'
      AND created_at > '2026-04-21 09:00:00';
COMMIT;

-- 3. 通知受影响用户
-- 发送补偿方案
```

---

## 📊 监控仪表板

### 关键指标

```go
type HealthMetrics struct {
    // 业务指标
    ActiveAgents       int           // 活跃 agent 数
    SessionsPerSecond  float64       // 每秒新建 session 数
    AvgResponseTime    time.Duration // 平均响应时间
    ErrorRate          float64       // 错误率（%）

    // 系统指标
    MemoryUsage        uint64        // 内存占用（字节）
    GoroutineCount     int           // Goroutine 数
    DatabaseConnPool   int           // 数据库连接
    CacheHitRate       float64       // 缓存命中率（%）

    // 隔离指标
    DataIsolationOK    bool          // 数据隔离正常
    CrossTenantLeaks   int           // 跨租户泄露数
}

// 暴露指标端点
func (hm *HealthMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    data := map[string]interface{}{
        "timestamp": time.Now().Unix(),
        "metrics": hm,
        "status": hm.GetStatus(),  // "healthy", "warning", "critical"
    }
    json.NewEncoder(w).Encode(data)
}
```

### 告警规则

```yaml
# prometheus alerts.yml
groups:
  - name: pomclaw
    rules:
      - alert: HighErrorRate
        expr: error_rate > 0.001  # 0.1%
        for: 5m
        action: page

      - alert: HighMemoryUsage
        expr: memory_usage > 1000000000  # 1GB
        for: 2m
        action: page

      - alert: DataIsolationFailure
        expr: cross_tenant_leaks > 0
        for: 1m
        action: page_critical  # 立即通知

      - alert: CacheFullness
        expr: cache_usage > 95
        for: 1m
        action: warn
```

---

## 🔒 安全检查清单

### 发布前验证

- [ ] **数据隔离**
  ```bash
  # 验证 5 个随机 agent 的数据不混乱
  ./test-isolation.sh --agents 5 --iterations 100
  ```

- [ ] **访问控制**
  ```bash
  # 验证 A 用户不能访问 B 用户数据
  ./test-access-control.sh
  ```

- [ ] **加密**
  ```bash
  # 验证敏感数据加密
  ./test-encryption.sh
  ```

- [ ] **审计日志**
  ```bash
  # 验证所有关键操作已记录
  ./test-audit-logs.sh
  ```

### 持续监控

- [ ] **异常检测**
  - 监控跨租户访问
  - 监控大量数据导出
  - 监控权限变更

---

## 📞 应急联系

在发生严重问题时：

1. **立即通知**
   - 技术主管
   - DevOps 团队
   - 客户支持

2. **根因分析**
   - 查看错误日志
   - 查看监控数据
   - 检查最近改动

3. **应急响应**
   - 如 < 5min 内能修复：hotfix
   - 否则：回滚到上一稳定版本

4. **事后**
   - 写 RCA 报告
   - 改进监控和测试
   - 更新应急手册

---

## ✅ 发布前最终检查

```bash
#!/bin/bash

echo "=== Pre-Release Safety Check ==="

# 1. 单元测试
echo "Running unit tests..."
go test -race ./pkg/... || exit 1
echo "✓ Unit tests passed"

# 2. 数据隔离验证
echo "Verifying data isolation..."
./test-isolation.sh || exit 1
echo "✓ Data isolation verified"

# 3. 性能基准
echo "Running performance baseline..."
./test-performance.sh || exit 1
echo "✓ Performance baseline OK"

# 4. 向后兼容性
echo "Checking backward compatibility..."
./test-backward-compat.sh || exit 1
echo "✓ Backward compatibility OK"

# 5. 监控告警
echo "Verifying monitoring..."
./test-monitoring.sh || exit 1
echo "✓ Monitoring OK"

# 6. 回滚方案测试
echo "Testing rollback procedure..."
./test-rollback.sh || exit 1
echo "✓ Rollback procedure OK"

echo ""
echo "=== ✅ All safety checks passed! Ready for release ==="
```

---

**发布负责人**：_______________
**日期**：_________________
**风险评估**：[ ] 低  [ ] 中  [ ] 高
**已获批准**：[ ]
