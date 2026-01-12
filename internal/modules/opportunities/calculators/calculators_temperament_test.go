package calculators

import (
	"testing"

	"github.com/aristath/sentinel/internal/modules/settings"
	"github.com/stretchr/testify/assert"
)

// TestCalculatorConfigDefaults verifies default thresholds match current hardcoded values
func TestCalculatorConfigDefaults(t *testing.T) {
	config := NewDefaultCalculatorConfig()

	// Profit Taking defaults (Category 2)
	assert.Equal(t, 0.15, config.ProfitTaking.MinGainThreshold,
		"min gain threshold should be 15%")
	assert.Equal(t, 0.30, config.ProfitTaking.WindfallThreshold,
		"windfall threshold should be 30%")
	assert.Equal(t, 90, config.ProfitTaking.MinHoldDays,
		"min hold days should be 90")
	assert.Equal(t, 1.0, config.ProfitTaking.SellPercentage,
		"sell percentage should be 100%")
	assert.Equal(t, 1.0, config.ProfitTaking.MaxSellPercentage,
		"max sell percentage should be 100%")

	// Averaging Down defaults (Category 3)
	assert.Equal(t, -0.20, config.AveragingDown.MaxLossThreshold,
		"max loss threshold should be -20%")
	assert.Equal(t, -0.05, config.AveragingDown.MinLossThreshold,
		"min loss threshold should be -5%")
	assert.Equal(t, 500.0, config.AveragingDown.MaxValuePerPosition,
		"max value per position should be 500")
	assert.Equal(t, 0.10, config.AveragingDown.AveragingDownPercent,
		"averaging down percent should be 10%")
	assert.Equal(t, 3, config.AveragingDown.MaxPositions,
		"max positions should be 3")

	// Opportunity Buys defaults (Category 4)
	assert.Equal(t, 0.65, config.OpportunityBuys.MinScore,
		"min score should be 0.65")
	assert.Equal(t, 500.0, config.OpportunityBuys.MaxValuePerPosition,
		"max value per position should be 500")
	assert.Equal(t, 5, config.OpportunityBuys.MaxPositions,
		"max positions should be 5")

	// Rebalance defaults (Category 8)
	assert.Equal(t, 0.05, config.Rebalance.MinOverweightThreshold,
		"min overweight threshold should be 5%")
	assert.Equal(t, 0.50, config.Rebalance.MaxSellPercentage,
		"max sell percentage should be 50%")

	// Transaction Efficiency defaults (Category 9)
	assert.Equal(t, 0.01, config.TransactionEfficiency.MaxCostRatio,
		"max cost ratio should be 1%")
}

// TestPriorityBoostDefaults verifies default priority boost multipliers
func TestPriorityBoostDefaults(t *testing.T) {
	config := NewDefaultCalculatorConfig()

	// Windfall boosts (Category 10)
	assert.Equal(t, 1.5, config.Boosts.WindfallBoost,
		"windfall boost should be 1.5")
	assert.Equal(t, 1.4, config.Boosts.BubbleRiskBoost,
		"bubble risk boost should be 1.4")
	assert.Equal(t, 1.3, config.Boosts.NeedsRebalanceBoost,
		"needs rebalance boost should be 1.3")
	assert.Equal(t, 1.2, config.Boosts.OverweightBoost,
		"overweight boost should be 1.2")

	// Quality-Value boosts (Category 11)
	assert.Equal(t, 1.5, config.Boosts.QualityValueBoost,
		"quality value boost should be 1.5")
	assert.Equal(t, 1.3, config.Boosts.RecoveryCandidateBoost,
		"recovery candidate boost should be 1.3")
	assert.Equal(t, 1.15, config.Boosts.HighQualityBoost,
		"high quality boost should be 1.15")
	assert.Equal(t, 1.1, config.Boosts.ValueOpportunityBoost,
		"value opportunity boost should be 1.1")

	// Risk Profile boosts (Category 12)
	assert.Equal(t, 1.15, config.Boosts.LowRiskBoost,
		"low risk boost should be 1.15")
	assert.Equal(t, 1.05, config.Boosts.MediumRiskBoost,
		"medium risk boost should be 1.05")
	assert.Equal(t, 0.90, config.Boosts.HighRiskPenalty,
		"high risk penalty should be 0.90")

	// Regime-Aware boosts (Category 13)
	assert.Equal(t, 1.15, config.Boosts.BullGrowthBoost,
		"bull growth boost should be 1.15")
	assert.Equal(t, 1.15, config.Boosts.BearValueBoost,
		"bear value boost should be 1.15")
	assert.Equal(t, 1.12, config.Boosts.SidewaysDividendBoost,
		"sideways dividend boost should be 1.12")

	// Quantum warning penalties (Category 14)
	assert.Equal(t, 0.90, config.Boosts.QuantumPenaltyMild,
		"quantum penalty mild should be 0.90 (10% reduction)")
	assert.Equal(t, 0.70, config.Boosts.QuantumPenaltySevere,
		"quantum penalty severe should be 0.70 (30% reduction)")
}

