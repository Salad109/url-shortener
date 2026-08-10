package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"url-shortener/transcoding"

	db "url-shortener/db/generated"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

//go:embed static/index.html
var indexPage []byte

//go:embed static/404.html
var notFoundPage []byte

//go:embed static/app.css
var styleSheet []byte

// server holds the dependencies shared by every handler.
type server struct {
	queries    *db.Queries
	ttlSeconds int32
}

func (s *server) routes() http.Handler {
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

func (s *server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(indexPage); err != nil {
		log.Println("Failed to write response:", err)
	}
}

func (s *server) handleStyles(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	if _, err := w.Write(styleSheet); err != nil {
		log.Println("Failed to write response:", err)
	}
}

func (s *server) handleNotFound(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	if _, err := w.Write(notFoundPage); err != nil {
		log.Println("Failed to write response:", err)
	}
}

func (s *server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(HealthResponse{Status: "ok"}); err != nil {
		log.Println("Failed to write response:", err)
	}
}

type HealthResponse struct {
	Status string `json:"status"`
}

func (s *server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shortCode := r.PathValue("shortCode")

	// Decode short code into ID
	id, err := transcoding.Decode(shortCode)
	if err != nil {
		s.handleNotFound(w, r) // 404, don't leak implementation details
		return
	}

	// Record the click, extend the TTL and fetch the original URL
	originalUrl, err := s.queries.ProcessClick(ctx, db.ProcessClickParams{ID: id, TtlSeconds: s.ttlSeconds})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			log.Println("Failed to look up", shortCode, err)
			http.Error(w, "Failed to resolve short URL", http.StatusInternalServerError)
			return
		}
		s.handleNotFound(w, r)
		return
	}

	log.Println("Redirecting to", originalUrl)
	http.Redirect(w, r, originalUrl, http.StatusFound)
}

func (s *server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateUrlRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate URL input
	u, err := url.Parse(req.OriginalURL)
	if err != nil || u.Scheme == "" || u.Host == "" || len(req.OriginalURL) > 2048 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Insert URL into the database
	row, err := s.queries.AddUrl(r.Context(), db.AddUrlParams{OriginalUrl: req.OriginalURL, TtlSeconds: s.ttlSeconds})
	if err != nil {
		http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
		return
	}

	// Encode ID into the short code
	resp := CreateUrlResponse{
		ShortCode: transcoding.Encode(row.ID),
		ExpiresAt: row.ExpiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Println("Failed to write response:", err)
	}
}

type CreateUrlRequest struct {
	OriginalURL string `json:"original_url"`
}

type CreateUrlResponse struct {
	ShortCode string             `json:"short_code"`
	ExpiresAt pgtype.Timestamptz `json:"expires_at"`
}

func (s *server) handleStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shortCode := r.PathValue("shortCode")

	id, err := transcoding.Decode(shortCode)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	row, err := s.queries.GetStatsById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		log.Println("Failed to retrieve stats for", shortCode, err)
		http.Error(w, "Failed to retrieve stats", http.StatusInternalServerError)
		return
	}

	resp := GetStatsResponse{
		ShortCode:     shortCode,
		OriginalURL:   row.OriginalUrl,
		CreatedAt:     row.CreatedAt,
		Clicks:        row.ClickCount,
		LastClickedAt: row.LastClickedAt,
		ExpiresAt:     row.ExpiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Println("Failed to write response:", err)
	}
}

type GetStatsResponse struct {
	ShortCode     string             `json:"short_code"`
	OriginalURL   string             `json:"original_url"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	Clicks        int64              `json:"clicks"`
	LastClickedAt pgtype.Timestamptz `json:"last_clicked_at"`
	ExpiresAt     pgtype.Timestamptz `json:"expires_at"`
}
