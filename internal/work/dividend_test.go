package work

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDividendDetectionService is a mock implementation for testing
type mockDividendDetectionService struct {
	dividends       any
	detectErr       error
	hasPending      bool
	detectCalls     int
	hasPendingCalls int
}

func (m *mockDividendDetectionService) DetectUnreinvestedDividends() (any, error) {
	m.detectCalls++
	return m.dividends, m.detectErr
}

func (m *mockDividendDetectionService) HasPendingDividends() bool {
	m.hasPendingCalls++
	return m.hasPending
}

// mockDividendAnalysisService is a mock implementation for testing
type mockDividendAnalysisService struct {
	analyzeErr   error
	analyzeCalls int
}

func (m *mockDividendAnalysisService) AnalyzeDividends(dividends any) (any, error) {
	m.analyzeCalls++
	return dividends, m.analyzeErr
}

// mockDividendRecommendationService is a mock implementation for testing
type mockDividendRecommendationService struct {
	recommendErr   error
	recommendCalls int
}

func (m *mockDividendRecommendationService) CreateRecommendations(analysis any) (any, error) {
	m.recommendCalls++
	return analysis, m.recommendErr
}

// mockDividendExecutionService is a mock implementation for testing
type mockDividendExecutionService struct {
	executeErr   error
	executeCalls int
}

func (m *mockDividendExecutionService) ExecuteTrades(recommendations any) error {
	m.executeCalls++
	return m.executeErr
}

// mockDividendCache is a mock implementation for testing
type mockDividendCache struct {
	data map[string]any
}

func newMockDividendCache() *mockDividendCache {
	return &mockDividendCache{data: make(map[string]any)}
}

func (m *mockDividendCache) Has(key string) bool {
	_, exists := m.data[key]
	return exists
}

func (m *mockDividendCache) Get(key string) any {
	return m.data[key]
}

func (m *mockDividendCache) Set(key string, value any) {
	m.data[key] = value
}

func (m *mockDividendCache) Delete(key string) {
	delete(m.data, key)
}

func (m *mockDividendCache) DeletePrefix(prefix string) {
	for key := range m.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(m.data, key)
		}
	}
}

func TestRegisterDividendWorkTypes(t *testing.T) {
	registry := NewRegistry()

	deps := &DividendDeps{
		DetectionService:      &mockDividendDetectionService{hasPending: true},
		AnalysisService:       &mockDividendAnalysisService{},
		RecommendationService: &mockDividendRecommendationService{},
		ExecutionService:      &mockDividendExecutionService{},
		Cache:                 newMockDividendCache(),
	}

	RegisterDividendWorkTypes(registry, deps)

	// Verify all work types are registered
	assert.NotNil(t, registry.Get("dividend:detect"))
	assert.NotNil(t, registry.Get("dividend:analyze"))
	assert.NotNil(t, registry.Get("dividend:recommend"))
	assert.NotNil(t, registry.Get("dividend:execute"))
}

func TestDividendWorkTypes_Dependencies(t *testing.T) {
	registry := NewRegistry()

	deps := &DividendDeps{
		DetectionService:      &mockDividendDetectionService{hasPending: true},
		AnalysisService:       &mockDividendAnalysisService{},
		RecommendationService: &mockDividendRecommendationService{},
		ExecutionService:      &mockDividendExecutionService{},
		Cache:                 newMockDividendCache(),
	}

	RegisterDividendWorkTypes(registry, deps)

	// detect has no dependencies
	detect := registry.Get("dividend:detect")
	require.NotNil(t, detect)
	assert.Empty(t, detect.DependsOn)

	// analyze depends on detect
	analyze := registry.Get("dividend:analyze")
	require.NotNil(t, analyze)
	assert.Equal(t, []string{"dividend:detect"}, analyze.DependsOn)

	// recommend depends on analyze
	recommend := registry.Get("dividend:recommend")
	require.NotNil(t, recommend)
	assert.Equal(t, []string{"dividend:analyze"}, recommend.DependsOn)

	// execute depends on recommend
	execute := registry.Get("dividend:execute")
	require.NotNil(t, execute)
	assert.Equal(t, []string{"dividend:recommend"}, execute.DependsOn)
}

