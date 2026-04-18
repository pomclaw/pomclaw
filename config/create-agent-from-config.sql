-- 基于你的本地配置创建Agent
-- 配置来自: C:\Users\Administrator\.pomclaw\config.json

-- 1. 创建默认Agent配置（基于config.json的agents.defaults）
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
    'cfg_default_001',
    'admin',
    'Default Agent',
    'default',                          -- CLI使用的默认agent
    'glm-5',                           -- 你配置的模型
    'openai',                          -- 你配置的provider
    8192,                              -- 你配置的max_tokens
    0.7,                               -- 你配置的temperature
    20,                                -- 你配置的max_iterations
    NULL,                              -- 使用默认系统提示
    '~/.pomclaw/workspace',            -- 你配置的workspace
    true,                              -- 你配置的restrict_to_workspace
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- 2. 创建一个代码审查Agent（示例）
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
    'cfg_code_reviewer_001',
    'admin',
    'Code Review Agent',
    'code_reviewer',
    'glm-5',                           -- 使用相同的模型
    'openai',                          -- 使用相同的provider
    8192,
    0.5,                               -- 更低的temperature，更确定性的输出
    20,
    'You are an expert code reviewer. Focus on:
1. Security vulnerabilities
2. Performance issues
3. Code quality and best practices
4. Potential bugs
Always provide specific, actionable feedback.',
    '~/.pomclaw/workspace',
    true,
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- 3. 创建一个客服Agent（示例）
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
    'cfg_customer_service_001',
    'admin',
    'Customer Service Agent',
    'customer_service',
    'glm-5',
    'openai',
    8192,
    0.8,                               -- 更高的temperature，更有创造性
    20,
    'You are a helpful and friendly customer service representative.
Be polite, patient, and professional.
Always try to solve customer problems efficiently.',
    '~/.pomclaw/workspace',
    true,
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- 查看创建的Agents
SELECT
    agent_id,
    agent_name,
    model,
    provider,
    temperature,
    max_tokens,
    is_active,
    created_at
FROM POM_AGENT_CONFIGS
WHERE is_active = true
ORDER BY created_at DESC;

-- 验证default agent
SELECT * FROM POM_AGENT_CONFIGS WHERE agent_id = 'default';
