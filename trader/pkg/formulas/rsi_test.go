package formulas

import (
	"math"
	"testing"
)

func TestCalculateRSI(t *testing.T) {
	tests := []struct {
		name        string
		closes      []float64
		length      int
		expectedNil bool
		expectedMin float64
		expectedMax float64
		description string
	}{
		{
			name:        "insufficient data",
			closes:      []float64{100.0, 105.0},
			length:      14,
			expectedNil: true,
			description: "Need at least length+1 prices",
		},
		{
			name:        "exact minimum",
			closes:      generateRSIPrices(100.0, 15),
			length:      14,
			expectedNil: false,
			expectedMin: 0.0,
			expectedMax: 100.0,
			description: "Exactly 15 prices for 14-period RSI",
		},
		{
			name:        "sufficient data",
			closes:      generateRSIPrices(100.0, 50),
			length:      14,
			expectedNil: false,
			expectedMin: 0.0,
			expectedMax: 100.0,
			description: "More than enough data",
		},
		{
			name:        "empty closes",
			closes:      []float64{},
			length:      14,
			expectedNil: true,
			description: "Empty prices should return nil",
		},
		{
			name:        "rising prices",
			closes:      generateRisingPrices(100.0, 30),
			length:      14,
			expectedNil: false,
			expectedMin: 0.0,
			expectedMax: 100.0,
			description: "Rising prices should give high RSI",
		},
		{
			name:        "falling prices",
			closes:      generateFallingPrices(100.0, 30),
			length:      14,
			expectedNil: false,
			expectedMin: 0.0,
			expectedMax: 100.0,
			description: "Falling prices should give low RSI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRSI(tt.closes, tt.length)
			if (result == nil) != tt.expectedNil {
				t.Errorf("CalculateRSI() = %v, expected nil: %v - %s", result, tt.expectedNil, tt.description)
				return
			}
			if result != nil {
				// RSI should be between 0 and 100
				if *result < tt.expectedMin || *result > tt.expectedMax {
					t.Errorf("RSI = %v, want between %v and %v", *result, tt.expectedMin, tt.expectedMax)
				}
				// Should not be NaN
				if math.IsNaN(*result) {
					t.Errorf("RSI is NaN")
				}
			}
		})
	}
}

func TestIsNaN(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected bool
	}{
		{"normal number", 5.0, false},
		{"zero", 0.0, false},
		{"negative", -5.0, false},
		{"infinity", math.Inf(1), false},
		{"negative infinity", math.Inf(-1), false},
		{"NaN", math.NaN(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNaN(tt.value)
			if result != tt.expected {
				t.Errorf("isNaN(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

// Helper functions for RSI tests
func generateRSIPrices(start float64, count int) []float64 {
	prices := make([]float64, count)
	for i := 0; i < count; i++ {
		// Alternating small changes to simulate price movement
		if i%2 == 0 {
			prices[i] = start + float64(i)*0.5
		} else {
			prices[i] = start + float64(i)*0.5 - 0.3
		}
	}
	return prices
}

func generateRisingPrices(start float64, count int) []float64 {
	prices := make([]float64, count)
	for i := 0; i < count; i++ {
		prices[i] = start + float64(i)*1.0
	}
	return prices
}

func generateFallingPrices(start float64, count int) []float64 {
	prices := make([]float64, count)
	for i := 0; i < count; i++ {
		prices[i] = start - float64(i)*1.0
	}
	return prices
}
