package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all static configuration for the CLI.
type Config struct {
	StoreServerURL  string        // e.g. "http://localhost:8080"
	DefaultTTL      time.Duration // parsed from STORE_DEFAULT_TTL
	CleanUpInterval time.Duration
	APIToken        string
}

// Load reads .env (if present) and then environment variables,
// parses them, and returns a Config.
func Load() (*Config, error) {
	// Load .env into the process environment (no-op if none)
	_ = godotenv.Load()

	// Required: STORE_SERVER
	url := os.Getenv("STORE_SERVER")
	if url == "" {
		return nil, fmt.Errorf("STORE_SERVER environment variable is required")
	}

	// Optional: STORE_DEFAULT_TTL (default to "60s")
	ttlStr := os.Getenv("STORE_DEFAULT_TTL")
	if ttlStr == "" {
		ttlStr = "60s"
	}
	defaultTTL, err := time.ParseDuration(ttlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid STORE_DEFAULT_TTL %q: %w", ttlStr, err)
	}

	interval := os.Getenv("CLEANUP_INTERVAL")
	if interval == "" {
		interval = "300s"
	}

	cleanUpInterval, err := time.ParseDuration(interval)
	if err != nil {
		return nil, fmt.Errorf("invalid CLEANUP_INTERVAL %q: %w", interval, err)
	}

	token := os.Getenv("STORE_API_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("STORE_API_TOKEN is required for token auth")
	}

	return &Config{
		StoreServerURL:  url,
		DefaultTTL:      defaultTTL,
		APIToken:        token,
		CleanUpInterval: cleanUpInterval,
	}, nil
}
