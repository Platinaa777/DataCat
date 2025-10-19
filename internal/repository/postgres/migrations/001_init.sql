CREATE SCHEMA IF NOT EXISTS datacat;
CREATE SEQUENCE IF NOT EXISTS datacat.resource_id;
CREATE TABLE IF NOT EXISTS datacat.resources (
    namespace text NOT NULL,
    kind text NOT NULL,
    id text NOT NULL,
    name text NOT NULL DEFAULT '',
    data jsonb NOT NULL,
    version bigint NOT NULL DEFAULT 1,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (namespace, kind, id)
);
CREATE UNIQUE INDEX IF NOT EXISTS resources_name ON datacat.resources(namespace, kind, name) WHERE name <> '';
CREATE INDEX IF NOT EXISTS resources_kind ON datacat.resources(kind);
CREATE TABLE IF NOT EXISTS datacat.outbox (
    id bigserial PRIMARY KEY,
    namespace text NOT NULL,
    channel_id text NOT NULL,
    message text NOT NULL,
    attempts integer NOT NULL DEFAULT 0,
    available_at timestamptz NOT NULL DEFAULT now(),
    delivered_at timestamptz,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS outbox_pending ON datacat.outbox(available_at) WHERE delivered_at IS NULL;
CREATE INDEX IF NOT EXISTS resources_alert_due ON datacat.resources((data->>'nextExecutionTime')) WHERE kind='alert';
