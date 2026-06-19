-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pgcrypto; -- for gen_random_uuid()

CREATE TABLE customers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    email       TEXT UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE location_type AS ENUM (
    'area', 'building', 'floor', 'room', 'rack',
    'cabinet', 'box', 'compartment', 'vehicle', 'other'
);

CREATE TABLE locations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    parent_id   UUID REFERENCES locations(id) ON DELETE RESTRICT,
    type        location_type NOT NULL DEFAULT 'other',
    name        TEXT NOT NULL,
    notes       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_locations_customer ON locations(customer_id);
CREATE INDEX idx_locations_parent ON locations(parent_id);

CREATE TYPE item_condition AS ENUM ('new', 'good', 'fair', 'poor');

CREATE TABLE items (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id          UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    location_id          UUID NOT NULL REFERENCES locations(id) ON DELETE RESTRICT,
    name                 TEXT NOT NULL,
    notes                TEXT,
    quantity             INTEGER NOT NULL DEFAULT 1 CHECK (quantity >= 0),
    fill_level           INTEGER CHECK (fill_level BETWEEN 0 AND 100),
    capacity             NUMERIC,
    condition            item_condition,
    date_acquired        DATE,
    purchase_price_cents BIGINT CHECK (purchase_price_cents >= 0),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_items_customer ON items(customer_id);
CREATE INDEX idx_items_location ON items(location_id);
-- Trigram index supports ILIKE search; swap for tsvector later if needed.
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_items_name_trgm ON items USING gin (name gin_trgm_ops);

CREATE TYPE file_kind AS ENUM ('image', 'receipt', 'document');

CREATE TABLE files (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    item_id     UUID REFERENCES items(id) ON DELETE CASCADE,
    kind        file_kind NOT NULL DEFAULT 'image',
    filename    TEXT NOT NULL,
    path        TEXT NOT NULL,
    mime_type   TEXT,
    size_bytes  BIGINT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_files_item ON files(item_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE files;
DROP TYPE file_kind;
DROP TABLE items;
DROP TYPE item_condition;
DROP TABLE locations;
DROP TYPE location_type;
DROP TABLE customers;
-- +goose StatementEnd
