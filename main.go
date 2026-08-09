package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"

	db "url-shortener/db/generated"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed db/migrations/*.sql
var migrations embed.FS

func main() {
	ctx := context.Background()

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	runMigrations(ctx, pool)

	srv := &server{queries: db.New(pool)}

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}

	log.Println("Server is running on port " + appPort)
	log.Fatal(http.ListenAndServe(":"+appPort, srv.routes()))
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
