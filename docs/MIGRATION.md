# Agents 表结构迁移指南

## 概述
从简单的 10 字段结构迁移到 GoClaw 完整的 30+ 字段结构。

## 已完成

### 1. 数据库表结构 ✅
- 创建了新的表结构文件：`docs/sql/pom_agents_v2.sql`
- 创建了数据库迁移脚本：`docs/sql/migrate_agents_to_v2.sql`
- 包含完整的 GoClaw 字段（跳过 tenant_id，单租户模式）

### 2. Store 层 ✅
- 更新了 `internal/store/agents.go`
- Agent 结构体包含所有字段（30+）
- ListAgents: 完整字段查询
- CreateAgent: 支持所有字段插入
- GetAgent: 完整字段查询（支持 ID 或 agent_key）
- UpdateAgent: 动态字段更新（白名单验证）
- DeleteAgent: 软删除实现

### 3. Types 层 ✅  
- 更新了 `internal/types/types.go`
- Agent: 完整响应类型
- CreateAgentReq: 支持所有创建字段
- UpdateAgentReq: 支持所有可更新字段（指针类型，可选更新）

### 4. Logic 层 ✅
- `internal/logic/listagentslogic.go` - 完整字段映射
- `internal/logic/createagentlogic.go` - 支持所有新字段创建
- `internal/logic/updateagentlogic.go` - 支持动态字段更新
- `internal/logic/getagentlogic.go` - 完整字段映射
- `internal/logic/deleteagentlogic.go` - 软删除（无需修改）

### 2. 数据库迁移 ⚠️
执行迁移脚本（已包含数据转换逻辑）：

```bash
# 1. 备份当前数据库（重要！）
pg_dump -h localhost -U pomclaw -d pomclaw > backup_full_$(date +%Y%m%d).sql

# 2. 执行迁移脚本（自动处理字段映射）
psql -h localhost -U pomclaw -d pomclaw < docs/sql/migrate_agents_to_v2.sql

# 3. 验证迁移结果
psql -h localhost -U pomclaw -d pomclaw -c "SELECT COUNT(*) FROM pom_agents;"
psql -h localhost -U pomclaw -d pomclaw -c "SELECT id, agent_key, display_name, owner_id FROM pom_agents LIMIT 5;"

# 4. 确认无误后删除旧表（可选）
# psql -h localhost -U pomclaw -d pomclaw -c "DROP TABLE pom_agents_old;"
```

**迁移脚本自动处理的字段映射：**
- `user_id` → `owner_id`
- `name` → `display_name` + `agent_key` (自动生成 slug)
- `description` → 废弃
- `system_prompt` → `agent_description`
- `tools` (JSON array) → `tools_config` (JSONB)

### 3. 前端适配
前端需要更新 agents API 的字段引用：
- `user_id` → `owner_id`
- `name` → `display_name`
- `description` → `agent_description`
- 添加新字段：`agent_key`, `frontmatter`, `context_window` 等

## 新字段说明

### 核心字段
- `agent_key`: 智能体唯一标识符（slug），用于 URL 和 API 路由
- `display_name`: 用户界面显示名称
- `frontmatter`: 简短的专业领域描述（一句话）
- `agent_description`: 完整的 LLM 召唤提示词

### LLM 配置
- `provider`: LLM 提供商（openai, anthropic, openrouter 等）
- `context_window`: 上下文窗口大小（默认 200000）
- `max_tool_iterations`: 最大工具调用次数（默认 20）

### JSONB 配置
- `tools_config`: 工具策略配置
- `memory_config`: 记忆系统配置
- `compaction_config`: 上下文压缩配置
- `reasoning_config`: 推理配置
- `workspace_sharing`: 工作区共享配置

### V3 新增字段
- `emoji`: 智能体图标
- `thinking_level`: 思考深度级别
- `max_tokens`: 最大生成 token 数
- `self_evolve`: 是否启用自我进化
- `skill_evolve`: 是否启用技能进化
- `skill_nudge_interval`: 技能推荐间隔

## 注意事项

1. **保持向后兼容**：旧的 API 字段暂时保留，避免破坏现有前端
2. **渐进迁移**：可以先让新旧字段共存，逐步迁移前端代码
3. **数据完整性**：迁移前务必备份数据
4. **测试充分**：迁移后测试所有 agents 相关功能
