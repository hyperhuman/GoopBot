// internal/bot/features/features_test.go
package features

import (
	"context"
	"fmt"
	"testing"

	"GoopBot/internal/bot"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockFeature is a mock implementation of the Feature interface
type mockFeature struct {
	mock.Mock
	name        string
	initialized bool
}

func newMockFeature(name string) *mockFeature {
	return &mockFeature{name: name}
}

func (m *mockFeature) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockFeature) Initialize(b *bot.Bot) error {
	args := m.Called(b)
	m.initialized = true
	return args.Error(0)
}

func (m *mockFeature) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockFeature) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestFeatureManagerRegisterFeature(t *testing.T) {
	manager := NewFeatureManager()
	mockFeature := newMockFeature("test-feature")

	// Setup expectations
	mockFeature.On("Name").Return("test-feature")
	mockFeature.On("Initialize", mock.Anything).Return(nil)

	// Execute
	err := manager.RegisterFeature(mockFeature)

	// Verify
	assert.NoError(t, err)
	feature, err := manager.GetFeature("test-feature")
	assert.NoError(t, err)
	assert.NotNil(t, feature)
	mockFeature.AssertExpectations(t)
}

func TestFeatureManagerStartAll(t *testing.T) {
	manager := NewFeatureManager()
	mockFeature := newMockFeature("test-feature")

	// Setup expectations
	mockFeature.On("Start", mock.Anything).Return(nil)
	mockFeature.On("Initialize", mock.Anything).Return(nil)

	// Register feature
	manager.RegisterFeature(mockFeature)

	// Execute
	err := manager.StartAll(context.Background())

	// Verify
	assert.NoError(t, err)
	mockFeature.AssertExpectations(t)
}

func NewFeatureManager() *FeatureManager {
	return &FeatureManager{
		features: make(map[string]Feature),
	}
}

// RegisterFeature registers a feature with the manager.
func (fm *FeatureManager) RegisterFeature(feature Feature) error {
	name := feature.Name()
	if _, exists := fm.features[name]; exists {
		return fmt.Errorf("feature %s already registered", name)
	}
	if err := feature.Initialize(nil); err != nil {
		return err
	}
	fm.features[name] = feature
	return nil
}

// StartAll starts all registered features.
func (fm *FeatureManager) StartAll(ctx context.Context) error {
	for _, feature := range fm.features {
		if err := feature.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

// GetFeature returns the feature by name if it exists, otherwise returns an error.
func (fm *FeatureManager) GetFeature(name string) (Feature, error) {
	feature, ok := fm.features[name]
	if !ok {
		return nil, fmt.Errorf("feature %s not found", name)
	}
	return feature, nil
}
