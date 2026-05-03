create table pom_config
(
    config_key   varchar(255) not null,
    agent_id     varchar(64)  not null,
    config_value text,
    updated_at   timestamp with time zone default CURRENT_TIMESTAMP,
    constraint pom_config_pkey
        primary key (config_key, agent_id)
);


