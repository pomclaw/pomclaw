-- Rename is_builtin to is_shared on providers table
-- 统一分享字段命名：系统预置 → 用户分享

ALTER TABLE providers RENAME COLUMN is_builtin TO is_shared;

CREATE INDEX idx_providers_shared ON providers (is_shared) WHERE is_shared = TRUE;

COMMENT ON COLUMN providers.is_shared IS '用户分享标记：true = 其他用户可见此 provider';