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

func TestBuildGroupAllocations_NormalizesTargets(t *testing.T) {
	// Test that buildGroupAllocations correctly normalizes non-normalized targets
	groupValues := map[string]float64{
		"EU":   5000.0,
		"US":   5000.0,
		"Asia": 0.0,
	}

	// Non-normalized targets (sum = 2.0, not 1.0)
	// User set: EU=0.8, US=0.8, Asia=0.4
	groupTargets := map[string]float64{
		"EU":   0.8,
		"US":   0.8,
		"Asia": 0.4,
	}

	totalValue := 10000.0

	allocations := buildGroupAllocations(groupValues, groupTargets, totalValue)

	// Build map for easier assertions
	allocMap := make(map[string]GroupAllocation)
	for _, a := range allocations {
		allocMap[a.Name] = a
	}

	// Normalized targets should be: EU=0.4, US=0.4, Asia=0.2
	// Current values: EU=5000 (50%), US=5000 (50%), Asia=0 (0%)
	// Deviations: EU: 0.5 - 0.4 = 0.1, US: 0.5 - 0.4 = 0.1, Asia: 0 - 0.2 = -0.2

	assert.InDelta(t, 0.4, allocMap["EU"].TargetPct, 0.0001, "EU target should be normalized to 0.4")
	assert.InDelta(t, 0.4, allocMap["US"].TargetPct, 0.0001, "US target should be normalized to 0.4")
	assert.InDelta(t, 0.2, allocMap["Asia"].TargetPct, 0.0001, "Asia target should be normalized to 0.2")

	assert.InDelta(t, 0.5, allocMap["EU"].CurrentPct, 0.0001, "EU current should be 0.5")
	assert.InDelta(t, 0.5, allocMap["US"].CurrentPct, 0.0001, "US current should be 0.5")
	assert.InDelta(t, 0.0, allocMap["Asia"].CurrentPct, 0.0001, "Asia current should be 0.0")

	assert.InDelta(t, 0.1, allocMap["EU"].Deviation, 0.0001, "EU deviation should be 0.1")
	assert.InDelta(t, 0.1, allocMap["US"].Deviation, 0.0001, "US deviation should be 0.1")
	assert.InDelta(t, -0.2, allocMap["Asia"].Deviation, 0.0001, "Asia deviation should be -0.2")
}
