create table skill_grants
(
    id         serial primary key,
    skill_id   varchar(26)                            not null,
    agent_id   varchar(26)                            not null,
    version    integer                  default 1     not null,
    created_at timestamp with time zone default now() not null,
    unique (skill_id, agent_id)
);

create index skill_grants_agent_id_idx on skill_grants (agent_id);
create index skill_grants_created_at_idx on skill_grants (created_at);
