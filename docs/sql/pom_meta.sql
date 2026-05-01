create table pom_meta
(
    meta_key   varchar(255) not null
        constraint pom_meta_pkey
            primary key,
    meta_value varchar(4000),
    updated_at timestamp with time zone default CURRENT_TIMESTAMP
);


