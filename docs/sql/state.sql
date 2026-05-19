create table state
(
    id          serial primary key,
    state_key   varchar(255) not null,
    agent_id    varchar(64)  not null,
    state_value varchar(4000),
    updated_at  timestamp with time zone default CURRENT_TIMESTAMP,
    unique (state_key, agent_id)
);

create index idx_state_agent on state (agent_id);

