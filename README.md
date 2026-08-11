# URL Shortener

A URL shortener built with a focus on performance using Go. Takes long URLs and converts them into short codes that
expire from disuse, with click tracking and analytics.

https://very-very-long-url.com/it-sure-is-very-long-and-ugly/1234567890/goober becomes http://localhost:8080/kVOkZ

## Lore

This started as a Spring Boot-based performance optimization deep-dive, with a Redis store, protocol buffers instead of
JSON, and asynchronous stat updates. That implementation and its write-up are preserved in [
`README-spring.md`](README-spring.md)

The Go rewrite keeps the ID scrambling trick, but swaps pretty much the entire stack. Redis+protobufs became
Postgres+sqlc, and Spring Boot became standard Go. The performance focus remains.

## Getting started

```bash
cp .env.example .env
# Edit .env and fill in all required values
docker compose up -d --build
```

`test.http` has example requests for every endpoint.

## API

Shorten a URL:

```bash
curl -X POST http://localhost:8080/create -H "Content-Type: application/json" -d '{"original_url": "https://example.com"}'
```

Returns:

```json
{
  "short_code": "kVOkZ",
  "short_url": "http://localhost:8080/kVOkZ",
  "expires_at": "..."
}
```

Get redirected to the original URL, updating the stats and bumping the expiration date:

```bash
curl -i http://localhost:8080/kVOkZ
# 302 Location: https://example.com
```

Check the stats, without resetting the timer:

```bash
curl http://localhost:8080/stats/kVOkZ
```

Returns:

```json
{
  "short_code": "kVOkZ",
  "original_url": "https://example.com",
  "created_at": "...",
  "clicks": 21,
  "last_clicked_at": "...",
  "expires_at": "..."
}
```

There is also a browser UI on `/` that covers all features.

## How it works

### The ID scrambling

Instead of storing randomized strings, which wastes the 5-character space, or exposing sequential database IDs (1, 2,
3...), the app scrambles the `BIGSERIAL` id with reversible modular arithmetic and base62:

```
Auto-generated ID: 1
(multiply by a large prime, mod MaxValue)
687194767
(convert to base62)
kVOkZ
```

A code can be easily translated back and from its ID without producing visibly adjacent codes.

### Expiry

Links expire from disuse rather than fixed age. Creating or clicking a link resets its `expires_at` timestamp forward to
`now() + URL_TTL`, so a link that keeps getting traffic stays alive and one that goes quiet dies.

### Why no indexes

Because of the scrambling trick, the short code <u>is</u> the ID. The only index is the primary key, on a column that
never changes, and that is what keeps clicks cheap. Postgres can only do an in-place HOT update when no indexed column
changes. `expires_at` used to be indexed to speed up expiry checks, but since every click rewrites it, every click paid.
Dropping that index took HOT updates from 7% to 99.96% and halved WAL per click from 277 bytes to 139.

## Tech stack

- Go standard library
- Postgres + pgx
- sqlc for handwritten SQL
- goose for migrations
- Docker Compose
