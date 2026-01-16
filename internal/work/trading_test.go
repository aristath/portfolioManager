package work

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTradingExecutionService is a mock implementation for testing
type mockTradingExecutionService struct {
	executeErr   error
	hasPending   bool
	executeCalls int
	pendingCalls int
}

func (m *mockTradingExecutionService) ExecutePendingTrades() error {
	m.executeCalls++
	return m.executeErr
}

func (m *mockTradingExecutionService) HasPendingTrades() bool {
	m.pendingCalls++
	return m.hasPending
}

// mockTradingRetryService is a mock implementation for testing
type mockTradingRetryService struct {
	retryErr    error
	hasFailed   bool
	retryCalls  int
	failedCalls int
}

func (m *mockTradingRetryService) RetryFailedTrades() error {
	m.retryCalls++
	return m.retryErr
}

func (m *mockTradingRetryService) HasFailedTrades() bool {
	m.failedCalls++
	return m.hasFailed
}

func TestRegisterTradingWorkTypes(t *testing.T) {
	registry := NewRegistry()

	deps := &TradingDeps{
		ExecutionService: &mockTradingExecutionService{hasPending: true},
		RetryService:     &mockTradingRetryService{hasFailed: true},
	}

	RegisterTradingWorkTypes(registry, deps)

	// Verify all work types are registered
	assert.NotNil(t, registry.Get("trading:execute"))
	assert.NotNil(t, registry.Get("trading:retry"))
}

func TestTradingExecute_Execute(t *testing.T) {
	registry := NewRegistry()

	mock := &mockTradingExecutionService{hasPending: true}

	deps := &TradingDeps{
		ExecutionService: mock,
		RetryService:     &mockTradingRetryService{},
	}

	RegisterTradingWorkTypes(registry, deps)

	wt := registry.Get("trading:execute")
	require.NotNil(t, wt)

	err := wt.Execute(context.Background(), "", nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, mock.executeCalls)
}

func TestTradingExecute_ExecuteError(t *testing.T) {
	registry := NewRegistry()

	mock := &mockTradingExecutionService{
		hasPending: true,
		executeErr: errors.New("execution failed"),
	}

	deps := &TradingDeps{
		ExecutionService: mock,
		RetryService:     &mockTradingRetryService{},
	}

	RegisterTradingWorkTypes(registry, deps)

	wt := registry.Get("trading:execute")
	require.NotNil(t, wt)

	err := wt.Execute(context.Background(), "", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "execution failed")
}

func TestTradingExecute_FindSubjects(t *testing.T) {
	t.Run("returns global work when has pending trades", func(t *testing.T) {
		registry := NewRegistry()

		deps := &TradingDeps{
			ExecutionService: &mockTradingExecutionService{hasPending: true},
			RetryService:     &mockTradingRetryService{},
		}

		RegisterTradingWorkTypes(registry, deps)

		wt := registry.Get("trading:execute")
		require.NotNil(t, wt)

		subjects := wt.FindSubjects()
		assert.Equal(t, []string{""}, subjects)
	})

	t.Run("returns nil when no pending trades", func(t *testing.T) {
		registry := NewRegistry()

		deps := &TradingDeps{
			ExecutionService: &mockTradingExecutionService{hasPending: false},
			RetryService:     &mockTradingRetryService{},
		}

		RegisterTradingWorkTypes(registry, deps)

		wt := registry.Get("trading:execute")
		require.NotNil(t, wt)

		subjects := wt.FindSubjects()
		assert.Nil(t, subjects)
	})
}

func TestTradingRetry_Execute(t *testing.T) {
	registry := NewRegistry()

	mock := &mockTradingRetryService{hasFailed: true}

	deps := &TradingDeps{
		ExecutionService: &mockTradingExecutionService{},
		RetryService:     mock,
	}

	RegisterTradingWorkTypes(registry, deps)

	wt := registry.Get("trading:retry")
	require.NotNil(t, wt)

	err := wt.Execute(context.Background(), "", nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, mock.retryCalls)
}

func TestTradingRetry_FindSubjects(t *testing.T) {
	t.Run("returns global work when has failed trades", func(t *testing.T) {
		registry := NewRegistry()

		deps := &TradingDeps{
			ExecutionService: &mockTradingExecutionService{},
			RetryService:     &mockTradingRetryService{hasFailed: true},
		}

		RegisterTradingWorkTypes(registry, deps)

		wt := registry.Get("trading:retry")
		require.NotNil(t, wt)

		subjects := wt.FindSubjects()
		assert.Equal(t, []string{""}, subjects)
	})

	t.Run("returns nil when no failed trades", func(t *testing.T) {
		registry := NewRegistry()

		deps := &TradingDeps{
			ExecutionService: &mockTradingExecutionService{},
			RetryService:     &mockTradingRetryService{hasFailed: false},
		}

		RegisterTradingWorkTypes(registry, deps)

		wt := registry.Get("trading:retry")
		require.NotNil(t, wt)

		subjects := wt.FindSubjects()
		assert.Nil(t, subjects)
	})
}

func TestTradingWorkTypes_MarketTiming(t *testing.T) {
	registry := NewRegistry()

	deps := &TradingDeps{
		ExecutionService: &mockTradingExecutionService{hasPending: true},
		RetryService:     &mockTradingRetryService{hasFailed: true},
	}

	RegisterTradingWorkTypes(registry, deps)

	// Both trading work types should run DuringMarketOpen
	for _, id := range []string{"trading:execute", "trading:retry"} {
		wt := registry.Get(id)
		require.NotNil(t, wt, "work type %s should exist", id)
		assert.Equal(t, DuringMarketOpen, wt.MarketTiming, "work type %s should have DuringMarketOpen timing", id)
	}
}

func TestTradingWorkTypes_Priority(t *testing.T) {
	t.Skip("Priority removed in favor of FIFO registration order")
	registry := NewRegistry()

	deps := &TradingDeps{
		ExecutionService: &mockTradingExecutionService{hasPending: true},
		RetryService:     &mockTradingRetryService{hasFailed: true},
	}

	RegisterTradingWorkTypes(registry, deps)

	// Execute should be Critical
	execute := registry.Get("trading:execute")
	require.NotNil(t, execute)

	// Retry should be Medium
	retry := registry.Get("trading:retry")
	require.NotNil(t, retry)
}

func TestTradingWorkTypes_NoDependencies(t *testing.T) {
	registry := NewRegistry()

	deps := &TradingDeps{
		ExecutionService: &mockTradingExecutionService{hasPending: true},
		RetryService:     &mockTradingRetryService{hasFailed: true},
	}

	RegisterTradingWorkTypes(registry, deps)

	// Neither trading work type has dependencies
	execute := registry.Get("trading:execute")
	require.NotNil(t, execute)
	assert.Empty(t, execute.DependsOn)

	retry := registry.Get("trading:retry")
	require.NotNil(t, retry)
	assert.Empty(t, retry.DependsOn)
}
