-- +goose Up
DROP INDEX urls_expires_at_idx;

-- +goose Down
CREATE INDEX urls_expires_at_idx ON urls (expires_at);
