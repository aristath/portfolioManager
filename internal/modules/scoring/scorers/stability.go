package scorers

import (
	"math"

	"github.com/aristath/sentinel/pkg/formulas"
)

// StabilityScorer calculates financial stability from price behavior
// Replaces StabilityScorer which required external data
// Components:
// - Consistency (40%): 5-year vs 10-year CAGR similarity
// - Volatility (30%): Inverse of annualized volatility (lower = more stable)
// - Recovery (30%): Speed of recovery from drawdowns (using Ulcer Index)
type StabilityScorer struct{}

// StabilityScore represents the result of stability scoring
type StabilityScore struct {
	Components map[string]float64 `json:"components"`
	Score      float64            `json:"score"`
}

// NewStabilityScorer creates a new stability scorer
func NewStabilityScorer() *StabilityScorer {
	return &StabilityScorer{}
}

// Calculate computes stability score from price history only
// Uses internal data: monthly prices for CAGR consistency, daily prices for volatility/recovery
func (ss *StabilityScorer) Calculate(
	monthlyPrices []formulas.MonthlyPrice,
	dailyPrices []float64,
) StabilityScore {
	// Handle empty or insufficient data
	if len(monthlyPrices) == 0 && len(dailyPrices) == 0 {
		return StabilityScore{
			Score: 0.5,
			Components: map[string]float64{
				"consistency": 0.5,
				"volatility":  0.5,
				"recovery":    0.5,
			},
		}
	}

	if len(monthlyPrices) <= 1 && len(dailyPrices) <= 1 {
		return StabilityScore{
			Score: 0.5,
			Components: map[string]float64{
				"consistency": 0.5,
				"volatility":  0.5,
				"recovery":    0.5,
			},
		}
	}

	// 1. Consistency Score (40% weight)
	// Compare 5-year CAGR vs 10-year CAGR
	consistencyScore := calculateConsistencyFromPrices(monthlyPrices)

	// 2. Volatility Score (30% weight)
	// Lower volatility = higher stability
	volatilityScore := calculateVolatilityFromPrices(dailyPrices)

	// 3. Recovery Score (30% weight)
	// Faster recovery from drawdowns = higher stability
	recoveryScore := calculateRecoveryFromPrices(dailyPrices)

	// Calculate weighted total
	// Weights: Consistency 40%, Volatility 30%, Recovery 30%
	totalScore := consistencyScore*0.40 + volatilityScore*0.30 + recoveryScore*0.30

	// Ensure score is in valid range
	totalScore = math.Max(0.0, math.Min(1.0, totalScore))

	return StabilityScore{
		Score: round3(totalScore),
		Components: map[string]float64{
			"consistency": round3(consistencyScore),
			"volatility":  round3(volatilityScore),
			"recovery":    round3(recoveryScore),
		},
	}
}

// calculateConsistencyFromPrices calculates consistency score from monthly prices
func calculateConsistencyFromPrices(monthlyPrices []formulas.MonthlyPrice) float64 {
	if len(monthlyPrices) < 12 {
		return 0.5 // Neutral for insufficient data
	}

	// Calculate 5-year CAGR (60 months)
	cagr5y := formulas.CalculateCAGR(monthlyPrices, 60)
	if cagr5y == nil {
		// Fall back to all available data
		cagr5y = formulas.CalculateCAGR(monthlyPrices, len(monthlyPrices))
	}

	if cagr5y == nil {
		return 0.5
	}

	// Calculate 10-year CAGR (120 months) if available
	var cagr10y *float64
	if len(monthlyPrices) >= 120 {
		cagr10y = formulas.CalculateCAGR(monthlyPrices, 120)
	}

	return calculateConsistencyScore(*cagr5y, cagr10y)
}

