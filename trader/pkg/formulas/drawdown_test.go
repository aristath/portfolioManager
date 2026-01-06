package formulas

import (
	"math"
	"testing"
)

func TestCalculateMaxDrawdown(t *testing.T) {
	tests := []struct {
		name        string
		prices      []float64
		expectedNil bool
		expectedMax float64
		tolerance   float64
		description string
	}{
		{
			name:        "insufficient data",
			prices:      []float64{100.0},
			expectedNil: true,
			description: "Need at least 2 prices",
		},
		{
			name:        "no drawdown",
			prices:      []float64{100.0, 105.0, 110.0, 115.0},
			expectedNil: false,
			expectedMax: 0.0,
			tolerance:   0.0001,
			description: "Rising prices have zero drawdown",
		},
		{
			name:        "single drawdown",
			prices:      []float64{100.0, 110.0, 90.0, 95.0},
			expectedNil: false,
			expectedMax: -0.1818, // (90-110)/110 ≈ -0.1818
			tolerance:   0.01,
			description: "Single drawdown from peak",
		},
		{
			name:        "multiple drawdowns",
			prices:      []float64{100.0, 120.0, 90.0, 115.0, 85.0},
			expectedNil: false,
			expectedMax: -0.2916, // (85-120)/120 ≈ -0.2916
			tolerance:   0.01,
			description: "Maximum of multiple drawdowns",
		},
		{
			name:        "recovery after drawdown",
			prices:      []float64{100.0, 120.0, 80.0, 100.0},
			expectedNil: false,
			expectedMax: -0.3333, // (80-120)/120 ≈ -0.3333
			tolerance:   0.01,
			description: "Drawdown followed by recovery",
		},
		{
			name:        "empty prices",
			prices:      []float64{},
			expectedNil: true,
			description: "Empty prices should return nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateMaxDrawdown(tt.prices)
			if (result == nil) != tt.expectedNil {
				t.Errorf("CalculateMaxDrawdown() = %v, expected nil: %v - %s", result, tt.expectedNil, tt.description)
				return
			}
			if result != nil {
				// Max drawdown should be negative or zero
				if *result > 0.0001 {
					t.Errorf("Max drawdown should be <= 0, got %v", *result)
				}
				if math.Abs(*result-tt.expectedMax) > tt.tolerance {
					t.Errorf("MaxDrawdown() = %v, want %v (±%v) - %s", *result, tt.expectedMax, tt.tolerance, tt.description)
				}
			}
		})
	}
}

func TestCalculateDrawdownMetrics(t *testing.T) {
	tests := []struct {
		name        string
		prices      []float64
		expectedNil bool
		description string
	}{
		{
			name:        "insufficient data",
			prices:      []float64{100.0},
			expectedNil: true,
			description: "Need at least 2 prices",
		},
		{
			name:        "valid data",
			prices:      []float64{100.0, 110.0, 90.0, 95.0},
			expectedNil: false,
			description: "Valid prices should return metrics",
		},
		{
			name:        "empty prices",
			prices:      []float64{},
			expectedNil: true,
			description: "Empty prices should return nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateDrawdownMetrics(tt.prices)
			if (result == nil) != tt.expectedNil {
				t.Errorf("CalculateDrawdownMetrics() = %v, expected nil: %v - %s", result, tt.expectedNil, tt.description)
				return
			}
			if result != nil {
				// Verify metrics are valid
				if result.DaysInDrawdown < 0 {
					t.Errorf("DaysInDrawdown should be >= 0, got %v", result.DaysInDrawdown)
				}
				if result.PeakValue <= 0 {
					t.Errorf("PeakValue should be > 0, got %v", result.PeakValue)
				}
				if result.CurrentValue <= 0 {
					t.Errorf("CurrentValue should be > 0, got %v", result.CurrentValue)
				}
				// Max drawdown should be negative or zero
				if result.MaxDrawdown > 0.0001 {
					t.Errorf("MaxDrawdown should be <= 0, got %v", result.MaxDrawdown)
				}
				// Current drawdown should be <= 0
				if result.CurrentDrawdown > 0.0001 {
					t.Errorf("CurrentDrawdown should be <= 0, got %v", result.CurrentDrawdown)
				}
			}
		})
	}
}

func TestCalculateUlcerIndex(t *testing.T) {
	tests := []struct {
		name        string
		prices      []float64
		period      int
		expectedNil bool
		description string
	}{
		{
			name:        "insufficient data",
			prices:      []float64{100.0, 105.0},
			period:      10,
			expectedNil: true,
			description: "Need at least period prices",
		},
		{
			name:        "valid data",
			prices:      generatePrices(100.0, 1.0, 20),
			period:      10,
			expectedNil: false,
			description: "Valid prices should calculate Ulcer Index",
		},
		{
			name:        "empty prices",
			prices:      []float64{},
			period:      10,
			expectedNil: true,
			description: "Empty prices should return nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateUlcerIndex(tt.prices, tt.period)
			if (result == nil) != tt.expectedNil {
				t.Errorf("CalculateUlcerIndex() = %v, expected nil: %v - %s", result, tt.expectedNil, tt.description)
				return
			}
			if result != nil {
				// Ulcer Index should be >= 0
				if *result < 0 {
					t.Errorf("UlcerIndex should be >= 0, got %v", *result)
				}
				if math.IsNaN(*result) || math.IsInf(*result, 0) {
					t.Errorf("UlcerIndex is NaN or Inf: %v", *result)
				}
			}
		})
	}
}
