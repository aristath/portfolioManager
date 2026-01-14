package calculations

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPriceProvider implements PriceProvider for testing
type mockPriceProvider struct {
	prices map[string][]float64
	err    error
}

func (m *mockPriceProvider) GetDailyPrices(isin string, days int) ([]float64, error) {
	if m.err != nil {
		return nil, m.err
	}
	prices, ok := m.prices[isin]
	if !ok {
		return []float64{}, nil
	}
	// Return up to 'days' prices from the end
	if len(prices) <= days {
		return prices, nil
	}
	return prices[len(prices)-days:], nil
}

// generatePrices creates a slice of test prices with slight variation
func generatePrices(count int) []float64 {
	prices := make([]float64, count)
	basePrice := 100.0
	for i := 0; i < count; i++ {
		// Create a gradual upward trend with some variation
		prices[i] = basePrice + float64(i)*0.1 + float64(i%10)*0.5
	}
	return prices
}

func testLogger() zerolog.Logger {
	return zerolog.Nop()
}

func TestTechnicalCalculator_CalculateForISIN(t *testing.T) {
	cache := setupTestCache(t)
	priceProvider := &mockPriceProvider{
		prices: map[string][]float64{
			"TEST_ISIN": generatePrices(300),
		},
	}

	calc := NewTechnicalCalculator(cache, priceProvider, testLogger())

	err := calc.CalculateForISIN("TEST_ISIN")
	require.NoError(t, err)

	// Verify all metrics were cached
	_, ok := cache.GetTechnical("TEST_ISIN", "ema", 200)
	assert.True(t, ok, "EMA-200 should be cached")

	_, ok = cache.GetTechnical("TEST_ISIN", "ema", 50)
	assert.True(t, ok, "EMA-50 should be cached")

	_, ok = cache.GetTechnical("TEST_ISIN", "rsi", 14)
	assert.True(t, ok, "RSI-14 should be cached")

	_, ok = cache.GetTechnical("TEST_ISIN", "sharpe", 0)
	assert.True(t, ok, "Sharpe ratio should be cached")

	_, ok = cache.GetTechnical("TEST_ISIN", "max_drawdown", 0)
	assert.True(t, ok, "Max drawdown should be cached")

	_, ok = cache.GetTechnical("TEST_ISIN", "52w_high", 0)
	assert.True(t, ok, "52-week high should be cached")

	_, ok = cache.GetTechnical("TEST_ISIN", "52w_low", 0)
	assert.True(t, ok, "52-week low should be cached")
}

func TestTechnicalCalculator_InsufficientData(t *testing.T) {
	cache := setupTestCache(t)
	priceProvider := &mockPriceProvider{
		prices: map[string][]float64{
			"TEST_ISIN": generatePrices(10), // Only 10 prices
		},
	}

	calc := NewTechnicalCalculator(cache, priceProvider, testLogger())

	err := calc.CalculateForISIN("TEST_ISIN")
	require.NoError(t, err, "Should not error on insufficient data")

	// EMA-200 requires 200 prices, should not be cached
	_, ok := cache.GetTechnical("TEST_ISIN", "ema", 200)
	assert.False(t, ok, "EMA-200 should not be cached with insufficient data")

	// EMA-50 requires 50 prices, should not be cached
	_, ok = cache.GetTechnical("TEST_ISIN", "ema", 50)
	assert.False(t, ok, "EMA-50 should not be cached with insufficient data")
}

func TestTechnicalCalculator_GetCachedValues(t *testing.T) {
	cache := setupTestCache(t)
	priceProvider := &mockPriceProvider{
		prices: map[string][]float64{
			"TEST_ISIN": generatePrices(300),
		},
	}

	calc := NewTechnicalCalculator(cache, priceProvider, testLogger())

	err := calc.CalculateForISIN("TEST_ISIN")
	require.NoError(t, err)

	// Verify cached values are reasonable
	ema200, ok := cache.GetTechnical("TEST_ISIN", "ema", 200)
	assert.True(t, ok)
	assert.Greater(t, ema200, 0.0, "EMA-200 should be positive")

	ema50, ok := cache.GetTechnical("TEST_ISIN", "ema", 50)
	assert.True(t, ok)
	assert.Greater(t, ema50, 0.0, "EMA-50 should be positive")

	rsi, ok := cache.GetTechnical("TEST_ISIN", "rsi", 14)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, rsi, 0.0, "RSI should be >= 0")
	assert.LessOrEqual(t, rsi, 100.0, "RSI should be <= 100")

	mdd, ok := cache.GetTechnical("TEST_ISIN", "max_drawdown", 0)
	assert.True(t, ok)
	assert.LessOrEqual(t, mdd, 0.0, "Max drawdown should be <= 0")

	high52w, ok := cache.GetTechnical("TEST_ISIN", "52w_high", 0)
	assert.True(t, ok)
	assert.Greater(t, high52w, 0.0, "52-week high should be positive")

	low52w, ok := cache.GetTechnical("TEST_ISIN", "52w_low", 0)
	assert.True(t, ok)
	assert.Greater(t, low52w, 0.0, "52-week low should be positive")
	assert.LessOrEqual(t, low52w, high52w, "52-week low should be <= high")
}

