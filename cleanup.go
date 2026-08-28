package main

import (
	"context"
	"log"
	"time"

	db "github.com/Salad109/url-shortener/db/generated"
)

const (
	cleanupBatchSize = 1000
	cleanupTimeout   = 10 * time.Second
)

// runCleanup periodically deletes expired URLs until the process exits.
func runCleanup(ctx context.Context, queries *db.Queries, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		deleteExpired(ctx, queries)
	}
}

// deleteExpired removes expired URLs in batches until none are left.
func deleteExpired(ctx context.Context, queries *db.Queries) {
	for {
		queryCtx, cancel := context.WithTimeout(ctx, cleanupTimeout)
		deleted, err := queries.DeleteExpiredURLs(queryCtx, cleanupBatchSize)
		cancel()

		if err != nil {
			log.Println("Failed to delete expired URLs:", err)
			return
		}

		if deleted > 0 {
			log.Println("Deleted", deleted, "expired URLs")
		}

		if deleted < cleanupBatchSize {
			return
		}
	}
}
