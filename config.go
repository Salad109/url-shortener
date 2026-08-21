package main

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// config holds every setting the app reads from the environment variables.
type config struct {
	databaseURL     string
	baseURL         string
	urlTTL          time.Duration
	cleanupInterval time.Duration
	requestTimeout  time.Duration
	cacheSize       int64
}

// loadConfig reads and validates all environment variables.
func loadConfig() (config, error) {
	var dbErr, baseURLErr, ttlErr, cleanupErr, timeoutErr, cacheErr error

	cfg := config{databaseURL: os.Getenv("DATABASE_URL")}
	if cfg.databaseURL == "" {
		dbErr = errors.New("DATABASE_URL is not set")
	}

	cfg.baseURL, baseURLErr = envURL("BASE_URL", "http://localhost:8080")

	cfg.urlTTL, ttlErr = envDuration("URL_TTL", 5*time.Minute)
	if ttlErr == nil && (cfg.urlTTL < time.Second || cfg.urlTTL.Seconds() > math.MaxInt32) {
		ttlErr = fmt.Errorf("URL_TTL: %s must be between 1s and about 68 years", cfg.urlTTL)
	}

	cfg.cleanupInterval, cleanupErr = envDuration("CLEANUP_INTERVAL", time.Minute)

	cfg.requestTimeout, timeoutErr = envDuration("REQUEST_TIMEOUT", 5*time.Second)

	cfg.cacheSize, cacheErr = envInt64("CACHE_SIZE", 4*1024*1024)

	return cfg, errors.Join(dbErr, baseURLErr, ttlErr, cleanupErr, timeoutErr, cacheErr)
}

func (c config) ttlSeconds() int32 {
	return int32(c.urlTTL.Seconds())
}

// envURL reads an absolute http(s) url with no trailing slash.
func envURL(key, fallback string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		value = fallback
	}

	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf(`%s: %q is not an absolute http(s) URL (try "https://example.com")`, key, value)
	}

	return strings.TrimSuffix(value, "/"), nil
}

// envDuration reads a Go duration string variable.
func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf(`%s: %q is not a duration (try "5m" or "90s")`, key, value)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s: %q must be positive", key, value)
	}

	return duration, nil
}

// envInt64 reads a positive int64 variable.
func envInt64(key string, fallback int64) (int64, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %q is not an integer", key, value)
	}
	if i <= 0 {
		return 0, fmt.Errorf("%s: %q must be positive", key, value)
	}

	return i, nil
}
