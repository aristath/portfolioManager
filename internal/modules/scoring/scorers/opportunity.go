package scorers

import (
	"math"

	"github.com/aristath/sentinel/internal/modules/scoring"
	"github.com/aristath/sentinel/pkg/formulas"
)

// OpportunityScorer calculates opportunity (value/dip) score
// Scoring is based purely on price distance from 52-week high (P/E ratio removed)
type OpportunityScorer struct{}

// OpportunityScore represents the result of opportunity scoring
type OpportunityScore struct {
	Components map[string]float64 `json:"components"`
	Score      float64            `json:"score"`
}

// NewOpportunityScorer creates a new opportunity scorer
func NewOpportunityScorer() *OpportunityScorer {
	return &OpportunityScorer{}
}

// Calculate calculates the opportunity score from daily prices and stability metrics
// Components:
// - Below 52-week High (100%): Distance from peak - dip opportunity
// P/E ratio component was removed (no longer available without external data)
func (os *OpportunityScorer) Calculate(
	dailyPrices []float64,
) OpportunityScore {
	return os.CalculateWithQualityGate(dailyPrices, nil, nil, "UNKNOWN")
}

// CalculateWithQualityGate calculates opportunity score with quality gates to prevent value traps
// Quality gates ensure we don't buy cheap but declining quality securities.
// Scoring is 100% based on distance from 52-week high (P/E ratio removed).
//
// Args:
//   - dailyPrices: Daily price history
//   - stabilityScore: Stability score (optional, for quality gate)
//   - longTermScore: Long-term score (optional, for quality gate)
//   - productType: Product type (EQUITY, ETF, MUTUALFUND, ETC, CASH, UNKNOWN)
//
// Returns:
//   - OpportunityScore with quality gate applied
func (os *OpportunityScorer) CalculateWithQualityGate(
	dailyPrices []float64,
	stabilityScore *float64,
	longTermScore *float64,
	productType string,
) OpportunityScore {
	if len(dailyPrices) < scoring.MinDaysForOpportunity {
		// Insufficient data - return neutral score
		return OpportunityScore{
			Score: 0.5,
			Components: map[string]float64{
				"below_52w_high":     0.5,
				"below_52w_high_raw": 0.0,
			},
		}
	}

	currentPrice := dailyPrices[len(dailyPrices)-1]

	// Calculate 52-week high distance score (100% weight - no P/E component)
	high52w := formulas.Calculate52WeekHigh(dailyPrices)
	below52wScore := scoreBelow52WeekHigh(currentPrice, high52w)

	// Base score is 100% below_52w_high (P/E ratio removed)
	baseScore := below52wScore

	// Apply quality gate: if opportunity score is high but quality is low, reduce score
	// This prevents buying value traps (cheap but declining quality)
	qualityPenalty := calculateQualityPenalty(baseScore, stabilityScore, longTermScore)
	finalScore := baseScore * (1.0 - qualityPenalty)
	finalScore = math.Min(1.0, finalScore)

	// Build components map with both scored and raw values
	components := map[string]float64{
		"below_52w_high": round3(below52wScore),
	}

	// Store quality gate penalty if applied
	if qualityPenalty > 0 {
		components["quality_penalty"] = round3(qualityPenalty)
	}

	// Store raw below_52w_high percentage for database storage
	// Percentage below 52-week high: (high - current) / high
	if high52w != nil && *high52w > 0 {
		below52wPct := (*high52w - currentPrice) / *high52w
		components["below_52w_high_raw"] = below52wPct
	} else {
		components["below_52w_high_raw"] = 0.0
	}

	return OpportunityScore{
		Score:      round3(finalScore),
		Components: components,
	}
}

// calculateQualityPenalty calculates penalty for low quality when opportunity score is high
// Prevents buying value traps: cheap but declining quality securities.
//
// Quality gate thresholds:
//   - minStabilityThreshold: 0.6 (stability must be decent)
//   - minLongTermThreshold: 0.5 (long-term must be acceptable)
//
// Penalty logic:
//   - If opportunity score > 0.7 and quality is below thresholds: apply 30% penalty
//   - If opportunity score > 0.5 and quality is very low: apply 15% penalty
//
// Returns:
//   - Penalty factor (0.0 to 0.3)
func calculateQualityPenalty(
	opportunityScore float64,
	stabilityScore *float64,
	longTermScore *float64,
) float64 {
	// If no quality data available, don't penalize (can't detect value trap)
	if stabilityScore == nil && longTermScore == nil {
		return 0.0
	}

	// Quality gate thresholds
	minStabilityThreshold := 0.6
	minLongTermThreshold := 0.5

	// Check if quality is below thresholds
	stabilityBelowThreshold := stabilityScore != nil && *stabilityScore < minStabilityThreshold
	longTermBelowThreshold := longTermScore != nil && *longTermScore < minLongTermThreshold

	// If both are below threshold, it's likely a value trap
	isValueTrap := stabilityBelowThreshold || longTermBelowThreshold

	if !isValueTrap {
		return 0.0 // Quality is acceptable, no penalty
	}

	// Apply penalty based on opportunity score level
	if opportunityScore > 0.7 {
		// High opportunity score but low quality = value trap
		return 0.30 // 30% penalty
	} else if opportunityScore > 0.5 {
		// Moderate opportunity score but low quality = potential value trap
		return 0.15 // 15% penalty
	}

	return 0.0 // Low opportunity score, no penalty needed
}

// scoreBelow52WeekHigh scores based on distance below 52-week high
// Further below = HIGHER score (buying opportunity)
func scoreBelow52WeekHigh(currentPrice float64, high52w *float64) float64 {
	if high52w == nil || *high52w <= 0 {
		return 0.5
	}

	pctBelow := (*high52w - currentPrice) / *high52w

	if pctBelow <= 0 {
		// At or above high = expensive
		return 0.2
	} else if pctBelow < scoring.BelowHighOK { // 0-10%
		return 0.2 + (pctBelow/scoring.BelowHighOK)*0.3 // 0.2-0.5
	} else if pctBelow < scoring.BelowHighGood { // 10-20%
		return 0.5 + ((pctBelow-scoring.BelowHighOK)/0.10)*0.3 // 0.5-0.8
	} else if pctBelow < scoring.BelowHighExcellent { // 20-30%
		return 0.8 + ((pctBelow-scoring.BelowHighGood)/0.10)*0.2 // 0.8-1.0
	} else { // 30%+ below
		return 1.0
	}
}

// Removed: scorePERatio function (P/E ratio no longer available without external data)

// IsPriceTooHigh checks if price is too close to 52-week high for buying
// Guardrail to prevent chasing all-time highs
func IsPriceTooHigh(currentPrice float64, high52w *float64, maxPriceVs52wHigh float64) bool {
	if high52w == nil || *high52w <= 0 {
		return false // No data, allow trade
	}
	return currentPrice >= *high52w*maxPriceVs52wHigh
}
