package symbolic_regression

// RegimeRange represents a regime range for formula discovery
type RegimeRange struct {
	Min  float64
	Max  float64
	Name string // e.g., "bear", "neutral", "bull"
}

// DefaultRegimeRanges returns the default regime ranges for splitting training data
func DefaultRegimeRanges() []RegimeRange {
	return []RegimeRange{
		{Min: -1.0, Max: -0.3, Name: "bear"},
		{Min: -0.3, Max: 0.3, Name: "neutral"},
		{Min: 0.3, Max: 1.0, Name: "bull"},
	}
}

// SplitByRegime splits training examples by regime ranges
func SplitByRegime(examples []TrainingExample, ranges []RegimeRange) map[RegimeRange][]TrainingExample {
	result := make(map[RegimeRange][]TrainingExample)

	// Initialize result map with empty slices
	for _, r := range ranges {
		result[r] = make([]TrainingExample, 0)
	}

	// Split examples by regime score
	for _, ex := range examples {
		regimeScore := ex.Inputs.RegimeScore

		// Find matching regime range
		for _, r := range ranges {
			// Check if regime score falls within this range
			// For the last range (max == 1.0), include the max value
			if r.Max == 1.0 {
				if regimeScore >= r.Min && regimeScore <= r.Max {
					result[r] = append(result[r], ex)
					break
				}
			} else {
				if regimeScore >= r.Min && regimeScore < r.Max {
					result[r] = append(result[r], ex)
					break
				}
			}
		}
	}

	return result
}

// FilterByRegimeRange filters examples to a specific regime range
func FilterByRegimeRange(examples []TrainingExample, min, max float64) []TrainingExample {
	var filtered []TrainingExample
	for _, ex := range examples {
		regimeScore := ex.Inputs.RegimeScore
		if regimeScore >= min && regimeScore <= max {
			filtered = append(filtered, ex)
		}
	}
	return filtered
}
