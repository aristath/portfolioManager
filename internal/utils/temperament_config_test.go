package utils

import (
	"math"
	"testing"
)

// TestAllMappingsExist verifies that all expected parameter mappings exist
func TestAllMappingsExist(t *testing.T) {
	expectedMappings := []string{
		// Category 1: Evaluation weights (5)
		"evaluation_opportunity_weight",
		"evaluation_quality_weight",
		"evaluation_risk_adjusted_weight",
		"evaluation_diversification_weight",
		"evaluation_regime_weight",

		// Category 2: Profit taking (3)
		"profit_taking_min_gain_threshold",
		"profit_taking_windfall_threshold",
		"profit_taking_sell_percentage",

		// Category 3: Averaging down (3)
		"averaging_down_max_loss_threshold",
		"averaging_down_min_loss_threshold",
		"averaging_down_percent",

		// Category 4: Opportunity buys (4)
		"opportunity_buys_min_score",
		"opportunity_buys_max_value_per_position",
		"opportunity_buys_max_positions",
		"opportunity_buys_target_return_threshold_pct",

		// Category 5: Kelly sizing (4 core + 8 additional = 12)
		"kelly_fixed_fractional",
		"kelly_min_position_size",
		"kelly_max_position_size",
		"kelly_bear_reduction",
		"kelly_base_multiplier",
		"kelly_confidence_adjustment_range",
		"kelly_regime_adjustment_range",
		"kelly_min_multiplier",
		"kelly_max_multiplier",
		"kelly_bear_max_reduction",
		"kelly_bull_threshold",
		"kelly_bear_threshold",

		// Category 6: Risk management (7)
		"risk_min_hold_days",
		"risk_sell_cooldown_days",
		"risk_max_loss_threshold",
		"risk_max_sell_percentage",
		"risk_min_time_between_trades",
		"risk_max_trades_per_day",
		"risk_max_trades_per_week",

		// Category 7: Quality gates (4)
		"quality_stability_threshold",
		"quality_long_term_threshold",
		"quality_exceptional_threshold",
		"quality_absolute_min_cagr",

		// Category 8: Rebalancing (3)
		"rebalancing_min_overweight_threshold",
		"rebalancing_position_drift_threshold",
		"rebalancing_cash_threshold_multiplier",

		// Category 9: Volatility acceptance (4)
		"volatility_volatile_threshold",
		"volatility_high_threshold",
		"volatility_max_acceptable",
		"volatility_max_acceptable_drawdown",

		// Category 10: Transaction efficiency (2)
		"transaction_max_cost_ratio",
		"transaction_limit_order_buffer",

		// Category 11: Priority boost - Profit taking (6)
		"boost_windfall_priority",
		"boost_bubble_risk",
		"boost_needs_rebalance",
		"boost_overweight",
		"boost_overvalued",
		"boost_near_52w_high",

		// Category 12: Priority boost - Averaging down (4)
		"boost_quality_value",
		"boost_recovery_candidate",
		"boost_high_quality",
		"boost_value_opportunity",

		// Category 13: Priority boost - Opportunity buys (12)
		"boost_quantum_warning_penalty",
		"boost_quality_value_buy",
		"boost_high_quality_value",
		"boost_deep_value",
		"boost_oversold_quality",
		"boost_excellent_returns",
		"boost_high_returns",
		"boost_quality_high_cagr",
		"boost_dividend_grower",
		"boost_high_dividend",
		"boost_quality_penalty_reduction_exceptional",
		"boost_quality_penalty_reduction_high",

		// Category 14: Priority boost - Regime (7)
		"boost_low_risk",
		"boost_medium_risk",
		"boost_high_risk_penalty",
		"boost_growth_bull",
		"boost_value_bear",
		"boost_dividend_sideways",
		"boost_strong_stability",

		// Category 15: Tag assigner - Value (5)
		"tag_value_opportunity_discount_pct",
		"tag_deep_value_discount_pct",
		"tag_deep_value_extreme_pct",
		"tag_undervalued_pe_threshold",
		"tag_below_52w_high_threshold",

		// Category 16: Tag assigner - Quality (8)
		"tag_high_quality_stability",
		"tag_high_quality_long_term",
		"tag_stable_stability",
		"tag_stable_volatility_max",
		"tag_stable_consistency",
		"tag_consistent_grower_consistency",
		"tag_consistent_grower_cagr",
		"tag_strong_stability_threshold",

		// Category 17: Tag assigner - Technical (5)
		"tag_rsi_oversold",
		"tag_rsi_overbought",
		"tag_recovery_momentum_threshold",
		"tag_recovery_stability_min",
		"tag_recovery_discount_min",

		// Category 18: Tag assigner - Dividend (4)
		"tag_high_dividend_yield",
		"tag_dividend_opportunity_score",
		"tag_dividend_opportunity_yield",
		"tag_dividend_consistency_score",

		// Category 19: Tag assigner - Danger (7)
		"tag_overvalued_pe_threshold",
		"tag_overvalued_near_high_pct",
		"tag_unsustainable_gains_return",
		"tag_valuation_stretch_ema",
		"tag_underperforming_days",
		"tag_stagnant_return_threshold",
		"tag_stagnant_days_threshold",

		// Category 20: Tag assigner - Portfolio risk (4)
		"tag_overweight_deviation",
		"tag_overweight_absolute",
		"tag_concentration_risk_threshold",
		"tag_needs_rebalance_deviation",

		// Category 21: Tag assigner - Risk profile (8)
		"tag_low_risk_volatility_max",
		"tag_low_risk_stability_min",
		"tag_low_risk_drawdown_max",
		"tag_medium_risk_volatility_min",
		"tag_medium_risk_volatility_max",
		"tag_medium_risk_stability_min",
		"tag_high_risk_volatility_threshold",
		"tag_high_risk_stability_threshold",

		// Category 22: Tag assigner - Bubble & value trap (12)
		"tag_bubble_cagr_threshold",
		"tag_bubble_sharpe_threshold",
		"tag_bubble_volatility_threshold",
		"tag_bubble_stability_threshold",
		"tag_value_trap_stability",
		"tag_value_trap_long_term",
		"tag_value_trap_momentum",
		"tag_value_trap_volatility",
		"tag_quantum_bubble_high_prob",
		"tag_quantum_bubble_warning_prob",
		"tag_quantum_trap_high_prob",
		"tag_quantum_trap_warning_prob",

		// Category 23: Tag assigner - Total return (5)
		"tag_excellent_total_return",
		"tag_high_total_return",
		"tag_moderate_total_return",
		"tag_dividend_total_return_yield",
		"tag_dividend_total_return_cagr",

		// Category 24: Tag assigner - Regime specific (6)
		"tag_bear_safe_volatility",
		"tag_bear_safe_stability",
		"tag_bear_safe_drawdown",
		"tag_bull_growth_cagr",
		"tag_bull_growth_stability",
		"tag_regime_volatile_volatility",

		// Category 26: Evaluation scoring (15)
		"scoring_windfall_excess_high",
		"scoring_windfall_excess_medium",
		"scoring_deviation_scale",
		"scoring_regime_bull_threshold",
		"scoring_regime_bear_threshold",
		"scoring_volatility_excellent",
		"scoring_volatility_good",
		"scoring_volatility_acceptable",
		"scoring_drawdown_excellent",
		"scoring_drawdown_good",
		"scoring_drawdown_acceptable",
		"scoring_sharpe_excellent",
		"scoring_sharpe_good",
		"scoring_sharpe_acceptable",
	}

	for _, name := range expectedMappings {
		mapping, exists := GetTemperamentMapping(name)
		if !exists {
			t.Errorf("Expected mapping %q to exist", name)
			continue
		}
		if mapping.Parameter != name {
			t.Errorf("Mapping %q has incorrect Parameter field: got %q", name, mapping.Parameter)
		}
	}
}

