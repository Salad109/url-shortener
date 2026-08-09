-- +goose Up
CREATE TABLE urls
(
    id              BIGSERIAL PRIMARY KEY,
    original_url    TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    click_count     BIGINT      NOT NULL DEFAULT 0,
    last_clicked_at TIMESTAMPTZ
);

-- +goose Down
DROP TABLE urls;