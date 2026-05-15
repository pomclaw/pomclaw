create table skills
(
    id          varchar(26)                            not null
        constraint skills_pkey primary key,
    user_id     varchar(26)                            not null,
    name        varchar(255)                           not null,
    slug        varchar(128)                           not null,
    description text,
    enabled     boolean                  default true  not null,
    status      varchar(32)              default 'active':: character varying not null,
    version     integer                  default 1     not null,
    created_at  timestamp with time zone default now() not null,
    updated_at  timestamp with time zone default now() not null,
    constraint skills_user_id_slug_key unique (user_id, slug)
);

create index skills_user_id_idx on skills (user_id);
create index skills_enabled_idx on skills (enabled);
create index skills_status_idx on skills (status);
