package testing

import (
	"errors"
	"testing"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/aristath/sentinel/internal/modules/allocation"
	"github.com/aristath/sentinel/internal/modules/portfolio"
	"github.com/aristath/sentinel/internal/modules/trading"
	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMockPositionRepository tests that NewMockPositionRepository creates a valid mock
func TestNewMockPositionRepository(t *testing.T) {
	mock := NewMockPositionRepository()
	require.NotNil(t, mock)

	// Verify it implements the interface
	var _ portfolio.PositionRepositoryInterface = mock

	// Test GetAll returns empty slice by default
	positions, err := mock.GetAll()
	require.NoError(t, err)
	assert.NotNil(t, positions)
	assert.Equal(t, 0, len(positions))
}

// TestNewMockPositionRepository_WithPositions tests that mock can be configured with positions
func TestNewMockPositionRepository_WithPositions(t *testing.T) {
	mock := NewMockPositionRepository()
	testPositions := []portfolio.Position{
		{ISIN: "US0378331005", Symbol: "AAPL", Quantity: 10, AvgPrice: 150.0},
		{ISIN: "US5949181045", Symbol: "MSFT", Quantity: 5, AvgPrice: 200.0},
	}

	// Set positions
	mock.SetPositions(testPositions)

	// Verify GetAll returns configured positions
	positions, err := mock.GetAll()
	require.NoError(t, err)
	require.Equal(t, 2, len(positions))
	assert.Equal(t, "AAPL", positions[0].Symbol)
	assert.Equal(t, "MSFT", positions[1].Symbol)
}

// TestNewMockPositionRepository_GetBySymbol tests GetBySymbol functionality
func TestNewMockPositionRepository_GetBySymbol(t *testing.T) {
	mock := NewMockPositionRepository()
	testPositions := []portfolio.Position{
		{ISIN: "US0378331005", Symbol: "AAPL", Quantity: 10},
	}
	mock.SetPositions(testPositions)

	// Test GetBySymbol
	pos, err := mock.GetBySymbol("AAPL")
	require.NoError(t, err)
	require.NotNil(t, pos)
	assert.Equal(t, "AAPL", pos.Symbol)

	// Test non-existent symbol
	pos, err = mock.GetBySymbol("INVALID")
	require.NoError(t, err)
	assert.Nil(t, pos)
}

// TestNewMockPositionRepository_ErrorInjection tests that mock can simulate errors
func TestNewMockPositionRepository_ErrorInjection(t *testing.T) {
	mock := NewMockPositionRepository()
	testErr := errors.New("simulated error")
	mock.SetError(testErr)

	// Verify error is returned
	_, err := mock.GetAll()
	require.Error(t, err)
	assert.Equal(t, testErr, err)
}

// TestNewMockTradeRepository tests that NewMockTradeRepository creates a valid mock
func TestNewMockTradeRepository(t *testing.T) {
	mock := NewMockTradeRepository()
	require.NotNil(t, mock)

	// Verify it implements the interface
	var _ trading.TradeRepositoryInterface = mock

	// Test GetHistory returns empty slice by default
	trades, err := mock.GetHistory(10)
	require.NoError(t, err)
	assert.NotNil(t, trades)
	assert.Equal(t, 0, len(trades))
}

// TestNewMockSecurityRepository tests that NewMockSecurityRepository creates a valid mock
func TestNewMockSecurityRepository(t *testing.T) {
	mock := NewMockSecurityRepository()
	require.NotNil(t, mock)

	// Verify it implements the interface
	var _ universe.SecurityRepositoryInterface = mock

	// Test GetBySymbol returns nil by default
	security, err := mock.GetBySymbol("AAPL")
	require.NoError(t, err)
	assert.Nil(t, security)
}

// TestNewMockSecurityRepository_WithSecurities tests mock with configured securities
func TestNewMockSecurityRepository_WithSecurities(t *testing.T) {
	mock := NewMockSecurityRepository()
	testSecurity := &universe.Security{
		ISIN:   "US0378331005",
		Symbol: "AAPL",
		Name:   "Apple Inc.",
	}

	mock.SetSecurity(testSecurity)

	// Test GetBySymbol
	security, err := mock.GetBySymbol("AAPL")
	require.NoError(t, err)
	require.NotNil(t, security)
	assert.Equal(t, "AAPL", security.Symbol)
	assert.Equal(t, "Apple Inc.", security.Name)
}

// TestNewMockCashManager tests that NewMockCashManager creates a valid mock
func TestNewMockCashManager(t *testing.T) {
	mock := NewMockCashManager()
	require.NotNil(t, mock)

	// Verify it implements the interface
	var _ domain.CashManager = mock

	// Test GetAllCashBalances returns empty map by default
	balances, err := mock.GetAllCashBalances()
	require.NoError(t, err)
	assert.NotNil(t, balances)
	assert.Equal(t, 0, len(balances))
}

// TestNewMockCashManager_WithBalances tests mock with configured balances
func TestNewMockCashManager_WithBalances(t *testing.T) {
	mock := NewMockCashManager()
	testBalances := map[string]float64{
		"EUR": 1000.0,
		"USD": 500.0,
	}
	mock.SetBalances(testBalances)

	// Test GetAllCashBalances
	balances, err := mock.GetAllCashBalances()
	require.NoError(t, err)
	require.Equal(t, 2, len(balances))
	assert.Equal(t, 1000.0, balances["EUR"])
	assert.Equal(t, 500.0, balances["USD"])
}

// TestNewMockTradernetClient tests that NewMockTradernetClient creates a valid mock
func TestNewMockTradernetClient(t *testing.T) {
	mock := NewMockTradernetClient()
	require.NotNil(t, mock)

	// Verify it implements the interface
	var _ domain.TradernetClientInterface = mock

	// Test IsConnected returns false by default
	assert.False(t, mock.IsConnected())
}

// TestNewMockTradernetClient_ConfigureConnection tests mock connection configuration
func TestNewMockTradernetClient_ConfigureConnection(t *testing.T) {
	mock := NewMockTradernetClient()

	// Set connected
	mock.SetConnected(true)
	assert.True(t, mock.IsConnected())

	// Set disconnected
	mock.SetConnected(false)
	assert.False(t, mock.IsConnected())
}

// TestNewMockCurrencyExchangeService tests that NewMockCurrencyExchangeService creates a valid mock
func TestNewMockCurrencyExchangeService(t *testing.T) {
	mock := NewMockCurrencyExchangeService()
	require.NotNil(t, mock)

	// Verify it implements the interface
	var _ domain.CurrencyExchangeServiceInterface = mock

	// Test GetRate returns error by default
	rate, err := mock.GetRate("USD", "EUR")
	require.Error(t, err)
	assert.Equal(t, 0.0, rate)
}

// TestNewMockCurrencyExchangeService_WithRates tests mock with configured rates
func TestNewMockCurrencyExchangeService_WithRates(t *testing.T) {
	mock := NewMockCurrencyExchangeService()
	testRates := map[string]map[string]float64{
		"USD": {
			"EUR": 0.85,
			"GBP": 0.73,
		},
	}
	mock.SetRates(testRates)

	// Test GetRate
	rate, err := mock.GetRate("USD", "EUR")
	require.NoError(t, err)
	assert.Equal(t, 0.85, rate)

	// Test reverse rate
	rate, err = mock.GetRate("EUR", "USD")
	require.NoError(t, err)
	assert.InDelta(t, 1.176, rate, 0.001) // 1/0.85
}

// TestNewMockAllocationTargetProvider tests that NewMockAllocationTargetProvider creates a valid mock
func TestNewMockAllocationTargetProvider(t *testing.T) {
	mock := NewMockAllocationTargetProvider()
	require.NotNil(t, mock)

	// Verify it implements the interface
	var _ domain.AllocationTargetProvider = mock

	// Test GetAll returns empty slice by default
	targets, err := mock.GetAll()
	require.NoError(t, err)
	assert.NotNil(t, targets)
	assert.Equal(t, 0, len(targets))
}

// TestNewMockAllocationTargetProvider_WithTargets tests mock with configured targets
func TestNewMockAllocationTargetProvider_WithTargets(t *testing.T) {
	mock := NewMockAllocationTargetProvider()
	testTargets := []allocation.AllocationTarget{
		{Type: "country", Name: "EU", TargetPct: 0.40},
		{Type: "country", Name: "US", TargetPct: 0.60},
	}
	mock.SetTargets(testTargets)

	// Test GetAll
	targets, err := mock.GetAll()
	require.NoError(t, err)
	require.Equal(t, 2, len(targets))
	assert.Equal(t, 0.40, targets["country:EU"])
	assert.Equal(t, 0.60, targets["country:US"])
}

// TestNewMockPortfolioSummaryProvider tests that NewMockPortfolioSummaryProvider creates a valid mock
func TestNewMockPortfolioSummaryProvider(t *testing.T) {
	mock := NewMockPortfolioSummaryProvider()
	require.NotNil(t, mock)

	// Verify it implements the interface
	var _ domain.PortfolioSummaryProvider = mock

	// Test GetPortfolioSummary returns zero summary by default
	summary, err := mock.GetPortfolioSummary()
	require.NoError(t, err)
	assert.Equal(t, 0.0, summary.TotalValue)
	assert.Equal(t, 0, len(summary.CountryAllocations))
}

// TestNewMockPortfolioSummaryProvider_WithSummary tests mock with configured summary
func TestNewMockPortfolioSummaryProvider_WithSummary(t *testing.T) {
	mock := NewMockPortfolioSummaryProvider()
	testSummary := domain.PortfolioSummary{
		TotalValue:  10000.0,
		CashBalance: 1000.0,
		CountryAllocations: []domain.PortfolioAllocation{
			{Name: "EU", CurrentPct: 0.40},
			{Name: "US", CurrentPct: 0.60},
		},
	}
	mock.SetSummary(testSummary)

	// Test GetPortfolioSummary
	summary, err := mock.GetPortfolioSummary()
	require.NoError(t, err)
	assert.Equal(t, 10000.0, summary.TotalValue)
	assert.Equal(t, 2, len(summary.CountryAllocations))
}

// TestNewMockPlannerRepository tests that NewMockPlannerRepository creates a valid mock
func TestNewMockPlannerRepository(t *testing.T) {
	// This test will be implemented when the mock is created
	// For now, it's a placeholder to define the expected interface
	t.Skip("Mock implementation pending")
}

// TestNewMockRecommendationRepository tests that NewMockRecommendationRepository creates a valid mock
func TestNewMockRecommendationRepository(t *testing.T) {
	// This test will be implemented when the mock is created
	// For now, it's a placeholder to define the expected interface
	t.Skip("Mock implementation pending")
}

// TestNewMockEventManager tests that NewMockEventManager creates a valid mock
func TestNewMockEventManager(t *testing.T) {
	// This test will be implemented when the mock is created
	// For now, it's a placeholder to define the expected interface
	t.Skip("Mock implementation pending")
}

// TestNewMockEventManager_EventTracking tests that mock tracks published events
func TestNewMockEventManager_EventTracking(t *testing.T) {
	// This test will be implemented when the mock is created
	// For now, it's a placeholder to define the expected interface
	t.Skip("Mock implementation pending")
}
