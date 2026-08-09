-- name: AddUrl :one
INSERT INTO urls (original_url)
VALUES ($1) RETURNING *;

-- name: ProcessClick :one
UPDATE urls
SET click_count     = click_count + 1,
    last_clicked_at = now()
WHERE id = $1 RETURNING original_url;

-- name: GetStatsById :one
SELECT *
FROM urls
WHERE id = $1;