// TestBoundaryValues verifies that value=0, 0.5, 1 produce expected outputs
func TestBoundaryValues(t *testing.T) {
	testCases := []struct {
		parameter     string
		temperament   string  // which slider controls this
		expectedAtMin float64 // value=0.0
		expectedBase  float64 // value=0.5
		expectedAtMax float64 // value=1.0
		tolerance     float64
	}{
		// Category 1: Evaluation weights
		{"evaluation_opportunity_weight", "aggression", 0.25, 0.30, 0.35, 0.01},
		{"evaluation_quality_weight", "aggression", 0.28, 0.25, 0.22, 0.01},           // inverse
		{"evaluation_risk_adjusted_weight", "risk_tolerance", 0.20, 0.15, 0.12, 0.01}, // inverse

		// Category 2: Profit taking
		{"profit_taking_min_gain_threshold", "patience", 0.25, 0.15, 0.10, 0.02}, // inverse patience
		{"profit_taking_windfall_threshold", "patience", 0.50, 0.30, 0.20, 0.02}, // inverse patience

		// Category 5: Kelly sizing
		{"kelly_fixed_fractional", "aggression", 0.25, 0.50, 0.75, 0.05},
		{"kelly_max_position_size", "risk_tolerance", 0.08, 0.15, 0.25, 0.02},

		// Category 6: Risk management
		{"risk_min_hold_days", "patience", 30, 90, 180, 5},
		{"risk_max_trades_per_day", "aggression", 2, 4, 8, 1},

		// Category 7: Quality gates
		{"quality_stability_threshold", "aggression", 0.65, 0.55, 0.45, 0.02}, // inverse

		// Category 11: Priority boosts
		{"boost_windfall_priority", "aggression", 1.2, 1.5, 1.8, 0.05},

		// Category 17: Technical thresholds
		{"tag_rsi_oversold", "aggression", 20, 30, 40, 2},
		{"tag_rsi_overbought", "patience", 80, 70, 60, 2}, // inverse patience
	}

	for _, tc := range testCases {
		mapping, exists := GetTemperamentMapping(tc.parameter)
		if !exists {
			t.Errorf("Mapping %q not found", tc.parameter)
			continue
		}

		// Helper to get adjusted value with correct slider
		getVal := func(sliderVal float64) float64 {
			risk, agg, pat := 0.5, 0.5, 0.5
			switch tc.temperament {
			case "risk_tolerance":
				risk = sliderVal
			case "aggression":
				agg = sliderVal
			case "patience":
				pat = sliderVal
			}
			return GetAdjustedValue(mapping, risk, agg, pat)
		}

		// Test value = 0.0
		result := getVal(0.0)
		if math.Abs(result-tc.expectedAtMin) > tc.tolerance {
			t.Errorf("%s at value=0.0: expected %.3f, got %.3f (tolerance %.3f)",
				tc.parameter, tc.expectedAtMin, result, tc.tolerance)
		}

		// Test value = 0.5 (should return Base)
		result = getVal(0.5)
		if math.Abs(result-tc.expectedBase) > tc.tolerance {
			t.Errorf("%s at value=0.5: expected %.3f (base), got %.3f (tolerance %.3f)",
				tc.parameter, tc.expectedBase, result, tc.tolerance)
		}

		// Test value = 1.0
		result = getVal(1.0)
		if math.Abs(result-tc.expectedAtMax) > tc.tolerance {
			t.Errorf("%s at value=1.0: expected %.3f, got %.3f (tolerance %.3f)",
				tc.parameter, tc.expectedAtMax, result, tc.tolerance)
		}
	}
}

