// Package evaluation provides evaluation functionality for security sequences.
// This package delegates to internal/evaluation for the actual scoring logic
// to maintain a single source of truth for evaluation weights and algorithms.
package evaluation

import (
	"math"

	coreevaluation "github.com/aristath/sentinel/internal/evaluation"
)

// =============================================================================
// CONSTANTS - Re-exported from core evaluation
// =============================================================================

const (
	// Main evaluation component weights (pure end-state scoring)
	WeightPortfolioQuality         = coreevaluation.WeightPortfolioQuality
	WeightDiversificationAlignment = coreevaluation.WeightDiversificationAlignment
	WeightRiskAdjustedMetrics      = coreevaluation.WeightRiskAdjustedMetrics
	WeightEndStateImprovement      = coreevaluation.WeightEndStateImprovement

	GeoWeight             = 0.40
	IndustryWeight        = 0.30
	QualityWeight         = 0.30
	DeviationScale        = 0.3
	SecurityQualityWeight = 0.6
	DividendYieldWeight   = 0.4
)

// =============================================================================
// MAIN EVALUATION FUNCTIONS - Delegate to core evaluation
// =============================================================================

// GetRegimeAdaptiveWeights returns evaluation weights adjusted for market regime.
// Delegates to core evaluation package.
func GetRegimeAdaptiveWeights(regimeScore float64) map[string]float64 {
	return coreevaluation.GetRegimeAdaptiveWeights(regimeScore)
}

// CalculateTransactionCost calculates total transaction cost for a sequence.
// Delegates to core evaluation package.
func CalculateTransactionCost(
	sequence []ActionCandidate,
	transactionCostFixed float64,
	transactionCostPercent float64,
) float64 {
	return coreevaluation.CalculateTransactionCost(sequence, transactionCostFixed, transactionCostPercent)
}

// CalculateTransactionCostEnhanced calculates total transaction cost with all components.
// Delegates to core evaluation package.
func CalculateTransactionCostEnhanced(
	sequence []ActionCandidate,
	transactionCostFixed float64,
	transactionCostPercent float64,
	spreadCostPercent float64,
	slippagePercent float64,
	marketImpactPercent float64,
) float64 {
	return coreevaluation.CalculateTransactionCostEnhanced(
		sequence,
		transactionCostFixed,
		transactionCostPercent,
		spreadCostPercent,
		slippagePercent,
		marketImpactPercent,
	)
}

// EvaluateEndState evaluates the end state of a portfolio after executing a sequence.
// Delegates to core evaluation package for unified scoring.
// Now uses pure end-state scoring with start context for improvement calculation.
func EvaluateEndState(
	startContext PortfolioContext,
	endContext PortfolioContext,
	sequence []ActionCandidate,
	transactionCostFixed float64,
	transactionCostPercent float64,
	costPenaltyFactor float64,
) float64 {
	return coreevaluation.EvaluateEndState(
		startContext,
		endContext,
		sequence,
		transactionCostFixed,
		transactionCostPercent,
		costPenaltyFactor,
		nil, // Use default scoring config (temperament config would need to be passed through)
	)
}

// CalculateDiversificationScore calculates diversification score for a portfolio.
// Delegates to core evaluation package.
func CalculateDiversificationScore(ctx PortfolioContext) float64 {
	return coreevaluation.CalculateDiversificationScore(ctx)
}

// EvaluateSequence evaluates a complete sequence: simulate + score.
// Delegates to core evaluation package.
func EvaluateSequence(
	sequence []ActionCandidate,
	context EvaluationContext,
) SequenceEvaluationResult {
	return coreevaluation.EvaluateSequence(sequence, context)
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// sum is a helper function to sum a slice of floats
func sum(values []float64) float64 {
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total
}

// Note: CheckSequenceFeasibility and SimulateSequence are defined in simulation.go

// =============================================================================
// Expose internal scoring functions for tests
// =============================================================================

// calculateOptimizerAlignment is exposed for testing
func calculateOptimizerAlignment(ctx PortfolioContext, totalValue float64) float64 {
	if len(ctx.OptimizerTargetWeights) == 0 {
		return 0.5
	}

	var deviations []float64
	for symbol, targetWeight := range ctx.OptimizerTargetWeights {
		currentValue, hasPosition := ctx.Positions[symbol]
		currentWeight := 0.0
		if hasPosition {
			currentWeight = currentValue / totalValue
		}

		deviation := math.Abs(currentWeight - targetWeight)
		deviations = append(deviations, deviation)
	}

	if len(deviations) == 0 {
		return 0.5
	}

	avgDeviation := sum(deviations) / float64(len(deviations))
	return math.Max(0.0, 1.0-avgDeviation/0.20)
}

// calculatePortfolioQualityScore is exposed for testing
func calculatePortfolioQualityScore(ctx PortfolioContext) float64 {
	// Use same context for start and end to get base quality score
	return coreevaluation.EvaluateEndState(ctx, ctx, nil, 0, 0, 0, nil)
}

// calculateRiskAdjustedScore is exposed for testing - returns neutral if no data
func calculateRiskAdjustedScore(ctx PortfolioContext) float64 {
	if ctx.TotalValue <= 0 {
		return 0.5
	}

	// Check if we have any Sharpe data
	hasSharpe := false
	for _, v := range ctx.SecuritySharpe {
		if v != 0 {
			hasSharpe = true
			break
		}
	}

	if !hasSharpe || len(ctx.SecuritySharpe) == 0 {
		return 0.5
	}

	// Calculate weighted Sharpe
	weightedSharpe := 0.0
	for symbol, value := range ctx.Positions {
		weight := value / ctx.TotalValue
		if sharpe, ok := ctx.SecuritySharpe[symbol]; ok {
			weightedSharpe += sharpe * weight
		}
	}

	// Convert to score
	if weightedSharpe >= 2.0 {
		return 1.0
	} else if weightedSharpe >= 1.0 {
		return 0.7 + (weightedSharpe-1.0)*0.3
	} else if weightedSharpe >= 0.5 {
		return 0.4 + (weightedSharpe-0.5)*0.6
	} else if weightedSharpe >= 0 {
		return weightedSharpe * 0.8
	}
	return 0.0
}

// The new scoring is purely based on portfolio end state quality.
