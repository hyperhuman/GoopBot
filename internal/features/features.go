// Package features provides the feature management system for the bot.
package features

import (
	"fmt"
	"sync"

	"GoopBot/internal/bot"
)

type FeatureManager struct {
	features map[string]Feature
	mu       sync.RWMutex
}

func (m *FeatureManager) RegisterFeature(f Feature) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.features[f.Name()]; exists {
		return fmt.Errorf("feature %s already registered", f.Name())
	}

	m.features[f.Name()] = f
	return nil
}
func (m *FeatureManager) GetFeature(name string) (Feature, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	feature, exists := m.features[name]
	if !exists {
		return nil, fmt.Errorf("feature %s not found", name)
	}
	return feature, nil
}

func NewFeatureManager() *FeatureManager {
	return &FeatureManager{
		features: make(map[string]Feature),
	}
}
