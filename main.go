package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"time"

	db "github.com/Salad109/url-shortener/db/generated"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed db/migrations/*.sql
var migrations embed.FS

const maxHeaderBytes = 8 << 10

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("Invalid configuration:\n", err)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.databaseUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	runMigrations(ctx, pool)

	queries := db.New(pool)

	srv := &server{queries: queries, baseUrl: cfg.baseUrl, ttlSeconds: cfg.ttlSeconds()}

	go runCleanup(ctx, queries, cfg.cleanupInterval)

	log.Println("URL TTL is", cfg.urlTtl)
	log.Println("Expired URLs are deleted every", cfg.cleanupInterval)
	log.Println("Request timeout is", cfg.requestTimeout)
	log.Println("Base URL is", cfg.baseUrl)
	log.Println("Server is running on port 8080")

	handler := http.TimeoutHandler(srv.routes(), cfg.requestTimeout, "Server is busy, try again shortly")

	httpServer := &http.Server{
		Addr:           ":8080",
		Handler:        handler,
		ReadTimeout:    cfg.requestTimeout,
		WriteTimeout:   cfg.requestTimeout + time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
	}
	log.Fatal(httpServer.ListenAndServe())
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
