package calculators

import (
	"github.com/aristath/sentinel/internal/modules/planning/domain"
)

// QualityCheckResult contains the result of quality checks
type QualityCheckResult struct {
	PassesQualityGate  bool
	IsValueTrap        bool
	IsBubbleRisk       bool
	BelowMinimumReturn bool
	QualityGateReason  string

	// Quantum detection results
	QuantumValueTrapProb float64 // Quantum probability (0-1)
	IsQuantumValueTrap   bool    // Quantum detected trap (>0.7)
	IsQuantumWarning     bool    // Quantum early warning (0.5-0.7)
	IsEnsembleValueTrap  bool    // Ensemble decision (classical OR quantum)
}

// CheckQualityGates is deprecated - quality checks are now handled via tags.
// This function always returns early as tags are mandatory for all quality gates.
// Tags encode explicit quality judgments: quality-gate-fail, below-minimum-return, bubble-risk, value-trap, etc.
// All quality checks should use tags directly - see tag_assigner.go for quality logic.
// Accepts ISIN as PRIMARY identifier for efficient O(1) lookups (kept for backward compatibility).
func CheckQualityGates(
	ctx *domain.OpportunityContext,
	isin string,
	isNewPosition bool,
	config *domain.PlannerConfiguration,
) QualityCheckResult {
	// Return default result (pass) - tags are now mandatory and handle all quality checks
	// This maintains backward compatibility with callers, but all quality judgments use tags
	result := QualityCheckResult{
		PassesQualityGate:    true, // Default: pass - tags handle actual filtering
		IsValueTrap:          false,
		IsBubbleRisk:         false,
		BelowMinimumReturn:   false,
		QuantumValueTrapProb: 0.0,
		IsQuantumValueTrap:   false,
		IsQuantumWarning:     false,
		IsEnsembleValueTrap:  false,
	}

	// Tags are now mandatory - all quality checks are handled via tags (checked elsewhere)
	// REMOVED: All score-based quality check logic - tags are required for all quality gates
	return result // Tags will be checked elsewhere
}

// GetScoreFromContext safely retrieves a score from context maps by ISIN
// Exported for use in calculators that need direct score access
func GetScoreFromContext(ctx *domain.OpportunityContext, isin string, scoreMap map[string]float64) float64 {
	if scoreMap == nil || isin == "" {
		return 0.0
	}

	// Direct ISIN lookup - O(1) instead of O(n) iteration
	if score, hasScore := scoreMap[isin]; hasScore {
		return score
	}

	return 0.0
}
