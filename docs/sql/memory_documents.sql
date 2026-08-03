create table memory_documents
(
    id           serial primary key,
    agent_id     uuid         not null,
    path         varchar(500) not null,
    content      text                     default ''::text not null,
    hash         varchar(64)  not null,
    custom_scope varchar(255),
    created_at   timestamp with time zone default now(),
    updated_at   timestamp with time zone default now()
);

create index idx_memdoc_agent_user
    on memory_documents (agent_id);

