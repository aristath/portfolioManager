package scorers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortTermScorer_StoresRawDrawdown(t *testing.T) {
	scorer := NewShortTermScorer()

	// Create daily prices with a drawdown
	dailyPrices := make([]float64, 100)
	basePrice := 100.0
	for i := range dailyPrices {
		if i < 50 {
			// Rising prices
			dailyPrices[i] = basePrice + float64(i)*0.5
		} else {
			// Drawdown - prices drop
			dailyPrices[i] = basePrice + 25.0 - float64(i-50)*0.3
		}
	}

	// Calculate max drawdown (negative value, e.g., -0.15 for 15% drawdown)
	maxDrawdown := -0.15

	result := scorer.Calculate(dailyPrices, &maxDrawdown)

	// Verify raw drawdown is stored
	assert.Contains(t, result.Components, "drawdown_raw", "Components should contain drawdown_raw")

	// Verify scored drawdown is still present
	assert.Contains(t, result.Components, "drawdown", "Components should contain scored drawdown")

	// Verify raw value is the actual drawdown percentage (negative value)
	rawDrawdown := result.Components["drawdown_raw"]
	assert.Equal(t, -0.15, rawDrawdown, "Raw drawdown should match input value")

	// Verify scored value is between 0 and 1
	scoredDrawdown := result.Components["drawdown"]
	assert.GreaterOrEqual(t, scoredDrawdown, 0.0, "Scored drawdown should be >= 0")
	assert.LessOrEqual(t, scoredDrawdown, 1.0, "Scored drawdown should be <= 1")
}

func TestShortTermScorer_StoresRawDrawdown_WithNilDrawdown(t *testing.T) {
	scorer := NewShortTermScorer()

	dailyPrices := make([]float64, 100)
	for i := range dailyPrices {
		dailyPrices[i] = 100.0 + float64(i)*0.1
	}

	result := scorer.Calculate(dailyPrices, nil)

	// When drawdown is nil, drawdown_raw should be 0.0
	if rawDrawdown, exists := result.Components["drawdown_raw"]; exists {
		assert.Equal(t, 0.0, rawDrawdown, "Raw drawdown should be 0.0 when nil")
	}
}
