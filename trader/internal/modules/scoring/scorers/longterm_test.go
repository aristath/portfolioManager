package scorers

import (
	"testing"

	"github.com/aristath/arduino-trader/pkg/formulas"
	"github.com/stretchr/testify/assert"
)

func TestLongTermScorer_StoresRawSharpe(t *testing.T) {
	scorer := NewLongTermScorer()

	// Create daily prices that will produce a Sharpe ratio
	dailyPrices := make([]float64, 100)
	basePrice := 100.0
	for i := range dailyPrices {
		// Simulate upward trend with some volatility
		dailyPrices[i] = basePrice + float64(i)*0.1 + float64(i%10)*0.5
	}

	// Create monthly prices for CAGR calculation
	monthlyPrices := make([]formulas.MonthlyPrice, 60)
	for i := range monthlyPrices {
		monthlyPrices[i] = formulas.MonthlyPrice{
			YearMonth:   "2019-01",
			AvgAdjClose: 100.0 + float64(i)*0.5,
		}
	}

	result := scorer.Calculate(monthlyPrices, dailyPrices, nil, 0.11)

	// Verify raw Sharpe ratio is stored
	assert.Contains(t, result.Components, "sharpe_raw", "Components should contain sharpe_raw")

	// Verify scored Sharpe is still present
	assert.Contains(t, result.Components, "sharpe", "Components should contain scored sharpe")

	// Verify raw value is a reasonable Sharpe ratio (typically -2 to 5)
	rawSharpe := result.Components["sharpe_raw"]
	assert.Greater(t, rawSharpe, -5.0, "Raw Sharpe should be reasonable")
	assert.Less(t, rawSharpe, 10.0, "Raw Sharpe should be reasonable")

	// Verify scored value is between 0 and 1
	scoredSharpe := result.Components["sharpe"]
	assert.GreaterOrEqual(t, scoredSharpe, 0.0, "Scored Sharpe should be >= 0")
	assert.LessOrEqual(t, scoredSharpe, 1.0, "Scored Sharpe should be <= 1")
}

func TestLongTermScorer_StoresRawSharpe_WithNilSharpe(t *testing.T) {
	scorer := NewLongTermScorer()

	// Create insufficient daily prices (less than 50)
	dailyPrices := make([]float64, 20)
	for i := range dailyPrices {
		dailyPrices[i] = 100.0 + float64(i)*0.1
	}

	monthlyPrices := make([]formulas.MonthlyPrice, 60)
	for i := range monthlyPrices {
		monthlyPrices[i] = formulas.MonthlyPrice{
			YearMonth:   "2019-01",
			AvgAdjClose: 100.0 + float64(i)*0.5,
		}
	}

	result := scorer.Calculate(monthlyPrices, dailyPrices, nil, 0.11)

	// When Sharpe is nil, sharpe_raw should be 0 or not present
	// (implementation choice - we'll store 0.0 for nil)
	if rawSharpe, exists := result.Components["sharpe_raw"]; exists {
		assert.Equal(t, 0.0, rawSharpe, "Raw Sharpe should be 0.0 when calculation fails")
	}
}
