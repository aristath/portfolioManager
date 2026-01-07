package evaluation

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatShift(t *testing.T) {
	tests := []struct {
		name     string
		shift    float64
		expected string
	}{
		{
			name:     "zero shift",
			shift:    0.0,
			expected: "0",
		},
		{
			name:     "negative 10 percent",
			shift:    -0.10,
			expected: "-0.1",
		},
		{
			name:     "negative 5 percent",
			shift:    -0.05,
			expected: "-0.05",
		},
		{
			name:     "positive 5 percent",
			shift:    0.05,
			expected: "0.05",
		},
		{
			name:     "positive 10 percent",
			shift:    0.10,
			expected: "0.1",
		},
		{
			name:     "custom shift",
			shift:    0.15,
			expected: "0.15",
		},
		{
			name:     "negative custom shift",
			shift:    -0.15,
			expected: "-0.15",
		},
		{
			name:     "very small shift",
			shift:    0.001,
			expected: "0.00",
		},
		{
			name:     "large shift",
			shift:    0.25,
			expected: "0.25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatShift(tt.shift)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateRandomPrices(t *testing.T) {
	// Seed random for reproducibility
	// Note: This test checks that the function runs and produces valid output
	// The actual randomness is tested indirectly through integration tests

	symbols := []string{"AAPL", "GOOGL", "MSFT"}
	volatilities := map[string]float64{
		"AAPL":  0.25,
		"GOOGL": 0.30,
		// MSFT will use default 0.2
	}

	result := generateRandomPrices(symbols, volatilities)

	// Should have all symbols
	assert.Len(t, result, len(symbols))
	for _, symbol := range symbols {
		assert.Contains(t, result, symbol)
	}

	// Check that multipliers are in valid range [0.5, 2.0]
	for symbol, multiplier := range result {
		assert.GreaterOrEqual(t, multiplier, 0.5, "Multiplier for %s should be >= 0.5", symbol)
		assert.LessOrEqual(t, multiplier, 2.0, "Multiplier for %s should be <= 2.0", symbol)
		assert.False(t, math.IsNaN(multiplier), "Multiplier for %s should not be NaN", symbol)
		assert.False(t, math.IsInf(multiplier, 0), "Multiplier for %s should not be Inf", symbol)
	}

	// Test with empty symbols
	emptyResult := generateRandomPrices([]string{}, volatilities)
	assert.Len(t, emptyResult, 0)

	// Test with nil volatilities
	nilVolResult := generateRandomPrices(symbols, nil)
	assert.Len(t, nilVolResult, len(symbols))
	for symbol := range nilVolResult {
		multiplier := nilVolResult[symbol]
		assert.GreaterOrEqual(t, multiplier, 0.5)
		assert.LessOrEqual(t, multiplier, 2.0)
	}

	// Test with zero volatility (should use default)
	zeroVol := map[string]float64{"AAPL": 0.0}
	zeroVolResult := generateRandomPrices([]string{"AAPL"}, zeroVol)
	assert.Len(t, zeroVolResult, 1)
	assert.Contains(t, zeroVolResult, "AAPL")
}
