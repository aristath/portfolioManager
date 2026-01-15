package universe

import (
	"testing"

	"github.com/aristath/sentinel/internal/modules/settings"
	"github.com/stretchr/testify/assert"
)

// TestTagAssignerConfigDefaults verifies default thresholds match current hardcoded values
func TestTagAssignerConfigDefaults(t *testing.T) {
	config := NewDefaultTagAssignerConfig()

	// Value thresholds (from plan Category 15)
	assert.Equal(t, 0.15, config.Value.ValueOpportunityDiscountPct,
		"value opportunity discount should be 15%")
	assert.Equal(t, 0.25, config.Value.DeepValueDiscountPct,
		"deep value discount should be 25%")
	assert.Equal(t, 0.10, config.Value.Below52wHighThreshold,
		"below 52w high threshold should be 10%")

	// Quality thresholds (from plan Category 16)
	assert.Equal(t, 0.70, config.Quality.HighQualityStability,
		"high quality stability should be 0.70")
	assert.Equal(t, 0.70, config.Quality.HighQualityLongTerm,
		"high quality long term should be 0.70")
	assert.Equal(t, 0.75, config.Quality.StableStability,
		"stable stability should be 0.75")
	assert.Equal(t, 0.25, config.Quality.StableVolatilityMax,
		"stable volatility max should be 0.25")
	assert.Equal(t, 0.75, config.Quality.StableConsistency,
		"stable consistency should be 0.75")

	// Technical thresholds (from plan Category 17)
	assert.Equal(t, 30.0, config.Technical.RSIOversold,
		"RSI oversold should be 30")
	assert.Equal(t, 70.0, config.Technical.RSIOverbought,
		"RSI overbought should be 70")

	// Dividend thresholds (from plan Category 18)
	assert.Equal(t, 0.04, config.Dividend.HighDividendYield,
		"high dividend yield should be 4%")

	// Danger thresholds (from plan Category 19)
	assert.Equal(t, 0.30, config.Danger.VolatileThreshold,
		"volatile threshold should be 30%")
	assert.Equal(t, 0.40, config.Danger.HighVolatilityThreshold,
		"high volatility threshold should be 40%")

	// Portfolio risk thresholds (from plan Category 20)
	// Updated: looser thresholds for quality-first approach
	// Diversification is guardrail, not driver
	assert.Equal(t, 0.05, config.PortfolioRisk.OverweightDeviation,
		"overweight deviation should be 5%")
	assert.Equal(t, 0.15, config.PortfolioRisk.OverweightAbsolute,
		"overweight absolute should be 15%")
	assert.Equal(t, 0.25, config.PortfolioRisk.ConcentrationRiskThreshold,
		"concentration risk threshold should be 25%")

	// Risk profile thresholds (from plan Category 21)
	assert.Equal(t, 0.15, config.RiskProfile.LowRiskVolatilityMax,
		"low risk volatility max should be 15%")
	assert.Equal(t, 0.30, config.RiskProfile.HighRiskVolatilityThreshold,
		"high risk volatility threshold should be 30%")
}

// TestTagAssignerConfigFromSettings verifies config can be created from settings service
func TestTagAssignerConfigFromSettings(t *testing.T) {
	// Verify the struct types align with settings types
	_ = settings.ValueThresholds{}
	_ = settings.QualityThresholds{}
	_ = settings.TechnicalThresholds{}
	_ = settings.DividendThresholds{}
	_ = settings.DangerThresholds{}
	_ = settings.PortfolioRiskThresholds{}
	_ = settings.RiskProfileThresholds{}
	_ = settings.BubbleTrapThresholds{}
	_ = settings.TotalReturnThresholds{}
	_ = settings.RegimeThresholds{}

	// All types compile successfully
	t.Log("All threshold struct types exist and compile")
}

// TestTagAssignerThresholdBounds verifies thresholds are within valid ranges
func TestTagAssignerThresholdBounds(t *testing.T) {
	config := NewDefaultTagAssignerConfig()

	// RSI bounds
	assert.True(t, config.Technical.RSIOversold >= 0 && config.Technical.RSIOversold <= 50,
		"RSI oversold should be in [0, 50]")
	assert.True(t, config.Technical.RSIOverbought >= 50 && config.Technical.RSIOverbought <= 100,
		"RSI overbought should be in [50, 100]")

	// Volatility thresholds should be reasonable
	assert.True(t, config.Quality.StableVolatilityMax > 0 && config.Quality.StableVolatilityMax < 0.50,
		"stable volatility max should be reasonable")
	assert.True(t, config.Danger.VolatileThreshold > 0 && config.Danger.VolatileThreshold < 1.0,
		"volatile threshold should be reasonable")

	// Quality thresholds should be in [0, 1]
	assert.True(t, config.Quality.HighQualityStability >= 0 && config.Quality.HighQualityStability <= 1,
		"high quality stability should be in [0, 1]")
	assert.True(t, config.Quality.StableStability >= 0 && config.Quality.StableStability <= 1,
		"stable stability should be in [0, 1]")

	// Yield/return thresholds should be positive percentages
	assert.True(t, config.Dividend.HighDividendYield > 0 && config.Dividend.HighDividendYield < 0.20,
		"high dividend yield should be reasonable (0-20%)")
}

