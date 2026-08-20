create table public.products
(
    id              bigserial
        constraint products_pk
            primary key,
    name            text                                      not null,
    manufacturer_id integer     default 0                     not null,
    item_number     varchar(50) default ''::character varying not null
);

comment on column public.products.item_number is 'артикул';