func TestTechnicalCalculator_MissingISIN(t *testing.T) {
	cache := setupTestCache(t)
	priceProvider := &mockPriceProvider{
		prices: map[string][]float64{}, // No prices
	}

	calc := NewTechnicalCalculator(cache, priceProvider, testLogger())

	err := calc.CalculateForISIN("MISSING_ISIN")
	require.NoError(t, err, "Should not error on missing ISIN")

	// Nothing should be cached
	_, ok := cache.GetTechnical("MISSING_ISIN", "ema", 200)
	assert.False(t, ok)
}

func TestTechnicalCalculator_PartialData(t *testing.T) {
	cache := setupTestCache(t)
	// 100 prices - enough for EMA-50 and RSI-14, but not EMA-200 or 52-week
	priceProvider := &mockPriceProvider{
		prices: map[string][]float64{
			"TEST_ISIN": generatePrices(100),
		},
	}

	calc := NewTechnicalCalculator(cache, priceProvider, testLogger())

	err := calc.CalculateForISIN("TEST_ISIN")
	require.NoError(t, err)

	// EMA-50 should be cached (100 >= 50)
	_, ok := cache.GetTechnical("TEST_ISIN", "ema", 50)
	assert.True(t, ok, "EMA-50 should be cached with 100 prices")

	// RSI-14 should be cached
	_, ok = cache.GetTechnical("TEST_ISIN", "rsi", 14)
	assert.True(t, ok, "RSI-14 should be cached with 100 prices")

	// EMA-200 should NOT be cached (100 < 200)
	_, ok = cache.GetTechnical("TEST_ISIN", "ema", 200)
	assert.False(t, ok, "EMA-200 should not be cached with only 100 prices")

	// 52-week high/low should NOT be cached (100 < 252)
	_, ok = cache.GetTechnical("TEST_ISIN", "52w_high", 0)
	assert.False(t, ok, "52-week high should not be cached with only 100 prices")
}

func TestTechnicalCalculator_MultipleISINs(t *testing.T) {
	cache := setupTestCache(t)
	priceProvider := &mockPriceProvider{
		prices: map[string][]float64{
			"ISIN1": generatePrices(300),
			"ISIN2": generatePrices(300),
		},
	}

	calc := NewTechnicalCalculator(cache, priceProvider, testLogger())

	err := calc.CalculateForISIN("ISIN1")
	require.NoError(t, err)

	err = calc.CalculateForISIN("ISIN2")
	require.NoError(t, err)

	// Both should have cached values
	_, ok := cache.GetTechnical("ISIN1", "ema", 200)
	assert.True(t, ok)

	_, ok = cache.GetTechnical("ISIN2", "ema", 200)
	assert.True(t, ok)
}

func TestTechnicalCalculator_Idempotent(t *testing.T) {
	cache := setupTestCache(t)
	priceProvider := &mockPriceProvider{
		prices: map[string][]float64{
			"TEST_ISIN": generatePrices(300),
		},
	}

	calc := NewTechnicalCalculator(cache, priceProvider, testLogger())

	// Calculate twice
	err := calc.CalculateForISIN("TEST_ISIN")
	require.NoError(t, err)

	ema1, _ := cache.GetTechnical("TEST_ISIN", "ema", 200)

	err = calc.CalculateForISIN("TEST_ISIN")
	require.NoError(t, err)

	ema2, _ := cache.GetTechnical("TEST_ISIN", "ema", 200)

	// Values should be the same (same input data)
	assert.Equal(t, ema1, ema2, "Repeated calculations should produce same result")
}
