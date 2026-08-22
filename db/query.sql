-- name: AddURL :one
INSERT INTO urls (original_url, expires_at)
VALUES ($1, now() + (sqlc.arg(ttl_seconds)::int * INTERVAL '1 second'))
RETURNING *;

-- name: ProcessClick :one
UPDATE urls
SET click_count     = click_count + 1,
    last_clicked_at = now(),
    expires_at      = now() + (sqlc.arg(ttl_seconds)::int * INTERVAL '1 second')
WHERE id = sqlc.arg(id)
  AND expires_at > now()
RETURNING original_url;

-- name: BatchUpdateStats :exec
UPDATE urls
SET click_count     = click_count + arr.clicks,
    last_clicked_at = arr.timestamp,
    expires_at      = arr.timestamp + (sqlc.arg(ttl_seconds)::int * INTERVAL '1 second')
FROM (SELECT unnest(@ids::bigint[])             AS id,
             unnest(@clicks::bigint[])          AS clicks,
             unnest(@timestamps::timestamptz[]) AS timestamp) AS arr
WHERE urls.id = arr.id
  AND urls.expires_at > arr.timestamp;

-- name: GetStatsByID :one
SELECT *
FROM urls
WHERE id = $1
  AND expires_at > now();

-- name: DeleteExpiredURLs :execrows
DELETE
FROM urls USING (SELECT id
                 FROM urls
                 WHERE expires_at <= now()
                 LIMIT $1 FOR UPDATE SKIP LOCKED) AS expired
WHERE urls.id = expired.id;
