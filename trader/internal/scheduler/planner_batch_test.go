package scheduler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertCAGRScoreToCAGR(t *testing.T) {
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
			name:     "below floor",
			input:    0.10,
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
			expected: 0.0592, // Linear interpolation: 0.0 + (0.5-0.15)*(0.11-0.0)/(0.8-0.15) ≈ 0.05923
		},
		{
			name:     "above target (0.9)",
			input:    0.9,
			expected: 0.155, // Linear interpolation: 0.11 + (0.9-0.8)*(0.20-0.11)/(1.0-0.8)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertCAGRScoreToCAGR(tt.input)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  float64
		wantError bool
	}{
		{
			name:      "valid integer string",
			input:     "123",
			expected:  123.0,
			wantError: false,
		},
		{
			name:      "valid decimal string",
			input:     "123.45",
			expected:  123.45,
			wantError: false,
		},
		{
			name:      "negative number",
			input:     "-123.45",
			expected:  -123.45,
			wantError: false,
		},
		{
			name:      "zero",
			input:     "0",
			expected:  0.0,
			wantError: false,
		},
		{
			name:      "scientific notation",
			input:     "1.23e2",
			expected:  123.0,
			wantError: false,
		},
		{
			name:      "invalid string",
			input:     "not a number",
			expected:  0.0,
			wantError: true,
		},
		{
			name:      "empty string",
			input:     "",
			expected:  0.0,
			wantError: true,
		},
		{
			name:      "partial number",
			input:     "123.",
			expected:  123.0,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFloat(tt.input)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expected, result, 0.0001)
			}
		})
	}
}
