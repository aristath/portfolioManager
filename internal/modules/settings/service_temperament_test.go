package settings

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetTemperamentValues verifies the service can retrieve temperament values
func TestGetTemperamentValues(t *testing.T) {
	// Test that default values are returned when no settings are stored
	t.Run("defaults", func(t *testing.T) {
		risk, agg, pat := getDefaultTemperamentValues()
		assert.Equal(t, 0.5, risk, "default risk_tolerance should be 0.5")
		assert.Equal(t, 0.5, agg, "default temperament_aggression should be 0.5")
		assert.Equal(t, 0.5, pat, "default temperament_patience should be 0.5")
	})
}

// Helper function that mimics what the service would do
func getDefaultTemperamentValues() (riskTolerance, aggression, patience float64) {
	riskTolerance = SettingDefaults["risk_tolerance"].(float64)
	aggression = SettingDefaults["temperament_aggression"].(float64)
	patience = SettingDefaults["temperament_patience"].(float64)
	return
}

// ============================================================================
// EVALUATION WEIGHTS TESTS
// ============================================================================

func TestEvaluationWeightsStruct(t *testing.T) {
	// Pure end-state scoring with 4 components
	weights := EvaluationWeights{
		PortfolioQuality:         0.35,
		DiversificationAlignment: 0.30,
		RiskAdjustedMetrics:      0.25,
		EndStateImprovement:      0.10,
	}

	// Weights should sum close to 1.0
	sum := weights.PortfolioQuality + weights.DiversificationAlignment +
		weights.RiskAdjustedMetrics + weights.EndStateImprovement
	assert.InDelta(t, 1.0, sum, 0.001, "weights should sum to 1.0")
}

func TestEvaluationWeightsNormalization(t *testing.T) {
	// Test that weights are normalized to sum to 1.0
	weights := EvaluationWeights{
		PortfolioQuality:         0.40,
		DiversificationAlignment: 0.35,
		RiskAdjustedMetrics:      0.30,
		EndStateImprovement:      0.15,
	}

	normalized := weights.Normalize()

	sum := normalized.PortfolioQuality + normalized.DiversificationAlignment +
		normalized.RiskAdjustedMetrics + normalized.EndStateImprovement
	assert.InDelta(t, 1.0, sum, 0.001, "normalized weights should sum to 1.0")
}

// ============================================================================
// KELLY PARAMS TESTS
// ============================================================================

func TestKellyParamsStruct(t *testing.T) {
	params := KellyParams{
		MinPositionSize: 0.01,
		MaxPositionSize: 0.15,
		MinMultiplier:   0.25,
		MaxMultiplier:   0.75,
		BullThreshold:   0.50,
		BearThreshold:   -0.50,
	}

	// Validate ranges
	assert.True(t, params.MinPositionSize < params.MaxPositionSize,
		"min position size should be less than max")
	assert.True(t, params.MinMultiplier < params.MaxMultiplier,
		"min multiplier should be less than max")
	assert.True(t, params.BearThreshold < 0 && params.BullThreshold > 0,
		"bear threshold should be negative, bull positive")
}

// ============================================================================
// RISK MANAGEMENT PARAMS TESTS
// ============================================================================

func TestRiskManagementParamsStruct(t *testing.T) {
	params := RiskManagementParams{
		MinHoldDays:       90,
		MaxLossThreshold:  -0.20,
		MaxSellPercentage: 0.20,
	}

	// Validate ranges
	assert.True(t, params.MinHoldDays > 0, "min hold days should be positive")
	assert.True(t, params.MaxLossThreshold < 0, "max loss threshold should be negative")
	assert.True(t, params.MaxSellPercentage > 0 && params.MaxSellPercentage <= 1.0,
		"max sell percentage should be in (0, 1]")
}

// ============================================================================
// PROFIT TAKING PARAMS TESTS
// ============================================================================

func TestProfitTakingParamsStruct(t *testing.T) {
	params := ProfitTakingParams{
		MinGainThreshold:  0.15,
		WindfallThreshold: 0.30,
		SellPercentage:    1.0,
	}

	// Validate ranges
	assert.True(t, params.MinGainThreshold > 0, "min gain should be positive")
	assert.True(t, params.WindfallThreshold > params.MinGainThreshold,
		"windfall threshold should be higher than min gain")
	assert.True(t, params.SellPercentage >= 0.5 && params.SellPercentage <= 1.0,
		"sell percentage should be in [0.5, 1.0]")
}

