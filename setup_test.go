// setup_test.go owns the Postgres container and the shared helpers.

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	db "github.com/Salad109/url-shortener/db/generated"
	"github.com/Salad109/url-shortener/transcoding"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	testBaseURL       = "http://test.local"
	liveTTLSeconds    = 3600
	shortTTLSeconds   = 60
	expiredTTLSeconds = -1
	contentHTML       = "text/html; charset=utf-8"
	contentCSS        = "text/css; charset=utf-8"
	contentJSON       = "application/json"
)

// testPool and testQueries are written once by TestMain and only read afterward.
var (
	testPool    *pgxpool.Pool
	testQueries *db.Queries
)

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

// runTests owns the container lifecycle so its defers survive os.Exit.
func runTests(m *testing.M) int {
	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
		"postgres:18",
		postgres.WithDatabase("url-shortener-db"),
		postgres.WithUsername("url-shortener-user"),
		postgres.WithPassword("url-shortener-password"),
		postgres.BasicWaitStrategies(),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return 1
	}

	dbURL, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("failed to build connection string: %s", err)
		return 1
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Printf("failed to open pool: %s", err)
		return 1
	}
	defer pool.Close()

	runMigrations(ctx, pool)
	testPool = pool
	testQueries = db.New(pool)

	return m.Run()
}

// newHandler returns a routed handler over the shared database, with its own TTL.
// Negative ttlSeconds inserts rows that are already expired.
func newHandler(ttlSeconds int32) http.Handler {
	s := &shortener{queries: testQueries, baseURL: testBaseURL, ttlSeconds: ttlSeconds}
	return s.routes()
}

// serve sends one request through the handler.
func serve(h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, target, strings.NewReader(body)))
	return rec
}

// checkResponse asserts the status and content type, plus a body substring when given one.
func checkResponse(t *testing.T, rec *httptest.ResponseRecorder, status int, contentType, contains string) {
	t.Helper()

	if rec.Code != status {
		t.Errorf("status = %d, want %d (body %s)", rec.Code, status, rec.Body)
	}
	if got := rec.Header().Get("Content-Type"); got != contentType {
		t.Errorf("content type = %q, want %q", got, contentType)
	}
	if contains != "" && !strings.Contains(rec.Body.String(), contains) {
		t.Errorf("body = %s, want it to contain %q", rec.Body, contains)
	}
}

// createURL shortens originalURL and fails the test unless it returns 201.
func createURL(t *testing.T, h http.Handler, originalURL string) CreateURLResponse {
	t.Helper()

	body, err := json.Marshal(CreateURLRequest{OriginalURL: originalURL})
	if err != nil {
		t.Fatalf("create %q: encode request: %s", originalURL, err)
	}

	rec := serve(h, http.MethodPost, "/create", string(body))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create %q: status = %d, want %d (body %s)", originalURL, rec.Code, http.StatusCreated, rec.Body)
	}

	var resp CreateURLResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("create %q: decode response: %s", originalURL, err)
	}

	return resp
}

// countRows counts how many of the ids are in the table, expired or not.
func countRows(t *testing.T, ids []int64) int {
	t.Helper()

	var count int
	if err := testPool.QueryRow(t.Context(), "SELECT count(*) FROM urls WHERE id = ANY($1)", ids).Scan(&count); err != nil {
		t.Fatalf("count rows: %s", err)
	}

	return count
}

// codeExists reports whether the short code has a row, expired or not.
func codeExists(t *testing.T, shortCode string) bool {
	t.Helper()

	id, err := transcoding.Decode(shortCode)
	if err != nil {
		t.Fatalf("decode %q: %s", shortCode, err)
	}

	return countRows(t, []int64{id}) == 1
}
