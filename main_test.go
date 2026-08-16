package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// startServer runs handler on a loopback port under the real server settings.
func startServer(t *testing.T, handler http.Handler, timeout time.Duration) *httptest.Server {
	t.Helper()

	srv := httptest.NewUnstartedServer(handler)
	srv.Config = newHTTPServer(handler, timeout)
	srv.Start()
	t.Cleanup(srv.Close)

	return srv
}

// getHealth requests /health with padding bytes of header and returns the response status.
func getHealth(t *testing.T, srv *httptest.Server, padding int) int {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+"/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Padding", strings.Repeat("a", padding))

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode
}

func TestOversizeHeaders(t *testing.T) {
	t.Parallel()

	unreachable := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("handler ran, want the header limit to reject the request before dispatch")
	})
	srv := startServer(t, unreachable, 5*time.Second)

	if got := getHealth(t, srv, 2*maxHeaderBytes); got != http.StatusRequestHeaderFieldsTooLarge {
		t.Errorf("status = %d, want %d", got, http.StatusRequestHeaderFieldsTooLarge)
	}
}

func TestRequestTimeout(t *testing.T) {
	t.Parallel()

	blocked := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})
	srv := startServer(t, blocked, 100*time.Millisecond)

	if got := getHealth(t, srv, 0); got != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", got, http.StatusServiceUnavailable)
	}
}
