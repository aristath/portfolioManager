package formulas

import (
	"math"
	"testing"
)

func TestCalculateBollingerBands(t *testing.T) {
	tests := []struct {
		name             string
		closes           []float64
		length           int
		stdDevMultiplier float64
		expectedNil      bool
		description      string
	}{
		{
			name:             "insufficient data",
			closes:           []float64{100.0, 105.0},
			length:           20,
			stdDevMultiplier: 2.0,
			expectedNil:      true,
			description:      "Need at least 20 prices for 20-period bands",
		},
		{
			name:             "exact length",
			closes:           generatePrices(100.0, 1.0, 20),
			length:           20,
			stdDevMultiplier: 2.0,
			expectedNil:      false,
			description:      "Exactly 20 prices should work",
		},
		{
			name:             "sufficient data",
			closes:           generatePrices(100.0, 1.0, 50),
			length:           20,
			stdDevMultiplier: 2.0,
			expectedNil:      false,
			description:      "More than enough data",
		},
		{
			name:             "empty prices",
			closes:           []float64{},
			length:           20,
			stdDevMultiplier: 2.0,
			expectedNil:      true,
			description:      "Empty prices should return nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateBollingerBands(tt.closes, tt.length, tt.stdDevMultiplier)
			if (result == nil) != tt.expectedNil {
				t.Errorf("CalculateBollingerBands() = %v, expected nil: %v - %s", result, tt.expectedNil, tt.description)
			}
			if result != nil {
				// Verify bands are valid
				if math.IsNaN(result.Upper) || math.IsNaN(result.Middle) || math.IsNaN(result.Lower) {
					t.Errorf("BollingerBands contains NaN values")
				}
				// Upper should be >= Middle >= Lower
				if result.Upper < result.Middle || result.Middle < result.Lower {
					t.Errorf("Invalid band order: Upper=%v, Middle=%v, Lower=%v", result.Upper, result.Middle, result.Lower)
				}
			}
		})
	}
}

func TestCalculateBollingerPosition(t *testing.T) {
	// Create test data with sufficient prices
	closes := generatePrices(100.0, 1.0, 30)
	length := 20
	stdDevMultiplier := 2.0

	tests := []struct {
		name          string
		closes        []float64
		length        int
		multiplier    float64
		expectedNil   bool
		checkPosition bool
		minPosition   float64
		maxPosition   float64
	}{
		{
			name:        "empty closes",
			closes:      []float64{},
			length:      20,
			multiplier:  2.0,
			expectedNil: true,
		},
		{
			name:        "insufficient data",
			closes:      []float64{100.0, 105.0},
			length:      20,
			multiplier:  2.0,
			expectedNil: true,
		},
		{
			name:          "valid data",
			closes:        closes,
			length:        length,
			multiplier:    stdDevMultiplier,
			expectedNil:   false,
			checkPosition: true,
			minPosition:   0.0,
			maxPosition:   1.0,
		},
		{
			name:          "price at middle",
			closes:        append(generatePrices(100.0, 0.0, 25), 100.0), // Constant prices, last at middle
			length:        20,
			multiplier:    2.0,
			expectedNil:   false,
			checkPosition: true,
			minPosition:   0.0,
			maxPosition:   1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateBollingerPosition(tt.closes, tt.length, tt.multiplier)
			if (result == nil) != tt.expectedNil {
				t.Errorf("CalculateBollingerPosition() = %v, expected nil: %v", result, tt.expectedNil)
				return
			}
			if result != nil && tt.checkPosition {
				// Position should be between 0.0 and 1.0
				if result.Position < tt.minPosition || result.Position > tt.maxPosition {
					t.Errorf("Position = %v, want between %v and %v", result.Position, tt.minPosition, tt.maxPosition)
				}
				// Bands should be valid
				if result.Bands.Upper < result.Bands.Middle || result.Bands.Middle < result.Bands.Lower {
					t.Errorf("Invalid band order in result")
				}
			}
		})
	}
}

// Helper function to generate price series
func generatePrices(start float64, increment float64, count int) []float64 {
	prices := make([]float64, count)
	for i := 0; i < count; i++ {
		prices[i] = start + float64(i)*increment
	}
	return prices
}
