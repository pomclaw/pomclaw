# Schema Migration - LLM Providers & Skills

## 概述

此迁移为 pomclaw 添加了 LLM 提供商管理和 Agent 技能授权功能。

## 新增表

### 1. `providers` - LLM 提供商表
存储用户配置的各种 LLM 提供商（OpenAI、Claude CLI、Ollama 等）

**主要字段:**
- `id` (PK): 提供商唯一标识
- `user_id` (FK): 所有者用户 ID
- `name`: 提供商名称（用户侧唯一）
- `provider_type`: 提供商类型（openai, claude-cli, ollama 等）
- `api_base`: API 基址（可选）
- `api_key`: API 密钥
- `display_name`: 显示名称
- `enabled`: 是否启用
- `settings`: JSONB 格式的自定义设置

**约束:**
- 主键: `id`
- 外键: `user_id` → `users(id)` ON DELETE CASCADE
- 唯一: `(user_id, name)` - 同一用户不能有同名提供商

**索引:**
- `user_id` - 快速查询用户的提供商
- `enabled` - 快速过滤启用的提供商

### 2. `skills` - 技能表
存储可用于 Agent 的各种技能

**主要字段:**
- `id` (PK): 技能唯一标识
- `user_id` (FK): 创建者用户 ID
- `name`: 技能名称
- `slug`: URL 友好的标识符
- `description`: 技能描述
- `enabled`: 是否启用
- `status`: 技能状态（active/archived）
- `version`: 版本号（用于版本管理）

**约束:**
- 主键: `id`
- 外键: `user_id` → `users(id)` ON DELETE CASCADE
- 唯一: `(user_id, slug)` - 同一用户不能有同 slug 技能

**索引:**
- `user_id` - 快速查询用户的技能
- `enabled` - 快速过滤启用的技能
- `status` - 快速查询特定状态的技能

### 3. `skill_grants` - 技能授予关系表
存储 Agent 与 Skill 的多对多关系

**主要字段:**
- `skill_id` (PK/FK): 技能 ID
- `agent_id` (PK/FK): Agent ID
- `version`: 授予的技能版本
- `created_at`: 授予时间

**约束:**
- 主键: `(skill_id, agent_id)` - 组合主键
- 外键: `skill_id` → `skills(id)` ON DELETE CASCADE
- 外键: `agent_id` → `agents(id)` ON DELETE CASCADE

**索引:**
- `agent_id` - 快速查询 Agent 的所有技能
- `created_at` - 按时间排序查询

## 执行顺序

运行迁移脚本时，**必须按以下顺序执行**（外键依赖关系）：

```bash
# 1. 假设 users 和 agents 已存在

# 2. 创建提供商表
psql -U user -d pomclaw < providers.sql

# 3. 创建技能表
psql -U user -d pomclaw < skills.sql

# 4. 创建技能授予关系表
psql -U user -d pomclaw < skill_grants.sql
```

## 外键依赖关系图

```
users
    ↓
    ├→ providers (user_id FK)
    ├→ skills (user_id FK)
    └→ agents (user_id FK)
           ↓
           ← skill_grants (agent_id FK)
                   ↑
                   ├ skills (skill_id FK)
```

## 回滚

如果需要回滚这些表：

```sql
-- 按相反顺序删除（由于外键约束）
DROP TABLE IF EXISTS skill_grants CASCADE;
DROP TABLE IF EXISTS skills CASCADE;
DROP TABLE IF EXISTS providers CASCADE;
```

## API 端点

### Providers (LLM 提供商)
```
GET    /v1/providers
POST   /v1/providers
GET    /v1/providers/:id
PUT    /v1/providers/:id
DELETE /v1/providers/:id
```

### Skills (技能管理)
```
GET    /v1/agents/:agent_id/skills      # 列出 Agent 可用技能
GET    /v1/skills
POST   /v1/skills
GET    /v1/skills/:id
POST   /v1/skills/:id/grant             # 授予技能给 Agent
DELETE /v1/skills/:id/revoke/:agent_id  # 撤销技能
```

## 注意事项

1. **API Key 安全**: API Key 在所有响应中自动掩码为 `***`
2. **用户隔离**: 所有查询都按 `user_id` 过滤，实现完全的用户隔离
3. **级联删除**: 删除用户时，相关的提供商和技能也会被删除
4. **版本管理**: 技能支持版本号，便于跟踪更新

## 相关文件

- 存储层: `internal/store/providers.go`, `internal/store/skills.go`
- 业务逻辑: `internal/logic/providerlogic.go`, `internal/logic/skilllogic.go`
- HTTP 处理: `internal/handler/*provider*.go`, `internal/handler/*skill*.go`
- 数据模型: `internal/types/provider.go`, `internal/types/skill.go`
