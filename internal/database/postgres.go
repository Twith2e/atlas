package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// NewPostgresDB creates a PostgreSQL connection pool and verifies it is reachable.
func NewPostgresDB(connStr string) (*sql.DB, error) {
	// sql.Open configures the pool; it does not establish a connection immediately.
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Configure limits for connections retained and created by the pool.
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	// Bound the final health check so startup does not wait indefinitely.
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return nil, err
	}

	return db, nil
}