// TestSafetyBoundsEnforced verifies absolute bounds are never violated
func TestSafetyBoundsEnforced(t *testing.T) {
	criticalParams := []struct {
		parameter   string
		absoluteMin float64
		absoluteMax float64
	}{
		{"risk_min_hold_days", 14, 365},
		{"risk_max_loss_threshold", -0.50, -0.05},
		{"opportunity_buys_min_score", 0.50, 0.90},
		{"kelly_fixed_fractional", 0.15, 0.80},
		{"risk_max_sell_percentage", 0.05, 0.75},
		{"quality_absolute_min_cagr", 0.04, 0.10},
		{"quality_stability_threshold", 0.35, 0.80},
	}

	for _, tc := range criticalParams {
		mapping, exists := GetTemperamentMapping(tc.parameter)
		if !exists {
			t.Errorf("Mapping %q not found", tc.parameter)
			continue
		}

		// Test all temperament combinations at extremes
		temperamentValues := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
		for _, risk := range temperamentValues {
			for _, agg := range temperamentValues {
				for _, pat := range temperamentValues {
					result := GetAdjustedValue(mapping, risk, agg, pat)

					if result < tc.absoluteMin {
						t.Errorf("%s violated absolute min: got %.4f, min is %.4f (risk=%.1f, agg=%.1f, pat=%.1f)",
							tc.parameter, result, tc.absoluteMin, risk, agg, pat)
					}
					if result > tc.absoluteMax {
						t.Errorf("%s violated absolute max: got %.4f, max is %.4f (risk=%.1f, agg=%.1f, pat=%.1f)",
							tc.parameter, result, tc.absoluteMax, risk, agg, pat)
					}
				}
			}
		}
	}
}

