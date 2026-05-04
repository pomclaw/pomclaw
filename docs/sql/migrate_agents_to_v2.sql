-- 迁移脚本：从 pom_agents v1 到 v2
-- 执行前请先备份数据库！

-- Step 1: 重命名旧表
ALTER TABLE pom_agents RENAME TO pom_agents_old;

-- Step 2: 创建新表结构
CREATE TABLE pom_agents (
    -- 基础字段
    id                    VARCHAR(26) PRIMARY KEY,
    agent_key             VARCHAR(100) NOT NULL UNIQUE,
    display_name          VARCHAR(255),
    frontmatter           TEXT,
    owner_id              VARCHAR(255) NOT NULL DEFAULT 'system',

    -- LLM 配置
    provider              VARCHAR(50) NOT NULL DEFAULT 'openrouter',
    model                 VARCHAR(200) NOT NULL,
    context_window        INTEGER NOT NULL DEFAULT 200000,
    max_tool_iterations   INTEGER NOT NULL DEFAULT 20,

    -- 工作区配置
    workspace             TEXT NOT NULL DEFAULT '.',
    restrict_to_workspace BOOLEAN NOT NULL DEFAULT TRUE,

    -- 类型与状态
    agent_type            VARCHAR(20) NOT NULL DEFAULT 'predefined',
    is_default            BOOLEAN NOT NULL DEFAULT FALSE,
    status                VARCHAR(20) DEFAULT 'active',

    -- 预算限制（可选）
    budget_monthly_cents  INTEGER,

    -- JSONB 配置字段
    tools_config          JSONB NOT NULL DEFAULT '{}',
    sandbox_config        JSONB,
    subagents_config      JSONB,
    memory_config         JSONB,
    compaction_config     JSONB,
    context_pruning       JSONB,
    other_config          JSONB NOT NULL DEFAULT '{}',

    -- V3 新增字段
    emoji                 VARCHAR(10),
    agent_description     TEXT,
    thinking_level        VARCHAR(20),
    max_tokens            INTEGER DEFAULT 0,
    self_evolve           BOOLEAN DEFAULT FALSE,
    skill_evolve          BOOLEAN DEFAULT FALSE,
    skill_nudge_interval  INTEGER DEFAULT 0,
    reasoning_config      JSONB,
    workspace_sharing     JSONB,
    chatgpt_oauth_routing JSONB,
    shell_deny_groups     JSONB,
    kg_dedup_config       JSONB,

    -- 时间戳
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    deleted_at            TIMESTAMP WITH TIME ZONE
);

-- Step 3: 迁移数据（字段映射）
INSERT INTO pom_agents (
    id, agent_key, display_name, owner_id, provider, model,
    context_window, max_tool_iterations, workspace, restrict_to_workspace,
    agent_type, status, tools_config, agent_description,
    created_at, updated_at
)
SELECT
    id,
    -- agent_key: 从 name 生成 slug
    LOWER(REGEXP_REPLACE(REGEXP_REPLACE(name, '[^a-zA-Z0-9\s-]', '', 'g'), '\s+', '-', 'g')),
    -- display_name: 使用原 name
    name,
    -- owner_id: 从 user_id 迁移
    user_id,
    'openrouter',  -- 默认 provider
    model,
    200000,  -- 默认 context_window
    20,      -- 默认 max_tool_iterations
    '.',     -- 默认 workspace
    true,    -- restrict_to_workspace
    'predefined',  -- agent_type
    status,
    -- tools: 将 JSON array 转换为 JSONB 的 tools_config
    COALESCE(tools::jsonb, '{}'::jsonb),
    -- agent_description: 从 system_prompt 迁移
    system_prompt,
    created_at,
    updated_at
FROM pom_agents_old;

-- Step 4: 创建索引
CREATE INDEX idx_pom_agents_owner ON pom_agents(owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_pom_agents_status ON pom_agents(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_pom_agents_agent_key_active ON pom_agents(agent_key) WHERE deleted_at IS NULL;
CREATE INDEX idx_pom_agents_updated ON pom_agents(updated_at DESC);

-- Step 5: 验证数据迁移
SELECT
    '旧表记录数' as type, COUNT(*) as count FROM pom_agents_old
UNION ALL
SELECT
    '新表记录数' as type, COUNT(*) as count FROM pom_agents;

-- Step 6: (可选) 删除旧表
-- 确认数据无误后执行：
-- DROP TABLE pom_agents_old;

COMMENT ON TABLE pom_agents IS 'AI 智能体配置表 (v2)';
