create table pom_state
(
    state_key   varchar(255) not null,
    agent_id    varchar(64)  not null,
    state_value varchar(4000),
    updated_at  timestamp with time zone default CURRENT_TIMESTAMP,
    constraint pom_state_pkey
        primary key (state_key, agent_id)
);


create index idx_pom_state_agent
    on pom_state (agent_id);

