package optimization

import (
	"testing"

	"github.com/aristath/sentinel/internal/modules/settings"
	"github.com/stretchr/testify/assert"
)

// TestKellyParamsStruct verifies the Kelly params struct from settings
func TestKellyParamsStruct(t *testing.T) {
	params := settings.KellyParams{
		FixedFractional: 0.50,
		MinPositionSize: 0.01,
		MaxPositionSize: 0.15,
		MinMultiplier:   0.25,
		MaxMultiplier:   0.75,
		BullThreshold:   0.50,
		BearThreshold:   -0.50,
	}

	// Verify constraints are valid
	assert.True(t, params.MinPositionSize < params.MaxPositionSize,
		"min position size should be less than max")
	assert.True(t, params.MinMultiplier < params.MaxMultiplier,
		"min multiplier should be less than max")
	assert.True(t, params.BearThreshold < 0 && params.BullThreshold > 0,
		"bear threshold should be negative, bull positive")
	assert.True(t, params.FixedFractional > 0 && params.FixedFractional <= 1.0,
		"fixed fractional should be in (0, 1]")
}

// TestKellyConfigDefaults verifies default values match current hardcoded values
func TestKellyConfigDefaults(t *testing.T) {
	config := NewDefaultKellyConfig()

	// These should match the current hardcoded values in kelly_sizer.go
	assert.Equal(t, 0.5, config.Params.BaseMultiplier, "base multiplier should be 0.5")
	assert.Equal(t, 0.15, config.Params.ConfidenceAdjustmentRange, "confidence range should be 0.15")
	assert.Equal(t, 0.10, config.Params.RegimeAdjustmentRange, "regime range should be 0.10")
	assert.Equal(t, 0.25, config.Params.MinMultiplier, "min multiplier should be 0.25")
	assert.Equal(t, 0.75, config.Params.MaxMultiplier, "max multiplier should be 0.75")
	assert.Equal(t, 0.50, config.Params.BullThreshold, "bull threshold should be 0.5")
	assert.Equal(t, -0.50, config.Params.BearThreshold, "bear threshold should be -0.5")
	assert.Equal(t, 0.25, config.Params.BearMaxReduction, "bear max reduction should be 0.25")
}

// TestAdaptiveFractionalMultiplier verifies the fractional multiplier calculation
func TestAdaptiveFractionalMultiplier(t *testing.T) {
	config := NewDefaultKellyConfig()

	testCases := []struct {
		name        string
		regimeScore float64
		confidence  float64
		minExpected float64
		maxExpected float64
	}{
		{"neutral_balanced", 0.0, 0.5, 0.45, 0.55},
		{"bull_high_confidence", 0.8, 0.9, 0.60, 0.76},
		{"bear_low_confidence", -0.8, 0.2, 0.24, 0.40},
		{"neutral_low_confidence", 0.0, 0.2, 0.30, 0.50},
		{"neutral_high_confidence", 0.0, 0.9, 0.50, 0.70},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			multiplier := config.GetFractionalMultiplier(tc.regimeScore, tc.confidence)
			assert.GreaterOrEqual(t, multiplier, tc.minExpected,
				"multiplier should be >= %f", tc.minExpected)
			assert.LessOrEqual(t, multiplier, tc.maxExpected,
				"multiplier should be <= %f", tc.maxExpected)
		})
	}
}

// TestKellyConfigRegimeAdjustment verifies regime-based position size adjustment
func TestKellyConfigRegimeAdjustment(t *testing.T) {
	config := NewDefaultKellyConfig()

	testCases := []struct {
		name         string
		kellyFrac    float64
		regimeScore  float64
		shouldReduce bool
	}{
		{"bull_market", 0.5, 0.8, false},
		{"neutral_market", 0.5, 0.0, false},
		{"mild_bear", 0.5, -0.3, true},
		{"strong_bear", 0.5, -0.9, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adjusted := config.ApplyRegimeAdjustment(tc.kellyFrac, tc.regimeScore)
			if tc.shouldReduce {
				assert.Less(t, adjusted, tc.kellyFrac,
					"should reduce position in bear market")
			} else {
				assert.Equal(t, tc.kellyFrac, adjusted,
					"should not reduce in bull/neutral market")
			}
		})
	}
}

// TestConstraintsApplied verifies min/max constraints are applied
func TestConstraintsApplied(t *testing.T) {
	config := NewDefaultKellyConfig()

	testCases := []struct {
		name      string
		kellyFrac float64
		expected  float64
	}{
		{"below_min", 0.005, config.Params.MinPositionSize},
		{"above_max", 0.50, config.Params.MaxPositionSize},
		{"within_bounds", 0.10, 0.10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := config.ApplyConstraints(tc.kellyFrac)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestTemperamentAffectsKellyConfig verifies temperament changes Kelly behavior
func TestTemperamentAffectsKellyConfig(t *testing.T) {
	// At balanced temperament (0.5), config should use base values
	// This is a design contract test

	config := NewDefaultKellyConfig()

	// The config should reflect the base values
	assert.Equal(t, 0.5, config.Params.BaseMultiplier,
		"at balanced temperament, base multiplier should be 0.5")

	// When we implement temperament-aware config, aggressive temperament
	// would increase the base multiplier toward the max
}

// TestKellyMultiplierBounds verifies multiplier stays within safe bounds
func TestKellyMultiplierBounds(t *testing.T) {
	config := NewDefaultKellyConfig()

	// Test all extreme combinations
	extremes := []float64{-1.0, 0.0, 1.0}
	for _, regime := range extremes {
		for _, confidence := range extremes {
			if confidence < 0 {
				continue // Confidence is [0, 1]
			}
			multiplier := config.GetFractionalMultiplier(regime, confidence)

			assert.GreaterOrEqual(t, multiplier, config.Params.MinMultiplier,
				"multiplier should be >= min at regime=%f, confidence=%f", regime, confidence)
			assert.LessOrEqual(t, multiplier, config.Params.MaxMultiplier,
				"multiplier should be <= max at regime=%f, confidence=%f", regime, confidence)
		}
	}
}

// TestBearReductionFactor verifies bear market reduction is bounded
func TestBearReductionFactor(t *testing.T) {
	config := NewDefaultKellyConfig()
	initialKelly := 0.5

	// Test progressive bear reduction
	bearScores := []float64{0.0, -0.25, -0.5, -0.75, -1.0}
	previousAdjusted := initialKelly

	for i, regime := range bearScores {
		adjusted := config.ApplyRegimeAdjustment(initialKelly, regime)

		// Should be monotonically decreasing (or equal for regime >= 0)
		if regime < 0 && i > 0 {
			assert.LessOrEqual(t, adjusted, previousAdjusted,
				"adjustment should decrease as bear gets stronger")
		}

		// Should never go below minimum reduction factor
		minAdjusted := initialKelly * (1.0 - config.Params.BearMaxReduction)
		assert.GreaterOrEqual(t, adjusted, minAdjusted,
			"adjusted should not go below minimum (bear max reduction)")

		previousAdjusted = adjusted
	}
}
