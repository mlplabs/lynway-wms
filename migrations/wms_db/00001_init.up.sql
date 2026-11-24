CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 1. Склады
CREATE TABLE public.warehouses (
                                   id   serial CONSTRAINT warehouses_pk PRIMARY KEY,
                                   name varchar(100) NOT NULL
);

-- 2. Зоны склада
CREATE TABLE public.zones (
                              id           serial CONSTRAINT zones_pk PRIMARY KEY,
                              name         varchar(100)          NOT NULL,
                              type         varchar(20)           NOT NULL, -- 'RECEIVING' | 'STORAGE' | 'DISPATCH'
                              is_system    boolean DEFAULT false NOT NULL,
                              warehouse_id integer               NOT NULL REFERENCES public.warehouses(id)
);

-- 3. Ячейки / Адреса хранения
CREATE TABLE public.cells (
                              id              bigserial CONSTRAINT cells_pk PRIMARY KEY,
                              name            varchar(50)   DEFAULT ''::character varying NOT NULL,
                              whs_id          integer       DEFAULT 0                     NOT NULL,
                              zone_id         integer       DEFAULT 0                     NOT NULL REFERENCES public.zones(id),
                              passage_id      integer       DEFAULT 0                     NOT NULL,
                              rack_id         integer       DEFAULT 0                     NOT NULL,
                              section_id      integer       DEFAULT 0                     NOT NULL,
                              floor           integer       DEFAULT 0                     NOT NULL,
                              number          integer       DEFAULT 0                     NOT NULL,
                              sz_length       integer       DEFAULT 0                     NOT NULL,
                              sz_width        integer       DEFAULT 0                     NOT NULL,
                              sz_height       integer       DEFAULT 0                     NOT NULL,
                              sz_volume       numeric(8, 3) DEFAULT 0                     NOT NULL,
                              sz_uf_volume    numeric(8, 3) DEFAULT 0                     NOT NULL,
                              sz_weight       numeric(8, 3) DEFAULT 0                     NOT NULL,
                              is_size_free    boolean       DEFAULT false                 NOT NULL,
                              is_weight_free  boolean       DEFAULT false                 NOT NULL,
                              not_allowed_in  boolean       DEFAULT false                 NOT NULL,
                              not_allowed_out boolean       DEFAULT false                 NOT NULL,
                              is_service      boolean       DEFAULT false                 NOT NULL
);

CREATE INDEX idx_cells_zone_service ON public.cells (zone_id, is_service);
CREATE INDEX idx_cells_hierarchy ON public.cells (whs_id, zone_id, passage_id, rack_id, section_id, floor, number);

-- 4. Контейнеры / Тара (Математика тар)
CREATE TABLE public.containers (
                                   id        bigserial CONSTRAINT containers_pk PRIMARY KEY,
                                   name      text,
                                   type      integer DEFAULT 0 NOT NULL,
                                   parent_id integer DEFAULT 0 NOT NULL,
                                   root_id   integer,
                                   depth     integer DEFAULT 0 NOT NULL
);

CREATE INDEX idx_containers_root ON public.containers (root_id);
CREATE INDEX idx_containers_parent ON public.containers (parent_id);

-- 5. Производители и Номенклатура
CREATE TABLE public.manufacturers (
                                      id   serial CONSTRAINT manufacturers_pk PRIMARY KEY,
                                      name text UNIQUE NOT NULL
);

CREATE INDEX idx_manufacturers_name_trgm ON public.manufacturers USING gin (name gin_trgm_ops);

CREATE TABLE public.items (
                              id              bigserial CONSTRAINT items_pk PRIMARY KEY,
                              name            text                            NOT NULL,
                              manufacturer_id integer                         REFERENCES public.manufacturers(id),
                              sku             text,
                              barcode         text UNIQUE,
                              tags            text[]      DEFAULT '{}'        NOT NULL,
                              attributes      jsonb       DEFAULT '{}'::jsonb NOT NULL,
                              created_at      timestamptz DEFAULT now()       NOT NULL,
                              updated_at      timestamptz DEFAULT now()       NOT NULL
);

