package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"time"
)

// config holds every setting the app reads from the environment variables.
type config struct {
	databaseUrl     string
	urlTtl          time.Duration
	cleanupInterval time.Duration
	requestTimeout  time.Duration
}

// loadConfig reads and validates all environment variables.
func loadConfig() (config, error) {
	var dbErr, ttlErr, cleanupErr, timeoutErr error

	cfg := config{databaseUrl: os.Getenv("DATABASE_URL")}
	if cfg.databaseUrl == "" {
		dbErr = errors.New("DATABASE_URL is not set")
	}

	cfg.urlTtl, ttlErr = envDuration("URL_TTL", 5*time.Minute)
	if ttlErr == nil && (cfg.urlTtl < time.Second || cfg.urlTtl.Seconds() > math.MaxInt32) {
		ttlErr = fmt.Errorf("URL_TTL: %s must be between 1s and about 68 years", cfg.urlTtl)
	}

	cfg.cleanupInterval, cleanupErr = envDuration("CLEANUP_INTERVAL", time.Minute)

	cfg.requestTimeout, timeoutErr = envDuration("REQUEST_TIMEOUT", 5*time.Second)

	return cfg, errors.Join(dbErr, ttlErr, cleanupErr, timeoutErr)
}

func (c config) ttlSeconds() int32 {
	return int32(c.urlTtl.Seconds())
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
