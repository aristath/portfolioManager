package symbolic_regression

import (
	"math"
)

// NormalizeFeatures normalizes feature values to [0, 1] range (except regime which stays [-1, 1])
// Uses min-max normalization per feature across all examples
func NormalizeFeatures(examples []TrainingExample) []TrainingExample {
	if len(examples) == 0 {
		return examples
	}

	// Find min/max for each feature
	minMax := findMinMax(examples)

	// Normalize each example
	normalized := make([]TrainingExample, len(examples))
	for i, ex := range examples {
		normalized[i] = ex
		normalized[i].Inputs = normalizeInputs(ex.Inputs, minMax)
	}

	return normalized
}

// FeatureMinMax holds min/max values for normalization
type FeatureMinMax struct {
	LongTermScore        [2]float64 // [min, max]
	FundamentalsScore    [2]float64
	DividendsScore       [2]float64
	OpportunityScore     [2]float64
	ShortTermScore       [2]float64
	TechnicalsScore      [2]float64
	OpinionScore         [2]float64
	DiversificationScore [2]float64
	TotalScore           [2]float64
	CAGR                 [2]float64
	DividendYield        [2]float64
	Volatility           [2]float64
	SharpeRatio          [2]float64
	SortinoRatio         [2]float64
	RSI                  [2]float64
	MaxDrawdown          [2]float64
	// RegimeScore is NOT normalized (stays in [-1, 1])
}

// findMinMax finds min/max values for each feature across all examples
func findMinMax(examples []TrainingExample) FeatureMinMax {
	mm := FeatureMinMax{
		LongTermScore:        [2]float64{math.MaxFloat64, -math.MaxFloat64},
		FundamentalsScore:    [2]float64{math.MaxFloat64, -math.MaxFloat64},
		DividendsScore:       [2]float64{math.MaxFloat64, -math.MaxFloat64},
		OpportunityScore:     [2]float64{math.MaxFloat64, -math.MaxFloat64},
		ShortTermScore:       [2]float64{math.MaxFloat64, -math.MaxFloat64},
		TechnicalsScore:      [2]float64{math.MaxFloat64, -math.MaxFloat64},
		OpinionScore:         [2]float64{math.MaxFloat64, -math.MaxFloat64},
		DiversificationScore: [2]float64{math.MaxFloat64, -math.MaxFloat64},
		TotalScore:           [2]float64{math.MaxFloat64, -math.MaxFloat64},
		CAGR:                 [2]float64{math.MaxFloat64, -math.MaxFloat64},
		DividendYield:        [2]float64{math.MaxFloat64, -math.MaxFloat64},
		Volatility:           [2]float64{math.MaxFloat64, -math.MaxFloat64},
		SharpeRatio:          [2]float64{math.MaxFloat64, -math.MaxFloat64},
		SortinoRatio:         [2]float64{math.MaxFloat64, -math.MaxFloat64},
		RSI:                  [2]float64{math.MaxFloat64, -math.MaxFloat64},
		MaxDrawdown:          [2]float64{math.MaxFloat64, -math.MaxFloat64},
	}

	for _, ex := range examples {
		updateMinMax(&mm.LongTermScore, ex.Inputs.LongTermScore)
		updateMinMax(&mm.FundamentalsScore, ex.Inputs.FundamentalsScore)
		updateMinMax(&mm.DividendsScore, ex.Inputs.DividendsScore)
		updateMinMax(&mm.OpportunityScore, ex.Inputs.OpportunityScore)
		updateMinMax(&mm.ShortTermScore, ex.Inputs.ShortTermScore)
		updateMinMax(&mm.TechnicalsScore, ex.Inputs.TechnicalsScore)
		updateMinMax(&mm.OpinionScore, ex.Inputs.OpinionScore)
		updateMinMax(&mm.DiversificationScore, ex.Inputs.DiversificationScore)
		updateMinMax(&mm.TotalScore, ex.Inputs.TotalScore)
		updateMinMax(&mm.CAGR, ex.Inputs.CAGR)
		updateMinMax(&mm.DividendYield, ex.Inputs.DividendYield)
		updateMinMax(&mm.Volatility, ex.Inputs.Volatility)

		if ex.Inputs.SharpeRatio != nil {
			updateMinMax(&mm.SharpeRatio, *ex.Inputs.SharpeRatio)
		}
		if ex.Inputs.SortinoRatio != nil {
			updateMinMax(&mm.SortinoRatio, *ex.Inputs.SortinoRatio)
		}
		if ex.Inputs.RSI != nil {
			updateMinMax(&mm.RSI, *ex.Inputs.RSI)
		}
		if ex.Inputs.MaxDrawdown != nil {
			updateMinMax(&mm.MaxDrawdown, *ex.Inputs.MaxDrawdown)
		}
	}

	// Set defaults for features with no data
	setDefaultMinMax(&mm)

	return mm
}

