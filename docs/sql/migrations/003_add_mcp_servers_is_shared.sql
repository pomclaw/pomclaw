-- Add is_shared column to mcp_servers table
-- 允许用户分享 MCP server 给其他用户

ALTER TABLE mcp_servers ADD COLUMN is_shared boolean NOT NULL DEFAULT false;

CREATE INDEX idx_mcp_servers_shared ON mcp_servers (is_shared) WHERE is_shared = TRUE;

COMMENT ON COLUMN mcp_servers.is_shared IS '用户分享标记：true = 其他用户可见此 MCP server';