// internal/bot/features/features.go
package features

import (
	"context"
	"fmt"
	"sync"

	"GoopBot/internal/bot/bot"
	"GoopBot/internal/bot/config"
	"GoopBot/internal/bot/logger"
)

// FeatureManager coordinates all bot features
type FeatureManager struct {
	features map[string]bot.Feature
	mu       sync.RWMutex
	bot      *bot.Bot
}

// NewFeatureManager creates a new feature manager instance
func NewFeatureManager(b *bot.Bot) *FeatureManager {
	return &FeatureManager{
		features: make(map[string]bot.Feature),
		bot:      b,
	}
}

// RegisterFeature adds a new feature to the manager
func (m *FeatureManager) RegisterFeature(f bot.Feature) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.features[f.Name()]; exists {
		return fmt.Errorf("feature %s already registered", f.Name())
	}

	// Initialize the feature
	if err := f.Initialize(m.bot); err != nil {
		return fmt.Errorf("initializing feature %s: %w", f.Name(), err)
	}

	m.features[f.Name()] = f
	return nil
}

// GetFeature retrieves a registered feature
func (m *FeatureManager) GetFeature(name string) (bot.Feature, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	f, ok := m.features[name]
	if !ok {
		return nil, fmt.Errorf("feature %s not found", name)
	}

	return f, nil
}

// StartAll starts all registered features
func (m *FeatureManager) StartAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var wg sync.WaitGroup
	errors := make(chan error, len(m.features))

	for _, feature := range m.features {
		wg.Add(1)
		go func(f bot.Feature) {
			defer wg.Done()
			if err := f.Start(ctx); err != nil {
				errors <- fmt.Errorf("starting feature %s: %w", f.Name(), err)
			}
		}(feature)
	}

	go func() {
		wg.Wait()
		close(errors)
	}()

	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}

// StopAll stops all registered features
func (m *FeatureManager) StopAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var wg sync.WaitGroup
	errors := make(chan error, len(m.features))

	for _, feature := range m.features {
		wg.Add(1)
		go func(f bot.Feature) {
			defer wg.Done()
			if err := f.Stop(ctx); err != nil {
				errors <- fmt.Errorf("stopping feature %s: %w", f.Name(), err)
			}
		}(feature)
	}

	go func() {
		wg.Wait()
		close(errors)
	}()

	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}
