package main

import (
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5"
)

// The sweeps are global, so these tests must be run sequentially.

// TestDeleteExpired tests that a single expired row is deleted, and a live row is not.
func TestDeleteExpired(t *testing.T) {
	expired := createURL(t, newHandler(expiredTTLSeconds), "https://example.com/imexpired").ShortCode
	live := createURL(t, newHandler(liveTTLSeconds), "https://example.com/imlive").ShortCode

	deleteExpired(t.Context(), testQueries)

	if codeExists(t, expired) {
		t.Error("expired row survived the sweep")
	}
	if !codeExists(t, live) {
		t.Error("live row was swept")
	}
}

// TestDeleteExpiredBatches inserts one row over the batch limit, so the delete sweep has to loop.
func TestDeleteExpiredBatches(t *testing.T) {
	rows, err := testPool.Query(t.Context(),
		`INSERT INTO urls (original_url, expires_at)
		 SELECT 'https://example.com/batch/' || n, now() - INTERVAL '1 hour'
		 FROM generate_series(1, $1) AS n
		 RETURNING id`,
		cleanupBatchSize+1)
	if err != nil {
		t.Fatalf("insert expired rows: %s", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		t.Fatalf("read ids: %s", err)
	}

	deleteExpired(t.Context(), testQueries)

	if survivors := countRows(t, ids); survivors != 0 {
		t.Errorf("%d of %d rows survived a multi-batch sweep", survivors, len(ids))
	}
}

// TestLookupAfterSweep tests that a swept link is still a 404, not a 500.
func TestLookupAfterSweep(t *testing.T) {
	resp := createURL(t, newHandler(expiredTTLSeconds), "https://example.com/sweep/lookup")

	deleteExpired(t.Context(), testQueries)

	h := newHandler(liveTTLSeconds)
	checkResponse(t, serve(h, http.MethodGet, "/"+resp.ShortCode, ""), http.StatusNotFound, contentHTML, "")
	checkResponse(t, serve(h, http.MethodGet, "/stats/"+resp.ShortCode, ""), http.StatusNotFound, contentJSON, "URL not found")
}
