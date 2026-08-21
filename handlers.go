package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	db "github.com/Salad109/url-shortener/db/generated"
	"github.com/Salad109/url-shortener/transcoding"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

//go:embed static/index.html
var indexPage []byte

//go:embed static/404.html
var notFoundPage []byte

//go:embed static/error.html
var errorPage []byte

//go:embed static/app.css
var styleSheet []byte

// maxRequestBytes prevents accepting comically large URLs.
const maxRequestBytes = 2048

// shortener holds the dependencies shared by every handler.
type shortener struct {
	queries    *db.Queries
	cache      *ristretto.Cache[string, string]
	baseURL    string
	ttlSeconds int32
}

func (s *shortener) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /static/app.css", s.handleStyles)
	mux.HandleFunc("GET /", s.handleNotFound)
	mux.HandleFunc("GET /{shortCode}", s.handleRedirect)
	mux.HandleFunc("POST /create", s.handleCreate)
	mux.HandleFunc("GET /stats/{shortCode}", s.handleStats)
	mux.HandleFunc("GET /health", s.handleHealth)
	return mux
}

// writeJSON sends v as the response body.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("Failed to write response:", err)
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

// writeHTML sends an embedded page to a browser-facing route.
func writeHTML(w http.ResponseWriter, status int, page []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(page); err != nil {
		log.Println("Failed to write response:", err)
	}
}

func (s *shortener) handleIndex(w http.ResponseWriter, _ *http.Request) {
	writeHTML(w, http.StatusOK, indexPage)
}

func (s *shortener) handleStyles(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	if _, err := w.Write(styleSheet); err != nil {
		log.Println("Failed to write response:", err)
	}
}

func (s *shortener) handleNotFound(w http.ResponseWriter, _ *http.Request) {
	writeHTML(w, http.StatusNotFound, notFoundPage)
}

type HealthResponse struct {
	Status string `json:"status"`
}

func (s *shortener) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

func (s *shortener) handleRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shortCode := r.PathValue("shortCode")

	// Decode short code into ID
	id, err := transcoding.Decode(shortCode)
	if err != nil {
		s.handleNotFound(w, r) // 404, don't leak implementation details
		return
	}

	// Check cache first
	if originalURL, found := s.cache.Get(shortCode); found {
		// Update stats asynchronously on hit
		asyncCtx := context.WithoutCancel(ctx)
		go func() {
			if err := s.queries.UpdateStats(asyncCtx, db.UpdateStatsParams{ID: id, TTLSeconds: s.ttlSeconds}); err != nil {
				log.Println("Failed to update stats:", err)
			}
		}()
		http.Redirect(w, r, originalURL, http.StatusFound)
		return
	}

	// Record the click, extend the TTL and fetch the original URL
	originalURL, err := s.queries.ProcessClick(ctx, db.ProcessClickParams{ID: id, TTLSeconds: s.ttlSeconds})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.handleNotFound(w, r)
			return
		}
		if ctx.Err() == nil {
			log.Printf("Failed to look up %q: %s", shortCode, err)
		}
		writeHTML(w, http.StatusInternalServerError, errorPage)
		return
	}

	// Cache the result
	s.cache.SetWithTTL(shortCode, originalURL, int64(len(originalURL)), time.Duration(s.ttlSeconds)*time.Second)

	http.Redirect(w, r, originalURL, http.StatusFound)
}

type CreateURLRequest struct {
	OriginalURL string `json:"original_url"`
}

type CreateURLResponse struct {
	ShortCode string             `json:"short_code"`
	ShortURL  string             `json:"short_url"`
	ExpiresAt pgtype.Timestamptz `json:"expires_at"`
}

func (s *shortener) handleCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)

	var req CreateURLRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "Request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate URL input
	u, err := url.Parse(req.OriginalURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		writeError(w, http.StatusBadRequest, "Invalid URL")
		return
	}

	// Insert URL into the database
	row, err := s.queries.AddURL(ctx, db.AddURLParams{OriginalURL: req.OriginalURL, TTLSeconds: s.ttlSeconds})
	if err != nil {
		if ctx.Err() == nil {
			log.Printf("Failed to shorten %q: %s", req.OriginalURL, err)
		}
		writeError(w, http.StatusInternalServerError, "Failed to create short URL")
		return
	}

	// Encode ID into the short code
	shortCode, err := transcoding.Encode(row.ID)
	if err != nil {
		log.Println("Short code space exhausted:", err)
		writeError(w, http.StatusServiceUnavailable, "Failed to create short URL")
		return
	}

	// Cache the new code
	s.cache.SetWithTTL(shortCode, req.OriginalURL, int64(len(req.OriginalURL)), time.Duration(s.ttlSeconds)*time.Second)

	resp := CreateURLResponse{
		ShortCode: shortCode,
		ShortURL:  s.baseURL + "/" + shortCode,
		ExpiresAt: row.ExpiresAt,
	}

	writeJSON(w, http.StatusCreated, resp)
}

type GetStatsResponse struct {
	ShortCode     string             `json:"short_code"`
	OriginalURL   string             `json:"original_url"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	Clicks        int64              `json:"clicks"`
	LastClickedAt pgtype.Timestamptz `json:"last_clicked_at"`
	ExpiresAt     pgtype.Timestamptz `json:"expires_at"`
}

func (s *shortener) handleStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shortCode := r.PathValue("shortCode")

	id, err := transcoding.Decode(shortCode)
	if err != nil {
		writeError(w, http.StatusNotFound, "URL not found")
		return
	}

	row, err := s.queries.GetStatsByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "URL not found")
			return
		}
		if ctx.Err() == nil {
			log.Printf("Failed to retrieve stats for %q: %s", shortCode, err)
		}
		writeError(w, http.StatusInternalServerError, "Failed to retrieve stats")
		return
	}

	resp := GetStatsResponse{
		ShortCode:     shortCode,
		OriginalURL:   row.OriginalURL,
		CreatedAt:     row.CreatedAt,
		Clicks:        row.ClickCount,
		LastClickedAt: row.LastClickedAt,
		ExpiresAt:     row.ExpiresAt,
	}

	writeJSON(w, http.StatusOK, resp)
}
