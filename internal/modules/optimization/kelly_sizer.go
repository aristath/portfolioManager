package optimization

import (
	"fmt"
	"math"

	"github.com/aristath/sentinel/internal/market_regime"
	"github.com/rs/zerolog"
)

// KellyPositionSizer calculates optimal position sizes using the Kelly Criterion
// with constraints and adaptive fractional Kelly based on regime and confidence.
type KellyPositionSizer struct {
	riskFreeRate    float64
	fixedFractional float64
	minPositionSize float64
	maxPositionSize float64
	fractionalMode  string // "fixed" or "adaptive"
	returnsCalc     *ReturnsCalculator
	riskBuilder     *RiskModelBuilder
	regimeDetector  *market_regime.MarketRegimeDetector
	settingsService KellySettingsService // Optional: for temperament-adjusted parameters
	log             zerolog.Logger
}

// KellySizeResult contains the result of Kelly sizing calculation.
type KellySizeResult struct {
	KellyFraction        float64
	ConstrainedFraction  float64
	FractionalMultiplier float64
	RegimeAdjustment     float64
	FinalSize            float64
}

// NewKellyPositionSizer creates a new Kelly position sizer.
func NewKellyPositionSizer(
	riskFreeRate float64,
	fixedFractional float64,
	minPositionSize float64,
	maxPositionSize float64,
	returnsCalc *ReturnsCalculator,
	riskBuilder *RiskModelBuilder,
	regimeDetector *market_regime.MarketRegimeDetector,
) *KellyPositionSizer {
	log := zerolog.Nop()
	if returnsCalc != nil {
		log = returnsCalc.log
	}

	return &KellyPositionSizer{
		riskFreeRate:    riskFreeRate,
		fixedFractional: fixedFractional,
		minPositionSize: minPositionSize,
		maxPositionSize: maxPositionSize,
		fractionalMode:  "adaptive", // Default to adaptive
		returnsCalc:     returnsCalc,
		riskBuilder:     riskBuilder,
		regimeDetector:  regimeDetector,
		log:             log.With().Str("component", "kelly_sizer").Logger(),
	}
}

// SetFractionalMode sets the fractional Kelly mode.
func (ks *KellyPositionSizer) SetFractionalMode(mode string) {
	if mode == "fixed" || mode == "adaptive" {
		ks.fractionalMode = mode
	}
}

// SetSettingsService sets the settings service for temperament-aware configuration.
// When set, the Kelly sizer will use temperament-adjusted parameters from the settings service.
func (ks *KellyPositionSizer) SetSettingsService(settingsService KellySettingsService) {
	ks.settingsService = settingsService
	ks.log.Info().Msg("Settings service configured for Kelly sizer (temperament-aware)")
}

// KellySettingsService interface for temperament-aware Kelly parameters.
// This interface is implemented by settings.Service.
type KellySettingsService interface {
	GetAdjustedKellyParams() KellyParamsConfig
}

// KellyParamsConfig holds temperament-adjusted Kelly parameters.
type KellyParamsConfig struct {
	FixedFractional           float64
	MinPositionSize           float64
	MaxPositionSize           float64
	BearReduction             float64
	BaseMultiplier            float64
	ConfidenceAdjustmentRange float64
	RegimeAdjustmentRange     float64
	MinMultiplier             float64
	MaxMultiplier             float64
	BearMaxReduction          float64
	BullThreshold             float64
	BearThreshold             float64
}

// getKellyParams returns temperament-adjusted Kelly parameters if settings service is available,
// otherwise returns the constructor-provided defaults that match original hardcoded behavior.
func (ks *KellyPositionSizer) getKellyParams() KellyParamsConfig {
	if ks.settingsService != nil {
		return ks.settingsService.GetAdjustedKellyParams()
	}
	// Fallback to constructor-provided values + original hardcoded defaults
	// These match the original hardcoded values in the methods before temperament integration
	return KellyParamsConfig{
		FixedFractional:           ks.fixedFractional,
		MinPositionSize:           ks.minPositionSize,
		MaxPositionSize:           ks.maxPositionSize,
		BearReduction:             0.5,  // Not used in current logic
		BaseMultiplier:            0.5,  // Original: 0.5 (half-Kelly)
		ConfidenceAdjustmentRange: 0.3,  // Original: (confidence - 0.5) * 0.3 = ±0.15
		RegimeAdjustmentRange:     0.2,  // Original: ±0.10
		MinMultiplier:             0.25, // Original: clamp min 0.25
		MaxMultiplier:             0.75, // Original: clamp max 0.75
		BearMaxReduction:          0.25, // Original: 1.0 - 0.25 * |regime| = max 25% reduction
		BullThreshold:             0.5,  // Original: regimeScore > 0.5
		BearThreshold:             -0.5, // Original: regimeScore < -0.5
	}
}

