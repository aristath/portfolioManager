package allocation

// NormalizeWeights converts arbitrary weights to percentages summing to 1.0.
// This function is exported for use by the optimization package and other consumers.
//
// Example: {A: 0.8, B: 0.8, C: 0.4} -> {A: 0.4, B: 0.4, C: 0.2}
//
// If all weights are zero, returns the input unchanged to avoid division by zero.
func NormalizeWeights(weights map[string]float64) map[string]float64 {
	var total float64
	for _, w := range weights {
		total += w
	}

	if total == 0 {
		return weights
	}

	result := make(map[string]float64)
	for name, w := range weights {
		result[name] = w / total
	}
	return result
}