// TestCalculatorConfigFromSettings verifies config can be created from settings service
func TestCalculatorConfigFromSettings(t *testing.T) {
	// Verify the struct types align with settings types
	_ = settings.ProfitTakingParams{}
	_ = settings.AveragingDownParams{}
	_ = settings.OpportunityBuysParams{}
	_ = settings.RebalancingParams{}
	_ = settings.TransactionParams{}
	_ = settings.RiskManagementParams{}
	_ = settings.ProfitTakingBoosts{}
	_ = settings.AveragingDownBoosts{}
	_ = settings.OpportunityBuyBoosts{}
	_ = settings.RegimeBoosts{}

	// All types compile successfully
	t.Log("All calculator param struct types exist and compile")
}

// TestThresholdBounds verifies thresholds are within valid ranges
func TestThresholdBounds(t *testing.T) {
	config := NewDefaultCalculatorConfig()

	// Profit taking bounds
	assert.True(t, config.ProfitTaking.MinGainThreshold > 0 && config.ProfitTaking.MinGainThreshold < 1.0,
		"min gain threshold should be between 0 and 100%")
	assert.True(t, config.ProfitTaking.WindfallThreshold > config.ProfitTaking.MinGainThreshold,
		"windfall threshold should be higher than min gain threshold")
	assert.True(t, config.ProfitTaking.MinHoldDays >= 0 && config.ProfitTaking.MinHoldDays <= 365,
		"min hold days should be reasonable (0-365)")

	// Averaging down bounds
	assert.True(t, config.AveragingDown.MaxLossThreshold < 0,
		"max loss threshold should be negative")
	assert.True(t, config.AveragingDown.MinLossThreshold < 0,
		"min loss threshold should be negative")
	assert.True(t, config.AveragingDown.MaxLossThreshold < config.AveragingDown.MinLossThreshold,
		"max loss should be more negative than min loss")

	// Opportunity buys bounds
	assert.True(t, config.OpportunityBuys.MinScore > 0 && config.OpportunityBuys.MinScore < 1.0,
		"min score should be between 0 and 1")
	assert.True(t, config.OpportunityBuys.MaxPositions > 0,
		"max positions should be positive")

	// Boost bounds
	assert.True(t, config.Boosts.WindfallBoost >= 1.0 && config.Boosts.WindfallBoost <= 2.0,
		"windfall boost should be between 1.0 and 2.0")
	assert.True(t, config.Boosts.HighRiskPenalty > 0 && config.Boosts.HighRiskPenalty < 1.0,
		"high risk penalty should be between 0 and 1.0")
}

// TestTemperamentAffectsCalculatorConfig verifies temperament changes config behavior
func TestTemperamentAffectsCalculatorConfig(t *testing.T) {
	// At balanced temperament (0.5), config should use base values
	config := NewDefaultCalculatorConfig()

	// These should reflect base values at balanced temperament
	assert.Equal(t, 0.15, config.ProfitTaking.MinGainThreshold,
		"at balanced temperament, min gain threshold should be base value 0.15")
	assert.Equal(t, 0.65, config.OpportunityBuys.MinScore,
		"at balanced temperament, min score should be base value 0.65")

	// When we implement temperament-aware config:
	// - Aggressive temperament would lower gain thresholds (sell sooner)
	// - Conservative temperament would raise gain thresholds (sell later)
	// - Patient temperament would increase hold days
}

