create table pom_sessions
(
    session_key varchar(255) not null
        constraint pom_sessions_pkey
            primary key,
    agent_id    varchar(64)  not null,
    messages    text,
    summary     text,
    created_at  timestamp with time zone default CURRENT_TIMESTAMP,
    updated_at  timestamp with time zone default CURRENT_TIMESTAMP
);


create index idx_pom_sessions_agent
    on pom_sessions (agent_id);