// CalculateOptimalSize calculates the optimal position size using Kelly Criterion
// with constraints and adaptive adjustments.
//
// Args:
//   - expectedReturn: Expected return for the security (annualized)
//   - variance: Variance of returns (annualized)
//   - confidence: Confidence level in the expected return (0.0 to 1.0)
//   - regimeScore: Current market regime score (-1.0 to +1.0)
//
// Returns:
//   - Optimal position size as fraction of portfolio (0.0 to 1.0)
func (ks *KellyPositionSizer) CalculateOptimalSize(
	expectedReturn float64,
	variance float64,
	confidence float64,
	regimeScore float64,
) float64 {
	// Step 1: Calculate raw Kelly fraction
	kellyFraction := ks.calculateKellyFraction(expectedReturn, ks.riskFreeRate, variance)

	// Step 2: Apply fractional Kelly (adaptive or fixed)
	fractionalMultiplier := ks.getFractionalMultiplier(regimeScore, confidence)
	fractionalKelly := kellyFraction * fractionalMultiplier

	// Step 3: Apply regime adjustment (more conservative in bear markets)
	regimeAdjusted := ks.applyRegimeAdjustment(fractionalKelly, regimeScore)

	// Step 4: Apply constraints (min/max bounds)
	finalSize := ks.applyConstraints(regimeAdjusted)

	return finalSize
}

// CalculateOptimalSizeForISIN calculates optimal size for a security by ISIN.
// This is a convenience method that looks up expected return and variance.
// Uses temperament-adjusted parameters when settings service is available.
func (ks *KellyPositionSizer) CalculateOptimalSizeForISIN(
	isin string,
	expectedReturns map[string]float64, // ISIN-keyed
	covMatrix [][]float64,
	isins []string, // ISIN array
	confidence float64,
	regimeScore float64,
) (float64, error) {
	params := ks.getKellyParams()

	// Get expected return
	expectedReturn, hasReturn := expectedReturns[isin]
	if !hasReturn {
		return params.MinPositionSize, fmt.Errorf("no expected return for ISIN %s", isin)
	}

	// Get variance from covariance matrix diagonal
	variance, err := ks.getVarianceFromCovMatrix(isin, covMatrix, isins)
	if err != nil {
		return params.MinPositionSize, fmt.Errorf("failed to get variance for ISIN %s: %w", isin, err)
	}

	// Calculate optimal size
	optimalSize := ks.CalculateOptimalSize(expectedReturn, variance, confidence, regimeScore)

	return optimalSize, nil
}

// CalculateOptimalSizeForSymbol calculates optimal size for a security by symbol.
// Uses temperament-adjusted parameters when settings service is available.
func (ks *KellyPositionSizer) CalculateOptimalSizeForSymbol(
	symbol string,
	expectedReturns map[string]float64,
	covMatrix [][]float64,
	symbols []string,
	confidence float64,
	regimeScore float64,
) (float64, error) {
	params := ks.getKellyParams()

	// Get expected return
	expectedReturn, hasReturn := expectedReturns[symbol]
	if !hasReturn {
		return params.MinPositionSize, fmt.Errorf("no expected return for symbol %s", symbol)
	}

	// Get variance from covariance matrix diagonal
	variance, err := ks.getVarianceFromCovMatrix(symbol, covMatrix, symbols)
	if err != nil {
		return params.MinPositionSize, fmt.Errorf("failed to get variance for %s: %w", symbol, err)
	}

	// Calculate optimal size
	optimalSize := ks.CalculateOptimalSize(expectedReturn, variance, confidence, regimeScore)

	return optimalSize, nil
}