CREATE INDEX idx_items_name_trgm ON public.items USING gin (name gin_trgm_ops);
CREATE INDEX idx_items_tags ON public.items USING gin (tags);
CREATE INDEX idx_items_attributes ON public.items USING gin (attributes);
CREATE UNIQUE INDEX idx_items_barcode_unique
    ON public.items (LOWER(TRIM(barcode)))
    WHERE barcode IS NOT NULL AND TRIM(barcode) <> '';

-- 6. Единый регистр движений
CREATE TABLE public.storage (
                                row_id       varchar(36)              DEFAULT ''::character varying NOT NULL,
                                row_time     timestamp with time zone DEFAULT now()                 NOT NULL,
                                whs_id       integer,
                                zone_id      integer,
                                cell_id      integer REFERENCES public.cells(id),
                                container_id integer,
                                item_id      integer,
                                quantity     integer,
                                CONSTRAINT chk_storage_location CHECK (whs_id IS NOT NULL
                                    AND zone_id IS NOT NULL
                                    AND (container_id IS NOT NULL OR item_id IS NOT NULL))
);

CREATE INDEX idx_storage_balance ON public.storage (zone_id, cell_id, container_id, item_id) INCLUDE (quantity);
CREATE INDEX idx_storage_item ON public.storage (item_id) INCLUDE (zone_id, cell_id, container_id, quantity);
CREATE INDEX idx_storage_container ON public.storage (container_id) INCLUDE (zone_id, cell_id, item_id, quantity);
CREATE INDEX idx_storage_row_id ON public.storage (row_id);

-- 7. Подсистема печати (CUPS-агент), Пользователи и Идемпотентность
CREATE TABLE public.agents
(
    id    bigserial PRIMARY KEY,
    name  varchar(100),
    token text
        constraint agents_pk_token
            unique
);

comment on column public.agents.name is 'токен доступа, c которым регистрируется агент';

CREATE TABLE public.printers (
                                 id           serial PRIMARY KEY,
                                 agent_id     text                  NOT NULL,
                                 cups_name    text                  NOT NULL,
                                 display_name text                  NOT NULL,
                                 is_online    boolean DEFAULT true  NOT NULL,
                                 last_seen_at timestamptz DEFAULT now() NOT NULL,
                                 UNIQUE (agent_id, cups_name)
);

CREATE TABLE public.print_jobs (
                                   id             bigserial PRIMARY KEY,
                                   printer_id     integer REFERENCES public.printers(id),
                                   print_type     integer, -- 1 = 'A4', 2 = 'TERMO'
                                   payload_type   varchar(10),
                                   payload_data   bytea,
                                   entity_type_id integer,
                                   entity_id      integer,
                                   copies         integer DEFAULT 1 NOT NULL,
                                   status         integer DEFAULT 0 NOT NULL,
                                   error_text     text,
                                   created_at     timestamptz DEFAULT now() NOT NULL
);

CREATE INDEX idx_print_jobs_status ON public.print_jobs (printer_id, status, created_at);

CREATE TABLE public.users (
                              id              serial PRIMARY KEY,
                              username        varchar(50) UNIQUE NOT NULL,
                              password_hash   text               NOT NULL,
                              role            varchar(20)        NOT NULL,
                              def_printer_id  integer REFERENCES public.printers(id),
                              def_tprinter_id integer REFERENCES public.printers(id)
);

CREATE TABLE public.operations_inbox (
                                         key           varchar(64) PRIMARY KEY,
                                         response_code integer     NOT NULL,
                                         response_body jsonb       DEFAULT '{}'::jsonb NOT NULL,
                                         created_at    timestamptz DEFAULT now() NOT NULL
);

CREATE INDEX idx_operations_inbox_created_at ON public.operations_inbox (created_at);
