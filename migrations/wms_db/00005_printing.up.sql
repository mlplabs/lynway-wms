create table agents
(
    id    bigserial not null
        constraint agents_pk
            primary key,
    name  varchar(100),
    token text
);

comment on column agents.name is 'токен доступа';

alter table agents
    add constraint agents_pk_2
        unique (token);



create table printers
(
    id       bigserial not null
        constraint printers_pk
            primary key,
    name     text      not null,
    agent_id integer
);

alter table printers
    add constraint printers_pk_2
        unique (agent_id, name);

comment on column printers.name is 'наименование принтера';
comment on column printers.agent_id is 'ID агента';


create table print_queue
(
    id         bigserial         not null
        constraint table_name_pk
            primary key,
    print_type integer default 0 not null,
    printer_id integer default 0 not null,
    entity_type_id  integer           not null,
    entity_id    integer           not null,
    created_at timestamptz default now() not null,
    status     integer     default 0 not null
);

