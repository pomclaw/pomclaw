create table agent_context_files
(
    id         serial primary key,
    user_id    varchar(26)  not null,
    agent_id   varchar(64)  not null,
    file_name  varchar(255) not null,
    content    text                     default ''::text not null,
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone default now(),

    constraint agent_context_files_agent_id_file_name_key
        unique (agent_id, file_name)
);


create index idx_agent_context_files_tenant
    on agent_context_files (agent_id);
