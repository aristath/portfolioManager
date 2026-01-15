// Package scorers provides security scoring implementations.
package scorers

import (
	"fmt"

	"github.com/aristath/sentinel/internal/utils"
)

// ConcentrationScorer checks for dangerous concentration levels in the portfolio.
// This is a guardrail/penalty-only scorer - it does NOT reward filling gaps,
// it only blocks dangerous concentration.
type ConcentrationScorer struct{}

// ConcentrationCheck represents the result of a concentration guardrail check.
type ConcentrationCheck struct {
	Passes          bool    // True if within acceptable limits
	PositionWeight  float64 // Current/proposed position weight (0-1)
	GeographyWeight float64 // Current geography allocation (0-1)
	Reason          string  // Explanation if blocked (empty if passes)
}

// ConcentrationThresholds defines hard limits for concentration.
// These are guardrails, not targets - we don't try to hit them, we only prevent exceeding them.
type ConcentrationThresholds struct {
	MaxPositionWeight  float64 // Maximum single position as fraction of portfolio (default: 0.20 = 20%)
	MaxGeographyWeight float64 // Maximum single geography as fraction of portfolio (default: 0.40 = 40%)
}

// ConcentrationContext provides the data needed for concentration checking.
// This is a minimal interface that can be satisfied by different context types.
type ConcentrationContext struct {
	Positions            map[string]float64 // Position values by ISIN
	TotalValue           float64            // Total portfolio value
	GeographyAllocations map[string]float64 // Current geography allocations (0-1)
}

// DefaultConcentrationThresholds returns the default concentration thresholds.
// These represent hard limits for concentration risk.
func DefaultConcentrationThresholds() ConcentrationThresholds {
	return ConcentrationThresholds{
		MaxPositionWeight:  0.20, // 20% max single position
		MaxGeographyWeight: 0.40, // 40% max single geography
	}
}

// NewConcentrationScorer creates a new concentration scorer.
func NewConcentrationScorer() *ConcentrationScorer {
	return &ConcentrationScorer{}
}

// CheckConcentration checks if a proposed buy would exceed concentration limits.
// Returns passes=true if within limits, passes=false with reason if blocked.
//
// Key behavior:
// - Position concentration (>20%) is ALWAYS checked
// - Geography concentration (>40%) is only checked for NEW positions
// - Adding to existing positions in over-concentrated geographies is allowed
//
// This implements "diversification as guardrail, not driver":
// - We don't boost priority for filling gaps
// - We only block dangerous concentration
func (cs *ConcentrationScorer) CheckConcentration(
	isin string,
	geography string,
	proposedValueEUR float64,
	ctx *ConcentrationContext,
	thresholds ConcentrationThresholds,
) ConcentrationCheck {
	// Handle nil context gracefully - pass by default
	if ctx == nil {
		return ConcentrationCheck{Passes: true}
	}

	// Handle empty portfolio - first position always passes
	if ctx.TotalValue <= 0 {
		return ConcentrationCheck{Passes: true}
	}

	// Calculate current position value and proposed new value
	currentPositionValue := 0.0
	if ctx.Positions != nil {
		currentPositionValue = ctx.Positions[isin]
	}
	isExistingPosition := currentPositionValue > 0

	// Calculate what the position weight would be after this buy
	newPositionValue := currentPositionValue + proposedValueEUR
	newTotalValue := ctx.TotalValue + proposedValueEUR
	newPositionWeight := newPositionValue / newTotalValue

	// Check position concentration
	if newPositionWeight > thresholds.MaxPositionWeight {
		return ConcentrationCheck{
			Passes:          false,
			PositionWeight:  newPositionWeight,
			GeographyWeight: 0,
			Reason: fmt.Sprintf("position concentration: %.1f%% > %.0f%% threshold",
				newPositionWeight*100, thresholds.MaxPositionWeight*100),
		}
	}

	// Check geography concentration (only for NEW positions)
	// Existing positions are allowed to continue growing in over-concentrated geographies
	if !isExistingPosition {
		geographies := utils.ParseCSV(geography)
		for _, geo := range geographies {
			geoAllocation := 0.0
			if ctx.GeographyAllocations != nil {
				geoAllocation = ctx.GeographyAllocations[geo]
			}

			if geoAllocation > thresholds.MaxGeographyWeight {
				return ConcentrationCheck{
					Passes:          false,
					PositionWeight:  newPositionWeight,
					GeographyWeight: geoAllocation,
					Reason: fmt.Sprintf("geography concentration: %s at %.1f%% > %.0f%% threshold",
						geo, geoAllocation*100, thresholds.MaxGeographyWeight*100),
				}
			}
		}
	}

	// All checks passed
	return ConcentrationCheck{
		Passes:          true,
		PositionWeight:  newPositionWeight,
		GeographyWeight: 0,
		Reason:          "",
	}
}
