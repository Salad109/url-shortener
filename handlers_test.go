package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Salad109/url-shortener/transcoding"

	"github.com/dgraph-io/ristretto/v2"
)

// click requests shortCode and fails the test unless it returns 302.
func click(t *testing.T, h http.Handler, shortCode string) *httptest.ResponseRecorder {
	t.Helper()

	rec := serve(h, http.MethodGet, "/"+shortCode, "")
	if rec.Code != http.StatusFound {
		t.Fatalf("click %q: status = %d, want %d (body %s)", shortCode, rec.Code, http.StatusFound, rec.Body)
	}

	return rec
}

// fetchStats reads the stats for shortCode and fails the test unless it returns 200.
func fetchStats(t *testing.T, h http.Handler, shortCode string) GetStatsResponse {
	t.Helper()

	rec := serve(h, http.MethodGet, "/stats/"+shortCode, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("stats %q: status = %d, want %d (body %s)", shortCode, rec.Code, http.StatusOK, rec.Body)
	}

	var resp GetStatsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("stats %q: decode response: %s", shortCode, err)
	}

	return resp
}

func TestCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		originalURL string
	}{
		{"https", "https://example.com"},
		{"http", "http://example.com"},
		{"query and fragment", "https://example.com/path?query=1&idk=2#fragment"},
		{"userinfo and port", "https://user:pass@example.com:8443/idk/idk"},
	}

	h := newHandler(liveTTLSeconds)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp := createURL(t, h, tt.originalURL)

			if _, err := transcoding.Decode(resp.ShortCode); err != nil {
				t.Errorf("short code %q does not decode: %s", resp.ShortCode, err)
			}
			if want := testBaseURL + "/" + resp.ShortCode; resp.ShortURL != want {
				t.Errorf("short url = %q, want %q", resp.ShortURL, want)
			}
			if !resp.ExpiresAt.Valid {
				t.Error("expires_at is null")
			}
			if until := time.Until(resp.ExpiresAt.Time); until < 55*time.Minute || until > 65*time.Minute {
				t.Errorf("expires_at is %s away, want about an hour", until)
			}
		})
	}
}

func TestCreateRejectsBadRequests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		body    string
		status  int
		message string
	}{
		{"malformed json", `{"original_url":`, http.StatusBadRequest, "Invalid JSON"},
		{"empty body", "", http.StatusBadRequest, "Invalid JSON"},
		{"wrong field type", `{"original_url":42}`, http.StatusBadRequest, "Invalid JSON"},
		{"missing field", `{}`, http.StatusBadRequest, "Invalid URL"},
		{"empty url", `{"original_url":""}`, http.StatusBadRequest, "Invalid URL"},
		{"no scheme", `{"original_url":"example.com"}`, http.StatusBadRequest, "Invalid URL"},
		{"no host", `{"original_url":"https://"}`, http.StatusBadRequest, "Invalid URL"},
		{"relative path", `{"original_url":"/glorp/test"}`, http.StatusBadRequest, "Invalid URL"},
		{"ftp scheme", `{"original_url":"ftp://example.com"}`, http.StatusBadRequest, "Invalid URL"},
		{"javascript scheme", `{"original_url":"javascript:alert(1)"}`, http.StatusBadRequest, "Invalid URL"},
		{"mailto scheme", `{"original_url":"mailto:someone@example.com"}`, http.StatusBadRequest, "Invalid URL"},
		{"oversize body", `{"original_url":"https://example.com/` + strings.Repeat("a", maxRequestBytes) + `"}`, http.StatusRequestEntityTooLarge, "Request body too large"},
	}

	h := newHandler(liveTTLSeconds)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			checkResponse(t, serve(h, http.MethodPost, "/create", tt.body), tt.status, contentJSON, tt.message)
		})
	}
}

func TestRedirect(t *testing.T) {
	t.Parallel()

	const originalURL = "https://en.wikipedia.org/wiki/Trollface"

	h := newHandler(liveTTLSeconds)
	resp := createURL(t, h, originalURL)

	rec := click(t, h, resp.ShortCode)
	if got := rec.Header().Get("Location"); got != originalURL {
		t.Errorf("location = %q, want %q", got, originalURL)
	}
}

func TestRedirectCountsClicks(t *testing.T) {
	t.Parallel()

	h := newHandler(liveTTLSeconds)
	resp := createURL(t, h, "https://example.com/")

	before := fetchStats(t, h, resp.ShortCode)
	if before.Clicks != 0 {
		t.Errorf("clicks before any redirect = %d, want 0", before.Clicks)
	}
	if before.LastClickedAt.Valid {
		t.Error("last_clicked_at is set before any redirect")
	}

	for range 3 {
		click(t, h, resp.ShortCode)
	}

	after := fetchStats(t, h, resp.ShortCode)
	if after.Clicks != 3 {
		t.Errorf("clicks = %d, want 3", after.Clicks)
	}
	if !after.LastClickedAt.Valid {
		t.Error("last_clicked_at is still null after 3 redirects")
	}
}

func TestRedirectExtendsTTL(t *testing.T) {
	t.Parallel()

	shortLived := newHandler(shortTTLSeconds)
	longLived := newHandler(liveTTLSeconds)

	resp := createURL(t, shortLived, "https://example.com/")
	before := resp.ExpiresAt.Time

	click(t, longLived, resp.ShortCode)

	want := time.Duration(liveTTLSeconds-shortTTLSeconds) * time.Second
	after := fetchStats(t, longLived, resp.ShortCode).ExpiresAt.Time
	if extension := after.Sub(before); extension < want || extension > want+time.Minute {
		t.Errorf("expires_at moved by %s, want about %s", extension, want)
	}
}

