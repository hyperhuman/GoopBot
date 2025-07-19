// Package redisutil provides utilities for interacting with Redis, including client creation and stream status caching.
package redisutil

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis configuration
type Config struct {
	Addr     string
	Password string
	DB       int
}

// NewRedisClient creates a new Redis client
func NewRedisClient(ctx context.Context, cfg Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Connected to Redis successfully")
	return rdb, nil
}

// StreamStatus represents stream status cache entry
type StreamStatus struct {
	Username string
	IsLive   bool
	LastSeen time.Time
}

// SetStreamStatus caches stream status
func SetStreamStatus(ctx context.Context, rdb *redis.Client, username string, isLive bool) error {
	status := StreamStatus{
		Username: username,
		IsLive:   isLive,
		LastSeen: time.Now(),
	}

	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	if err := rdb.Set(ctx, username, data, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to cache stream status: %w", err)
	}

	return nil
}
