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

### Capacity

The 5-character Base62 space holds 916,132,830 links, which is every value the scrambler can produce.

Five characters is also the most a 64-bit multiply can handle. `Scramble` computes `id * LargePrime` before reducing it
mod `MaxValue`, and that intermediate has to fit in an `int64`, which requires `MaxValue^2 < 2^63`. Past it the multiply
overflows and `Encode` hands back an empty string.

Widening the space is a relatively straightforward change. It just needs a 128-bit intermediate and a new set of primes.

| Code length | Capacity          | Scramble             |
|-------------|-------------------|----------------------|
| 5 (current) | 916,132,830       | bare int64           |
| 6           | 56,800,235,582    | 128-bit intermediate |
| 7           | 3,521,614,606,206 | 128-bit intermediate |

916 million is plenty for now.

## Optimizations

**No index except the primary key** - the scrambling trick makes the short code the ID, so the only index needed is the
primary key, on a column that never changes. Postgres can only do an in-place HOT update when no indexed column changes,
and every click rewrites `expires_at`, which used to be indexed to speed up expiry checks. Dropping it took HOT updates
from 0% to ~99% and WAL per click from ~380 B to ~200 B. Measured:

| Redirects/s | p50 indexless | p50 indexed | p95 indexless | p95 indexed |
|-------------|---------------|-------------|---------------|-------------|
| 5,000       | 1.30 ms       | 1.31 ms     | 1.53 ms       | 1.53 ms     |
| 15,000      | 1.40 ms       | 1.43 ms     | 2.19 ms       | 3.25 ms     |
| 20,000      | 1.44 ms       | 1.52 ms     | 3.74 ms       | 5.54 ms     |
| 25,000      | 1.63 ms       | 1.78 ms     | 18.68 ms      | 36.23 ms    |
| 30,000      | 29.49 ms      | 129.48 ms   | 91.19 ms      | 152.20 ms   |

Below 15k RPS they are indistinguishable, so under light load the index is free. Where it actually costs is in the
headroom: at 30k RPS the indexless build degrades to 29 ms, while the indexed one degrades to 129 ms.

**Soft expiry before deletion** - every read carries `AND expires_at > now()`, so a link appears dead on time, no matter
when the scheduled deletion sweep runs.

**One statement per click** - `ProcessClick` increments the counter, stamps `last_clicked_at`, pushes `expires_at`
forward and returns the URL all in a single query.

## Tech stack

- Go standard library
- Postgres + pgx
- sqlc for handwritten SQL
- goose for migrations
- Docker Compose
