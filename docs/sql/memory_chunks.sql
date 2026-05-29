create table memory_chunks
(
    id           serial primary key,
    document_id  uuid,
    agent_id     varchar(64)                        not null,
    user_id      varchar(26)                        not null,
    path         text                               not null,
    start_line   integer                  default 0 not null,
    end_line     integer                  default 0 not null,
    hash         varchar(64)                        not null,
    text         text                               not null,
    embedding    vector(1536),
    tsv          tsvector,
    custom_scope varchar(255),
    created_at   timestamp with time zone default now(),
    updated_at   timestamp with time zone default now()
);

create index idx_mem_agent_user
    on memory_chunks (agent_id, user_id);

create index idx_mem_global
    on memory_chunks (agent_id) where (user_id IS NULL);

create index idx_mem_document
    on memory_chunks (document_id);

create index idx_mem_tsv
    on memory_chunks using gin (tsv);

create index idx_mem_vec
    on memory_chunks using hnsw (embedding vector_cosine_ops);
