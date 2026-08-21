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
from 0% to ~99% and WAL per click from ~345 B to ~185 B. Measured:

| Requests/s | p50 indexless | p50 indexed | p95 indexless | p95 indexed | dropped indexless | dropped indexed |
|------------|---------------|-------------|---------------|-------------|-------------------|-----------------|
| 10,000     | 1.37 ms       | 1.41 ms     | 2.15 ms       | 2.39 ms     | 0%                | 0%              |
| 15,000     | 1.49 ms       | 1.54 ms     | 3.72 ms       | 5.40 ms     | 0%                | 0%              |
| 20,000     | 1.65 ms       | 1.77 ms     | 32.91 ms      | 99.84 ms    | 0%                | 0%              |
| 25,000     | 10.53 ms      | 487.95 ms   | 351.04 ms     | 553.14 ms   | 0%                | 1.15%           |
| 30,000     | 729.88 ms     | 764.73 ms   | 829.81 ms     | 923.86 ms   | 11.02%            | 15.93%          |

Below 20k they are indistinguishable, so under light load the index is free. It costs headroom instead: the highest rate
served without dropping is 24k indexless against 22k indexed, and at 25k the indexless build still answers in 10 ms
while the indexed one is already at 488 ms. The extra index maintenance roughly doubles Postgres CPU usage. By 30k both
are saturated.

**In-memory cache in front of lookups** - a short code's URL is cached on creation and every database hit, with the
cache entries having the same TTL as the database rows. A cache hit skips the query entirely, short-circuiting the
redirect and recording the click asynchronously in the background. Measured:

| Requests/s | p50 cacheless | p50 cached | p95 cacheless | p95 cached | dropped cacheless | dropped cached |
|------------|---------------|------------|---------------|------------|-------------------|----------------|
| 20,000     | 1.65 ms       | 0.07 ms    | 32.91 ms      | 0.32 ms    | 0%                | 0%             |
| 25,000     | 10.53 ms      | 0.08 ms    | 351.04 ms     | 0.63 ms    | 0%                | 0%             |
| 30,000     | 729.88 ms     | 0.12 ms    | 829.81 ms     | 28.63 ms   | 11.02%            | 4.97%          |
| 35,000     | 739.43 ms     | 0.13 ms    | 836.99 ms     | 67.76 ms   | 23.92%            | 12.12%         |
| 40,000     | 729.03 ms     | 0.15 ms    | 833.24 ms     | 207.16 ms  | 32.82%            | 27.21%         |

A hit returns the URL from memory and makes the click a background write, so the redirect is decoupled from the
bookkeeping. That explains why latency falls 20x to 130x while throughput barely moves: essentially the same query still
runs, it's just the redirect stops waiting for the database return. The drop-free ceiling goes from 24k to 26k
requests/s.

**Soft expiry before deletion** - every read carries `AND expires_at > now()`, so a link appears dead on time, no matter
when the scheduled deletion sweep runs.

**One statement per click** - `ProcessClick` increments the counter, stamps `last_clicked_at`, pushes `expires_at`
forward and returns the URL all in a single query.

## Tech stack

- Go standard library
- Postgres + pgx
- ristretto for the lookup cache
- sqlc for handwritten SQL
- goose for migrations
- Docker Compose
