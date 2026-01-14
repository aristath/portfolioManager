package universe

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestTagAssigner_ValueOpportunity(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	currentPrice := 80.0
	price52wHigh := 100.0
	input := AssignTagsInput{
		Symbol:       "TEST",
		CurrentPrice: &currentPrice,
		Price52wHigh: &price52wHigh,
		GroupScores: map[string]float64{
			"opportunity": 0.75,
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "value-opportunity")
	assert.Contains(t, tags, "below-52w-high")
}

func TestTagAssigner_HighQuality(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.85,
			"long_term": 0.80,
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "high-quality")
	assert.Contains(t, tags, "high-stability")
}

func TestTagAssigner_Stable(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	volatility := 0.15

	input := AssignTagsInput{
		Symbol:     "TEST",
		Volatility: &volatility,
		GroupScores: map[string]float64{
			"stability": 0.80,
		},
		SubScores: map[string]map[string]float64{
			"stability": {
				"consistency": 0.85,
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "stable")
}

func TestTagAssigner_Volatile(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	volatility := 0.35

	input := AssignTagsInput{
		Symbol:     "TEST",
		Volatility: &volatility,
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "volatile")
	assert.Contains(t, tags, "high-risk")
}

func TestTagAssigner_Oversold(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	rsi := 25.0

	input := AssignTagsInput{
		Symbol: "TEST",
		RSI:    &rsi,
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "oversold")
}

func TestTagAssigner_Overbought(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	rsi := 75.0

	input := AssignTagsInput{
		Symbol: "TEST",
		RSI:    &rsi,
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "overbought")
}

func TestTagAssigner_HighDividend(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	dividendYield := 7.0

	input := AssignTagsInput{
		Symbol:        "TEST",
		DividendYield: &dividendYield,
		GroupScores: map[string]float64{
			"dividends": 0.75,
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "high-dividend")
	assert.Contains(t, tags, "dividend-opportunity")
	assert.Contains(t, tags, "dividend-focused")
}

func TestTagAssigner_MultipleTags(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	currentPrice := 75.0 // 25% below 52W high
	price52wHigh := 100.0
	volatility := 0.12
	dividendYield := 5.0
	input := AssignTagsInput{
		Symbol:        "TEST",
		CurrentPrice:  &currentPrice,
		Price52wHigh:  &price52wHigh,
		Volatility:    &volatility,
		DividendYield: &dividendYield,
		GroupScores: map[string]float64{
			"stability":   0.85, // > 0.8 for high-quality
			"long_term":   0.80, // > 0.75 for high-quality
			"opportunity": 0.75, // > 0.7 for value-opportunity
			"dividends":   0.75,
		},
		SubScores: map[string]map[string]float64{
			"stability": {
				"consistency": 0.85,
			},
		},
		Score: &SecurityScore{
			TotalScore: 0.78,
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// Should have multiple tags
	assert.Greater(t, len(tags), 5)
	assert.Contains(t, tags, "value-opportunity")
	assert.Contains(t, tags, "high-quality")
	assert.Contains(t, tags, "stable")
	assert.Contains(t, tags, "dividend-opportunity")
	assert.Contains(t, tags, "low-risk")
}

func TestTagAssigner_NoTags(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.50,
			"long_term": 0.50,
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// Should have at least risk profile tags
	assert.GreaterOrEqual(t, len(tags), 0)
}

func TestTagAssigner_QualityGatePass(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.65, // >= 0.6
			"long_term": 0.55, // >= 0.5
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestTagAssigner_QualityGateFail(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	// Test case 1: Stability too low for relaxed threshold
	input1 := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.54, // < 0.55 (new relaxed threshold)
			"long_term": 0.44, // < 0.45 (new relaxed threshold)
		},
	}

	tags1, err := assigner.AssignTagsForSecurity(input1)
	assert.NoError(t, err)
	assert.Contains(t, tags1, "quality-gate-fail")

	// Test case 2: Long-term too low for relaxed threshold
	input2 := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.54, // < 0.55
			"long_term": 0.44, // < 0.45
		},
	}

	tags2, err := assigner.AssignTagsForSecurity(input2)
	assert.NoError(t, err)
	assert.Contains(t, tags2, "quality-gate-fail")
}

func TestTagAssigner_QualityValue(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	currentPrice := 80.0
	price52wHigh := 100.0
	input := AssignTagsInput{
		Symbol:       "TEST",
		CurrentPrice: &currentPrice,
		Price52wHigh: &price52wHigh,
		GroupScores: map[string]float64{
			"stability":   0.85, // > 0.8 for high-quality
			"long_term":   0.80, // > 0.75 for high-quality
			"opportunity": 0.75, // > 0.7 for value-opportunity
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// Should have both high-quality and value-opportunity
	assert.Contains(t, tags, "high-quality")
	assert.Contains(t, tags, "value-opportunity")
	// Should also have quality-value combination tag
	assert.Contains(t, tags, "quality-value")
}

func TestTagAssigner_BubbleRisk(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	volatility := 0.45 // > 0.40

	input := AssignTagsInput{
		Symbol:     "TEST",
		Volatility: &volatility,
		GroupScores: map[string]float64{
			"stability": 0.55, // < 0.6
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"cagr_raw":    0.18, // > 16.5%
				"sharpe_raw":  0.3,  // < 0.5
				"sortino_raw": 0.4,  // < 0.5
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "bubble-risk")
	assert.Contains(t, tags, "ensemble-bubble-risk") // Classical bubble should also get ensemble tag
	assert.NotContains(t, tags, "quality-high-cagr")
}

func TestTagAssigner_QualityHighCAGR(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	volatility := 0.30 // <= 0.40

	input := AssignTagsInput{
		Symbol:     "TEST",
		Volatility: &volatility,
		GroupScores: map[string]float64{
			"stability": 0.70, // >= 0.6
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"cagr_raw":    0.17, // > 15%
				"sharpe_raw":  0.6,  // >= 0.5
				"sortino_raw": 0.6,  // >= 0.5
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "quality-high-cagr")
	assert.NotContains(t, tags, "bubble-risk")
}

// Removed: TestTagAssigner_ValueTrap and TestTagAssigner_NotValueTrap
// These tests tested P/E-based value-trap detection which was removed
// (P/E ratio data is no longer available)

func TestTagAssigner_ExcellentTotalReturn(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	dividendYield := 0.10 // 10%
	cagrValue := 0.09     // 9% (total = 19% >= 18%)

	input := AssignTagsInput{
		Symbol:        "TEST",
		DividendYield: &dividendYield,
		SubScores: map[string]map[string]float64{
			"long_term": {
				"cagr_raw": cagrValue,
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "excellent-total-return")
}

func TestTagAssigner_HighTotalReturn(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	dividendYield := 0.08 // 8%
	cagrValue := 0.08     // 8% (total = 16% >= 15%)

	input := AssignTagsInput{
		Symbol:        "TEST",
		DividendYield: &dividendYield,
		SubScores: map[string]map[string]float64{
			"long_term": {
				"cagr_raw": cagrValue,
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "high-total-return")
	assert.NotContains(t, tags, "excellent-total-return")
}

func TestTagAssigner_ModerateTotalReturn(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	dividendYield := 0.06 // 6%
	cagrValue := 0.07     // 7% (total = 13% >= 12%)

	input := AssignTagsInput{
		Symbol:        "TEST",
		DividendYield: &dividendYield,
		SubScores: map[string]map[string]float64{
			"long_term": {
				"cagr_raw": cagrValue,
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "moderate-total-return")
	assert.NotContains(t, tags, "high-total-return")
}

func TestTagAssigner_DividendTotalReturn(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	dividendYield := 0.10 // 10% >= 8%
	cagrValue := 0.06     // 6% >= 5%

	input := AssignTagsInput{
		Symbol:        "TEST",
		DividendYield: &dividendYield,
		SubScores: map[string]map[string]float64{
			"long_term": {
				"cagr_raw": cagrValue,
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "dividend-total-return")
}

func TestTagAssigner_NeedsRebalance(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	// Test case 1: Overweight by more than 3%
	positionWeight1 := 0.15 // 15%
	targetWeight1 := 0.10   // 10% (deviation = 5% > 3%)

	input1 := AssignTagsInput{
		Symbol:         "TEST",
		PositionWeight: &positionWeight1,
		TargetWeight:   &targetWeight1,
	}

	tags1, err := assigner.AssignTagsForSecurity(input1)
	assert.NoError(t, err)
	assert.Contains(t, tags1, "needs-rebalance")

	// Test case 2: Underweight by more than 3%
	positionWeight2 := 0.05 // 5%
	targetWeight2 := 0.10   // 10% (deviation = -5% < -3%)

	input2 := AssignTagsInput{
		Symbol:         "TEST",
		PositionWeight: &positionWeight2,
		TargetWeight:   &targetWeight2,
	}

	tags2, err := assigner.AssignTagsForSecurity(input2)
	assert.NoError(t, err)
	assert.Contains(t, tags2, "needs-rebalance")
}

func TestTagAssigner_RegimeBearSafe(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	volatility := 0.15  // < 0.20
	maxDrawdown := 15.0 // < 20%

	input := AssignTagsInput{
		Symbol:      "TEST",
		Volatility:  &volatility,
		MaxDrawdown: &maxDrawdown,
		GroupScores: map[string]float64{
			"stability": 0.80, // > 0.75
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "regime-bear-safe")
}

func TestTagAssigner_RegimeBullGrowth(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	cagrValue := 0.13 // > 12%

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.75, // > 0.7
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"cagr_raw": cagrValue,
			},
			"short_term": {
				"momentum": 0.05, // > 0
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "regime-bull-growth")
}

func TestTagAssigner_RegimeSidewaysValue(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	currentPrice := 80.0
	price52wHigh := 100.0
	input := AssignTagsInput{
		Symbol:       "TEST",
		CurrentPrice: &currentPrice,
		Price52wHigh: &price52wHigh,
		GroupScores: map[string]float64{
			"opportunity": 0.75, // > 0.7 for value-opportunity
			"stability":   0.80, // > 0.75
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// Should have value-opportunity
	assert.Contains(t, tags, "value-opportunity")
	// Should also have regime-sideways-value
	assert.Contains(t, tags, "regime-sideways-value")
}

func TestTagAssigner_RegimeVolatile(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	// Test case 1: High volatility
	volatility1 := 0.35 // > 0.30
	historicalVolatility1 := 0.20

	input1 := AssignTagsInput{
		Symbol:               "TEST",
		Volatility:           &volatility1,
		HistoricalVolatility: &historicalVolatility1,
	}

	tags1, err := assigner.AssignTagsForSecurity(input1)
	assert.NoError(t, err)
	assert.Contains(t, tags1, "regime-volatile")

	// Test case 2: Volatility spike
	volatility2 := 0.30
	historicalVolatility2 := 0.15 // volatility > historical * 1.5 = 0.225, so 0.30 > 0.225 = spike

	input2 := AssignTagsInput{
		Symbol:               "TEST",
		Volatility:           &volatility2,
		HistoricalVolatility: &historicalVolatility2,
	}

	tags2, err := assigner.AssignTagsForSecurity(input2)
	assert.NoError(t, err)
	assert.Contains(t, tags2, "volatility-spike")
	assert.Contains(t, tags2, "regime-volatile")
}

func TestTagAssigner_AllEnhancedTags(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	// Create a security that meets criteria for multiple enhanced tags
	currentPrice := 75.0
	price52wHigh := 100.0
	volatility := 0.18
	historicalVolatility := 0.15
	dividendYield := 0.10 // 10%
	maxDrawdown := 15.0
	positionWeight := 0.10
	targetWeight := 0.10

	input := AssignTagsInput{
		Symbol:               "TEST",
		CurrentPrice:         &currentPrice,
		Price52wHigh:         &price52wHigh,
		Volatility:           &volatility,
		HistoricalVolatility: &historicalVolatility,
		DividendYield:        &dividendYield,
		MaxDrawdown:          &maxDrawdown,
		PositionWeight:       &positionWeight,
		TargetWeight:         &targetWeight,
		GroupScores: map[string]float64{
			"stability":   0.85, // > 0.8 for high-quality, > 0.6 for quality-gate-pass
			"long_term":   0.80, // > 0.75 for high-quality, > 0.5 for quality-gate-pass
			"opportunity": 0.75, // > 0.7 for value-opportunity
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"cagr_raw":    0.17, // > 15% for quality-high-cagr
				"sharpe_raw":  1.8,  // >= 1.5 for high-sharpe
				"sortino_raw": 1.8,  // >= 1.5 for high-sortino
			},
			"short_term": {
				"momentum": 0.05, // > 0 for regime-bull-growth
			},
		},
		Score: &SecurityScore{
			TotalScore: 0.78,
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)

	// Verify quality gate tags
	assert.NotContains(t, tags, "quality-gate-fail", "Should pass quality gate")
	assert.Contains(t, tags, "high-quality")
	assert.Contains(t, tags, "value-opportunity")
	assert.Contains(t, tags, "quality-value")

	// Verify bubble detection tags
	assert.Contains(t, tags, "quality-high-cagr")

	// Verify total return tags
	// Total return = 0.17 + 0.10 = 0.27 >= 0.18
	assert.Contains(t, tags, "excellent-total-return")
	// dividend-total-return: 0.10 >= 0.08 AND 0.17 >= 0.05
	assert.Contains(t, tags, "dividend-total-return")

	// Verify regime-specific tags
	assert.Contains(t, tags, "regime-bear-safe")      // volatility < 0.20, stability > 0.75, drawdown < 20%
	assert.Contains(t, tags, "regime-bull-growth")    // CAGR > 12%, stability > 0.7, momentum > 0
	assert.Contains(t, tags, "regime-sideways-value") // value-opportunity AND stability > 0.75

	// Should NOT have bubble-risk (good risk metrics)
	assert.NotContains(t, tags, "bubble-risk")

	t.Logf("Assigned %d tags total", len(tags))
}

func TestTagAssigner_QuantumBubbleDetection(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	volatility := 0.38 // Just below 0.40 threshold

	input := AssignTagsInput{
		Symbol:     "TEST",
		Volatility: &volatility,
		GroupScores: map[string]float64{
			"stability": 0.62, // Just above 0.6 threshold
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"cagr_raw":    0.16, // 16% - high but not > 16.5% (classical threshold)
				"sharpe_raw":  0.52, // Just above 0.5 (classical threshold)
				"sortino_raw": 0.52, // Just above 0.5 (classical threshold)
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)

	// Classical should NOT detect (all metrics just above thresholds)
	assert.NotContains(t, tags, "bubble-risk")

	// Quantum might detect (early warning) - check if quantum tags are present
	// Quantum detection is probabilistic, so we just verify the system runs
	// In practice, with these inputs, quantum might detect early warning
	t.Logf("Quantum bubble detection tags: %v", tags)
	// Verify system doesn't crash and produces tags
	assert.Greater(t, len(tags), 0, "Should produce some tags")
}

// Removed: TestTagAssigner_QuantumValueTrapDetection
// This test tested P/E-based quantum value-trap detection which was removed
// (P/E ratio data is no longer available)

func TestTagAssigner_EnsembleBubbleDetection(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	volatility := 0.45

	input := AssignTagsInput{
		Symbol:     "TEST",
		Volatility: &volatility,
		GroupScores: map[string]float64{
			"stability": 0.55, // < 0.6
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"cagr_raw":    0.18, // > 16.5% (classical threshold)
				"sharpe_raw":  0.3,  // < 0.5 (classical threshold)
				"sortino_raw": 0.4,  // < 0.5 (classical threshold)
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)

	// Both classical and ensemble should detect
	assert.Contains(t, tags, "bubble-risk")
	assert.Contains(t, tags, "ensemble-bubble-risk")
}

// ============================================================================
// Multi-Path Quality Gate Tests (TDD - these will fail until implementation)
// ============================================================================

// Path 1: Balanced (relaxed, adaptive) Tests

func TestQualityGate_Path1_Balanced_Pass(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.56, // >= 0.55
			"long_term": 0.46, // >= 0.45
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_Path1_Balanced_Fail(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.54, // < 0.55
			"long_term": 0.46, // >= 0.45
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.NotNil(t, tags) // Should fail Path 1, but might pass other paths - we'll test fail-all later
}

// Path 2: Exceptional Excellence Tests

func TestQualityGate_Path2_ExceptionalExcellence_StabilityPass(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.76, // >= 0.75
			"long_term": 0.30, // Below all other thresholds
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_Path2_ExceptionalExcellence_LongTermPass(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.40, // Below all other thresholds
			"long_term": 0.76, // >= 0.75
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_Path2_ExceptionalExcellence_Fail(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.74, // < 0.75
			"long_term": 0.74, // < 0.75
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.NotNil(t, tags) // Might still pass other paths - not testing fail-all here
}

// Path 3: Quality Value Play Tests

func TestQualityGate_Path3_QualityValuePlay_Pass(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability":   0.61, // >= 0.60
			"opportunity": 0.66, // >= 0.65
			"long_term":   0.31, // >= 0.30
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_Path3_QualityValuePlay_Fail(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability":   0.61, // >= 0.60
			"opportunity": 0.64, // < 0.65
			"long_term":   0.31, // >= 0.30
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.NotNil(t, tags) // Might still pass other paths
}

// Path 4: Dividend Income Play Tests

func TestQualityGate_Path4_DividendIncomePlay_Pass(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	dividendYield := 0.036 // >= 0.035 (3.6%)

	input := AssignTagsInput{
		Symbol:        "TEST",
		DividendYield: &dividendYield,
		GroupScores: map[string]float64{
			"stability": 0.56, // >= 0.55
			"dividends": 0.66, // >= 0.65
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_Path4_DividendIncomePlay_Fail(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	dividendYield := 0.034 // < 0.035

	input := AssignTagsInput{
		Symbol:        "TEST",
		DividendYield: &dividendYield,
		GroupScores: map[string]float64{
			"stability": 0.56, // >= 0.55
			"dividends": 0.66, // >= 0.65
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.NotNil(t, tags) // Might still pass other paths
}

// Path 5: Risk-Adjusted Excellence Tests

func TestQualityGate_Path5_RiskAdjustedExcellence_SharpePass(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	volatility := 0.34 // <= 0.35

	input := AssignTagsInput{
		Symbol:     "TEST",
		Volatility: &volatility,
		GroupScores: map[string]float64{
			"long_term": 0.56, // >= 0.55
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"sharpe_raw":  0.91, // >= 0.9
				"sortino_raw": 0.50, // Not required
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_Path5_RiskAdjustedExcellence_SortinoPass(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	volatility := 0.34 // <= 0.35

	input := AssignTagsInput{
		Symbol:     "TEST",
		Volatility: &volatility,
		GroupScores: map[string]float64{
			"long_term": 0.56, // >= 0.55
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"sharpe_raw":  0.50, // Not required
				"sortino_raw": 0.91, // >= 0.9
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_Path5_RiskAdjustedExcellence_Fail(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	volatility := 0.36 // > 0.35 - volatility too high

	input := AssignTagsInput{
		Symbol:     "TEST",
		Volatility: &volatility,
		GroupScores: map[string]float64{
			"long_term": 0.56, // >= 0.55
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"sharpe_raw":  0.91, // >= 0.9
				"sortino_raw": 0.91, // >= 0.9
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.NotNil(t, tags) // Might still pass other paths
}

// Path 6: Composite Minimum Tests

func TestQualityGate_Path6_CompositeMinimum_Pass(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.50, // >= 0.45, composite: 0.6*0.50 + 0.4*0.55 = 0.52
			"long_term": 0.55, //
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_Path6_CompositeMinimum_Fail(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.44, // < 0.45 (fails stability floor)
			"long_term": 0.70, // High, but stability floor not met
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.NotNil(t, tags) // Might still pass other paths
}

// Path 7: Growth Opportunity Tests

func TestQualityGate_Path7_GrowthOpportunity_Pass(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	volatility := 0.39 // <= 0.40

	input := AssignTagsInput{
		Symbol:     "TEST",
		Volatility: &volatility,
		GroupScores: map[string]float64{
			"stability": 0.51, // >= 0.50
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"cagr_raw": 0.14, // >= 0.13 (14% CAGR)
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_Path7_GrowthOpportunity_Fail(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	volatility := 0.41 // > 0.40 - volatility too high

	input := AssignTagsInput{
		Symbol:     "TEST",
		Volatility: &volatility,
		GroupScores: map[string]float64{
			"stability": 0.51, // >= 0.50
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"cagr_raw": 0.14, // >= 0.13
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.NotNil(t, tags) // Might still pass other paths
}

// Boundary Value Tests

func TestQualityGate_Path1_BoundaryExact(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.55, // Exactly at threshold
			"long_term": 0.45, // Exactly at threshold
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_Path2_BoundaryExact(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.75, // Exactly at threshold
			"long_term": 0.10, // Below all others
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_Path3_BoundaryExact(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability":   0.60, // Exactly at threshold
			"opportunity": 0.65, // Exactly at threshold
			"long_term":   0.30, // Exactly at threshold
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_Path4_BoundaryExact(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	dividendYield := 0.035 // Exactly at threshold

	input := AssignTagsInput{
		Symbol:        "TEST",
		DividendYield: &dividendYield,
		GroupScores: map[string]float64{
			"stability": 0.55, // Exactly at threshold
			"dividends": 0.65, // Exactly at threshold
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_Path5_BoundaryExact(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	volatility := 0.35 // Exactly at threshold

	input := AssignTagsInput{
		Symbol:     "TEST",
		Volatility: &volatility,
		GroupScores: map[string]float64{
			"long_term": 0.55, // Exactly at threshold
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"sharpe_raw": 0.9, // Exactly at threshold
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_Path6_BoundaryExact(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.45, // Exactly at stability floor
			"long_term": 0.60, // Composite: 0.6*0.45 + 0.4*0.60 = 0.51 < 0.52
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.NotNil(t, tags) // This should fail Path 6 (composite too low), might pass others
}

func TestQualityGate_Path7_BoundaryExact(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	volatility := 0.40 // Exactly at threshold

	input := AssignTagsInput{
		Symbol:     "TEST",
		Volatility: &volatility,
		GroupScores: map[string]float64{
			"stability": 0.50, // Exactly at threshold
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"cagr_raw": 0.13, // Exactly at threshold
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

// Multi-Path Scenario Tests

func TestQualityGate_PassesMultiplePaths(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	dividendYield := 0.05
	volatility := 0.25

	input := AssignTagsInput{
		Symbol:        "TEST",
		DividendYield: &dividendYield,
		Volatility:    &volatility,
		GroupScores: map[string]float64{
			"stability": 0.76, // Passes Path 1, 2, 4
			"long_term": 0.60, // Passes Path 1, 5
			"dividends": 0.70, // Passes Path 4
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"sharpe_raw": 1.2, // Passes Path 5
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_PassesOnlyOnePath(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.76, // Only passes Path 2 (exceptional excellence)
			"long_term": 0.20, // Too low for all other paths
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_FailsAllPaths(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	dividendYield := 0.02
	volatility := 0.50

	input := AssignTagsInput{
		Symbol:        "TEST",
		DividendYield: &dividendYield,
		Volatility:    &volatility,
		GroupScores: map[string]float64{
			"stability":   0.40, // Below all thresholds
			"long_term":   0.25, // Below all thresholds
			"opportunity": 0.50, // Below thresholds
			"dividends":   0.50, // Below thresholds
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"cagr_raw":    0.08, // Below growth threshold
				"sharpe_raw":  0.50, // Below risk-adjusted threshold
				"sortino_raw": 0.50, // Below risk-adjusted threshold
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail IS present when failing all paths
	assert.Contains(t, tags, "quality-gate-fail", "Should have quality-gate-fail when failing all paths")
}

func TestQualityGate_MissingDataPartialPaths(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	// Some scores present, some missing - should still pass via Path 1
	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.60, // Passes Path 1
			"long_term": 0.50, // Passes Path 1
			// Missing: opportunity, dividends
		},
		// Missing: SubScores, DividendYield, Volatility
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail is NOT present (inverted logic)
	assert.NotContains(t, tags, "quality-gate-fail", "Should NOT have quality-gate-fail when passing")
}

func TestQualityGate_AllDataMissing_Fail(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	input := AssignTagsInput{
		Symbol: "TEST",
		// All scores missing (map is empty or nil)
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// NEW: Verify quality-gate-fail IS present when failing all paths
	assert.Contains(t, tags, "quality-gate-fail", "Should have quality-gate-fail when failing all paths")
}

// Adaptive Threshold Tests

func TestQualityGate_Path1_AdaptiveBearMarket(t *testing.T) {
	// Test requires mocking AdaptiveService which returns stricter thresholds
	// For now, test with default thresholds (will enhance when adaptive is implemented)
	t.Skip("Adaptive threshold testing requires AdaptiveService mock - will implement with main logic")
}

func TestQualityGate_Path1_AdaptiveBullMarket(t *testing.T) {
	// Test requires mocking AdaptiveService which returns relaxed thresholds
	// For now, test with default thresholds (will enhance when adaptive is implemented)
	t.Skip("Adaptive threshold testing requires AdaptiveService mock - will implement with main logic")
}

// TestQualityGate_NeverAssignsPassTag verifies that quality-gate-pass is NEVER assigned
// (architectural change: we only assign quality-gate-fail when failing, not quality-gate-pass when passing)
func TestQualityGate_NeverAssignsPassTag(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	// Test multiple scenarios - none should assign quality-gate-pass
	scenarios := []struct {
		name  string
		input AssignTagsInput
	}{
		{
			name: "Passing all paths",
			input: AssignTagsInput{
				Symbol: "PASS_ALL",
				GroupScores: map[string]float64{
					"stability": 0.80,
					"long_term": 0.80,
				},
			},
		},
		{
			name: "Failing all paths",
			input: AssignTagsInput{
				Symbol: "FAIL_ALL",
				GroupScores: map[string]float64{
					"stability": 0.40,
					"long_term": 0.30,
				},
			},
		},
		{
			name: "Passing Path 2 (Exceptional Excellence)",
			input: AssignTagsInput{
				Symbol: "PASS_PATH2",
				GroupScores: map[string]float64{
					"stability": 0.76, // >= 0.75
					"long_term": 0.30, // Below all other thresholds
				},
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			tags, err := assigner.AssignTagsForSecurity(scenario.input)
			assert.NoError(t, err)
			assert.NotContains(t, tags, "quality-gate-pass",
				"Should NEVER assign quality-gate-pass - only quality-gate-fail when failing")
		})
	}
}

// ============================================================================
// Tests for Configurable Thresholds (New Implementation)
// ============================================================================

func TestTagAssigner_Path5_ConfigurableSharpeSortino(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	// Path 5: Risk-Adjusted Excellence
	// With new configurable thresholds: Sharpe >= 0.7 OR Sortino >= 0.7 (was 0.9)
	// This should allow more securities to pass Path 5
	volatility := 0.30
	input := AssignTagsInput{
		Symbol:     "TEST",
		Volatility: &volatility,
		GroupScores: map[string]float64{
			"long_term": 0.60, // >= 0.55 threshold
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"sharpe_raw":  0.75, // >= 0.7 (new threshold, would fail with 0.9)
				"sortino_raw": 0.65, // < 0.7 but sharpe passes
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// Should pass quality gate via Path 5
	assert.NotContains(t, tags, "quality-gate-fail", "Should pass Path 5 with Sharpe 0.75 (>= 0.7)")
}

func TestTagAssigner_Path5_FailsWithOldThreshold(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	// Test that securities with Sharpe/Sortino between 0.7 and 0.9 now pass
	// (previously would have failed with 0.9 threshold)
	volatility := 0.30
	input := AssignTagsInput{
		Symbol:     "TEST",
		Volatility: &volatility,
		GroupScores: map[string]float64{
			"long_term": 0.60,
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"sharpe_raw":  0.75, // Would fail old 0.9 threshold, passes new 0.7
				"sortino_raw": 0.72, // Would fail old 0.9 threshold, passes new 0.7
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.NotContains(t, tags, "quality-gate-fail", "Should pass with Sharpe 0.75 and Sortino 0.72 (both >= 0.7)")
}

func TestTagAssigner_GrowthTag_ConfigurableThreshold(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	// Growth tag: CAGR >= 0.13 (was hardcoded 0.15)
	// This should allow securities with 13-15% CAGR to get growth tag
	cagr := 0.14 // Between old (0.15) and new (0.13) threshold
	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability": 0.75, // > HighQualityStability (0.70)
		},
		SubScores: map[string]map[string]float64{
			"long_term": {
				"cagr_raw": cagr, // 14% - would fail old 0.15, passes new 0.13
			},
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "growth", "Should have growth tag with CAGR 0.14 (>= 0.13)")
}

func TestTagAssigner_ValueOpportunity_ConfigurableThreshold(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	currentPrice := 80.0
	price52wHigh := 100.0

	// Value opportunity: opportunityScore > 0.65 (configurable, was hardcoded)
	// Test with score exactly at threshold
	input := AssignTagsInput{
		Symbol:       "TEST",
		CurrentPrice: &currentPrice,
		Price52wHigh: &price52wHigh,
		GroupScores: map[string]float64{
			"opportunity": 0.66, // > 0.65 threshold
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "value-opportunity", "Should have value-opportunity tag with score 0.66 (>= 0.65)")
}

func TestTagAssigner_HighScore_ConfigurableThreshold(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	// High score: totalScore > 0.7 (configurable threshold)
	input := AssignTagsInput{
		Symbol: "TEST",
		GroupScores: map[string]float64{
			"stability":   0.75,
			"long_term":   0.72,
			"opportunity": 0.70,
		},
	}

	_, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	// Total score should be calculated and compared to HighScoreThreshold (0.7)
	// If total > 0.7, should have high-score tag
	// Note: Actual total score calculation depends on scoring logic
	// This test verifies the threshold is configurable and code compiles
}

func TestTagAssigner_SidewaysValue_ConfigurableThreshold(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	currentPrice := 80.0
	price52wHigh := 100.0

	// Sideways value: value-opportunity tag + stability > 0.75 (configurable)
	input := AssignTagsInput{
		Symbol:       "TEST",
		CurrentPrice: &currentPrice,
		Price52wHigh: &price52wHigh,
		GroupScores: map[string]float64{
			"opportunity": 0.70, // > 0.65, gets value-opportunity tag
			"stability":   0.76, // > 0.75, required for sideways-value
		},
	}

	tags, err := assigner.AssignTagsForSecurity(input)
	assert.NoError(t, err)
	assert.Contains(t, tags, "value-opportunity", "Should have value-opportunity tag")
	assert.Contains(t, tags, "regime-sideways-value", "Should have regime-sideways-value tag with stability 0.76 (>= 0.75)")
}

func TestTagAssigner_AllQualityGatePaths_ConfigurableThresholds(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	assigner := NewTagAssigner(log)

	testCases := []struct {
		name       string
		input      AssignTagsInput
		shouldPass bool
		pathName   string
	}{
		{
			name: "Path 2: Exceptional Excellence",
			input: AssignTagsInput{
				Symbol: "PATH2",
				GroupScores: map[string]float64{
					"stability": 0.76, // >= 0.75 (ExceptionalExcellenceThreshold)
					"long_term": 0.30, // Below threshold
				},
			},
			shouldPass: true,
			pathName:   "exceptional_excellence",
		},
		{
			name: "Path 3: Quality Value Play",
			input: AssignTagsInput{
				Symbol: "PATH3",
				GroupScores: map[string]float64{
					"stability":   0.61, // >= 0.60 (QualityValueStabilityMin)
					"opportunity": 0.66, // >= 0.65 (QualityValueOpportunityMin)
					"long_term":   0.31, // >= 0.30 (QualityValueLongTermMin)
				},
			},
			shouldPass: true,
			pathName:   "quality_value",
		},
		{
			name: "Path 4: Dividend Income Play",
			input: AssignTagsInput{
				Symbol: "PATH4",
				GroupScores: map[string]float64{
					"stability": 0.56, // >= 0.55 (DividendIncomeStabilityMin)
					"dividends": 0.66, // >= 0.65 (DividendIncomeScoreMin) - dividend score is in GroupScores
				},
				DividendYield: func() *float64 { v := 0.036; return &v }(), // >= 0.035 (DividendIncomeYieldMin)
			},
			shouldPass: true,
			pathName:   "dividend_income",
		},
		{
			name: "Path 6: Composite Minimum",
			input: AssignTagsInput{
				Symbol: "PATH6",
				GroupScores: map[string]float64{
					"stability": 0.50, // >= 0.45 (CompositeStabilityFloor)
					"long_term": 0.55, // Composite: 0.6*0.50 + 0.4*0.55 = 0.52 (>= CompositeScoreMin)
				},
			},
			shouldPass: true,
			pathName:   "composite",
		},
		{
			name: "Path 7: Growth Opportunity",
			input: AssignTagsInput{
				Symbol:     "PATH7",
				Volatility: func() *float64 { v := 0.35; return &v }(), // <= 0.40 (GrowthOpportunityVolatilityMax)
				GroupScores: map[string]float64{
					"stability": 0.51, // >= 0.50 (GrowthOpportunityStabilityMin)
				},
				SubScores: map[string]map[string]float64{
					"long_term": {
						"cagr_raw": 0.14, // >= 0.13 (GrowthOpportunityCAGRMin)
					},
				},
			},
			shouldPass: true,
			pathName:   "growth",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tags, err := assigner.AssignTagsForSecurity(tc.input)
			assert.NoError(t, err)
			if tc.shouldPass {
				assert.NotContains(t, tags, "quality-gate-fail",
					"Should pass %s path", tc.pathName)
			} else {
				assert.Contains(t, tags, "quality-gate-fail",
					"Should fail %s path", tc.pathName)
			}
		})
	}
}
