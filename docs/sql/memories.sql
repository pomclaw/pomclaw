create table memories
(
    memory_id    varchar(64) not null
        constraint memories_pkey
            primary key,
    agent_id     varchar(64) not null,
    content      text,
    embedding    vector(1536),
    importance   numeric(3, 2)            default 0.5,
    category     varchar(255),
    access_count integer                  default 0,
    created_at   timestamp with time zone default CURRENT_TIMESTAMP,
    accessed_at  timestamp with time zone,
    updated_at   timestamp with time zone default CURRENT_TIMESTAMP
);

create index idx_memories_agent
    on memories (agent_id);

create index idx_memories_agent_cat
    on memories (agent_id, category);

create index idx_memories_vec
    on memories using ivfflat (embedding vector_cosine_ops);
