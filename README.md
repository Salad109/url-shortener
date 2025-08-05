# URL Shortener

A simple, lightweight URL shortener built with Spring Boot. Takes long URLs and converts them into short codes with
click tracking and analytics.

## Features

Submit a long URL, get back an up to 5-character long code. Click the short link, get redirected to your original URL.
Check stats to see how many times it's been clicked.

```
https://very-very-long-url.com/it-sure-is-very-long-and-ugly/1234567890/goober 
becomes 
http://localhost:8080/kVOkZ
```

## Getting Started

**Just the URL shortener:**

```bash
./mvnw spring-boot:run
```

App available at `http://localhost:8080`

**With monitoring dashboard:**

```bash
docker-compose up --build
```

- App: `http://localhost:8080`
- Grafana: `http://localhost:3000` (admin/admin)

## The ID scrambling

Instead of using randomized strings, which is inefficient, or exposing sequential database IDs (1, 2, 3...), the app
scrambles them using reversible modulo arithmetic:

```
Auto-generated database ID: 1 
(scramble with large number operations)
687194767
(convert to base62)
kVOkZ
```

This prevents people from guessing other URLs by incrementing the code. The scrambling is reversible, so `kVOkZ` always
maps back to database ID 1.

## API

**Create short URL:**

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"originalUrl": "https://example.com"}'
# Returns: kVOkZ
```

**Use short URL:**

```bash
curl http://localhost:8080/kVOkZ
# Redirects to https://example.com
```

**Check stats:**

```bash
curl http://localhost:8080/stats/kVOkZ
# Returns: click count, creation time, last click time
```

Auto-cleanup service removes unused URLs after 5 minutes since last click or creation for demonstration purposes.

See `test.http` for example requests you can run directly.

## Tech Stack

- Java 21 + Spring Boot
- SQLite
- Prometheus + Grafana
- Docker Compose
- JUnit 5 + AssertJ + Mockito