// updateMinMax updates min/max for a feature
func updateMinMax(minMax *[2]float64, value float64) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return
	}
	if value < minMax[0] {
		minMax[0] = value
	}
	if value > minMax[1] {
		minMax[1] = value
	}
}

// setDefaultMinMax sets default min/max for features with no data
func setDefaultMinMax(mm *FeatureMinMax) {
	setDefaultIfNeeded(&mm.LongTermScore)
	setDefaultIfNeeded(&mm.FundamentalsScore)
	setDefaultIfNeeded(&mm.DividendsScore)
	setDefaultIfNeeded(&mm.OpportunityScore)
	setDefaultIfNeeded(&mm.ShortTermScore)
	setDefaultIfNeeded(&mm.TechnicalsScore)
	setDefaultIfNeeded(&mm.OpinionScore)
	setDefaultIfNeeded(&mm.DiversificationScore)
	setDefaultIfNeeded(&mm.TotalScore)
	setDefaultIfNeeded(&mm.CAGR)
	setDefaultIfNeeded(&mm.DividendYield)
	setDefaultIfNeeded(&mm.Volatility)
	setDefaultIfNeeded(&mm.SharpeRatio)
	setDefaultIfNeeded(&mm.SortinoRatio)
	setDefaultIfNeeded(&mm.RSI)
	setDefaultIfNeeded(&mm.MaxDrawdown)
}

// setDefaultIfNeeded sets [0, 1] as default if minMax hasn't been set
func setDefaultIfNeeded(minMax *[2]float64) {
	if minMax[0] == math.MaxFloat64 {
		minMax[0] = 0.0
		minMax[1] = 1.0
	}
	// If min == max, add small range to avoid division by zero
	if minMax[0] == minMax[1] {
		minMax[1] = minMax[0] + 0.001
	}
}

