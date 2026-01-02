package formulas

import (
	"math"
	"testing"
)

func TestCalculateAnnualReturn(t *testing.T) {
	tests := []struct {
		name      string
		returns   []float64
		expected  float64
		tolerance float64
	}{
		{
			name:      "empty returns",
			returns:   []float64{},
			expected:  0.0,
			tolerance: 0.0,
		},
		{
			name:      "one year of small positive returns",
			returns:   makeReturns(0.001, 252), // 252 daily returns of 0.1%
			expected:  0.286,                   // Approximately 28.6% annualized
			tolerance: 0.01,
		},
		{
			name:      "half year of returns",
			returns:   makeReturns(0.002, 126), // 126 days (half year) of 0.2% returns
			expected:  0.654,                   // CAGR: (1.002^126)^(252/126) - 1 ≈ 65.4%
			tolerance: 0.01,
		},
		{
			name:      "one year of negative returns",
			returns:   makeReturns(-0.001, 252),
			expected:  -0.221, // Negative annualized return
			tolerance: 0.01,
		},
		{
			name:      "very short period",
			returns:   []float64{0.01, 0.02},
			expected:  0.0302, // Simple cumulative for very short periods
			tolerance: 0.001,
		},
		{
			name:      "mixed returns",
			returns:   []float64{0.01, -0.005, 0.02, -0.01, 0.015},
			expected:  3.44, // CAGR over 5 days: (1.0303)^(252/5) - 1 ≈ 3.44
			tolerance: 0.1,
		},
		{
			name:      "zero returns",
			returns:   makeReturns(0.0, 252),
			expected:  0.0,
			tolerance: 0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateAnnualReturn(tt.returns)
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("CalculateAnnualReturn() = %v, want %v (±%v)", result, tt.expected, tt.tolerance)
			}
		})
	}
}

func TestAnnualizedVolatility(t *testing.T) {
	tests := []struct {
		name      string
		returns   []float64
		expected  float64
		tolerance float64
	}{
		{
			name:      "empty returns",
			returns:   []float64{},
			expected:  0.0,
			tolerance: 0.0,
		},
		{
			name:      "constant returns",
			returns:   makeReturns(0.001, 252),
			expected:  0.0, // No volatility when all returns are same
			tolerance: 0.001,
		},
		{
			name:      "mixed returns",
			returns:   []float64{0.01, -0.01, 0.02, -0.02, 0.015, -0.015},
			expected:  0.244, // Some volatility
			tolerance: 0.05,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnnualizedVolatility(tt.returns)
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("AnnualizedVolatility() = %v, want %v (±%v)", result, tt.expected, tt.tolerance)
			}
		})
	}
}

// Helper function to create a slice of identical returns
func makeReturns(value float64, count int) []float64 {
	returns := make([]float64, count)
	for i := range returns {
		returns[i] = value
	}
	return returns
}
