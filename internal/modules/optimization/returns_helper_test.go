package optimization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClamp(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		min      float64
		max      float64
		expected float64
	}{
		{
			name:     "value within range",
			value:    5.0,
			min:      0.0,
			max:      10.0,
			expected: 5.0,
		},
		{
			name:     "value below min",
			value:    -5.0,
			min:      0.0,
			max:      10.0,
			expected: 0.0,
		},
		{
			name:     "value above max",
			value:    15.0,
			min:      0.0,
			max:      10.0,
			expected: 10.0,
		},
		{
			name:     "value equals min",
			value:    0.0,
			min:      0.0,
			max:      10.0,
			expected: 0.0,
		},
		{
			name:     "value equals max",
			value:    10.0,
			min:      0.0,
			max:      10.0,
			expected: 10.0,
		},
		{
			name:     "negative range",
			value:    -15.0,
			min:      -10.0,
			max:      -5.0,
			expected: -10.0,
		},
		{
			name:     "zero range",
			value:    5.0,
			min:      10.0,
			max:      10.0,
			expected: 10.0,
		},
		{
			name:     "large positive value",
			value:    1000.0,
			min:      0.0,
			max:      10.0,
			expected: 10.0,
		},
		{
			name:     "large negative value",
			value:    -1000.0,
			min:      0.0,
			max:      10.0,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clamp(tt.value, tt.min, tt.max)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}

func TestConvertCAGRScoreToCAGR_Optimization(t *testing.T) {
	// This is a duplicate function in the optimization package
	// Test it here to ensure consistency
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{
			name:     "zero score",
			input:    0.0,
			expected: 0.0,
		},
		{
			name:     "negative score",
			input:    -0.5,
			expected: 0.0,
		},
		{
			name:     "floor score (0.15)",
			input:    0.15,
			expected: 0.0,
		},
		{
			name:     "target score (0.8)",
			input:    0.8,
			expected: 0.11,
		},
		{
			name:     "maximum score (1.0)",
			input:    1.0,
			expected: 0.20,
		},
		{
			name:     "mid-range score (0.5)",
			input:    0.5,
			expected: 0.0592, // Linear interpolation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertCAGRScoreToCAGR(tt.input)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}
