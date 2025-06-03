package config

import (
	"fmt"
	"os"
)

// Config holds all application settings.
type Config struct {
	DatabaseURL   string // e.g. "postgresql://user:pass@host:port/dbname"
	ServerAddress string // e.g. ":8080"
}

// Load reads environment variables and returns a Config.
func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("env var DATABASE_URL is required")
	}

	// SERVER_ADDRESS is optional—default to ":8080" if unset.
	addr := os.Getenv("SERVER_ADDRESS")
	if addr == "" {
		addr = ":5000"
	}

	return &Config{
		DatabaseURL:   dbURL,
		ServerAddress: addr,
	}, nil
}
