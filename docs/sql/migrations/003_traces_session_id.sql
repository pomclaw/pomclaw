-- 003: traces 表 session 关联从 session_key(text) 改为 session_id(bigint)
-- 与 model tracesmodel_gen.go 的 SessionId(sql.NullInt64, db:"session_id") 对齐。
-- 代码已切到 session_id，但旧库仅有 session_key，导致 flush 时
-- ERROR: column "session_id" of relation "traces" does not exist (SQLSTATE 42703)。

-- 移除旧的字符串 session key 列与索引（已无代码引用）
drop index if exists idx_traces_session;
alter table traces drop column if exists session_key;

-- 新增 session_id（可空，语义为 sessions.id）
alter table traces add column if not exists session_id bigint;

create index if not exists idx_traces_session
    on traces (session_id asc, created_at desc) where (session_id is not null);
