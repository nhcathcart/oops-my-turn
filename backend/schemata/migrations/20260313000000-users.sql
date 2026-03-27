-- +migrate Up
CREATE TABLE users (
    id          TEXT        PRIMARY KEY,
    google_id   TEXT        UNIQUE NOT NULL,
    email       TEXT        NOT NULL,
    first_name  TEXT        NOT NULL,
    last_name   TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +migrate Down
DROP TABLE IF EXISTS users;