// normalizeInputs normalizes a single input set
func normalizeInputs(inputs TrainingInputs, minMax FeatureMinMax) TrainingInputs {
	normalized := inputs

	normalized.LongTermScore = normalizeValue(inputs.LongTermScore, minMax.LongTermScore)
	normalized.FundamentalsScore = normalizeValue(inputs.FundamentalsScore, minMax.FundamentalsScore)
	normalized.DividendsScore = normalizeValue(inputs.DividendsScore, minMax.DividendsScore)
	normalized.OpportunityScore = normalizeValue(inputs.OpportunityScore, minMax.OpportunityScore)
	normalized.ShortTermScore = normalizeValue(inputs.ShortTermScore, minMax.ShortTermScore)
	normalized.TechnicalsScore = normalizeValue(inputs.TechnicalsScore, minMax.TechnicalsScore)
	normalized.OpinionScore = normalizeValue(inputs.OpinionScore, minMax.OpinionScore)
	normalized.DiversificationScore = normalizeValue(inputs.DiversificationScore, minMax.DiversificationScore)
	normalized.TotalScore = normalizeValue(inputs.TotalScore, minMax.TotalScore)
	normalized.CAGR = normalizeValue(inputs.CAGR, minMax.CAGR)
	normalized.DividendYield = normalizeValue(inputs.DividendYield, minMax.DividendYield)
	normalized.Volatility = normalizeValue(inputs.Volatility, minMax.Volatility)

	// Optional metrics - normalize if present
	if inputs.SharpeRatio != nil {
		val := normalizeValue(*inputs.SharpeRatio, minMax.SharpeRatio)
		normalized.SharpeRatio = &val
	}
	if inputs.SortinoRatio != nil {
		val := normalizeValue(*inputs.SortinoRatio, minMax.SortinoRatio)
		normalized.SortinoRatio = &val
	}
	if inputs.RSI != nil {
		val := normalizeValue(*inputs.RSI, minMax.RSI)
		normalized.RSI = &val
	}
	if inputs.MaxDrawdown != nil {
		val := normalizeValue(*inputs.MaxDrawdown, minMax.MaxDrawdown)
		normalized.MaxDrawdown = &val
	}

	// RegimeScore is NOT normalized (stays in [-1, 1])
	// Use default for missing values
	if normalized.LongTermScore == 0 && inputs.LongTermScore == 0 {
		normalized.LongTermScore = 0.5
	}
	if normalized.FundamentalsScore == 0 && inputs.FundamentalsScore == 0 {
		normalized.FundamentalsScore = 0.5
	}

	return normalized
}

// normalizeValue normalizes a single value using min-max normalization
func normalizeValue(value float64, minMax [2]float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0.0
	}

	min, max := minMax[0], minMax[1]
	if max == min {
		return 0.5 // Default to middle if no range
	}

	normalized := (value - min) / (max - min)

	// Clamp to [0, 1]
	if normalized < 0 {
		return 0.0
	}
	if normalized > 1 {
		return 1.0
	}

	return normalized
}

// ExtractFeatureNames extracts all available feature names from inputs
func ExtractFeatureNames(inputs TrainingInputs) []string {
	features := []string{
		"long_term",
		"fundamentals",
		"dividends",
		"opportunity",
		"short_term",
		"technicals",
		"opinion",
		"diversification",
		"total_score",
		"cagr",
		"dividend_yield",
		"volatility",
		"regime",
	}

	// Add optional features if present
	if inputs.SharpeRatio != nil {
		features = append(features, "sharpe")
	}
	if inputs.SortinoRatio != nil {
		features = append(features, "sortino")
	}
	if inputs.RSI != nil {
		features = append(features, "rsi")
	}
	if inputs.MaxDrawdown != nil {
		features = append(features, "max_drawdown")
	}

	return features
}

// GetFeatureValue gets the value of a feature by name
func GetFeatureValue(inputs TrainingInputs, featureName string) float64 {
	switch featureName {
	case "long_term":
		return inputs.LongTermScore
	case "fundamentals":
		return inputs.FundamentalsScore
	case "dividends":
		return inputs.DividendsScore
	case "opportunity":
		return inputs.OpportunityScore
	case "short_term":
		return inputs.ShortTermScore
	case "technicals":
		return inputs.TechnicalsScore
	case "opinion":
		return inputs.OpinionScore
	case "diversification":
		return inputs.DiversificationScore
	case "total_score":
		return inputs.TotalScore
	case "cagr":
		return inputs.CAGR
	case "dividend_yield":
		return inputs.DividendYield
	case "volatility":
		return inputs.Volatility
	case "regime":
		return inputs.RegimeScore
	case "sharpe":
		if inputs.SharpeRatio != nil {
			return *inputs.SharpeRatio
		}
		return 0.0
	case "sortino":
		if inputs.SortinoRatio != nil {
			return *inputs.SortinoRatio
		}
		return 0.0
	case "rsi":
		if inputs.RSI != nil {
			return *inputs.RSI
		}
		return 0.0
	case "max_drawdown":
		if inputs.MaxDrawdown != nil {
			return *inputs.MaxDrawdown
		}
		return 0.0
	default:
		return 0.0
	}
}
