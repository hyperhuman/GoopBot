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

	key := fmt.Sprintf("stream:%s", username)
	if err := rdb.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to cache stream status: %w", err)
	}

	return nil
}

// GetStreamStatus retrieves cached stream status
func GetStreamStatus(ctx context.Context, rdb *redis.Client, username string) (*StreamStatus, error) {
	key := fmt.Sprintf("stream:%s", username)
	data, err := rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get stream status: %w", err)
	}

	var status StreamStatus
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status: %w", err)
	}

	return &status, nil
}

// SetStreamCheckCooldown sets a cooldown to prevent too frequent API calls
func SetStreamCheckCooldown(ctx context.Context, rdb *redis.Client, username string, duration time.Duration) error {
	key := fmt.Sprintf("cooldown:%s", username)
	return rdb.Set(ctx, key, "checked", duration).Err()
}

// IsStreamCheckOnCooldown checks if a stream check is on cooldown
func IsStreamCheckOnCooldown(ctx context.Context, rdb *redis.Client, username string) bool {
	key := fmt.Sprintf("cooldown:%s", username)
	_, err := rdb.Get(ctx, key).Result()
	return err != redis.Nil // If key exists, it's on cooldown
}
