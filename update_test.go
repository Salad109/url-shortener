package main

import (
	"testing"
	"time"

	db "github.com/Salad109/url-shortener/db/generated"

	"github.com/jackc/pgx/v5/pgtype"
)

// addURL inserts a row straight through the query layer and returns its id.
func addURL(t *testing.T, ttlSeconds int32) int64 {
	t.Helper()

	row, err := testQueries.AddURL(t.Context(), db.AddURLParams{
		OriginalURL: "https://example.com/batch",
		TTLSeconds:  ttlSeconds,
	})
	if err != nil {
		t.Fatalf("add url: %s", err)
	}

	return row.ID
}

// fetchRow reads the stat columns for id, expired or not.
func fetchRow(t *testing.T, id int64) (clicks int64, lastClickedAt, expiresAt pgtype.Timestamptz) {
	t.Helper()

	const query = "SELECT click_count, last_clicked_at, expires_at FROM urls WHERE id = $1"
	if err := testPool.QueryRow(t.Context(), query, id).Scan(&clicks, &lastClickedAt, &expiresAt); err != nil {
		t.Fatalf("fetch row %d: %s", id, err)
	}

	return clicks, lastClickedAt, expiresAt
}

func TestMergeUpdates(t *testing.T) {
	t.Parallel()

	early := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	late := early.Add(time.Minute)

	tests := []struct {
		name     string
		existing StatUpdate
		update   StatUpdate
		want     StatUpdate
	}{
		{"no existing entry", StatUpdate{}, StatUpdate{ID: 1, Clicks: 1, Timestamp: late}, StatUpdate{ID: 1, Clicks: 1, Timestamp: late}},
		{"sums clicks", StatUpdate{ID: 1, Clicks: 3, Timestamp: early}, StatUpdate{ID: 1, Clicks: 2, Timestamp: late}, StatUpdate{ID: 1, Clicks: 5, Timestamp: late}},
		{"keeps the later existing timestamp", StatUpdate{ID: 1, Clicks: 1, Timestamp: late}, StatUpdate{ID: 1, Clicks: 1, Timestamp: early}, StatUpdate{ID: 1, Clicks: 2, Timestamp: late}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := mergeUpdates(tt.existing, tt.update)

			if got.ID != tt.want.ID {
				t.Errorf("id = %d, want %d", got.ID, tt.want.ID)
			}
			if got.Clicks != tt.want.Clicks {
				t.Errorf("clicks = %d, want %d", got.Clicks, tt.want.Clicks)
			}
			if !got.Timestamp.Equal(tt.want.Timestamp) {
				t.Errorf("timestamp = %s, want %s", got.Timestamp, tt.want.Timestamp)
			}
		})
	}
}

func TestRecordClickCountsDrops(t *testing.T) {
	// Capacity one makes the channel full after a single click.
	full := make(chan StatUpdate, 1)

	recordClick(full, 1)
	recordClick(full, 2)

	if queued := len(full); queued != 1 {
		t.Errorf("queued = %d, want 1", queued)
	}
	if dropped := droppedClicks.Swap(0); dropped != 1 {
		t.Errorf("dropped = %d, want 1", dropped)
	}
}

func TestFlushUpdates(t *testing.T) {
	t.Parallel()

	// Postgres stores microseconds, so a truncated timestamp survives the round trip exactly.
	now := time.Now().Truncate(time.Microsecond)

	tests := []struct {
		name   string
		clicks int64
		offset time.Duration
	}{
		{"one click", 1, -3 * time.Second},
		{"two clicks", 2, -2 * time.Second},
		{"coalesced clicks", 5, -time.Second},
	}

	// Distinct click counts and timestamps prove the three unnested arrays line up per row.
	ids := make([]int64, len(tests))
	updates := make(map[int64]StatUpdate, len(tests))
	for i, tt := range tests {
		ids[i] = addURL(t, liveTTLSeconds)
		updates[ids[i]] = StatUpdate{ID: ids[i], Clicks: tt.clicks, Timestamp: now.Add(tt.offset)}
	}

	flushUpdates(t.Context(), testQueries, liveTTLSeconds, updates)

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			clicks, lastClickedAt, expiresAt := fetchRow(t, ids[i])
			wantClickedAt := now.Add(tt.offset)

			if clicks != tt.clicks {
				t.Errorf("clicks = %d, want %d", clicks, tt.clicks)
			}
			if !lastClickedAt.Time.Equal(wantClickedAt) {
				t.Errorf("last clicked at = %s, want %s", lastClickedAt.Time, wantClickedAt)
			}
			if want := wantClickedAt.Add(liveTTLSeconds * time.Second); !expiresAt.Time.Equal(want) {
				t.Errorf("expires at = %s, want %s", expiresAt.Time, want)
			}
		})
	}
}

func TestFlushUpdatesSkips(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ttlSeconds int32
		updates    func(id int64) map[int64]StatUpdate
	}{
		{
			"expired row", expiredTTLSeconds,
			func(id int64) map[int64]StatUpdate {
				return map[int64]StatUpdate{id: {ID: id, Clicks: 1, Timestamp: time.Now()}}
			},
		},
		{
			"empty batch", liveTTLSeconds,
			func(int64) map[int64]StatUpdate { return map[int64]StatUpdate{} },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			id := addURL(t, tt.ttlSeconds)

			flushUpdates(t.Context(), testQueries, liveTTLSeconds, tt.updates(id))

			clicks, lastClickedAt, _ := fetchRow(t, id)
			if clicks != 0 {
				t.Errorf("clicks = %d, want 0", clicks)
			}
			if lastClickedAt.Valid {
				t.Errorf("last clicked at = %s, want null", lastClickedAt.Time)
			}
		})
	}
}
