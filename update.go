package main

import (
	"context"
	"log"
	"time"

	db "github.com/Salad109/url-shortener/db/generated"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	updateBatchSize = 1000
	updateInterval  = time.Second
	updateChanSize  = 100_000
)

type StatUpdate struct {
	ID        int64
	Clicks    int64
	Timestamp time.Time
}

// runStatUpdater batches and updates URL stats.
func runStatUpdater(ctx context.Context, queries *db.Queries, ttlSeconds int32, updateChan chan StatUpdate) {
	buf := make(map[int64]StatUpdate, updateBatchSize)

	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	for {
		var flush bool

		select {
		case update := <-updateChan:
			buf[update.ID] = mergeUpdates(buf[update.ID], update)
			flush = len(buf) >= updateBatchSize
		case <-ticker.C:
			flush = len(buf) > 0
		}

		if flush {
			flushUpdates(ctx, queries, ttlSeconds, buf)
			clear(buf)
		}
	}
}

// recordClick queues one click for id, dropping it when the batcher has fallen behind.
func recordClick(ch chan<- StatUpdate, id int64) {
	select {
	case ch <- StatUpdate{ID: id, Clicks: 1, Timestamp: time.Now()}:
	default:
	}
}

// mergeUpdates merges a StatUpdate into the existing one for the same ID, summing clicks and keeping the later timestamp.
func mergeUpdates(existing, update StatUpdate) StatUpdate {
	update.Clicks += existing.Clicks
	if existing.Timestamp.After(update.Timestamp) {
		update.Timestamp = existing.Timestamp
	}
	return update
}

// flushUpdates updates the database with the given batch of stat updates.
func flushUpdates(ctx context.Context, queries *db.Queries, ttlSeconds int32, updates map[int64]StatUpdate) {
	if len(updates) == 0 {
		return
	}

	ids := make([]int64, 0, len(updates))
	clicks := make([]int64, 0, len(updates))
	timestamps := make([]pgtype.Timestamptz, 0, len(updates))
	for id, update := range updates {
		ids = append(ids, id)
		clicks = append(clicks, update.Clicks)
		timestamps = append(timestamps, pgtype.Timestamptz{Time: update.Timestamp, Valid: true})
	}

	if err := queries.BatchUpdateStats(ctx, db.BatchUpdateStatsParams{
		TTLSeconds: ttlSeconds,
		Ids:        ids,
		Clicks:     clicks,
		Timestamps: timestamps,
	}); err != nil {
		log.Println("Failed to flush updates:", err)
	}
}
