create table sessions
(
    id             serial primary key,
    user_id        uuid   not null,
    agent_id       uuid   not null,
    messages       text,
    summary        text,
    label          varchar(64),
    messages_count int    not null,

    input_tokens   bigint not null,
    output_tokens  bigint not null,
    created_at     timestamp with time zone default CURRENT_TIMESTAMP,
    updated_at     timestamp with time zone default CURRENT_TIMESTAMP
);


create index idx_sessions_agent
    on sessions (agent_id);
