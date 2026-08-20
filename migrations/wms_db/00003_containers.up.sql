create table containers
(
    id        bigserial
        constraint containers_pk
            primary key,
    name      text,
    type      integer,
    parent_id integer default 0 not null,
    root_id   integer,
    depth     integer default 0 not null
);

comment on column containers.type is 'тип контейнера, упаковки';