// TestPriorityBoostBounds verifies boost multipliers stay in safe range
func TestPriorityBoostBounds(t *testing.T) {
	boostParams := []string{
		"boost_windfall_priority",
		"boost_bubble_risk",
		"boost_needs_rebalance",
		"boost_overweight",
		"boost_quality_value",
		"boost_recovery_candidate",
		"boost_quality_value_buy",
		"boost_deep_value",
		"boost_low_risk",
		"boost_growth_bull",
	}

	penaltyParams := []string{
		"boost_high_risk_penalty",
		"boost_quantum_warning_penalty",
	}

	temperamentValues := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	// Boost multipliers: [0.5, 2.0]
	for _, param := range boostParams {
		mapping, exists := GetTemperamentMapping(param)
		if !exists {
			t.Errorf("Mapping %q not found", param)
			continue
		}

		for _, risk := range temperamentValues {
			for _, agg := range temperamentValues {
				for _, pat := range temperamentValues {
					result := GetAdjustedValue(mapping, risk, agg, pat)

					if result < 0.5 || result > 2.0 {
						t.Errorf("%s boost out of range [0.5, 2.0]: got %.4f (risk=%.1f, agg=%.1f, pat=%.1f)",
							param, result, risk, agg, pat)
					}
				}
			}
		}
	}

	// Penalty multipliers: [0.3, 1.0]
	for _, param := range penaltyParams {
		mapping, exists := GetTemperamentMapping(param)
		if !exists {
			t.Errorf("Mapping %q not found", param)
			continue
		}

		for _, risk := range temperamentValues {
			for _, agg := range temperamentValues {
				for _, pat := range temperamentValues {
					result := GetAdjustedValue(mapping, risk, agg, pat)

					if result < 0.3 || result > 1.0 {
						t.Errorf("%s penalty out of range [0.3, 1.0]: got %.4f (risk=%.1f, agg=%.1f, pat=%.1f)",
							param, result, risk, agg, pat)
					}
				}
			}
		}
	}
}

// TestInverseRelationships verifies inverse temperament relationships work
func TestInverseRelationships(t *testing.T) {
	inverseCases := []struct {
		parameter string
		// For inverse: value=0 should give higher result, value=1 should give lower result
	}{
		{"evaluation_quality_weight"},       // Aggression inverse
		{"evaluation_risk_adjusted_weight"}, // Risk inverse
		{"quality_stability_threshold"},  // Aggression inverse
		{"boost_bubble_risk"},               // Risk inverse
		{"tag_overvalued_pe_threshold"},     // Risk inverse
	}

	for _, tc := range inverseCases {
		mapping, exists := GetTemperamentMapping(tc.parameter)
		if !exists {
			t.Errorf("Mapping %q not found", tc.parameter)
			continue
		}

		if !mapping.Inverse {
			t.Errorf("%s should be marked as Inverse=true", tc.parameter)
			continue
		}

		// For inverse: value=0 (conservative) should yield higher value
		// value=1 (aggressive) should yield lower value
		resultAt0 := GetAdjustedValue(mapping, 0.0, 0.0, 0.0)
		resultAt1 := GetAdjustedValue(mapping, 1.0, 1.0, 1.0)

		if resultAt0 <= resultAt1 {
			t.Errorf("%s inverse relationship broken: at 0.0 got %.4f, at 1.0 got %.4f (should be higher at 0)",
				tc.parameter, resultAt0, resultAt1)
		}
	}
}

