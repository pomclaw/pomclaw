create table pom_agents
(
    id            varchar(26)                            not null
        constraint pom_agents_pkey primary key,
    user_id       varchar(26)                            not null
        constraint pom_agents_user_id_fkey
            references pom_users
            on delete cascade,
    name          varchar(255)                           not null,
    description   text                     default ''::text not null,
    system_prompt text                     default ''::text not null,
    model         varchar(64)                            not null,
    tools         jsonb                    default '[]'::jsonb not null,
    status        varchar(16)              default 'active':: character varying not null,
    created_at    timestamp with time zone default now() not null,
    updated_at    timestamp with time zone default now() not null
);