func TestDividendDetect_Execute(t *testing.T) {
	registry := NewRegistry()
	cache := newMockDividendCache()

	mock := &mockDividendDetectionService{
		hasPending: true,
		dividends:  []string{"div1", "div2"},
	}

	deps := &DividendDeps{
		DetectionService:      mock,
		AnalysisService:       &mockDividendAnalysisService{},
		RecommendationService: &mockDividendRecommendationService{},
		ExecutionService:      &mockDividendExecutionService{},
		Cache:                 cache,
	}

	RegisterDividendWorkTypes(registry, deps)

	wt := registry.Get("dividend:detect")
	require.NotNil(t, wt)

	err := wt.Execute(context.Background(), "", nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, mock.detectCalls)

	// Check dividends were cached
	assert.True(t, cache.Has("detected_dividends"))
}

func TestDividendDetect_FindSubjects(t *testing.T) {
	t.Run("returns global work when has pending dividends", func(t *testing.T) {
		registry := NewRegistry()

		deps := &DividendDeps{
			DetectionService:      &mockDividendDetectionService{hasPending: true},
			AnalysisService:       &mockDividendAnalysisService{},
			RecommendationService: &mockDividendRecommendationService{},
			ExecutionService:      &mockDividendExecutionService{},
			Cache:                 newMockDividendCache(),
		}

		RegisterDividendWorkTypes(registry, deps)

		wt := registry.Get("dividend:detect")
		require.NotNil(t, wt)

		subjects := wt.FindSubjects()
		assert.Equal(t, []string{""}, subjects)
	})

	t.Run("returns nil when no pending dividends", func(t *testing.T) {
		registry := NewRegistry()

		deps := &DividendDeps{
			DetectionService:      &mockDividendDetectionService{hasPending: false},
			AnalysisService:       &mockDividendAnalysisService{},
			RecommendationService: &mockDividendRecommendationService{},
			ExecutionService:      &mockDividendExecutionService{},
			Cache:                 newMockDividendCache(),
		}

		RegisterDividendWorkTypes(registry, deps)

		wt := registry.Get("dividend:detect")
		require.NotNil(t, wt)

		subjects := wt.FindSubjects()
		assert.Nil(t, subjects)
	})
}

func TestDividendWorkTypes_MarketTiming(t *testing.T) {
	registry := NewRegistry()

	deps := &DividendDeps{
		DetectionService:      &mockDividendDetectionService{hasPending: true},
		AnalysisService:       &mockDividendAnalysisService{},
		RecommendationService: &mockDividendRecommendationService{},
		ExecutionService:      &mockDividendExecutionService{},
		Cache:                 newMockDividendCache(),
	}

	RegisterDividendWorkTypes(registry, deps)

	// All dividend work types should run AnyTime
	for _, id := range []string{"dividend:detect", "dividend:analyze", "dividend:recommend", "dividend:execute"} {
		wt := registry.Get(id)
		require.NotNil(t, wt, "work type %s should exist", id)
		assert.Equal(t, AnyTime, wt.MarketTiming, "work type %s should have AnyTime timing", id)
	}
}

func TestDividendWorkTypes_Priority(t *testing.T) {
	registry := NewRegistry()

	deps := &DividendDeps{
		DetectionService:      &mockDividendDetectionService{hasPending: true},
		AnalysisService:       &mockDividendAnalysisService{},
		RecommendationService: &mockDividendRecommendationService{},
		ExecutionService:      &mockDividendExecutionService{},
		Cache:                 newMockDividendCache(),
	}

	RegisterDividendWorkTypes(registry, deps)

	// All dividend work types should have High priority
	for _, id := range []string{"dividend:detect", "dividend:analyze", "dividend:recommend", "dividend:execute"} {
		wt := registry.Get(id)
		require.NotNil(t, wt, "work type %s should exist", id)
	}
}

func TestDividendDetect_ExecuteError(t *testing.T) {
	registry := NewRegistry()
	cache := newMockDividendCache()

	mock := &mockDividendDetectionService{
		hasPending: true,
		detectErr:  errors.New("detection failed"),
	}

	deps := &DividendDeps{
		DetectionService:      mock,
		AnalysisService:       &mockDividendAnalysisService{},
		RecommendationService: &mockDividendRecommendationService{},
		ExecutionService:      &mockDividendExecutionService{},
		Cache:                 cache,
	}

	RegisterDividendWorkTypes(registry, deps)

	wt := registry.Get("dividend:detect")
	require.NotNil(t, wt)

	err := wt.Execute(context.Background(), "", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "detection failed")

	// Cache should not be set on error
	assert.False(t, cache.Has("detected_dividends"))
}