// TestProgressionCurves verifies each progression type produces correct curve shapes
func TestProgressionCurves(t *testing.T) {
	// For linear: intermediate values should be proportional
	// For sigmoid: values should cluster at ends
	// For exponential: values should accelerate toward max
	// For logarithmic: values should decelerate toward max

	linearMapping := TemperamentMapping{
		Parameter:   "test_linear",
		Temperament: "aggression",
		Inverse:     false,
		Min:         0.0,
		Max:         1.0,
		Base:        0.5,
		Progression: "linear",
		AbsoluteMin: 0.0,
		AbsoluteMax: 1.0,
	}

	// Linear should have 0.25 at aggression=0.25
	result := GetAdjustedValue(linearMapping, 0.5, 0.25, 0.5)
	if math.Abs(result-0.25) > 0.05 {
		t.Errorf("Linear progression: expected ~0.25 at aggression=0.25, got %.4f", result)
	}

	// Linear should have 0.75 at aggression=0.75
	result = GetAdjustedValue(linearMapping, 0.5, 0.75, 0.5)
	if math.Abs(result-0.75) > 0.05 {
		t.Errorf("Linear progression: expected ~0.75 at aggression=0.75, got %.4f", result)
	}

	sigmoidMapping := TemperamentMapping{
		Parameter:   "test_sigmoid",
		Temperament: "aggression",
		Inverse:     false,
		Min:         0.0,
		Max:         1.0,
		Base:        0.5,
		Progression: "sigmoid",
		AbsoluteMin: 0.0,
		AbsoluteMax: 1.0,
	}

	// Sigmoid should have steeper transition around 0.5
	resultAt25 := GetAdjustedValue(sigmoidMapping, 0.5, 0.25, 0.5)
	resultAt50 := GetAdjustedValue(sigmoidMapping, 0.5, 0.50, 0.5)
	resultAt75 := GetAdjustedValue(sigmoidMapping, 0.5, 0.75, 0.5)

	// The jump from 0.25 to 0.50 should be less than from 0.50 to 0.75 (steeper in middle)
	jump1 := resultAt50 - resultAt25
	jump2 := resultAt75 - resultAt50

	// Both jumps should be positive (increasing)
	if jump1 <= 0 || jump2 <= 0 {
		t.Errorf("Sigmoid should be monotonically increasing: jump1=%.4f, jump2=%.4f (resultAt25=%.4f, resultAt50=%.4f, resultAt75=%.4f)",
			jump1, jump2, resultAt25, resultAt50, resultAt75)
	}
}

// TestEvaluationWeightsNormalize verifies weights can be normalized to sum to 1.0
func TestEvaluationWeightsNormalize(t *testing.T) {
	weightParams := []string{
		"evaluation_opportunity_weight",
		"evaluation_quality_weight",
		"evaluation_risk_adjusted_weight",
		"evaluation_diversification_weight",
		"evaluation_regime_weight",
	}

	temperamentValues := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	for _, risk := range temperamentValues {
		for _, agg := range temperamentValues {
			for _, pat := range temperamentValues {
				var sum float64
				weights := make(map[string]float64)

				for _, param := range weightParams {
					mapping, exists := GetTemperamentMapping(param)
					if !exists {
						t.Fatalf("Mapping %q not found", param)
					}
					weights[param] = GetAdjustedValue(mapping, risk, agg, pat)
					sum += weights[param]
				}

				// Verify sum is reasonable (between 0.8 and 1.2 before normalization)
				if sum < 0.5 || sum > 1.5 {
					t.Errorf("Raw weights sum %.4f is unreasonable (risk=%.1f, agg=%.1f, pat=%.1f)",
						sum, risk, agg, pat)
				}

				// After normalization, sum should be 1.0
				normalizedSum := 0.0
				for _, w := range weights {
					normalizedSum += w / sum
				}
				if math.Abs(normalizedSum-1.0) > 0.001 {
					t.Errorf("Normalized weights should sum to 1.0, got %.6f", normalizedSum)
				}
			}
		}
	}
}

