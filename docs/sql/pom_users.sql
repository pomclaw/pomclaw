create table pom_users
(
    id         varchar(26)                            not null
        constraint pom_users_pkey
            primary key,
    username   varchar(64)                            not null
        constraint pom_users_username_key
            unique,
    email      varchar(255)                           not null
        constraint pom_users_email_key
            unique,
    password   varchar(255)                           not null,
    status     varchar(16)              default 'active':: character varying not null,
    created_at timestamp with time zone default now() not null,
    updated_at timestamp with time zone default now() not null
);


