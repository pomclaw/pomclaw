create table pom_transcripts
(
    id           serial not null
        constraint pom_transcripts_pkey
            primary key,
    session_key  varchar(255),
    agent_id     varchar(64),
    sequence_num integer,
    role         varchar(32),
    content      text,
    created_at   timestamp with time zone default CURRENT_TIMESTAMP
);

create index idx_pom_transcripts_session
    on pom_transcripts (session_key);

