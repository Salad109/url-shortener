package main

import (
	"context"
	"log"
	"net/http"
	"os"

	db "url-shortener/db/generated"

	"github.com/jackc/pgx/v5/pgxpool"
)

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

	srv := &server{queries: db.New(pool)}

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}

	log.Println("Server is running on port " + appPort)
	log.Fatal(http.ListenAndServe(":"+appPort, srv.routes()))
}
