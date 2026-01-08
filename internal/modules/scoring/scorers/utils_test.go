package scorers

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRound1(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"zero", 0.0, 0.0},
		{"integer", 5.0, 5.0},
		{"round down", 5.34, 5.3},
		{"round up", 5.36, 5.4},
		{"round half up", 5.35, 5.4},
		{"negative", -5.36, -5.4},
		{"large number", 123.456, 123.5},
		{"small number", 0.123, 0.1},
		{"many decimals", 3.14159, 3.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := round1(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRound2(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"zero", 0.0, 0.0},
		{"integer", 5.0, 5.0},
		{"round down", 5.334, 5.33},
		{"round up", 5.336, 5.34},
		{"round half up", 5.335, 5.34},
		{"negative", -5.336, -5.34},
		{"large number", 123.4567, 123.46},
		{"small number", 0.1234, 0.12},
		{"many decimals", 3.14159, 3.14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := round2(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoundPercent(t *testing.T) {
	tests := []struct {
		name     string
		label    string
		pct      float64
		expected string
	}{
		{"with label positive", "Gain", 15.5, "Gain: 16%"},
		{"with label zero", "Return", 0.0, "Return: 0%"},
		{"with label negative", "Loss", -10.3, "Loss: -10%"},
		{"empty label positive", "", 25.7, "26"},
		{"empty label zero", "", 0.0, "0"},
		{"empty label negative", "", -5.2, "-5"},
		{"large percentage", "Growth", 123.45, "Growth: 123%"},
		{"small percentage", "Drop", 0.34, "Drop: 0%"},
		{"rounding up", "Change", 99.6, "Change: 100%"},
		{"rounding down", "Change", 99.4, "Change: 99%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := roundPercent(tt.label, tt.pct)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRound1_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input float64
	}{
		{"NaN", math.NaN()},
		{"positive infinity", math.Inf(1)},
		{"negative infinity", math.Inf(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := round1(tt.input)
			// Should handle edge cases gracefully (not panic)
			assert.False(t, math.IsNaN(result) && !math.IsNaN(tt.input), "Should handle NaN")
		})
	}
}

func TestRound2_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input float64
	}{
		{"NaN", math.NaN()},
		{"positive infinity", math.Inf(1)},
		{"negative infinity", math.Inf(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := round2(tt.input)
			// Should handle edge cases gracefully (not panic)
			assert.False(t, math.IsNaN(result) && !math.IsNaN(tt.input), "Should handle NaN")
		})
	}
}