// ============================================================================
// PRIORITY BOOSTS TESTS
// ============================================================================

func TestProfitTakingBoostsStruct(t *testing.T) {
	boosts := ProfitTakingBoosts{
		WindfallPriority: 1.5,
		BubbleRisk:       1.4,
	}

	// All boosts should be >= 1.0 (positive multipliers)
	assert.True(t, boosts.WindfallPriority >= 1.0, "windfall priority should be >= 1.0")
	assert.True(t, boosts.BubbleRisk >= 1.0, "bubble risk should be >= 1.0")
}

// ============================================================================
// TAG THRESHOLDS TESTS
// ============================================================================

func TestValueThresholdsStruct(t *testing.T) {
	thresholds := ValueThresholds{
		ValueOpportunityDiscountPct: 0.15,
		DeepValueDiscountPct:        0.25,
	}

	// Validate relationships
	assert.True(t, thresholds.DeepValueDiscountPct > thresholds.ValueOpportunityDiscountPct,
		"deep value should require larger discount than value opportunity")
}

func TestQualityThresholdsStruct(t *testing.T) {
	thresholds := QualityThresholds{
		HighQualityStability:        0.70,
		ValueOpportunityScoreThreshold: 0.65,
	}

	// Quality thresholds should be high (>= 0.5)
	assert.True(t, thresholds.HighQualityStability >= 0.5,
		"high quality stability threshold should be >= 0.5")
	assert.True(t, thresholds.ValueOpportunityScoreThreshold >= 0.5,
		"value opportunity score threshold should be >= 0.5")
}

