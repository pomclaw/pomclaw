-- 1. 先创建 pgvector 扩展
CREATE
EXTENSION IF NOT EXISTS vector;

create table daily_notes
(
    note_id    varchar(64) not null
        constraint daily_notes_pkey
            primary key,
    agent_id   varchar(64) not null,
    note_date  date        not null,
    content    text,
    embedding  vector(1536),
    created_at timestamp with time zone default CURRENT_TIMESTAMP,
    updated_at timestamp with time zone default CURRENT_TIMESTAMP
);


create index idx_daily_agent_date
    on daily_notes (agent_id, note_date);

create index idx_daily_notes_vec
    on daily_notes using ivfflat (embedding vector_cosine_ops);