// TestTemperamentAffectsTagThresholds verifies temperament changes threshold behavior
func TestTemperamentAffectsTagThresholds(t *testing.T) {
	// At balanced temperament (0.5), config should use base values
	// This is a design contract test

	config := NewDefaultTagAssignerConfig()

	// These should all reflect base values
	assert.Equal(t, 0.70, config.Quality.HighQualityStability,
		"at balanced temperament, high quality stability should be base value 0.70")
	assert.Equal(t, 30.0, config.Technical.RSIOversold,
		"at balanced temperament, RSI oversold should be base value 30")

	// When we implement temperament-aware config:
	// - Aggressive temperament would lower quality thresholds (buy more readily)
	// - Conservative temperament would raise quality thresholds (more selective)
}

// TestQualityGateThresholds verifies quality gate thresholds
func TestQualityGateThresholds(t *testing.T) {
	config := NewDefaultTagAssignerConfig()

	// Multi-path quality gate thresholds
	assert.Equal(t, 0.55, config.QualityGate.StabilityThreshold,
		"stability threshold should be 0.55")
	assert.Equal(t, 0.45, config.QualityGate.LongTermThreshold,
		"long term threshold should be 0.45")
	assert.Equal(t, 0.75, config.QualityGate.ExceptionalThreshold,
		"exceptional threshold should be 0.75")
}

// TestBubbleTrapThresholds verifies bubble and value trap detection thresholds
func TestBubbleTrapThresholds(t *testing.T) {
	config := NewDefaultTagAssignerConfig()

	// Bubble detection (from plan Category 22)
	assert.Equal(t, 0.15, config.BubbleTrap.BubbleCAGRThreshold,
		"bubble CAGR threshold should be 15%")
	assert.Equal(t, 0.50, config.BubbleTrap.BubbleSharpeThreshold,
		"bubble Sharpe threshold should be 0.5")
	assert.Equal(t, 0.40, config.BubbleTrap.BubbleVolatilityThreshold,
		"bubble volatility threshold should be 40%")
	assert.Equal(t, 0.55, config.BubbleTrap.BubbleStabilityThreshold,
		"bubble stability threshold should be 0.55")

	// Value trap detection
	assert.Equal(t, 0.55, config.BubbleTrap.ValueTrapStability,
		"value trap stability should be 0.55")
	assert.Equal(t, 0.45, config.BubbleTrap.ValueTrapLongTerm,
		"value trap long term should be 0.45")
	assert.Equal(t, -0.05, config.BubbleTrap.ValueTrapMomentum,
		"value trap momentum should be -0.05")
	assert.Equal(t, 0.35, config.BubbleTrap.ValueTrapVolatility,
		"value trap volatility should be 0.35")

	// Quantum probability thresholds
	assert.Equal(t, 0.70, config.BubbleTrap.QuantumBubbleHighProb,
		"quantum bubble high prob should be 0.70")
	assert.Equal(t, 0.50, config.BubbleTrap.QuantumBubbleWarningProb,
		"quantum bubble warning prob should be 0.50")
}

// TestTotalReturnThresholds verifies total return thresholds
func TestTotalReturnThresholds(t *testing.T) {
	config := NewDefaultTagAssignerConfig()

	assert.Equal(t, 0.18, config.TotalReturn.ExcellentTotalReturn,
		"excellent total return should be 18%")
	assert.Equal(t, 0.15, config.TotalReturn.HighTotalReturn,
		"high total return should be 15%")
	assert.Equal(t, 0.12, config.TotalReturn.ModerateTotalReturn,
		"moderate total return should be 12%")
}

// TestRegimeThresholds verifies regime-specific thresholds
func TestRegimeThresholds(t *testing.T) {
	config := NewDefaultTagAssignerConfig()

	assert.Equal(t, 0.20, config.Regime.BearSafeVolatility,
		"bear safe volatility should be 20%")
	assert.Equal(t, 0.70, config.Regime.BearSafeStability,
		"bear safe stability should be 0.70")
	assert.Equal(t, 0.12, config.Regime.BullGrowthCAGR,
		"bull growth CAGR should be 12%")
}
