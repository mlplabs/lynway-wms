create table manufacturers
(
    id   bigserial
        constraint manufacturers_pk
            primary key,
    name varchar(250) not null
);


create table barcodes
(
    id           bigserial
        constraint barcodes_pk
            primary key,
    name         text        default ''::text              not null,
    barcode_type integer     default 0                     not null,
    owner_id     integer     default 0                     not null,
    owner_type   integer     default 0                     not null
);

comment on column barcodes.owner_id is 'id подчиненного элемента';

comment on column barcodes.owner_type is 'таблица владельца шк';







