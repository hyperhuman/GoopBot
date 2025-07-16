package bot

import (
    "context"
    "sync"
)

type Bot struct {
    config     Config
    logger     Logger
    features   map[string]Feature
    mu         sync.RWMutex
}

func NewBot(ctx context.Context) (*Bot, error) {
    // Load configuration
    config, err := loadConfig()
    if err != nil {
        return nil, fmt.Errorf("loading config: %w", err)
    }

    // Initialize logger
    logger := newLogger(config.LogLevel)

    // Create feature manager
    features := make(map[string]Feature)

    return &Bot{
        config:   config,
        logger:   logger,
        features: features,
    }, nil
}