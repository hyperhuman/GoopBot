// Package bot provides the core bot functionality and feature management.
package bot

import (
	"context"
	"fmt"
	"sync"
)

type Feature interface {
	Name() string
	Initialize(*Bot) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Bot struct {
	config   Config
	logger   Logger
	features map[string]Feature
	mu       sync.RWMutex
}

func (b *Bot) Start(ctx context.Context) error {
	for _, feature := range b.features {
		if err := feature.Start(ctx); err != nil {
			return fmt.Errorf("starting feature %s: %w", feature.Name(), err)
		}
	}
	return nil
}

func (b *Bot) Stop(ctx context.Context) error {
	// Stop all features
	for _, feature := range b.features {
		if err := feature.Stop(ctx); err != nil {
			return fmt.Errorf("stopping feature %s: %w", feature.Name(), err)
		}
	}
	return nil
}

func (b *Bot) RegisterFeature(f Feature) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.features[f.Name()]; exists {
		return fmt.Errorf("feature %s already registered", f.Name())
	}

	if err := f.Initialize(b); err != nil {
		return fmt.Errorf("initializing feature %s: %w", f.Name(), err)
	}

	b.features[f.Name()] = f
	return nil
}

func NewBot(ctx context.Context) (*Bot, error) {
	// Initialize bot with default configuration and logger
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	logger := NewLogger(config.LogLevel)

	bot := &Bot{
		config:   config,
		logger:   logger,
		features: make(map[string]Feature),
	}

	// Register features
	if err := bot.RegisterFeature(NewTwitchFeature()); err != nil {
		return nil, fmt.Errorf("registering Twitch feature: %w", err)
	}

	return bot, nil
}

// TwitchFeature is a dummy implementation of the Feature interface.
type TwitchFeature struct{}

func (t *TwitchFeature) Name() string {
	return "Twitch"
}

func (t *TwitchFeature) Initialize(b *Bot) error {
	// Initialization logic here
	return nil
}

func (t *TwitchFeature) Start(ctx context.Context) error {
	// Start logic here
	return nil
}

func (t *TwitchFeature) Stop(ctx context.Context) error {
	// Stop logic here
	return nil
}

func NewTwitchFeature() Feature {
	return &TwitchFeature{}
}
