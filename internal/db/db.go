package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // Postgres driver
)

// Connect opens a database pool and verifies connectivity.
func Connect(databaseURL string) (*sql.DB, error) {
	// Open doesn’t establish any connections immediately;
	// it prepares the database handle.
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	// Set sensible defaults for the connection pool
	db.SetMaxOpenConns(20)                  // max concurrent connections
	db.SetMaxIdleConns(5)                   // idle connections to keep
	db.SetConnMaxLifetime(30 * time.Minute) // recycle connections every 30m

	// Verify that we can actually connect (and credentials are correct)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	return db, nil
}
