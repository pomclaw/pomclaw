ALTER TABLE skills
    ADD COLUMN IF NOT EXISTS content_url TEXT,
    ADD COLUMN IF NOT EXISTS file_hash   VARCHAR(64),
    ADD COLUMN IF NOT EXISTS metadata    JSONB DEFAULT '{}';

COMMENT ON COLUMN skills.content_url IS 'OSS CDN URL for skill zip package';
COMMENT ON COLUMN skills.file_hash IS 'SHA256 hash of uploaded zip for dedup';
COMMENT ON COLUMN skills.metadata IS 'JSON metadata extracted from SKILL.md frontmatter';
