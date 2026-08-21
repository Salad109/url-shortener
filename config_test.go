package main

import (
	"strings"
	"testing"
	"time"
)

// setEnv sets every config variable, clearing the ones env does not name.
func setEnv(t *testing.T, env map[string]string) {
	t.Helper()

	for _, key := range []string{"DATABASE_URL", "BASE_URL", "URL_TTL", "CLEANUP_INTERVAL", "REQUEST_TIMEOUT", "CACHE_SIZE"} {
		t.Setenv(key, env[key])
	}
}

func TestLoadConfigAccepts(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want config
	}{
		{
			"defaults",
			map[string]string{"DATABASE_URL": "postgres://localhost/db"},
			config{
				databaseURL:     "postgres://localhost/db",
				baseURL:         "http://localhost:8080",
				urlTTL:          5 * time.Minute,
				cleanupInterval: time.Minute,
				requestTimeout:  5 * time.Second,
				cacheSize:       4 * 1024 * 1024,
			},
		},
		{
			"overrides",
			map[string]string{
				"DATABASE_URL":     "postgres://localhost/db",
				"BASE_URL":         "https://example.com/",
				"URL_TTL":          "48h",
				"CLEANUP_INTERVAL": "30s",
				"REQUEST_TIMEOUT":  "2s",
				"CACHE_SIZE":       "8388608",
			},
			config{
				databaseURL:     "postgres://localhost/db",
				baseURL:         "https://example.com",
				urlTTL:          48 * time.Hour,
				cleanupInterval: 30 * time.Second,
				requestTimeout:  2 * time.Second,
				cacheSize:       8 * 1024 * 1024,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, tt.env)

			cfg, err := loadConfig()
			if err != nil {
				t.Fatalf("loadConfig: %s", err)
			}
			if cfg != tt.want {
				t.Errorf("config = %+v, want %+v", cfg, tt.want)
			}
		})
	}
}

func TestLoadConfigRejects(t *testing.T) {
	valid := "postgres://localhost/db"

	tests := []struct {
		name string
		env  map[string]string
		want []string
	}{
		{"missing database url", map[string]string{}, []string{"DATABASE_URL"}},
		{"base url without scheme", map[string]string{"DATABASE_URL": valid, "BASE_URL": "example.com"}, []string{"BASE_URL"}},
		{"base url without host", map[string]string{"DATABASE_URL": valid, "BASE_URL": "https://"}, []string{"BASE_URL"}},
		{"unparseable ttl", map[string]string{"DATABASE_URL": valid, "URL_TTL": "yes"}, []string{"URL_TTL"}},
		{"negative ttl", map[string]string{"DATABASE_URL": valid, "URL_TTL": "-5m"}, []string{"URL_TTL"}},
		{"ttl under a second", map[string]string{"DATABASE_URL": valid, "URL_TTL": "500ms"}, []string{"URL_TTL"}},
		{"ttl over the int32 ceiling", map[string]string{"DATABASE_URL": valid, "URL_TTL": "1000000h"}, []string{"URL_TTL"}},
		{"zero cleanup interval", map[string]string{"DATABASE_URL": valid, "CLEANUP_INTERVAL": "0s"}, []string{"CLEANUP_INTERVAL"}},
		{"unparseable timeout", map[string]string{"DATABASE_URL": valid, "REQUEST_TIMEOUT": "maybe"}, []string{"REQUEST_TIMEOUT"}},
		{"unparseable cache size", map[string]string{"DATABASE_URL": valid, "CACHE_SIZE": "4MB"}, []string{"CACHE_SIZE"}},
		{"zero cache size", map[string]string{"DATABASE_URL": valid, "CACHE_SIZE": "0"}, []string{"CACHE_SIZE"}},
		{"negative cache size", map[string]string{"DATABASE_URL": valid, "CACHE_SIZE": "-1"}, []string{"CACHE_SIZE"}},
		{
			"every variable at once",
			map[string]string{"BASE_URL": "no", "URL_TTL": "no", "CLEANUP_INTERVAL": "no", "REQUEST_TIMEOUT": "no", "CACHE_SIZE": "no"},
			[]string{"DATABASE_URL", "BASE_URL", "URL_TTL", "CLEANUP_INTERVAL", "REQUEST_TIMEOUT", "CACHE_SIZE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, tt.env)

			_, err := loadConfig()
			if err == nil {
				t.Fatal("loadConfig succeeded, want an error")
			}
			for _, want := range tt.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error does not mention %q:\n%s", want, err)
				}
			}
		})
	}
}