// calculateConsistencyScore calculates consistency score from CAGR values
// High consistency = similar 5-year and 10-year CAGRs
func calculateConsistencyScore(cagr5y float64, cagr10y *float64) float64 {
	if cagr10y == nil {
		// Without 10-year data, return neutral score
		// But give slight bonus for positive 5-year CAGR
		if cagr5y >= 0.10 {
			return 0.65
		} else if cagr5y >= 0.05 {
			return 0.6
		} else if cagr5y >= 0 {
			return 0.55
		}
		return 0.5
	}

	// Calculate absolute difference between 5-year and 10-year CAGR
	diff := math.Abs(cagr5y - *cagr10y)

	// Score based on difference:
	// diff < 2% → 1.0 (highly consistent)
	// diff < 5% → 0.8
	// diff < 10% → 0.6
	// diff < 15% → 0.4
	// diff >= 15% → 0.2
	var score float64
	switch {
	case diff < 0.02:
		score = 1.0
	case diff < 0.05:
		// Linear interpolation between 0.8 and 1.0
		score = 1.0 - (diff-0.02)/0.03*0.2
	case diff < 0.10:
		// Linear interpolation between 0.6 and 0.8
		score = 0.8 - (diff-0.05)/0.05*0.2
	case diff < 0.15:
		// Linear interpolation between 0.4 and 0.6
		score = 0.6 - (diff-0.10)/0.05*0.2
	default:
		// Large divergence
		score = math.Max(0.2, 0.4-(diff-0.15)/0.10*0.2)
	}

	return score
}

// calculateVolatilityFromPrices calculates volatility score from daily prices
func calculateVolatilityFromPrices(dailyPrices []float64) float64 {
	if len(dailyPrices) < 30 {
		return 0.5 // Neutral for insufficient data
	}

	volatility := formulas.CalculateVolatility(dailyPrices)
	return calculateVolatilityScore(volatility)
}

// calculateVolatilityScore converts volatility to a score
// Lower volatility = higher stability score
func calculateVolatilityScore(volatility *float64) float64 {
	if volatility == nil {
		return 0.5
	}

	vol := *volatility

	// Score inversely proportional to volatility
	// vol < 10% → 1.0 (very stable)
	// vol < 15% → 0.9
	// vol < 20% → 0.75
	// vol < 25% → 0.6
	// vol < 30% → 0.45
	// vol < 40% → 0.3
	// vol >= 40% → 0.15
	var score float64
	switch {
	case vol < 0.10:
		score = 1.0
	case vol < 0.15:
		score = 1.0 - (vol-0.10)/0.05*0.1
	case vol < 0.20:
		score = 0.9 - (vol-0.15)/0.05*0.15
	case vol < 0.25:
		score = 0.75 - (vol-0.20)/0.05*0.15
	case vol < 0.30:
		score = 0.6 - (vol-0.25)/0.05*0.15
	case vol < 0.40:
		score = 0.45 - (vol-0.30)/0.10*0.15
	default:
		score = math.Max(0.1, 0.3-(vol-0.40)/0.20*0.15)
	}

	return score
}

// calculateRecoveryFromPrices calculates recovery score from daily prices
// Uses Ulcer Index as a measure of drawdown depth and duration
func calculateRecoveryFromPrices(dailyPrices []float64) float64 {
	if len(dailyPrices) < 30 {
		return 0.5 // Neutral for insufficient data
	}

	// Calculate Ulcer Index (14-day period)
	ulcerIndex := formulas.CalculateUlcerIndex(dailyPrices, 14)
	return calculateRecoveryScore(ulcerIndex)
}

// calculateRecoveryScore converts Ulcer Index to a score
// Lower Ulcer Index = faster recovery = higher score
func calculateRecoveryScore(ulcerIndex *float64) float64 {
	if ulcerIndex == nil {
		return 0.5
	}

	ui := *ulcerIndex

	// Score inversely proportional to Ulcer Index
	// UI < 2 → 1.0 (very fast recovery)
	// UI < 5 → 0.85
	// UI < 8 → 0.7
	// UI < 12 → 0.5
	// UI < 18 → 0.35
	// UI >= 18 → 0.2
	var score float64
	switch {
	case ui < 2:
		score = 1.0
	case ui < 5:
		score = 1.0 - (ui-2)/3*0.15
	case ui < 8:
		score = 0.85 - (ui-5)/3*0.15
	case ui < 12:
		score = 0.7 - (ui-8)/4*0.2
	case ui < 18:
		score = 0.5 - (ui-12)/6*0.15
	default:
		score = math.Max(0.1, 0.35-(ui-18)/12*0.15)
	}

	return score
}
