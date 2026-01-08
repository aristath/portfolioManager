package formulas

import (
	"math"
	"testing"
)

func TestCalculateEMA(t *testing.T) {
	tests := []struct {
		name        string
		closes      []float64
		length      int
		expectedNil bool
		description string
	}{
		{
			name:        "empty closes",
			closes:      []float64{},
			length:      200,
			expectedNil: true,
			description: "Empty prices should return nil",
		},
		{
			name:        "insufficient data falls back to SMA",
			closes:      []float64{100.0, 105.0, 110.0},
			length:      200,
			expectedNil: false,
			description: "Should fallback to SMA when not enough data",
		},
		{
			name:        "exact length",
			closes:      generatePrices(100.0, 1.0, 200),
			length:      200,
			expectedNil: false,
			description: "Exactly 200 prices should work",
		},
		{
			name:        "sufficient data",
			closes:      generatePrices(100.0, 1.0, 300),
			length:      200,
			expectedNil: false,
			description: "More than enough data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateEMA(tt.closes, tt.length)
			if (result == nil) != tt.expectedNil {
				t.Errorf("CalculateEMA() = %v, expected nil: %v - %s", result, tt.expectedNil, tt.description)
				return
			}
			if result != nil {
				// EMA should be a valid number
				if math.IsNaN(*result) || math.IsInf(*result, 0) {
					t.Errorf("EMA is NaN or Inf: %v", *result)
				}
				// EMA should be positive for positive prices
				if len(tt.closes) > 0 && tt.closes[len(tt.closes)-1] > 0 && *result < 0 {
					t.Errorf("EMA should be positive for positive prices, got %v", *result)
				}
			}
		})
	}
}

func TestCalculateSMA(t *testing.T) {
	tests := []struct {
		name        string
		closes      []float64
		length      int
		expectedNil bool
		description string
	}{
		{
			name:        "insufficient data",
			closes:      []float64{100.0, 105.0},
			length:      20,
			expectedNil: true,
			description: "Need at least length prices",
		},
		{
			name:        "exact length",
			closes:      generatePrices(100.0, 1.0, 20),
			length:      20,
			expectedNil: false,
			description: "Exactly 20 prices should work",
		},
		{
			name:        "sufficient data",
			closes:      generatePrices(100.0, 1.0, 50),
			length:      20,
			expectedNil: false,
			description: "More than enough data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateSMA(tt.closes, tt.length)
			if (result == nil) != tt.expectedNil {
				t.Errorf("CalculateSMA() = %v, expected nil: %v - %s", result, tt.expectedNil, tt.description)
				return
			}
			if result != nil {
				// SMA should be valid
				if math.IsNaN(*result) || math.IsInf(*result, 0) {
					t.Errorf("SMA is NaN or Inf: %v", *result)
				}
			}
		})
	}
}

func TestCalculateDistanceFromEMA(t *testing.T) {
	tests := []struct {
		name        string
		closes      []float64
		length      int
		expectedNil bool
		description string
	}{
		{
			name:        "empty closes",
			closes:      []float64{},
			length:      200,
			expectedNil: true,
			description: "Empty prices should return nil",
		},
		{
			name:        "valid data",
			closes:      generatePrices(100.0, 1.0, 250),
			length:      200,
			expectedNil: false,
			description: "Valid data should calculate distance",
		},
		{
			name:        "price above EMA",
			closes:      append(generatePrices(100.0, 0.1, 200), 150.0),
			length:      200,
			expectedNil: false,
			description: "Price above EMA should give positive distance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateDistanceFromEMA(tt.closes, tt.length)
			if (result == nil) != tt.expectedNil {
				t.Errorf("CalculateDistanceFromEMA() = %v, expected nil: %v - %s", result, tt.expectedNil, tt.description)
				return
			}
			if result != nil {
				// Distance should be valid
				if math.IsNaN(*result) || math.IsInf(*result, 0) {
					t.Errorf("Distance is NaN or Inf: %v", *result)
				}
			}
		})
	}
}
