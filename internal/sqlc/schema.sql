CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS "transactions"
(
    "id"          UUID PRIMARY KEY         DEFAULT uuid_generate_v4(),
    "event_id"    UUID  NOT NULL,
    "event_name"  TEXT  NOT NULL,
    "opts"        JSONB NOT NULL,
    "payload"     JSONB NOT NULL,
    "status"      TEXT  NOT NULL,
    "total_retry" INTEGER                  DEFAULT 0,
    "started_at"  TIMESTAMP WITH TIME ZONE DEFAULT now(),
    "ended_at"    TIMESTAMP WITH TIME ZONE,
    "info"        JSONB,
    "updated_at"  TIMESTAMP WITH TIME ZONE DEFAULT now(),
    "created_at"  TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS "tx_sagas"
(
    "id"             UUID PRIMARY KEY         DEFAULT uuid_generate_v4(),
    "event_id"       UUID  NOT NULL UNIQUE,
    "transaction_id" UUID  NOT NULL,
    "event_name"     TEXT  NOT NULL,
    "opts"           JSONB NOT NULL,
    "payload"        JSONB NOT NULL,
    "status"         TEXT  NOT NULL,
    "total_retry"    INTEGER                  DEFAULT 0,
    "retries_errors" JSONB,
    "started_at"     TIMESTAMP WITH TIME ZONE DEFAULT now(),
    "ended_at"       TIMESTAMP WITH TIME ZONE,
    "info"           JSONB,
    "updated_at"     TIMESTAMP WITH TIME ZONE DEFAULT now(),
    "created_at"     TIMESTAMP WITH TIME ZONE DEFAULT now(),
    FOREIGN KEY (transaction_id) REFERENCES "transactions" (id)
);

CREATE INDEX IF NOT EXISTS "idx_transactions_event_id" ON "transactions" ("event_id");
CREATE INDEX IF NOT EXISTS "idx_transactions_status" ON "transactions" ("status");