// calculateKellyFraction calculates the raw Kelly fraction.
// Formula: (expectedReturn - riskFreeRate) / variance
func (ks *KellyPositionSizer) calculateKellyFraction(expectedReturn, riskFreeRate, variance float64) float64 {
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

// applyConstraints applies min/max constraints to Kelly fraction.
// Uses temperament-adjusted min/max position sizes when settings service is available.
func (ks *KellyPositionSizer) applyConstraints(kellyFraction float64) float64 {
	params := ks.getKellyParams()

	// Floor at minimum position size
	if kellyFraction < params.MinPositionSize {
		return params.MinPositionSize
	}

	// Cap at maximum position size
	if kellyFraction > params.MaxPositionSize {
		return params.MaxPositionSize
	}

	return kellyFraction
}

// applyFractionalKelly applies fractional Kelly multiplier.
func (ks *KellyPositionSizer) applyFractionalKelly(kellyFraction float64, regimeScore float64, confidence float64) float64 {
	multiplier := ks.getFractionalMultiplier(regimeScore, confidence)
	return kellyFraction * multiplier
}

// getFractionalMultiplier returns the fractional Kelly multiplier based on mode.
// Uses temperament-adjusted parameters when settings service is available.
func (ks *KellyPositionSizer) getFractionalMultiplier(regimeScore float64, confidence float64) float64 {
	params := ks.getKellyParams()

	if ks.fractionalMode == "fixed" {
		return params.FixedFractional
	}

	// Adaptive mode: multiplier based on regime and confidence
	// Base multiplier from temperament settings
	baseMultiplier := params.BaseMultiplier

	// Confidence adjustment: based on temperament-adjusted range
	// High confidence (0.8+) → +range/2, Low confidence (0.3-) → -range/2
	confidenceAdjustment := (confidence - 0.5) * params.ConfidenceAdjustmentRange

	// Regime adjustment: based on temperament-adjusted range
	// Bull (above BullThreshold) → +range/2, Bear (below BearThreshold) → -range/2
	regimeAdjustment := 0.0
	if regimeScore > params.BullThreshold {
		regimeAdjustment = params.RegimeAdjustmentRange / 2 // Bull market: more aggressive
	} else if regimeScore < params.BearThreshold {
		regimeAdjustment = -params.RegimeAdjustmentRange / 2 // Bear market: more conservative
	}

	// Calculate final multiplier
	multiplier := baseMultiplier + confidenceAdjustment + regimeAdjustment

	// Clamp to temperament-adjusted range
	if multiplier < params.MinMultiplier {
		multiplier = params.MinMultiplier
	}
	if multiplier > params.MaxMultiplier {
		multiplier = params.MaxMultiplier
	}

	return multiplier
}

// applyRegimeAdjustment applies regime-based adjustment to Kelly fraction.
// More conservative in bear markets. Uses temperament-adjusted parameters.
func (ks *KellyPositionSizer) applyRegimeAdjustment(kellyFraction float64, regimeScore float64) float64 {
	params := ks.getKellyParams()

	// Only reduce in bear markets (regimeScore < 0)
	if regimeScore >= 0 {
		return kellyFraction
	}

	// Reduction factor: 1.0 (no reduction) to (1 - BearMaxReduction) as regime goes 0 to -1.0
	// BearMaxReduction is temperament-adjusted (e.g., 0.25 means max 25% reduction)
	reductionFactor := 1.0 - params.BearMaxReduction*math.Abs(regimeScore)

	// Clamp reduction factor to minimum (1 - BearMaxReduction)
	minReductionFactor := 1.0 - params.BearMaxReduction
	if reductionFactor < minReductionFactor {
		reductionFactor = minReductionFactor
	}

	return kellyFraction * reductionFactor
}

// getVarianceFromCovMatrix extracts variance for an identifier (ISIN or Symbol) from covariance matrix.
func (ks *KellyPositionSizer) getVarianceFromCovMatrix(identifier string, covMatrix [][]float64, identifiers []string) (float64, error) {
	// Find identifier index
	identifierIndex := -1
	for i, id := range identifiers {
		if id == identifier {
			identifierIndex = i
			break
		}
	}

	if identifierIndex < 0 {
		return 0.0, fmt.Errorf("identifier %s not found in identifiers list", identifier)
	}

	if identifierIndex >= len(covMatrix) {
		return 0.0, fmt.Errorf("identifier index %d out of bounds for covariance matrix", identifierIndex)
	}

	if identifierIndex >= len(covMatrix[identifierIndex]) {
		return 0.0, fmt.Errorf("covariance matrix row %d has insufficient columns", identifierIndex)
	}

	// Variance is the diagonal element
	variance := covMatrix[identifierIndex][identifierIndex]

	if variance < 0 {
		return 0.0, fmt.Errorf("negative variance for identifier %s: %f", identifier, variance)
	}

	return variance, nil
}

// CalculateOptimalSizesForAll calculates optimal sizes for all securities.
// Uses temperament-adjusted parameters when settings service is available.
func (ks *KellyPositionSizer) CalculateOptimalSizesForAll(
	expectedReturns map[string]float64,
	covMatrix [][]float64,
	symbols []string,
	confidences map[string]float64,
	regimeScore float64,
) (map[string]float64, error) {
	params := ks.getKellyParams()
	result := make(map[string]float64, len(symbols))

	for _, symbol := range symbols {
		// Get confidence (default to 0.5 if not provided)
		confidence := 0.5
		if conf, hasConf := confidences[symbol]; hasConf {
			confidence = conf
		}

		optimalSize, err := ks.CalculateOptimalSizeForSymbol(
			symbol,
			expectedReturns,
			covMatrix,
			symbols,
			confidence,
			regimeScore,
		)
		if err != nil {
			ks.log.Warn().
				Str("symbol", symbol).
				Err(err).
				Msg("Failed to calculate Kelly size, using min size")
			optimalSize = params.MinPositionSize
		}

		result[symbol] = optimalSize
	}

	return result, nil
}
