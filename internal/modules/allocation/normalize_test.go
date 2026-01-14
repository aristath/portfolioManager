package allocation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeWeights(t *testing.T) {
	tests := []struct {
		name     string
		weights  map[string]float64
		expected map[string]float64
	}{
		{
			name: "already normalized",
			weights: map[string]float64{
				"A": 0.5,
				"B": 0.3,
				"C": 0.2,
			},
			expected: map[string]float64{
				"A": 0.5,
				"B": 0.3,
				"C": 0.2,
			},
		},
		{
			name: "needs normalization - sum > 1",
			weights: map[string]float64{
				"A": 0.8,
				"B": 0.8,
				"C": 0.4,
			},
			expected: map[string]float64{
				"A": 0.4,
				"B": 0.4,
				"C": 0.2,
			},
		},
		{
			name: "needs normalization - sum < 1",
			weights: map[string]float64{
				"A": 0.1,
				"B": 0.1,
			},
			expected: map[string]float64{
				"A": 0.5,
				"B": 0.5,
			},
		},
		{
			name: "arbitrary numbers",
			weights: map[string]float64{
				"EU":   2.0,
				"US":   3.0,
				"Asia": 5.0,
			},
			expected: map[string]float64{
				"EU":   0.2,
				"US":   0.3,
				"Asia": 0.5,
			},
		},
		{
			name:     "empty map",
			weights:  map[string]float64{},
			expected: map[string]float64{},
		},
		{
			name: "all zeros - returns input unchanged",
			weights: map[string]float64{
				"A": 0.0,
				"B": 0.0,
			},
			expected: map[string]float64{
				"A": 0.0,
				"B": 0.0,
			},
		},
		{
			name: "single item",
			weights: map[string]float64{
				"Only": 0.5,
			},
			expected: map[string]float64{
				"Only": 1.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeWeights(tt.weights)

			assert.Equal(t, len(tt.expected), len(result), "result should have same number of entries")

			for key, expectedValue := range tt.expected {
				assert.Contains(t, result, key, "result should contain key %s", key)
				assert.InDelta(t, expectedValue, result[key], 0.0001,
					"key %s: expected %.4f, got %.4f", key, expectedValue, result[key])
			}
		})
	}
}

func TestNormalizeWeights_PreservesInput(t *testing.T) {
	// Verify that the function doesn't modify the input map
	input := map[string]float64{
		"A": 0.8,
		"B": 0.8,
	}

	// Store original values
	originalA := input["A"]
	originalB := input["B"]

	_ = NormalizeWeights(input)

	// Verify input wasn't modified
	assert.Equal(t, originalA, input["A"], "input should not be modified")
	assert.Equal(t, originalB, input["B"], "input should not be modified")
}
