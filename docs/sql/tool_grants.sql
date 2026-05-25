create table tool_grants
(
    id         serial primary key,
    user_id    varchar(64)                            not null,
    tool_name  varchar(100)                           not null,
    enabled    boolean,
    settings   jsonb,
    updated_at timestamp with time zone default now() not null
);

create index tool_grants_user_id_idx on tool_grants (user_id);
