package work

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMarketRegimeService is a mock implementation for testing
type mockMarketRegimeService struct {
	analyzeErr   error
	needsWork    bool
	analyzeCalls int
}

func (m *mockMarketRegimeService) AnalyzeMarketRegime() error {
	m.analyzeCalls++
	return m.analyzeErr
}

func (m *mockMarketRegimeService) NeedsAnalysis() bool {
	return m.needsWork
}

func TestRegisterAnalysisWorkTypes(t *testing.T) {
	registry := NewRegistry()

	deps := &AnalysisDeps{
		MarketRegimeService: &mockMarketRegimeService{needsWork: true},
	}

	RegisterAnalysisWorkTypes(registry, deps)

	// Verify work type is registered
	wt := registry.Get("analysis:market-regime")
	require.NotNil(t, wt)

	assert.Equal(t, "analysis:market-regime", wt.ID)
	assert.Equal(t, AllMarketsClosed, wt.MarketTiming)
	assert.Equal(t, 24*time.Hour, wt.Interval)
}

func TestAnalysisMarketRegime_Execute(t *testing.T) {
	registry := NewRegistry()

	mock := &mockMarketRegimeService{needsWork: true}
	deps := &AnalysisDeps{
		MarketRegimeService: mock,
	}

	RegisterAnalysisWorkTypes(registry, deps)

	wt := registry.Get("analysis:market-regime")
	require.NotNil(t, wt)

	err := wt.Execute(context.Background(), "", nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, mock.analyzeCalls)
}

func TestAnalysisMarketRegime_FindSubjects(t *testing.T) {
	registry := NewRegistry()

	t.Run("returns global work when needs analysis", func(t *testing.T) {
		mock := &mockMarketRegimeService{needsWork: true}
		deps := &AnalysisDeps{
			MarketRegimeService: mock,
		}

		RegisterAnalysisWorkTypes(registry, deps)

		wt := registry.Get("analysis:market-regime")
		require.NotNil(t, wt)

		subjects := wt.FindSubjects()
		assert.Equal(t, []string{""}, subjects)
	})

	t.Run("returns nil when no analysis needed", func(t *testing.T) {
		registry := NewRegistry()
		mock := &mockMarketRegimeService{needsWork: false}
		deps := &AnalysisDeps{
			MarketRegimeService: mock,
		}

		RegisterAnalysisWorkTypes(registry, deps)

		wt := registry.Get("analysis:market-regime")
		require.NotNil(t, wt)

		subjects := wt.FindSubjects()
		assert.Nil(t, subjects)
	})
}
