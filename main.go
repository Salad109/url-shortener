package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"time"

	db "github.com/Salad109/url-shortener/db/generated"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed db/migrations/*.sql
var migrations embed.FS

const (
	maxHeaderBytes = 8 << 10
	listenAddr     = ":8080"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("Invalid configuration:\n", err)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	runMigrations(ctx, pool)

	queries := db.New(pool)

	cache, err := ristretto.NewCache(&ristretto.Config[string, string]{
		NumCounters: cfg.cacheSize / 100 * 10, // items 100 bytes on average, multiplied 10x
		MaxCost:     cfg.cacheSize,
		BufferItems: 64,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cache.Close()

	s := &shortener{queries: queries, baseURL: cfg.baseURL, ttlSeconds: cfg.ttlSeconds(), cache: cache}

	go runCleanup(ctx, queries, cfg.cleanupInterval)

	log.Println("URL TTL is", cfg.urlTTL)
	log.Println("Expired URLs are deleted every", cfg.cleanupInterval)
	log.Println("Request timeout is", cfg.requestTimeout)
	log.Println("Base URL is", cfg.baseURL)
	log.Println("Cache size is", cfg.cacheSize)
	log.Println("Server is running on", listenAddr)

	log.Fatal(newHTTPServer(s.routes(), cfg.requestTimeout).ListenAndServe())
}

// newHTTPServer wraps handler in the request timeout and applies the transport limits.
func newHTTPServer(handler http.Handler, timeout time.Duration) *http.Server {
	return &http.Server{
		Addr:           listenAddr,
		Handler:        http.TimeoutHandler(handler, timeout, "Server is busy, try again shortly"),
		ReadTimeout:    timeout,
		WriteTimeout:   timeout + time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
	}
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) {
	sqlDb := stdlib.OpenDBFromPool(pool)
	defer func() {
		if err := sqlDb.Close(); err != nil {
			log.Println("Failed to close migration handle:", err)
		}
	}()

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	if err := goose.UpContext(ctx, sqlDb, "db/migrations"); err != nil {
		log.Fatal("Failed to run migrations: ", err)
	}
}