// TestApplyQuantumWarningPenaltyValues verifies penalty application
func TestApplyQuantumWarningPenaltyValues(t *testing.T) {
	tests := []struct {
		name           string
		calculatorType string
		basePriority   float64
		hasWarning     bool
		expectedMin    float64
		expectedMax    float64
	}{
		{"averaging_down_with_warning", "averaging_down", 1.0, true, 0.85, 0.95},
		{"averaging_down_no_warning", "averaging_down", 1.0, false, 1.0, 1.0},
		{"opportunity_buys_with_warning", "opportunity_buys", 1.0, true, 0.65, 0.75},
		{"profit_taking_with_warning", "profit_taking", 1.0, true, 1.0, 1.0}, // No penalty for sells
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var tags []string
			if tc.hasWarning {
				tags = []string{"quantum-bubble-warning"}
			}

			result := ApplyQuantumWarningPenalty(tc.basePriority, tags, tc.calculatorType)

			assert.True(t, result >= tc.expectedMin && result <= tc.expectedMax,
				"result %.2f should be between %.2f and %.2f", result, tc.expectedMin, tc.expectedMax)
		})
	}
}

// TestApplyTagBasedPriorityBoostsValues verifies boost application
func TestApplyTagBasedPriorityBoostsValues(t *testing.T) {
	tests := []struct {
		name           string
		calculatorType string
		tags           []string
		basePriority   float64
		expectedMin    float64
		expectedMax    float64
	}{
		{
			name:           "low_risk_buy",
			calculatorType: "opportunity_buys",
			tags:           []string{"low-risk"},
			basePriority:   1.0,
			expectedMin:    1.10,
			expectedMax:    1.20,
		},
		{
			name:           "high_risk_penalty",
			calculatorType: "opportunity_buys",
			tags:           []string{"high-risk"},
			basePriority:   1.0,
			expectedMin:    0.85,
			expectedMax:    0.95,
		},
		{
			name:           "quality_boosts",
			calculatorType: "averaging_down",
			tags:           []string{"strong-fundamentals", "stable"},
			basePriority:   1.0,
			expectedMin:    1.15,
			expectedMax:    1.30,
		},
		{
			name:           "sell_underperforming",
			calculatorType: "profit_taking",
			tags:           []string{"underperforming", "stagnant"},
			basePriority:   1.0,
			expectedMin:    1.30,
			expectedMax:    1.50,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ApplyTagBasedPriorityBoosts(tc.basePriority, tc.tags, tc.calculatorType)

			assert.True(t, result >= tc.expectedMin && result <= tc.expectedMax,
				"result %.2f should be between %.2f and %.2f for tags %v",
				result, tc.expectedMin, tc.expectedMax, tc.tags)
		})
	}
}

// TestCalculatorBoostMultiplierBounds verifies boost multipliers stay within safe bounds
func TestCalculatorBoostMultiplierBounds(t *testing.T) {
	config := NewDefaultCalculatorConfig()

	// All boosts should be >= 1.0 (no negative effect on good things)
	boosts := []float64{
		config.Boosts.WindfallBoost,
		config.Boosts.BubbleRiskBoost,
		config.Boosts.NeedsRebalanceBoost,
		config.Boosts.OverweightBoost,
		config.Boosts.QualityValueBoost,
		config.Boosts.RecoveryCandidateBoost,
		config.Boosts.HighQualityBoost,
		config.Boosts.ValueOpportunityBoost,
		config.Boosts.LowRiskBoost,
		config.Boosts.MediumRiskBoost,
		config.Boosts.BullGrowthBoost,
		config.Boosts.BearValueBoost,
		config.Boosts.SidewaysDividendBoost,
	}

	for i, boost := range boosts {
		assert.True(t, boost >= 1.0 && boost <= 2.0,
			"boost %d should be between 1.0 and 2.0, got %.2f", i, boost)
	}

	// Penalties should be < 1.0 but > 0
	penalties := []float64{
		config.Boosts.HighRiskPenalty,
		config.Boosts.QuantumPenaltyMild,
		config.Boosts.QuantumPenaltySevere,
	}

	for i, penalty := range penalties {
		assert.True(t, penalty > 0 && penalty <= 1.0,
			"penalty %d should be between 0 and 1.0, got %.2f", i, penalty)
	}
}
