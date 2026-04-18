# 多Agent架构使用指南

## 概述

Pomclaw 现在支持多Agent架构，每个Agent有独立的：
- 配置（模型、温度、系统提示等）
- 数据存储（sessions、memories、state）
- 工作空间

## 配置文件说明

### config.json 不需要大改

原有的 `config.json` 中的 `agents.defaults` 配置**继续保留**，它作为：
1. **"default" agent** 的默认配置
2. CLI 直接使用时的配置
3. 新Agent配置的回退默认值

示例配置（保持不变）：
```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.pomclaw/workspace",
      "restrict_to_workspace": true,
      "provider": "ollama",
      "model": "qwen3.5:35b-a3b",
      "max_tokens": 8192,
      "temperature": 0.7,
      "max_tool_iterations": 20
    }
  }
}
```

## 创建新Agent

### 方法1：直接在数据库中插入（推荐测试）

```sql
-- PostgreSQL
INSERT INTO POM_AGENT_CONFIGS (
    config_id,
    user_id,
    agent_name,
    agent_id,
    model,
    provider,
    max_tokens,
    temperature,
    max_iterations,
    system_prompt,
    workspace,
    restrict_workspace,
    is_active
) VALUES (
    'cfg_test_001',
    'user_001',
    'Code Review Agent',
    'code_reviewer',              -- 重要：此ID用于路由消息
    'qwen3.5:35b-a3b',
    'ollama',
    8192,
    0.7,
    20,
    'You are an expert code reviewer. Focus on security, performance, and best practices.',
    '~/.pomclaw/workspace',
    true,
    true
);
```

详细SQL示例请查看：`docs/create-agent-example.sql`

### 方法2：通过代码创建（未来API）

将来实现API层后，可以通过HTTP接口创建。

## 使用Agent

### CLI使用（默认agent）

```bash
# CLI直接使用，自动使用 "default" agent
./pomclaw
> 你好

# 等同于使用 agent_id = "default"
```

### 通过消息总线指定Agent

在 `InboundMessage` 中指定 `agent_id`：

```go
msg := bus.InboundMessage{
    Channel:    "telegram",
    SenderID:   "user123",
    ChatID:     "chat456",
    Content:    "Review this code",
    SessionKey: "session_001",
    AgentID:    "code_reviewer",  // 指定使用哪个agent
}
msgBus.PublishInbound(msg)
```

### 不同Channel使用不同Agent

每个消息源（Telegram、Discord等）可以在发送消息时指定 `agent_id`，实现：
- Telegram机器人 → 使用 "customer_service" agent
- Discord机器人 → 使用 "code_reviewer" agent
- CLI → 使用 "default" agent

## 数据隔离

每个Agent的数据完全隔离：

| Agent ID | Sessions | Memories | State |
|----------|----------|----------|-------|
| default | ✓ | ✓ | ✓ |
| code_reviewer | ✓ | ✓ | ✓ |
| customer_service | ✓ | ✓ | ✓ |

数据通过 `agent_id` 字段隔离，互不影响。

## Agent配置字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| config_id | VARCHAR(64) | 配置记录ID（主键） |
| user_id | VARCHAR(64) | 所属用户ID |
| agent_name | VARCHAR(255) | Agent显示名称 |
| **agent_id** | VARCHAR(64) | **Agent唯一标识（消息路由使用）** |
| model | VARCHAR(255) | 使用的模型名称 |
| provider | VARCHAR(64) | LLM提供商 (openai/ollama/anthropic等) |
| max_tokens | INTEGER | 最大token数 |
| temperature | NUMERIC(3,2) | 温度参数 (0.0-1.0) |
| max_iterations | INTEGER | 最大工具调用次数 |
| system_prompt | TEXT | 自定义系统提示（可选） |
| workspace | VARCHAR(512) | 工作空间路径 |
| restrict_workspace | BOOLEAN | 是否限制文件访问在工作空间内 |
| is_active | BOOLEAN | 是否激活（软删除标记） |

## 测试步骤

### 1. 确保数据库已初始化

```bash
# 启动pomclaw会自动创建表结构
./pomclaw
```

### 2. 创建测试Agent

```sql
-- 连接到你的PostgreSQL数据库
psql -U your_user -d your_database

-- 执行创建语句（见 create-agent-example.sql）
INSERT INTO POM_AGENT_CONFIGS ...
```

### 3. 查看创建的Agent

```sql
SELECT agent_id, agent_name, model, provider, is_active
FROM POM_AGENT_CONFIGS
WHERE is_active = true;
```

### 4. 在代码中使用

修改你的Channel代码，在构造 `InboundMessage` 时指定 `AgentID`：

```go
// 示例：Telegram handler
msg := bus.InboundMessage{
    Channel:    "telegram",
    SenderID:   strconv.FormatInt(update.Message.From.ID, 10),
    ChatID:     strconv.FormatInt(update.Message.Chat.ID, 10),
    Content:    update.Message.Text,
    SessionKey: sessionKey,
    AgentID:    "code_reviewer",  // 使用特定agent
}
```

## 常见问题

**Q: CLI还能用吗？**  
A: 可以。CLI默认使用 "default" agent，配置来自 `config.json` 的 `agents.defaults`。

**Q: 如果指定的agent_id不存在会怎样？**  
A: 系统会使用默认配置（从config.json读取），agent_id仍然是你指定的值，数据会隔离存储。

**Q: 需要重启服务吗？**  
A: 不需要。Agent配置是动态加载的，创建新Agent后直接可用。

**Q: 如何删除Agent？**  
A: 软删除：`UPDATE POM_AGENT_CONFIGS SET is_active = false WHERE agent_id = 'xxx'`  
   硬删除需要同时清理关联的session/memory/state数据。

## 下一步

目前暂未实现API层，手动通过SQL创建Agent。后续计划：
- RESTful API接口（Agent CRUD）
- Web管理界面
- JWT认证
- 多租户支持
