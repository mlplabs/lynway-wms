create table if not exists public.users
(
    id     bigserial
        constraint users_pk
            primary key,
    login  varchar(100) not null
        constraint users_pk_login
            unique,
    password text         not null,
    code   varchar(10)
);
