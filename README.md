# URL Shortener

A URL shortener built with a focus on performance using Spring Boot. Takes long URLs and converts them into short codes
with
click tracking and analytics.

## Features

Submit a long URL, get back an up to 5-character long code. Click the short link, get redirected to your original URL.
Check stats to see how many times it's been clicked.

```
https://very-very-long-url.com/it-sure-is-very-long-and-ugly/1234567890/goober 
becomes 
http://localhost:8080/kVOkZ
```

## Getting started

```bash
docker-compose up --build
```

- App: `http://localhost:8080`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` (admin/admin)

## The ID scrambling

Instead of using randomized strings, which is a wasteful use of the 5-character space, or directly exposing sequential
database IDs (1, 2, 3...), the app scrambles them using reversible modulo arithmetic:

```
Auto-generated ID: 1 
(scramble with large number operations)
687194767
(convert to base62)
kVOkZ
```

This prevents people from guessing other URLs by incrementing the code. The scrambling is reversible, so `kVOkZ` always
maps back to ID 1.

## API

Shorten URL:

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"originalUrl": "https://example.com"}'
# Returns: kVOkZ
```

Use short URL:

```bash
curl http://localhost:8080/kVOkZ
# Redirects to https://example.com
```

Check stats:

```bash
curl http://localhost:8080/stats/kVOkZ
# Returns various stats
```

Auto-cleanup service removes unused URLs after 5-6 minutes since last usage (click or creation) for demonstration
purposes.

See `test.http` for example requests you can run directly.

## Tech stack

- Java 21 + Spring Boot
- Redis
- Protocol Buffers
- Prometheus + Grafana
- Docker Compose
- JUnit 5 + AssertJ + Mockito

## Technical choices and their impact on performance

- Redis - used for O(1) lookups. Comes with free TTL functionality which replaced the manual implementation.
  5-10x speedup on all endpoints compared to SQLite
- Protocol Buffers - replaced JSON with binary serialization. Lowered CPU usage by 15% and reduced
  endpoint latency by 10-25%
- Asynchronous stats updates - stats are updated asynchronously from the main redirection flow, speeding up the redirect
  endpoint by 60-70%
- And many more small optimizations, like enabling virtual threads, replacing regexes with direct comparisons, etc. -
  10-20% reduction in CPU usage and endpoint latency

![Dashboard screenshot](grafana/dashboard.webp)
Screenshot of the Grafana dashboard during load testing