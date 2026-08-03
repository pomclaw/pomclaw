create table providers
(
    id            serial primary key,
    user_id       uuid                                   not null,
    name          varchar(128)                           not null,
    description   varchar(255),
    provider_type varchar(64)                            not null,
    api_base      text,
    api_key       varchar(512)                           not null,
    display_name  varchar(255),
    enabled       boolean                  default false not null,
    settings      jsonb                    default '{}'::jsonb not null,
    is_shared     boolean                  default false not null,

    created_at    timestamp with time zone default now() not null,
    updated_at    timestamp with time zone default now() not null,
    constraint providers_user_id_name_key unique (user_id, name)
);

create index providers_user_id_idx on providers (user_id);
create index providers_enabled_idx on providers (enabled);
create index idx_providers_shared on providers (is_shared) where is_shared = true;
