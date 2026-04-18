-- 示例：在数据库中创建新的Agent配置

-- 1. 首先确保有用户记录（可选，如果还没实现用户管理，user_id 可以随意填）
-- INSERT INTO POM_USERS (user_id, username, password_hash, email, role)
-- VALUES ('user_001', 'admin', 'hash_placeholder', 'admin@example.com', 'admin');

-- 2. 创建Agent配置
-- PostgreSQL 示例
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
    is_active,
    created_at,
    updated_at
) VALUES (
    'cfg_' || substr(md5(random()::text), 1, 20),  -- 随机生成 config_id
    'user_001',                                     -- 用户ID
    'My Custom Agent',                              -- Agent 名称
    'my_agent_001',                                 -- Agent ID（重要！消息中使用此ID路由）
    'qwen3.5:35b-a3b',                             -- 模型名称
    'ollama',                                       -- Provider (openai/ollama/anthropic等)
    8192,                                          -- 最大tokens
    0.7,                                           -- Temperature
    20,                                            -- 最大工具调用次数
    'You are a helpful assistant specialized in code review.',  -- 自定义系统提示
    '~/.pomclaw/workspace',                        -- 工作空间路径
    true,                                          -- 是否限制在工作空间内
    true,                                          -- 是否激活
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Oracle 示例
-- INSERT INTO POM_AGENT_CONFIGS (
--     config_id, user_id, agent_name, agent_id,
--     model, provider, max_tokens, temperature, max_iterations,
--     system_prompt, workspace, restrict_workspace, is_active,
--     created_at, updated_at
-- ) VALUES (
--     'cfg_' || SUBSTR(DBMS_RANDOM.STRING('x', 20), 1, 20),
--     'user_001',
--     'My Custom Agent',
--     'my_agent_001',
--     'qwen3.5:35b-a3b',
--     'ollama',
--     8192,
--     0.7,
--     20,
--     'You are a helpful assistant specialized in code review.',
--     '~/.pomclaw/workspace',
--     1,  -- Oracle: 1 = true, 0 = false
--     1,
--     CURRENT_TIMESTAMP,
--     CURRENT_TIMESTAMP
-- );

-- 3. 查看已创建的Agents
SELECT agent_id, agent_name, model, provider, is_active, created_at
FROM POM_AGENT_CONFIGS
WHERE is_active = true
ORDER BY created_at DESC;

-- 4. 更新Agent配置
-- UPDATE POM_AGENT_CONFIGS
-- SET model = 'llama3:8b',
--     temperature = 0.5,
--     updated_at = CURRENT_TIMESTAMP
-- WHERE agent_id = 'my_agent_001';

-- 5. 停用Agent（软删除）
-- UPDATE POM_AGENT_CONFIGS
-- SET is_active = false,
--     updated_at = CURRENT_TIMESTAMP
-- WHERE agent_id = 'my_agent_001';
