-- Add is_shared column to skills table
-- 允许用户分享 skill 给其他用户

ALTER TABLE skills ADD COLUMN is_shared boolean NOT NULL DEFAULT false;

CREATE INDEX idx_skills_shared ON skills (is_shared) WHERE is_shared = TRUE;

COMMENT ON COLUMN skills.is_shared IS '用户分享标记：true = 其他用户可见此 skill';