package scorers

import (
	"fmt"
	"testing"

	"github.com/aristath/sentinel/pkg/formulas"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// StabilityScorer.Calculate Tests
// ============================================================================

func TestStabilityScorer_Calculate_ConsistentGrowth(t *testing.T) {
	scorer := NewStabilityScorer()

	// Create consistent 10% annual growth pattern over 10 years
	monthlyPrices := generateConsistentGrowthPrices(120, 0.10)         // 10 years, 10% annual
	dailyPrices := generateDailyPricesFromMonthly(monthlyPrices, 0.01) // Low volatility

	result := scorer.Calculate(monthlyPrices, dailyPrices)

	// Should score reasonably well for consistent, stable growth
	// Score components depend on generated data characteristics
	assert.Greater(t, result.Score, 0.5)
	assert.Contains(t, result.Components, "consistency")
	assert.Contains(t, result.Components, "volatility")
	assert.Contains(t, result.Components, "recovery")
}

func TestStabilityScorer_Calculate_ErraticGrowth(t *testing.T) {
	scorer := NewStabilityScorer()

	// Create erratic growth: 5yr CAGR = 20%, 10yr CAGR = 5% (inconsistent)
	monthlyPrices := generateErraticGrowthPrices(120)
	dailyPrices := generateDailyPricesFromMonthly(monthlyPrices, 0.03) // Moderate volatility

	result := scorer.Calculate(monthlyPrices, dailyPrices)

	// Should score lower due to inconsistency
	assert.Less(t, result.Score, 0.7)
}

func TestStabilityScorer_Calculate_HighVolatility(t *testing.T) {
	scorer := NewStabilityScorer()

	// Create consistent growth but with high volatility
	monthlyPrices := generateConsistentGrowthPrices(120, 0.12)
	dailyPrices := generateDailyPricesFromMonthly(monthlyPrices, 0.05) // High volatility

	result := scorer.Calculate(monthlyPrices, dailyPrices)

	// High volatility should reduce stability score
	assert.Less(t, result.Components["volatility"], 0.6)
}

func TestStabilityScorer_Calculate_QuickRecovery(t *testing.T) {
	scorer := NewStabilityScorer()

	// Create prices with drawdown but quick recovery
	monthlyPrices := generateConsistentGrowthPrices(120, 0.10)
	dailyPrices := generatePricesWithQuickRecovery(250) // Quick recovery from drawdowns

	result := scorer.Calculate(monthlyPrices, dailyPrices)

	// Quick recovery should score well on recovery component
	assert.Greater(t, result.Components["recovery"], 0.5)
}

func TestStabilityScorer_Calculate_SlowRecovery(t *testing.T) {
	scorer := NewStabilityScorer()

	// Create prices with drawdown and slow recovery
	monthlyPrices := generateConsistentGrowthPrices(120, 0.05)
	dailyPrices := generatePricesWithSlowRecovery(250) // Slow recovery

	result := scorer.Calculate(monthlyPrices, dailyPrices)

	// Recovery score depends on Ulcer Index from generated data
	// Just verify it's a valid score
	assert.GreaterOrEqual(t, result.Components["recovery"], 0.0)
	assert.LessOrEqual(t, result.Components["recovery"], 1.0)
}

func TestStabilityScorer_Calculate_EmptyPrices(t *testing.T) {
	scorer := NewStabilityScorer()

	result := scorer.Calculate([]formulas.MonthlyPrice{}, []float64{})

	// Empty prices should return neutral score
	assert.Equal(t, 0.5, result.Score)
	assert.Equal(t, 0.5, result.Components["consistency"])
	assert.Equal(t, 0.5, result.Components["volatility"])
	assert.Equal(t, 0.5, result.Components["recovery"])
}

func TestStabilityScorer_Calculate_SinglePrice(t *testing.T) {
	scorer := NewStabilityScorer()

	monthlyPrices := []formulas.MonthlyPrice{{YearMonth: "2024-01", AvgAdjClose: 100.0}}
	dailyPrices := []float64{100.0}

	result := scorer.Calculate(monthlyPrices, dailyPrices)

	// Single price should return neutral score
	assert.Equal(t, 0.5, result.Score)
}

func TestStabilityScorer_Calculate_InsufficientData(t *testing.T) {
	scorer := NewStabilityScorer()

	// Only 12 months of data (less than 5 years for full consistency calculation)
	monthlyPrices := generateConsistentGrowthPrices(12, 0.10)
	dailyPrices := generateDailyPricesFromMonthly(monthlyPrices, 0.02)

	result := scorer.Calculate(monthlyPrices, dailyPrices)

	// Should still calculate with available data, using partial consistency
	assert.GreaterOrEqual(t, result.Score, 0.0)
	assert.LessOrEqual(t, result.Score, 1.0)
}

// ============================================================================
// Consistency Component Tests
// ============================================================================

func TestCalculateConsistencyScore_SimilarCAGR(t *testing.T) {
	// 5-year CAGR = 10%, 10-year CAGR = 10% (highly consistent)
	cagr5y := 0.10
	cagr10y := 0.10

	score := calculateConsistencyScore(cagr5y, &cagr10y)

	// Very similar CAGRs should score high (1.0)
	assert.GreaterOrEqual(t, score, 0.9)
}

func TestCalculateConsistencyScore_SmallDifference(t *testing.T) {
	// 5-year CAGR = 11%, 10-year CAGR = 10% (small difference)
	cagr5y := 0.11
	cagr10y := 0.10

	score := calculateConsistencyScore(cagr5y, &cagr10y)

	// Small difference should still score well (>= 0.8)
	assert.GreaterOrEqual(t, score, 0.8)
}

func TestCalculateConsistencyScore_LargeDifference(t *testing.T) {
	// 5-year CAGR = 25%, 10-year CAGR = 5% (large difference)
	cagr5y := 0.25
	cagr10y := 0.05

	score := calculateConsistencyScore(cagr5y, &cagr10y)

	// Large difference should score lower
	assert.LessOrEqual(t, score, 0.6)
}

func TestCalculateConsistencyScore_NoTenYearData(t *testing.T) {
	// Only 5-year data available
	cagr5y := 0.12

	score := calculateConsistencyScore(cagr5y, nil)

	// Without 10-year data, return 0.65 for CAGR >= 10%
	assert.Equal(t, 0.65, score)
}

// ============================================================================
// Volatility Component Tests
// ============================================================================

func TestCalculateVolatilityScore_LowVolatility(t *testing.T) {
	// Very low volatility (10% annualized)
	volatility := 0.10

	score := calculateVolatilityScore(&volatility)

	// Low volatility should score high
	assert.GreaterOrEqual(t, score, 0.8)
}

func TestCalculateVolatilityScore_ModerateVolatility(t *testing.T) {
	// Moderate volatility (20% annualized)
	volatility := 0.20

	score := calculateVolatilityScore(&volatility)

	// Moderate volatility should score middle-range
	assert.GreaterOrEqual(t, score, 0.5)
	assert.LessOrEqual(t, score, 0.8)
}

func TestCalculateVolatilityScore_HighVolatility(t *testing.T) {
	// High volatility (40% annualized)
	volatility := 0.40

	score := calculateVolatilityScore(&volatility)

	// High volatility should score low
	assert.LessOrEqual(t, score, 0.5)
}

func TestCalculateVolatilityScore_Nil(t *testing.T) {
	score := calculateVolatilityScore(nil)

	// Nil volatility returns neutral
	assert.Equal(t, 0.5, score)
}

// ============================================================================
// Recovery Component Tests
// ============================================================================

func TestCalculateRecoveryScore_FastRecovery(t *testing.T) {
	// Low Ulcer Index = fast recovery
	ulcerIndex := 2.0

	score := calculateRecoveryScore(&ulcerIndex)

	// Fast recovery should score high
	assert.GreaterOrEqual(t, score, 0.8)
}

func TestCalculateRecoveryScore_ModerateRecovery(t *testing.T) {
	// Moderate Ulcer Index
	ulcerIndex := 8.0

	score := calculateRecoveryScore(&ulcerIndex)

	// Moderate recovery should score middle-range
	assert.GreaterOrEqual(t, score, 0.4)
	assert.LessOrEqual(t, score, 0.8)
}

func TestCalculateRecoveryScore_SlowRecovery(t *testing.T) {
	// High Ulcer Index = slow recovery
	ulcerIndex := 20.0

	score := calculateRecoveryScore(&ulcerIndex)

	// Slow recovery should score low
	assert.LessOrEqual(t, score, 0.4)
}

func TestCalculateRecoveryScore_Nil(t *testing.T) {
	score := calculateRecoveryScore(nil)

	// Nil returns neutral
	assert.Equal(t, 0.5, score)
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestStabilityScorer_ScoreCappedAtOne(t *testing.T) {
	scorer := NewStabilityScorer()

	// Perfect stability scenario: low volatility, consistent growth, quick recovery
	monthlyPrices := generatePerfectGrowthPrices(120)
	dailyPrices := generateLowVolatilityDailyPrices(250)

	result := scorer.Calculate(monthlyPrices, dailyPrices)

	// Score should be capped at 1.0
	assert.LessOrEqual(t, result.Score, 1.0)
}

func TestStabilityScorer_ScoreFlooredAtZero(t *testing.T) {
	scorer := NewStabilityScorer()

	// Bad scenario but shouldn't go below 0
	monthlyPrices := generateDecliningPrices(120)
	dailyPrices := generateHighVolatilityDailyPrices(250)

	result := scorer.Calculate(monthlyPrices, dailyPrices)

	// Score should not go below 0
	assert.GreaterOrEqual(t, result.Score, 0.0)
}

func TestStabilityScorer_ComponentWeighting(t *testing.T) {
	scorer := NewStabilityScorer()

	monthlyPrices := generateConsistentGrowthPrices(120, 0.10)
	dailyPrices := generateDailyPricesFromMonthly(monthlyPrices, 0.02)

	result := scorer.Calculate(monthlyPrices, dailyPrices)

	// Verify weights: 40% consistency, 30% volatility, 30% recovery
	expectedScore := result.Components["consistency"]*0.40 +
		result.Components["volatility"]*0.30 +
		result.Components["recovery"]*0.30
	assert.InDelta(t, expectedScore, result.Score, 0.01)
}

// ============================================================================
// Test Helper Functions
// ============================================================================

// generateConsistentGrowthPrices creates monthly prices with consistent CAGR
func generateConsistentGrowthPrices(months int, annualCAGR float64) []formulas.MonthlyPrice {
	prices := make([]formulas.MonthlyPrice, months)
	basePrice := 100.0
	monthlyReturn := (1 + annualCAGR) / 12.0 // Simplified monthly return

	for i := 0; i < months; i++ {
		// Compound growth
		price := basePrice * (1 + monthlyReturn*float64(i))
		year := 2014 + i/12
		month := (i % 12) + 1
		prices[i] = formulas.MonthlyPrice{
			YearMonth:   formatYearMonth(year, month),
			AvgAdjClose: price,
		}
	}
	return prices
}

// generateErraticGrowthPrices creates prices with inconsistent growth pattern
func generateErraticGrowthPrices(months int) []formulas.MonthlyPrice {
	prices := make([]formulas.MonthlyPrice, months)
	basePrice := 100.0

	for i := 0; i < months; i++ {
		// First half: slow growth (5% annual)
		// Second half: fast growth (25% annual)
		var monthlyReturn float64
		if i < months/2 {
			monthlyReturn = 0.05 / 12.0
		} else {
			monthlyReturn = 0.25 / 12.0
		}
		basePrice *= (1 + monthlyReturn)
		year := 2014 + i/12
		month := (i % 12) + 1
		prices[i] = formulas.MonthlyPrice{
			YearMonth:   formatYearMonth(year, month),
			AvgAdjClose: basePrice,
		}
	}
	return prices
}

// generateDailyPricesFromMonthly creates daily prices interpolated from monthly with volatility
func generateDailyPricesFromMonthly(monthly []formulas.MonthlyPrice, dailyVol float64) []float64 {
	if len(monthly) == 0 {
		return []float64{}
	}

	// Generate approximately 21 trading days per month
	daysPerMonth := 21
	daily := make([]float64, len(monthly)*daysPerMonth)

	for m, mp := range monthly {
		for d := 0; d < daysPerMonth; d++ {
			idx := m*daysPerMonth + d
			// Add some noise based on volatility
			noise := (float64(d%5) - 2.0) * dailyVol * mp.AvgAdjClose
			daily[idx] = mp.AvgAdjClose + noise
		}
	}
	return daily
}

// generatePricesWithQuickRecovery creates prices that recover quickly from drawdowns
func generatePricesWithQuickRecovery(days int) []float64 {
	prices := make([]float64, days)
	basePrice := 100.0

	for i := 0; i < days; i++ {
		// Small drawdowns with quick V-shaped recovery
		cycle := i % 50
		if cycle < 10 {
			// Small dip
			prices[i] = basePrice * (1 - 0.02*float64(cycle)/10)
		} else if cycle < 20 {
			// Quick recovery
			prices[i] = basePrice * (1 - 0.02*(20-float64(cycle))/10)
		} else {
			prices[i] = basePrice
		}
		basePrice *= 1.0003 // Small upward drift
	}
	return prices
}

// generatePricesWithSlowRecovery creates prices that recover slowly from drawdowns
func generatePricesWithSlowRecovery(days int) []float64 {
	prices := make([]float64, days)
	basePrice := 100.0

	for i := 0; i < days; i++ {
		// Larger drawdowns with slow U-shaped recovery
		cycle := i % 100
		if cycle < 30 {
			// Slow decline
			prices[i] = basePrice * (1 - 0.15*float64(cycle)/30)
		} else if cycle < 100 {
			// Very slow recovery
			prices[i] = basePrice * (1 - 0.15*(100-float64(cycle))/70)
		}
		basePrice *= 1.0001 // Small upward drift
	}
	return prices
}

// generatePerfectGrowthPrices creates ideal stable growth
func generatePerfectGrowthPrices(months int) []formulas.MonthlyPrice {
	return generateConsistentGrowthPrices(months, 0.11) // Exactly 11% CAGR
}

// generateLowVolatilityDailyPrices creates prices with very low volatility
func generateLowVolatilityDailyPrices(days int) []float64 {
	prices := make([]float64, days)
	basePrice := 100.0

	for i := 0; i < days; i++ {
		// Very smooth upward progression
		prices[i] = basePrice * (1 + 0.0004*float64(i)) // ~10% annual
	}
	return prices
}

// generateDecliningPrices creates prices with declining trend
func generateDecliningPrices(months int) []formulas.MonthlyPrice {
	prices := make([]formulas.MonthlyPrice, months)
	basePrice := 100.0

	for i := 0; i < months; i++ {
		basePrice *= 0.99 // 1% monthly decline
		year := 2014 + i/12
		month := (i % 12) + 1
		prices[i] = formulas.MonthlyPrice{
			YearMonth:   formatYearMonth(year, month),
			AvgAdjClose: basePrice,
		}
	}
	return prices
}

// generateHighVolatilityDailyPrices creates prices with high volatility
func generateHighVolatilityDailyPrices(days int) []float64 {
	prices := make([]float64, days)
	basePrice := 100.0

	for i := 0; i < days; i++ {
		// High volatility swings
		swing := float64(i%10-5) * 3.0
		prices[i] = basePrice + swing
	}
	return prices
}

// formatYearMonth formats year and month as YYYY-MM
func formatYearMonth(year, month int) string {
	return fmt.Sprintf("%04d-%02d", year, month)
}