func TestLookupNotFound(t *testing.T) {
	t.Parallel()

	routes := []struct {
		name        string
		prefix      string
		contentType string
		message     string
	}{
		{"redirect", "/", contentHTML, ""},
		{"stats", "/stats/", contentJSON, "URL not found"},
	}

	live := newHandler(liveTTLSeconds)
	expired := createURL(t, newHandler(expiredTTLSeconds), "https://example.com/")
	if !codeExists(t, expired.ShortCode) {
		t.Fatal("fixture row was deleted rather than expired, so its 404 would prove nothing")
	}

	unknown, err := transcoding.Encode(transcoding.MaxID)
	if err != nil {
		t.Fatalf("encode MaxID: %s", err)
	}

	shortCodes := []struct {
		name      string
		shortCode string
	}{
		{"unknown id", unknown},
		{"expired", expired.ShortCode},
		{"too long", "abcdef"},
		{"leading zero", "0abc"},
		{"invalid base62", "ab!c"},
	}

	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()

			for _, tt := range shortCodes {
				t.Run(tt.name, func(t *testing.T) {
					t.Parallel()

					rec := serve(live, http.MethodGet, route.prefix+tt.shortCode, "")
					checkResponse(t, rec, http.StatusNotFound, route.contentType, route.message)
				})
			}
		})
	}
}

func TestStats(t *testing.T) {
	t.Parallel()

	const originalURL = "https://example.com/stats"

	h := newHandler(liveTTLSeconds)
	resp := createURL(t, h, originalURL)
	stats := fetchStats(t, h, resp.ShortCode)

	if stats.ShortCode != resp.ShortCode {
		t.Errorf("short code = %q, want %q", stats.ShortCode, resp.ShortCode)
	}
	if stats.OriginalURL != originalURL {
		t.Errorf("original url = %q, want %q", stats.OriginalURL, originalURL)
	}
	if !stats.CreatedAt.Valid {
		t.Error("created_at is null")
	}
	if !stats.ExpiresAt.Valid {
		t.Error("expires_at is null")
	}
}

func TestStaticRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		target      string
		status      int
		contentType string
		contains    string
	}{
		{"index", "/", http.StatusOK, contentHTML, "<html"},
		{"stylesheet", "/static/app.css", http.StatusOK, contentCSS, ""},
		{"health", "/health", http.StatusOK, contentJSON, `"status":"ok"`},
		{"unknown path", "/unknown/path", http.StatusNotFound, contentHTML, ""},
		{"unknown static asset", "/static/idk.js", http.StatusNotFound, contentHTML, ""},
	}

	h := newHandler(liveTTLSeconds)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := serve(h, http.MethodGet, tt.target, "")
			checkResponse(t, rec, tt.status, tt.contentType, tt.contains)
		})
	}
}

// newCachingHandler returns a routed handler over a real cache, that cache, and the channel its clicks buffer into.
func newCachingHandler(t *testing.T, ttlSeconds int32) (http.Handler, *ristretto.Cache[string, string], chan StatUpdate) {
	t.Helper()

	cache, err := ristretto.NewCache(&ristretto.Config[string, string]{
		NumCounters: 1000,
		MaxCost:     1 << 20,
		BufferItems: 64,
		Metrics:     true,
	})
	if err != nil {
		t.Fatalf("new cache: %s", err)
	}
	t.Cleanup(cache.Close)

	updateChan := make(chan StatUpdate, updateBatchSize)
	s := &shortener{queries: testQueries, baseURL: testBaseURL, ttlSeconds: ttlSeconds, cache: cache, updateChan: updateChan}

	return s.routes(), cache, updateChan
}

// flushClicks merges everything buffered in ch and writes it, standing in for runStatUpdater.
func flushClicks(t *testing.T, ch chan StatUpdate, ttlSeconds int32) {
	t.Helper()

	buf := make(map[int64]StatUpdate)
	for len(ch) > 0 {
		update := <-ch
		buf[update.ID] = mergeUpdates(buf[update.ID], update)
	}

	flushUpdates(t.Context(), testQueries, ttlSeconds, buf)
}

func TestRedirectServesFromCache(t *testing.T) {
	t.Parallel()

	const originalURL = "https://example.com/cached"

	h, cache, updates := newCachingHandler(t, liveTTLSeconds)
	resp := createURL(t, h, originalURL)
	cache.Wait()

	rec := click(t, h, resp.ShortCode)
	if got := rec.Header().Get("Location"); got != originalURL {
		t.Errorf("location = %q, want %q", got, originalURL)
	}
	if hits := cache.Metrics.Hits(); hits != 1 {
		t.Errorf("cache hits = %d, want 1", hits)
	}

	flushClicks(t, updates, liveTTLSeconds)

	if got := fetchStats(t, h, resp.ShortCode).Clicks; got != 1 {
		t.Errorf("clicks = %d, want 1", got)
	}
}

func TestRedirectCachesDatabaseLookup(t *testing.T) {
	t.Parallel()

	const originalURL = "https://example.com/warmed"

	writer, _, _ := newCachingHandler(t, liveTTLSeconds)
	resp := createURL(t, writer, originalURL)

	// A second handler starts cold, so the first click has to reach the database.
	reader, cache, _ := newCachingHandler(t, liveTTLSeconds)
	click(t, reader, resp.ShortCode)
	if misses := cache.Metrics.Misses(); misses != 1 {
		t.Errorf("cache misses = %d, want 1", misses)
	}
	cache.Wait()

	click(t, reader, resp.ShortCode)
	if hits := cache.Metrics.Hits(); hits != 1 {
		t.Errorf("cache hits after warming = %d, want 1", hits)
	}
}
