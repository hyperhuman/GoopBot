// Package db provides database connection and migration utilities for GoopBot.
package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Config holds database configuration
type Config struct {
	Path string
}

// NewSQLiteDB creates a new SQLite database connection
func NewSQLiteDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Ping to verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// Migrate applies necessary schema migrations
func Migrate(db *sql.DB) error {
	queries := []string{
		`
        CREATE TABLE IF NOT EXISTS twitch_subscriptions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            discord_user_id TEXT NOT NULL,
            twitch_username TEXT NOT NULL UNIQUE,
            channel_id TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (channel_id) REFERENCES channels(id)
        );
        `,
		`
        CREATE TABLE IF NOT EXISTS channels (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            guild_id TEXT NOT NULL
        );
        `,
	}

	for _, q := range queries {
		_, err := db.Exec(q)
		if err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	log.Println("Database migration completed successfully")
	return nil
}
