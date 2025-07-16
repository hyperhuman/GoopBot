package bot

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"google.golang.org/api/calendar/v3"

	"GoopBot/internal/bot/config"
	"GoopBot/internal/bot/logger"
	"GoopBot/internal/features"
)

// Types and interfaces
type Config struct {
	LogLevel string `json:"log_level"`
	// Add other config fields as needed
}

type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

type Feature interface {
	Name() string
	Initialize(*Bot) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Bot represents the main bot structure
type Bot struct {
	config   Config
	logger   Logger
	features map[string]Feature
	mu       sync.RWMutex
}

// NewBot creates a new bot instance
func NewBot(ctx context.Context) (*Bot, error) {
	// Load configuration
	config, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	// Initialize logger
	logger := logger.NewLogger(config.LogLevel)

	// Create feature manager
	features := make(map[string]Feature)

	return &Bot{
		config:   config,
		logger:   logger,
		features: features,
	}, nil
}
