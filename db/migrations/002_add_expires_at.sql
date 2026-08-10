-- +goose Up
ALTER TABLE urls
    ADD COLUMN expires_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE urls
    ALTER COLUMN expires_at DROP DEFAULT;

CREATE INDEX urls_expires_at_idx ON urls (expires_at);

-- +goose Down
ALTER TABLE urls
DROP
COLUMN expires_at;
