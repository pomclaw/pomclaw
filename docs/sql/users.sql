create table users
(
    id         serial primary key,
    user_id    uuid                                   not null,
    username   varchar(64)                            not null
        constraint users_username_key
            unique,
    email      varchar(255)                           not null
        constraint users_email_key
            unique,
    password   varchar(255)                           not null,
    status     varchar(16)              default 'active':: character varying not null,
    created_at timestamp with time zone default now() not null,
    updated_at timestamp with time zone default now() not null
);


