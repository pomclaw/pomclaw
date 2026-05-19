-- Pomclaw Agents Table (精简版)
-- 只保留必需字段，减少维护负担

CREATE TABLE agents
(
    -- 基础字段
    id                    VARCHAR(26) PRIMARY KEY,
    agent_key             VARCHAR(100) NOT NULL UNIQUE,
    display_name          VARCHAR(255),
    frontmatter           TEXT, -- 简短的专业领域描述
    owner_id              VARCHAR(255) NOT NULL,

    -- LLM 配置
    provider              VARCHAR(50)  NOT NULL    DEFAULT 'openrouter',
    model                 VARCHAR(200) NOT NULL,
    context_window        INTEGER      NOT NULL    DEFAULT 200000,
    max_tool_iterations   INTEGER      NOT NULL    DEFAULT 20,

    -- 工作区配置
    workspace             TEXT         NOT NULL    DEFAULT '.',
    restrict_to_workspace BOOLEAN      NOT NULL    DEFAULT TRUE,

    -- JSONB 配置字段
    tools_config          JSONB        NOT NULL    DEFAULT '{}',
    memory_config         JSONB        NOT NULL    DEFAULT '{}',
    compaction_config     JSONB        NOT NULL    DEFAULT '{}',
    other_config          JSONB        NOT NULL    DEFAULT '{}',

    -- 显示与行为字段
    emoji                 VARCHAR(10),
    agent_description     TEXT,
    thinking_level        VARCHAR(20),
    max_tokens            INTEGER                  DEFAULT 0,
    self_evolve           BOOLEAN                  DEFAULT FALSE,
    skill_evolve          BOOLEAN                  DEFAULT FALSE,

    -- 时间戳
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    deleted_at            TIMESTAMP WITH TIME ZONE
);

-- 索引
CREATE INDEX idx_agents_owner ON agents (owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_agents_agent_key_active ON agents (agent_key) WHERE deleted_at IS NULL;
CREATE INDEX idx_agents_updated ON agents (updated_at DESC);

-- 注释
COMMENT
ON TABLE agents IS 'AI 智能体配置表';
COMMENT
ON COLUMN agents.agent_key IS '智能体唯一标识符（slug）';
COMMENT
ON COLUMN agents.display_name IS '显示名称';
COMMENT
ON COLUMN agents.frontmatter IS '专业领域简短描述';
COMMENT
ON COLUMN agents.agent_description IS 'LLM 召唤提示词';
COMMENT
ON COLUMN agents.tools_config IS '工具策略配置';
COMMENT
ON COLUMN agents.memory_config IS '记忆系统配置';
COMMENT
ON COLUMN agents.compaction_config IS '上下文压缩配置';
