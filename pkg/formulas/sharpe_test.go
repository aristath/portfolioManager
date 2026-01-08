package formulas

import (
	"math"
	"testing"
)

func TestCalculateSharpeRatio(t *testing.T) {
	tests := []struct {
		name           string
		returns        []float64
		riskFreeRate   float64
		periodsPerYear int
		expectedNil    bool
		description    string
	}{
		{
			name:           "insufficient data",
			returns:        []float64{0.01},
			riskFreeRate:   0.02,
			periodsPerYear: 252,
			expectedNil:    true,
			description:    "Need at least 2 returns",
		},
		{
			name:           "zero volatility",
			returns:        makeReturns(0.001, 10),
			riskFreeRate:   0.02,
			periodsPerYear: 252,
			expectedNil:    true,
			description:    "Zero standard deviation returns nil",
		},
		{
			name:           "valid daily returns",
			returns:        []float64{0.01, -0.005, 0.02, -0.01, 0.015},
			riskFreeRate:   0.02,
			periodsPerYear: 252,
			expectedNil:    false,
			description:    "Valid returns should calculate Sharpe",
		},
		{
			name:           "monthly returns",
			returns:        []float64{0.05, -0.02, 0.08, -0.03, 0.06},
			riskFreeRate:   0.03,
			periodsPerYear: 12,
			expectedNil:    false,
			description:    "Monthly returns with 12 periods per year",
		},
		{
			name:           "empty returns",
			returns:        []float64{},
			riskFreeRate:   0.02,
			periodsPerYear: 252,
			expectedNil:    true,
			description:    "Empty returns should return nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateSharpeRatio(tt.returns, tt.riskFreeRate, tt.periodsPerYear)
			if (result == nil) != tt.expectedNil {
				t.Errorf("CalculateSharpeRatio() = %v, expected nil: %v - %s", result, tt.expectedNil, tt.description)
				return
			}
			if result != nil {
				// Sharpe ratio should be a valid number
				if math.IsNaN(*result) || math.IsInf(*result, 0) {
					t.Errorf("Sharpe ratio is NaN or Inf: %v", *result)
				}
			}
		})
	}
}

func TestCalculateSharpeFromPrices(t *testing.T) {
	tests := []struct {
		name         string
		prices       []float64
		riskFreeRate float64
		expectedNil  bool
		description  string
	}{
		{
			name:         "insufficient prices",
			prices:       []float64{100.0},
			riskFreeRate: 0.02,
			expectedNil:  true,
			description:  "Need at least 2 prices",
		},
		{
			name:         "valid prices",
			prices:       []float64{100.0, 105.0, 110.0, 108.0, 115.0},
			riskFreeRate: 0.02,
			expectedNil:  false,
			description:  "Valid prices should calculate Sharpe",
		},
		{
			name:         "empty prices",
			prices:       []float64{},
			riskFreeRate: 0.02,
			expectedNil:  true,
			description:  "Empty prices should return nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateSharpeFromPrices(tt.prices, tt.riskFreeRate)
			if (result == nil) != tt.expectedNil {
				t.Errorf("CalculateSharpeFromPrices() = %v, expected nil: %v - %s", result, tt.expectedNil, tt.description)
				return
			}
			if result != nil {
				// Sharpe ratio should be valid
				if math.IsNaN(*result) || math.IsInf(*result, 0) {
					t.Errorf("Sharpe ratio is NaN or Inf: %v", *result)
				}
			}
		})
	}
}

func TestCalculateSortinoRatio(t *testing.T) {
	tests := []struct {
		name           string
		returns        []float64
		riskFreeRate   float64
		targetReturn   float64
		periodsPerYear int
		expectedNil    bool
		description    string
	}{
		{
			name:           "insufficient data",
			returns:        []float64{0.01},
			riskFreeRate:   0.02,
			targetReturn:   0.05,
			periodsPerYear: 252,
			expectedNil:    true,
			description:    "Need at least 2 returns",
		},
		{
			name:           "all returns above MAR",
			returns:        []float64{0.05, 0.06, 0.07, 0.08},
			riskFreeRate:   0.02,
			targetReturn:   0.03,
			periodsPerYear: 252,
			expectedNil:    true,
			description:    "No downside below MAR returns nil",
		},
		{
			name:           "valid with downside",
			returns:        []float64{0.01, -0.02, 0.015, -0.01, 0.02},
			riskFreeRate:   0.02,
			targetReturn:   0.05,
			periodsPerYear: 252,
			expectedNil:    false,
			description:    "Valid returns with downside should calculate Sortino",
		},
		{
			name:           "empty returns",
			returns:        []float64{},
			riskFreeRate:   0.02,
			targetReturn:   0.05,
			periodsPerYear: 252,
			expectedNil:    true,
			description:    "Empty returns should return nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateSortinoRatio(tt.returns, tt.riskFreeRate, tt.targetReturn, tt.periodsPerYear)
			if (result == nil) != tt.expectedNil {
				t.Errorf("CalculateSortinoRatio() = %v, expected nil: %v - %s", result, tt.expectedNil, tt.description)
				return
			}
			if result != nil {
				// Sortino ratio should be valid
				if math.IsNaN(*result) || math.IsInf(*result, 0) {
					t.Errorf("Sortino ratio is NaN or Inf: %v", *result)
				}
			}
		})
	}
}
