create table skill_grants
(
    id         serial primary key,
    user_id    uuid                                   not null,
    agent_id   uuid                                   not null,
    skill_id   integer                                not null,
    version    integer                  default 1     not null,
    enabled    boolean                  default true  not null,
    created_at timestamp with time zone default now() not null,
    unique (skill_id, agent_id)
);

create index skill_grants_agent_id_idx on skill_grants (agent_id);
create index skill_grants_created_at_idx on skill_grants (created_at);

-- Migration: ALTER TABLE skill_grants ADD COLUMN enabled boolean NOT NULL DEFAULT true;