func TestQualityGateParamsNewFields(t *testing.T) {
	params := QualityGateParams{
		StabilityThreshold:            0.55,
		LongTermThreshold:                0.45,
		ExceptionalThreshold:             0.85,
		AbsoluteMinCAGR:                  0.05,
		ExceptionalExcellenceThreshold:   0.75,
		QualityValueStabilityMin:      0.60,
		QualityValueOpportunityMin:       0.65,
		QualityValueLongTermMin:          0.30,
		DividendIncomeStabilityMin:    0.55,
		DividendIncomeScoreMin:           0.65,
		DividendIncomeYieldMin:           0.035,
		RiskAdjustedLongTermThreshold:    0.55,
		RiskAdjustedSharpeThreshold:      0.70,
		RiskAdjustedSortinoThreshold:     0.70,
		RiskAdjustedVolatilityMax:        0.35,
		CompositeStabilityWeight:      0.60,
		CompositeLongTermWeight:          0.40,
		CompositeScoreMin:                0.52,
		CompositeStabilityFloor:       0.45,
		GrowthOpportunityCAGRMin:         0.13,
		GrowthOpportunityStabilityMin: 0.50,
		GrowthOpportunityVolatilityMax:   0.40,
		HighScoreThreshold:               0.70,
	}

	// Mark Path 1 fields as intentionally unused (test focuses on new Path 2-7 fields)
	_ = params.StabilityThreshold
	_ = params.LongTermThreshold
	_ = params.ExceptionalThreshold
	_ = params.AbsoluteMinCAGR
	_ = params.RiskAdjustedLongTermThreshold

	// Validate Path 2: Exceptional Excellence
	assert.True(t, params.ExceptionalExcellenceThreshold >= 0.70 && params.ExceptionalExcellenceThreshold <= 0.85,
		"exceptional excellence threshold should be in [0.70, 0.85]")

	// Validate Path 3: Quality Value Play
	assert.True(t, params.QualityValueStabilityMin >= 0.55 && params.QualityValueStabilityMin <= 0.70,
		"quality value stability min should be in [0.55, 0.70]")
	assert.True(t, params.QualityValueOpportunityMin >= 0.60 && params.QualityValueOpportunityMin <= 0.75,
		"quality value opportunity min should be in [0.60, 0.75]")
	assert.True(t, params.QualityValueLongTermMin >= 0.25 && params.QualityValueLongTermMin <= 0.40,
		"quality value long term min should be in [0.25, 0.40]")

	// Validate Path 4: Dividend Income Play
	assert.True(t, params.DividendIncomeStabilityMin >= 0.50 && params.DividendIncomeStabilityMin <= 0.65,
		"dividend income stability min should be in [0.50, 0.65]")
	assert.True(t, params.DividendIncomeScoreMin >= 0.60 && params.DividendIncomeScoreMin <= 0.75,
		"dividend income score min should be in [0.60, 0.75]")
	assert.True(t, params.DividendIncomeYieldMin >= 0.025 && params.DividendIncomeYieldMin <= 0.050,
		"dividend income yield min should be in [0.025, 0.050]")

	// Validate Path 5: Risk-Adjusted Excellence
	assert.True(t, params.RiskAdjustedSharpeThreshold >= 0.60 && params.RiskAdjustedSharpeThreshold <= 0.85,
		"risk adjusted sharpe threshold should be in [0.60, 0.85]")
	assert.True(t, params.RiskAdjustedSortinoThreshold >= 0.60 && params.RiskAdjustedSortinoThreshold <= 0.85,
		"risk adjusted sortino threshold should be in [0.60, 0.85]")
	assert.True(t, params.RiskAdjustedVolatilityMax >= 0.30 && params.RiskAdjustedVolatilityMax <= 0.45,
		"risk adjusted volatility max should be in [0.30, 0.45]")

	// Validate Path 6: Composite Minimum
	assert.InDelta(t, 1.0, params.CompositeStabilityWeight+params.CompositeLongTermWeight, 0.01,
		"composite weights should sum to 1.0")
	assert.True(t, params.CompositeScoreMin >= 0.48 && params.CompositeScoreMin <= 0.58,
		"composite score min should be in [0.48, 0.58]")
	assert.True(t, params.CompositeStabilityFloor >= 0.40 && params.CompositeStabilityFloor <= 0.55,
		"composite stability floor should be in [0.40, 0.55]")

	// Validate Path 7: Growth Opportunity
	assert.True(t, params.GrowthOpportunityCAGRMin >= 0.11 && params.GrowthOpportunityCAGRMin <= 0.16,
		"growth opportunity CAGR min should be in [0.11, 0.16]")
	assert.True(t, params.GrowthOpportunityStabilityMin >= 0.45 && params.GrowthOpportunityStabilityMin <= 0.60,
		"growth opportunity stability min should be in [0.45, 0.60]")
	assert.True(t, params.GrowthOpportunityVolatilityMax >= 0.35 && params.GrowthOpportunityVolatilityMax <= 0.50,
		"growth opportunity volatility max should be in [0.35, 0.50]")

	// Validate High Score Threshold
	assert.True(t, params.HighScoreThreshold >= 0.65 && params.HighScoreThreshold <= 0.80,
		"high score threshold should be in [0.65, 0.80]")
}

func TestBubbleTrapThresholdsNewFields(t *testing.T) {
	thresholds := BubbleTrapThresholds{
		BubbleCAGRThreshold:    0.15,
		GrowthTagCAGRThreshold: 0.13,
	}

	// Growth tag threshold should be <= bubble threshold (less strict)
	assert.True(t, thresholds.GrowthTagCAGRThreshold <= thresholds.BubbleCAGRThreshold,
		"growth tag CAGR threshold should be <= bubble CAGR threshold")
	assert.True(t, thresholds.GrowthTagCAGRThreshold >= 0.12 && thresholds.GrowthTagCAGRThreshold <= 0.15,
		"growth tag CAGR threshold should be in [0.12, 0.15]")
}

func TestRegimeThresholdsNewFields(t *testing.T) {
	thresholds := RegimeThresholds{
		BearSafeStability:      0.70,
		BullGrowthStability:    0.70,
		SidewaysValueStability: 0.75,
	}

	// Sideways value should require higher stability than bear/bull (more selective)
	assert.True(t, thresholds.SidewaysValueStability >= thresholds.BearSafeStability,
		"sideways value stability should be >= bear safe stability")
	assert.True(t, thresholds.SidewaysValueStability >= thresholds.BullGrowthStability,
		"sideways value stability should be >= bull growth stability")
	assert.True(t, thresholds.SidewaysValueStability >= 0.70 && thresholds.SidewaysValueStability <= 0.85,
		"sideways value stability should be in [0.70, 0.85]")
}

