create table providers
(
    id            varchar(26)                            not null
        constraint providers_pkey primary key,
    user_id       varchar(26)                            not null,
    name          varchar(128)                           not null,
    provider_type varchar(64)                            not null,
    api_base      text,
    api_key       varchar(512)                           not null,
    display_name  varchar(255),
    enabled       boolean                  default true  not null,
    settings      jsonb                    default '{}'::jsonb not null,
    created_at    timestamp with time zone default now() not null,
    updated_at    timestamp with time zone default now() not null,
    constraint providers_user_id_name_key unique (user_id, name)
);

create index providers_user_id_idx on providers (user_id);
create index providers_enabled_idx on providers (enabled);
