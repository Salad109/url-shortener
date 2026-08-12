package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"

	db "github.com/Salad109/url-shortener/db/generated"
	"github.com/Salad109/url-shortener/transcoding"

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

// server holds the dependencies shared by every handler.
type server struct {
	queries    *db.Queries
	baseUrl    string
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

// writeJSON sends v as the response body.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("Failed to write response:", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// writeHTML sends an embedded page to a browser-facing route.
func writeHTML(w http.ResponseWriter, status int, page []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(page); err != nil {
		log.Println("Failed to write response:", err)
	}
}

func (s *server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	writeHTML(w, http.StatusOK, indexPage)
}

func (s *server) handleStyles(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	if _, err := w.Write(styleSheet); err != nil {
		log.Println("Failed to write response:", err)
	}
}

func (s *server) handleNotFound(w http.ResponseWriter, _ *http.Request) {
	writeHTML(w, http.StatusNotFound, notFoundPage)
}

func (s *server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
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
		if errors.Is(err, pgx.ErrNoRows) {
			s.handleNotFound(w, r)
			return
		}
		if ctx.Err() == nil {
			log.Println("Failed to look up", shortCode, err)
		}
		writeHTML(w, http.StatusInternalServerError, errorPage)
		return
	}

	http.Redirect(w, r, originalUrl, http.StatusFound)
}

func (s *server) handleCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateUrlRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate URL input
	u, err := url.Parse(req.OriginalURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || len(req.OriginalURL) > 2048 {
		writeError(w, http.StatusBadRequest, "Invalid URL")
		return
	}

	// Insert URL into the database
	row, err := s.queries.AddUrl(ctx, db.AddUrlParams{OriginalUrl: req.OriginalURL, TtlSeconds: s.ttlSeconds})
	if err != nil {
		if ctx.Err() == nil {
			log.Println("Failed to shorten", req.OriginalURL, err)
		}
		writeError(w, http.StatusInternalServerError, "Failed to create short URL")
		return
	}

	// Encode ID into the short code
	shortCode := transcoding.Encode(row.ID)
	resp := CreateUrlResponse{
		ShortCode: shortCode,
		ShortURL:  s.baseUrl + "/" + shortCode,
		ExpiresAt: row.ExpiresAt,
	}

	writeJSON(w, http.StatusCreated, resp)
}

type CreateUrlRequest struct {
	OriginalURL string `json:"original_url"`
}

type CreateUrlResponse struct {
	ShortCode string             `json:"short_code"`
	ShortURL  string             `json:"short_url"`
	ExpiresAt pgtype.Timestamptz `json:"expires_at"`
}

func (s *server) handleStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shortCode := r.PathValue("shortCode")

	id, err := transcoding.Decode(shortCode)
	if err != nil {
		writeError(w, http.StatusNotFound, "URL not found")
		return
	}

	row, err := s.queries.GetStatsById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "URL not found")
			return
		}
		if ctx.Err() == nil {
			log.Println("Failed to retrieve stats for", shortCode, err)
		}
		writeError(w, http.StatusInternalServerError, "Failed to retrieve stats")
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

	writeJSON(w, http.StatusOK, resp)
}

type GetStatsResponse struct {
	ShortCode     string             `json:"short_code"`
	OriginalURL   string             `json:"original_url"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	Clicks        int64              `json:"clicks"`
	LastClickedAt pgtype.Timestamptz `json:"last_clicked_at"`
	ExpiresAt     pgtype.Timestamptz `json:"expires_at"`
}
