-- Pomclaw MCP Servers Table
-- Model Context Protocol 服务器配置

create table mcp_servers
(
    id          serial primary key,
    user_id     uuid                                   not null,
    name        varchar(255)                           not null,
    description varchar(255),
    transport   varchar(50)                            not null, -- stdio, sse, streamable-http
    command     text,                                            -- stdio: command to spawn
    args        jsonb                    default '[]'::jsonb,    -- stdio: command arguments
    url         text,                                            -- sse/http: server URL
    headers     jsonb                    default '{}'::jsonb,    -- sse/http: HTTP headers
    env         jsonb                    default '{}'::jsonb,    -- stdio: environment variables
    api_key     varchar(512),
    tool_prefix varchar(50),
    timeout_sec integer                  default 60    not null,
    settings    jsonb                    default '{}'::jsonb not null,
    enabled     boolean                  default true  not null,
    is_shared   boolean                  default false not null,
    created_at  timestamp with time zone default now() not null,
    updated_at  timestamp with time zone default now() not null,
    constraint mcp_servers_name_key unique (name)
);

-- MCP Agent Grants: 授权哪些 Agent 可以使用此 MCP 服务器
create table mcp_agent_grants
(
    id               serial primary key,
    user_id          uuid                                   not null,
    server_id        integer                                not null,
    agent_id         uuid                                   not null,
    enabled          boolean                  default true  not null,
    tool_allow       jsonb, -- ["tool1", "tool2"] (null = all)
    tool_deny        jsonb, -- ["dangerous_tool"]
    config_overrides jsonb,
    created_by       uuid                                   not null,
    created_at       timestamp with time zone default now() not null,
    updated_at       timestamp with time zone default now() not null,
    constraint mcp_agent_grants_server_agent_key unique (server_id, agent_id)
);

-- 索引
create index mcp_servers_user_id_idx on mcp_servers (user_id);
create index mcp_servers_enabled_idx on mcp_servers (enabled);
create index idx_mcp_servers_shared on mcp_servers (is_shared) where is_shared = true;
create index mcp_agent_grants_server_id_idx on mcp_agent_grants (server_id);
create index mcp_agent_grants_agent_id_idx on mcp_agent_grants (agent_id);