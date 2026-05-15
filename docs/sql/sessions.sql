create table sessions
(
    session_key    varchar(255) not null
        constraint sessions_pkey
            primary key,
    agent_id       varchar(64)  not null,
    messages       text,
    summary        text,
    label          varchar(64),
    messages_count int          not null,

    input_tokens   bigint       not null,
    output_tokens  bigint       not null,
    created_at     timestamp with time zone default CURRENT_TIMESTAMP,
    updated_at     timestamp with time zone default CURRENT_TIMESTAMP
);


create index idx_sessions_agent
    on sessions (agent_id);
