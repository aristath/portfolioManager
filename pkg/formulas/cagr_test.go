package formulas

import (
	"math"
	"testing"
)

func TestCalculateCAGR(t *testing.T) {
	tests := []struct {
		name        string
		prices      []MonthlyPrice
		months      int
		expectedNil bool
		description string
	}{
		{
			name:        "insufficient data",
			prices:      makeMonthlyPrices(100.0, 4),
			months:      60,
			expectedNil: true,
			description: "Need at least 6 months",
		},
		{
			name:        "exact minimum",
			prices:      makeMonthlyPrices(100.0, 6),
			months:      6,
			expectedNil: false,
			description: "Exactly 6 months should work",
		},
		{
			name:        "valid data",
			prices:      makeMonthlyPrices(100.0, 60),
			months:      60,
			expectedNil: false,
			description: "5 years of data",
		},
		{
			name:        "empty prices",
			prices:      []MonthlyPrice{},
			months:      60,
			expectedNil: true,
			description: "Empty prices should return nil",
		},
		{
			name:        "zero start price",
			prices:      makeMonthlyPricesWithZero(100.0, 12, 0),
			months:      12,
			expectedNil: true,
			description: "Zero start price returns nil",
		},
		{
			name:        "zero end price",
			prices:      makeMonthlyPricesWithZero(100.0, 12, 11),
			months:      12,
			expectedNil: true,
			description: "Zero end price returns nil",
		},
		{
			name:        "requested months exceeds available",
			prices:      makeMonthlyPrices(100.0, 24),
			months:      60,
			expectedNil: false,
			description: "Should use all available months",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCAGR(tt.prices, tt.months)
			if (result == nil) != tt.expectedNil {
				t.Errorf("CalculateCAGR() = %v, expected nil: %v - %s", result, tt.expectedNil, tt.description)
				return
			}
			if result != nil {
				// CAGR should be valid
				if math.IsNaN(*result) || math.IsInf(*result, 0) {
					t.Errorf("CAGR is NaN or Inf: %v", *result)
				}
			}
		})
	}
}

func TestCalculateCAGRFromPrices(t *testing.T) {
	tests := []struct {
		name        string
		prices      []float64
		months      int
		expectedNil bool
		description string
	}{
		{
			name:        "empty prices",
			prices:      []float64{},
			months:      60,
			expectedNil: true,
			description: "Empty prices should return nil",
		},
		{
			name:        "insufficient data",
			prices:      makePriceSlice(100.0, 4),
			months:      60,
			expectedNil: true,
			description: "Need at least 6 prices",
		},
		{
			name:        "valid data",
			prices:      makePriceSlice(100.0, 60),
			months:      60,
			expectedNil: false,
			description: "Valid prices should calculate CAGR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCAGRFromPrices(tt.prices, tt.months)
			if (result == nil) != tt.expectedNil {
				t.Errorf("CalculateCAGRFromPrices() = %v, expected nil: %v - %s", result, tt.expectedNil, tt.description)
				return
			}
			if result != nil {
				// CAGR should be valid
				if math.IsNaN(*result) || math.IsInf(*result, 0) {
					t.Errorf("CAGR is NaN or Inf: %v", *result)
				}
			}
		})
	}
}

// Helper functions
func makeMonthlyPrices(start float64, count int) []MonthlyPrice {
	prices := make([]MonthlyPrice, count)
	for i := 0; i < count; i++ {
		prices[i] = MonthlyPrice{
			YearMonth:   "2020-01",
			AvgAdjClose: start + float64(i)*1.0,
		}
	}
	return prices
}

func makeMonthlyPricesWithZero(start float64, count int, zeroIndex int) []MonthlyPrice {
	prices := makeMonthlyPrices(start, count)
	if zeroIndex >= 0 && zeroIndex < len(prices) {
		prices[zeroIndex].AvgAdjClose = 0.0
	}
	return prices
}

func makePriceSlice(start float64, count int) []float64 {
	prices := make([]float64, count)
	for i := 0; i < count; i++ {
		prices[i] = start + float64(i)*1.0
	}
	return prices
}