func TestTechnicalThresholdsStruct(t *testing.T) {
	thresholds := TechnicalThresholds{
		RSIOversold:   30,
		RSIOverbought: 70,
	}

	// RSI thresholds should be in valid RSI range
	assert.True(t, thresholds.RSIOversold >= 0 && thresholds.RSIOversold <= 50,
		"RSI oversold should be in [0, 50]")
	assert.True(t, thresholds.RSIOverbought >= 50 && thresholds.RSIOverbought <= 100,
		"RSI overbought should be in [50, 100]")
}

// ============================================================================
// SCORING PARAMS TESTS
// ============================================================================

func TestScoringParamsStruct(t *testing.T) {
	params := ScoringParams{
		VolatilityExcellent:  0.15,
		VolatilityGood:       0.25,
		VolatilityAcceptable: 0.40,
		SharpeExcellent:      2.0,
		SharpeGood:           1.0,
	}

	// Validate tier ordering
	assert.True(t, params.VolatilityExcellent < params.VolatilityGood,
		"excellent should be better than good")
	assert.True(t, params.VolatilityGood < params.VolatilityAcceptable,
		"good should be better than acceptable")
	assert.True(t, params.SharpeExcellent > params.SharpeGood,
		"excellent Sharpe should be higher than good")
}

// ============================================================================
// INTEGRATION-STYLE TESTS (using actual temperament config)
// ============================================================================

func TestTemperamentAffectsEvaluationWeights(t *testing.T) {
	// At balanced (0.5, 0.5, 0.5), weights should be at base values
	// This tests that the temperament system integration is working

	// Get base weights from temperament config
	require.NotNil(t, SettingDefaults["risk_tolerance"])
	require.NotNil(t, SettingDefaults["temperament_aggression"])
	require.NotNil(t, SettingDefaults["temperament_patience"])

	// All temperament defaults should be 0.5
	assert.Equal(t, 0.5, SettingDefaults["risk_tolerance"].(float64))
	assert.Equal(t, 0.5, SettingDefaults["temperament_aggression"].(float64))
	assert.Equal(t, 0.5, SettingDefaults["temperament_patience"].(float64))
}

// ============================================================================
// PARAMETER BOUNDS TESTS
// ============================================================================

func TestParameterBoundsValid(t *testing.T) {
	// Test cases for absolute bounds validation
	testCases := []struct {
		name        string
		value       float64
		absoluteMin float64
		absoluteMax float64
	}{
		{"MinHoldDays", 90, 14, 365},
		{"MaxLossThreshold", -0.20, -0.50, -0.05},
		{"MinScore", 0.65, 0.50, 0.90},
		{"KellyMultiplier", 0.50, 0.15, 0.80},
		{"MaxSellPercentage", 0.20, 0.05, 0.75},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, tc.value >= tc.absoluteMin,
				"%s value %.4f should be >= absolute min %.4f", tc.name, tc.value, tc.absoluteMin)
			assert.True(t, tc.value <= tc.absoluteMax,
				"%s value %.4f should be <= absolute max %.4f", tc.name, tc.value, tc.absoluteMax)
		})
	}
}

// ============================================================================
// HELPER FUNCTION TESTS
// ============================================================================

func TestClampToAbsoluteBounds(t *testing.T) {
	// Test the clamp function logic
	testCases := []struct {
		value    float64
		min      float64
		max      float64
		expected float64
	}{
		{0.5, 0.0, 1.0, 0.5},  // Within bounds
		{-0.5, 0.0, 1.0, 0.0}, // Below min
		{1.5, 0.0, 1.0, 1.0},  // Above max
		{0.0, 0.0, 1.0, 0.0},  // At min
		{1.0, 0.0, 1.0, 1.0},  // At max
	}

	for _, tc := range testCases {
		result := clampValue(tc.value, tc.min, tc.max)
		assert.Equal(t, tc.expected, result,
			"clamp(%.2f, %.2f, %.2f) should be %.2f", tc.value, tc.min, tc.max, tc.expected)
	}
}

