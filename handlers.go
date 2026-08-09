package main

import (
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

// server holds the dependencies shared by every handler.
type server struct {
	queries *db.Queries
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{shortCode}", s.handleRedirect)
	mux.HandleFunc("POST /create", s.handleCreate)
	mux.HandleFunc("GET /stats/{shortCode}", s.handleStats)
	return mux
}

func (s *server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shortCode := r.PathValue("shortCode")

	// Decode short code into ID
	id, err := transcoding.Decode(shortCode)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound) // 404, don't leak implementation details
		return
	}

	// Fetch the original URL from the database using the decoded ID
	originalUrl, err := s.queries.GetOriginalUrlById(ctx, id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			log.Println("Failed to look up", shortCode, err)
			http.Error(w, "Failed to resolve short URL", http.StatusInternalServerError)
			return
		}
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	if err := s.queries.ProcessClick(ctx, id); err != nil {
		log.Println("Failed to record click for", shortCode, err)
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
	row, err := s.queries.AddUrl(r.Context(), req.OriginalURL)
	if err != nil {
		http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
		return
	}

	// Encode ID into the short code
	resp := CreateUrlResponse{
		ShortCode: transcoding.Encode(row.ID),
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
	ShortCode string `json:"short_code"`
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
		ShortCode:   shortCode,
		OriginalURL: row.OriginalUrl,
		CreatedAt:   row.CreatedAt,
		Clicks:      row.ClickCount,
		LastClick:   row.LastClickedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Println("Failed to write response:", err)
	}
}

type GetStatsResponse struct {
	ShortCode   string             `json:"shortCode"`
	OriginalURL string             `json:"original_url"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	Clicks      int64              `json:"clicks"`
	LastClick   pgtype.Timestamptz `json:"last_click"`
}
