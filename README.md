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

**Soft expiry before deletion** - every read carries `AND expires_at > now()`, so a link appears dead on time, no matter
when the scheduled deletion sweep runs.

**One statement per click** - `ProcessClick` increments the counter, stamps `last_clicked_at`, pushes `expires_at`
forward and returns the URL all in a single query.

**No index except the primary key** - the scrambling trick makes the short code the ID, so the only index needed is the
primary key, on a column that never changes. Postgres can only do an in-place HOT update when no indexed column changes,
and every click rewrites `expires_at`, which used to be indexed to speed up expiry checks. That turned every click into
a full tuple rewrite.

Measured at 5,000 requests/s: dropping the index takes HOT updates from 0% to 98% and WAL per click from 394 B to 234 B,
at the same Postgres CPU usage. Throughput improvement is minor, at under 5%. Dropping the index is database hygiene
more than performance gains.

**In-memory cache in front of lookups** - a short code's URL is cached on creation and every database hit, with the
cache entries having the same TTL as the database rows. A cache hit skips the query entirely, short-circuiting the
redirect and recording the click asynchronously in the background. Measured:

| Requests/s | p50 uncached | p50 cached | p95 uncached | p95 cached |
|------------|--------------|------------|--------------|------------|
| 2,000      | 1.28 ms      | 0.09 ms    | 1.97 ms      | 0.17 ms    |
| 10,000     | 2.08 ms      | 0.09 ms    | 6.49 ms      | 0.40 ms    |
| 14,500     | 4.62 ms      | 0.12 ms    | 78.49 ms     | 0.81 ms    |
| 15,000     | 113.69 ms    | 0.12 ms    | 163.27 ms    | 0.85 ms    |
| 20,000     | 750.32 ms    | 0.23 ms    | 857.41 ms    | 26.56 ms   |

The uncached build degrades gradually from 2,000 to 14,500 requests/s, then falls off a cliff as it approaches the
~16,000 commits a second Postgres can sustain. The cached build does not visibly break anywhere in this table, because
the database backlog is masked by a cache that redirects from memory.

The write is still one `UPDATE` per click, just fired from a detached goroutine, so every click still costs a
transaction. Counting clicks that actually reached the table against the load window plus the time the app needed to
catch up afterward, the cached build lands 16,216/s at 20,000 RPS, 16,020/s at 28,000 and 15,015/s at 40,000 - pinned,
while the hidden backlog grows from 7 s to 22 s to 47 s. Both builds top out around 15,000-16,000 clicks/s because both
are doing the exact same write to Postgres.

**Batched click stats** - a cache hit sends a click event to a batching channel. One goroutine writes them in a single
`unnest` query, flushing on 4000 distinct IDs or one second. Both builds serve everything offered up to 30,000, so the
axis is commits - and, higher up, whether the clicks survive at all. Measured:

| Requests/s | commits/s per-click | commits/s batched | clicks dropped batched |
|------------|---------------------|-------------------|------------------------|
| 2,000      | 1,829               | 2                 | 0                      |
| 20,000     | 15,788              | 6                 | 0                      |
| 40,000     | 14,783              | 10                | 0                      |
| 50,000     | 15,207              | 12                | 0                      |
| 60,000     | 15,656              | 5                 | 1,142,049              |
| 70,000     | 16,026              | 4                 | 1,391,129              |

The per-click build flatlines from 20,000 upward, hitting the same ceiling as the builds above. It cannot commit faster,
so every request above it goes into the backlog. Batching thousands of clicks per commit solves the problem, making WAL
per click go from 198 B to 115 B, and Postgres CPU from 104% to 37%.

The batched build's backlog is also bounded, unlike the per-click build. Holding 30,000 requests/s and increasing the
load window, the per-click build needs 25s, 51s and 109s to catch up after 30s, 60s and 120s of load, while the batched
build holds at 1s. Sustained throughput increased from ~15,000/s to ~50,000/s.

That limit comes from running the stack on a two-core machine, not the app. Serving HTTP crowds the run queue, so the
batcher's single goroutine stops being scheduled often enough to drain the channel and the non-blocking send starts
discarding. The same binary on four cores handles 60,000 without losing a click.

## Tech stack

- Go standard library
- Postgres + pgx
- ristretto for the lookup cache
- sqlc for handwritten SQL
- goose for migrations
- Docker Compose