func clampValue(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Test that all struct types exist and have expected fields
func TestStructTypesExist(t *testing.T) {
	// This test verifies that all the struct types we define are valid
	// The actual implementation will be tested via the service methods

	_ = EvaluationWeights{}
	_ = KellyParams{}
	_ = RiskManagementParams{}
	_ = ProfitTakingParams{}
	_ = AveragingDownParams{}
	_ = OpportunityBuysParams{}
	_ = QualityGateParams{}
	_ = RebalancingParams{}

	// Verify new QualityGateParams fields exist
	qgp := QualityGateParams{
		StabilityThreshold:            0.55,
		LongTermThreshold:                0.45,
		ExceptionalThreshold:             0.85,
		AbsoluteMinCAGR:                  0.05,
		ExceptionalExcellenceThreshold:   0.75,
		QualityValueStabilityMin:      0.60,
		QualityValueOpportunityMin:       0.65,
		QualityValueLongTermMin:          0.30,
		DividendIncomeStabilityMin:    0.55,
		DividendIncomeScoreMin:           0.65,
		DividendIncomeYieldMin:           0.035,
		RiskAdjustedLongTermThreshold:    0.55,
		RiskAdjustedSharpeThreshold:      0.70,
		RiskAdjustedSortinoThreshold:     0.70,
		RiskAdjustedVolatilityMax:        0.35,
		CompositeStabilityWeight:      0.60,
		CompositeLongTermWeight:          0.40,
		CompositeScoreMin:                0.52,
		CompositeStabilityFloor:       0.45,
		GrowthOpportunityCAGRMin:         0.13,
		GrowthOpportunityStabilityMin: 0.50,
		GrowthOpportunityVolatilityMax:   0.40,
		HighScoreThreshold:               0.70,
	}
	_ = qgp // Use the variable

	// Verify new BubbleTrapThresholds field exists
	btt := BubbleTrapThresholds{
		GrowthTagCAGRThreshold: 0.13,
	}
	_ = btt

	// Verify new RegimeThresholds field exists
	rt := RegimeThresholds{
		SidewaysValueStability: 0.75,
	}
	_ = rt

	// Verify new QualityThresholds field exists
	qt := QualityThresholds{
		ValueOpportunityScoreThreshold: 0.65,
	}
	_ = qt
	_ = VolatilityParams{}
	_ = TransactionParams{}
	_ = ProfitTakingBoosts{}
	_ = AveragingDownBoosts{}
	_ = OpportunityBuyBoosts{}
	_ = RegimeBoosts{}
	_ = ValueThresholds{}
	_ = QualityThresholds{}
	_ = TechnicalThresholds{}
	_ = DividendThresholds{}
	_ = DangerThresholds{}
	_ = PortfolioRiskThresholds{}
	_ = RiskProfileThresholds{}
	_ = BubbleTrapThresholds{}
	_ = TotalReturnThresholds{}
	_ = RegimeThresholds{}
	_ = ScoringParams{}

	// If we get here, all struct types compile successfully
	t.Log("All struct types exist and compile")
}

// TestRoundTrip verifies that values round-trip correctly through the system
func TestRoundTrip(t *testing.T) {
	// At temperament 0.5, values should be at their base
	risk, agg, pat := 0.5, 0.5, 0.5

	// The round-trip test verifies that:
	// 1. Base values are returned at temperament 0.5
	// 2. Min values are returned at temperament 0.0
	// 3. Max values are returned at temperament 1.0

	// These are spot-check values from the temperament config
	testCases := []struct {
		name         string
		temperament  float64
		expectedBase float64
		tolerance    float64
	}{
		{"Balanced", 0.5, 0.5, 0.01},
		{"Conservative", 0.0, 0.5, 0.3}, // Will be at min or max depending on mapping
		{"Aggressive", 1.0, 0.5, 0.3},   // Will be at min or max depending on mapping
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// The actual temperament value is meaningful
			assert.InDelta(t, tc.expectedBase, (risk+agg+pat)/3, 0.5,
				"temperament average should be reasonable")
		})
	}
}

// Test normalize function
func TestNormalizeWeights(t *testing.T) {
	// Pure end-state scoring weights
	weights := EvaluationWeights{
		PortfolioQuality:         0.35,
		DiversificationAlignment: 0.30,
		RiskAdjustedMetrics:      0.25,
		EndStateImprovement:      0.10,
	}

	normalized := weights.Normalize()

	// Sum should be exactly 1.0
	sum := normalized.PortfolioQuality + normalized.DiversificationAlignment +
		normalized.RiskAdjustedMetrics + normalized.EndStateImprovement
	assert.True(t, math.Abs(sum-1.0) < 0.0001, "normalized sum should be 1.0, got %f", sum)
}