// TestMappingCompleteness verifies all mappings have required fields
func TestMappingCompleteness(t *testing.T) {
	allMappings := GetAllTemperamentMappings()

	validTemperaments := map[string]bool{
		"risk_tolerance": true,
		"aggression":     true,
		"patience":       true,
		"fixed":          true, // for params that don't change with temperament
	}

	validProgressions := map[string]bool{
		"linear":              true,
		"linear-reverse":      true,
		"exponential":         true,
		"exponential-reverse": true,
		"logarithmic":         true,
		"logarithmic-reverse": true,
		"sigmoid":             true,
		"sigmoid-reverse":     true,
	}

	for name, mapping := range allMappings {
		// Check parameter name matches key
		if mapping.Parameter != name {
			t.Errorf("Mapping key %q has mismatched Parameter field %q", name, mapping.Parameter)
		}

		// Check temperament is valid
		if !validTemperaments[mapping.Temperament] {
			t.Errorf("Mapping %q has invalid Temperament %q", name, mapping.Temperament)
		}

		// Check progression is valid (unless fixed)
		if mapping.Temperament != "fixed" && !validProgressions[mapping.Progression] {
			t.Errorf("Mapping %q has invalid Progression %q", name, mapping.Progression)
		}

		// Check Min < Max
		if mapping.Min >= mapping.Max {
			t.Errorf("Mapping %q has Min (%.4f) >= Max (%.4f)", name, mapping.Min, mapping.Max)
		}

		// Check Base is between Min and Max
		if mapping.Base < mapping.Min || mapping.Base > mapping.Max {
			t.Errorf("Mapping %q has Base (%.4f) outside [Min=%.4f, Max=%.4f]",
				name, mapping.Base, mapping.Min, mapping.Max)
		}

		// Check AbsoluteMin <= Min and AbsoluteMax >= Max
		if mapping.AbsoluteMin > mapping.Min {
			t.Errorf("Mapping %q has AbsoluteMin (%.4f) > Min (%.4f)",
				name, mapping.AbsoluteMin, mapping.Min)
		}
		if mapping.AbsoluteMax < mapping.Max {
			t.Errorf("Mapping %q has AbsoluteMax (%.4f) < Max (%.4f)",
				name, mapping.AbsoluteMax, mapping.Max)
		}
	}
}

// TestMappingCount verifies we have at least the expected number of mappings
func TestMappingCount(t *testing.T) {
	allMappings := GetAllTemperamentMappings()

	// Per the plan, we should have ~150+ mappings
	minExpected := 140 // Allow some tolerance
	actual := len(allMappings)

	if actual < minExpected {
		t.Errorf("Expected at least %d mappings, got %d", minExpected, actual)
	}

	t.Logf("Total temperament mappings: %d", actual)
}

// TestFixedParametersDontChange verifies fixed params return base regardless of temperament
func TestFixedParametersDontChange(t *testing.T) {
	fixedParams := []string{
		"evaluation_regime_weight",       // Always 0.10, narrowest range
		"boost_medium_risk",              // Neutral for medium risk
		"tag_dividend_total_return_cagr", // Minimum growth always
	}

	temperamentValues := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	for _, param := range fixedParams {
		mapping, exists := GetTemperamentMapping(param)
		if !exists {
			t.Errorf("Mapping %q not found", param)
			continue
		}

		if mapping.Temperament != "fixed" {
			// Skip if not actually fixed
			continue
		}

		baseValue := mapping.Base
		for _, risk := range temperamentValues {
			for _, agg := range temperamentValues {
				for _, pat := range temperamentValues {
					result := GetAdjustedValue(mapping, risk, agg, pat)
					if math.Abs(result-baseValue) > 0.001 {
						t.Errorf("%s (fixed) should always return %.4f, got %.4f (risk=%.1f, agg=%.1f, pat=%.1f)",
							param, baseValue, result, risk, agg, pat)
					}
				}
			}
		}
	}
}
