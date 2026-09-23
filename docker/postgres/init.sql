CREATE TABLE IF NOT EXISTS products (
    id                UUID PRIMARY KEY,
    name              TEXT NOT NULL,
    price             DOUBLE PRECISION NOT NULL,
    stock             INTEGER NOT NULL,
    description       TEXT NOT NULL,
    short_description TEXT NOT NULL,
    hidden            BOOLEAN NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL,
    updated_at        TIMESTAMPTZ NOT NULL,
    images_urls       TEXT[] NOT NULL DEFAULT '{}',
    category_id       UUID NOT NULL
);

CREATE TABLE IF NOT EXISTS outbox (
    id            UUID PRIMARY KEY,
    aggregate_id  UUID NOT NULL,
    event_type    TEXT NOT NULL,
    payload       JSONB NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL,
    published     BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS outbox_unpublished_idx ON outbox (created_at)
    WHERE published = FALSE;
