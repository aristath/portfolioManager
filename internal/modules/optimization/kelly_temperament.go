package optimization

import (
	"math"

	"github.com/aristath/sentinel/internal/modules/settings"
)

// KellyConfig holds the configuration for temperament-aware Kelly sizing.
// This allows the Kelly sizer to use adjusted parameters from the settings service.
type KellyConfig struct {
	// Params holds the Kelly parameters adjusted by temperament
	Params settings.KellyParams
}

// NewKellyConfig creates a KellyConfig from the settings service.
// This is the recommended way to create a Kelly config as it respects temperament settings.
func NewKellyConfig(settingsService *settings.Service) KellyConfig {
	return KellyConfig{
		Params: settingsService.GetAdjustedKellyParams(),
	}
}

// NewDefaultKellyConfig creates a KellyConfig with default (non-temperament-adjusted) values.
// Use this when you don't have access to the settings service.
func NewDefaultKellyConfig() KellyConfig {
	return KellyConfig{
		Params: settings.KellyParams{
			// Core Kelly parameters
			FixedFractional: 0.50, // Half-Kelly
			MinPositionSize: 0.01, // 1% minimum
			MaxPositionSize: 0.15, // 15% maximum

			// Adaptive multiplier parameters
			BaseMultiplier:            0.50, // Base fractional Kelly
			ConfidenceAdjustmentRange: 0.15, // ±15% based on confidence
			RegimeAdjustmentRange:     0.10, // ±10% based on regime
			MinMultiplier:             0.25, // Floor at quarter-Kelly
			MaxMultiplier:             0.75, // Ceiling at three-quarter-Kelly

			// Regime thresholds
			BullThreshold: 0.50,  // Above this = bull market
			BearThreshold: -0.50, // Below this = bear market

			// Bear market adjustment
			BearReduction:    0.75, // Reduce to 75% in strong bear
			BearMaxReduction: 0.25, // Maximum 25% reduction
		},
	}
}

// GetFractionalMultiplier returns the fractional Kelly multiplier based on regime and confidence.
// This implements the adaptive Kelly sizing logic using temperament-adjusted parameters.
func (c KellyConfig) GetFractionalMultiplier(regimeScore, confidence float64) float64 {
	// Base multiplier (adjusted by temperament)
	multiplier := c.Params.BaseMultiplier

	// Confidence adjustment: maps confidence [0, 1] to [-range, +range]
	// High confidence (0.8+) increases multiplier, low confidence decreases it
	confidenceAdjustment := (confidence - 0.5) * c.Params.ConfidenceAdjustmentRange * 2

	// Regime adjustment: applies only in clear bull/bear markets
	regimeAdjustment := 0.0
	if regimeScore > c.Params.BullThreshold {
		regimeAdjustment = c.Params.RegimeAdjustmentRange // Bull market: more aggressive
	} else if regimeScore < c.Params.BearThreshold {
		regimeAdjustment = -c.Params.RegimeAdjustmentRange // Bear market: more conservative
	}

	// Calculate final multiplier
	multiplier = multiplier + confidenceAdjustment + regimeAdjustment

	// Clamp to safe range
	if multiplier < c.Params.MinMultiplier {
		multiplier = c.Params.MinMultiplier
	}
	if multiplier > c.Params.MaxMultiplier {
		multiplier = c.Params.MaxMultiplier
	}

	return multiplier
}

// ApplyRegimeAdjustment applies regime-based adjustment to Kelly fraction.
// More conservative in bear markets.
func (c KellyConfig) ApplyRegimeAdjustment(kellyFraction, regimeScore float64) float64 {
	// Only reduce in bear markets (regimeScore < 0)
	if regimeScore >= 0 {
		return kellyFraction
	}

	// Reduction factor: 1.0 (no reduction) to (1.0 - BearMaxReduction) as regime goes 0 to -1.0
	// Formula: 1.0 - BearMaxReduction * |regimeScore| for negative regime scores
	reductionFactor := 1.0 - c.Params.BearMaxReduction*math.Abs(regimeScore)

	// Clamp reduction factor to minimum
	minReduction := 1.0 - c.Params.BearMaxReduction
	if reductionFactor < minReduction {
		reductionFactor = minReduction
	}

	return kellyFraction * reductionFactor
}

// ApplyConstraints applies min/max constraints to Kelly fraction.
func (c KellyConfig) ApplyConstraints(kellyFraction float64) float64 {
	// Floor at minimum position size
	if kellyFraction < c.Params.MinPositionSize {
		return c.Params.MinPositionSize
	}

	// Cap at maximum position size
	if kellyFraction > c.Params.MaxPositionSize {
		return c.Params.MaxPositionSize
	}

	return kellyFraction
}

// CalculateOptimalSize calculates the optimal position size using Kelly Criterion
// with temperament-adjusted constraints and adaptive adjustments.
//
// Args:
//   - expectedReturn: Expected return for the security (annualized)
//   - variance: Variance of returns (annualized)
//   - riskFreeRate: Risk-free rate for excess return calculation
//   - confidence: Confidence level in the expected return (0.0 to 1.0)
//   - regimeScore: Current market regime score (-1.0 to +1.0)
//
// Returns:
//   - Optimal position size as fraction of portfolio (0.0 to 1.0)
func (c KellyConfig) CalculateOptimalSize(
	expectedReturn float64,
	variance float64,
	riskFreeRate float64,
	confidence float64,
	regimeScore float64,
) float64 {
	// Step 1: Calculate raw Kelly fraction
	kellyFraction := calculateKellyFractionCore(expectedReturn, riskFreeRate, variance)

	// Step 2: Apply fractional Kelly (adaptive based on regime and confidence)
	fractionalMultiplier := c.GetFractionalMultiplier(regimeScore, confidence)
	fractionalKelly := kellyFraction * fractionalMultiplier

	// Step 3: Apply regime adjustment (more conservative in bear markets)
	regimeAdjusted := c.ApplyRegimeAdjustment(fractionalKelly, regimeScore)

	// Step 4: Apply constraints (min/max bounds)
	finalSize := c.ApplyConstraints(regimeAdjusted)

	return finalSize
}

// calculateKellyFractionCore calculates the raw Kelly fraction.
// Formula: (expectedReturn - riskFreeRate) / variance
func calculateKellyFractionCore(expectedReturn, riskFreeRate, variance float64) float64 {
	// Edge = expected return - risk-free rate
	edge := expectedReturn - riskFreeRate

	// If no edge or negative edge, return 0
	if edge <= 0 {
		return 0.0
	}

	// If variance is zero or very small, return 0 (division by zero protection)
	if variance <= 1e-10 {
		return 0.0
	}

	// Kelly fraction = edge / variance
	kellyFraction := edge / variance

	// Ensure non-negative
	if kellyFraction < 0 {
		return 0.0
	}

	return kellyFraction
}
