package formulas

import (
	"math"
	"testing"
)

func TestCorrelationMatrixFromCovariance(t *testing.T) {
	tests := []struct {
		name        string
		cov         [][]float64
		expectedErr bool
		description string
	}{
		{
			name:        "empty matrix",
			cov:         [][]float64{},
			expectedErr: true,
			description: "Empty matrix should return error",
		},
		{
			name:        "not square matrix",
			cov:         [][]float64{{1.0, 2.0}, {3.0}},
			expectedErr: true,
			description: "Non-square matrix should return error",
		},
		{
			name:        "valid 2x2 matrix",
			cov:         [][]float64{{1.0, 0.5}, {0.5, 1.0}},
			expectedErr: false,
			description: "Valid covariance matrix",
		},
		{
			name:        "invalid variance on diagonal",
			cov:         [][]float64{{0.0, 0.5}, {0.5, 1.0}},
			expectedErr: true,
			description: "Zero variance should return error",
		},
		{
			name:        "NaN on diagonal",
			cov:         [][]float64{{math.NaN(), 0.5}, {0.5, 1.0}},
			expectedErr: true,
			description: "NaN variance should return error",
		},
		{
			name:        "3x3 matrix",
			cov:         [][]float64{{1.0, 0.5, 0.3}, {0.5, 1.0, 0.4}, {0.3, 0.4, 1.0}},
			expectedErr: false,
			description: "Valid 3x3 covariance matrix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CorrelationMatrixFromCovariance(tt.cov)
			if (err != nil) != tt.expectedErr {
				t.Errorf("CorrelationMatrixFromCovariance() error = %v, expectedErr %v - %s", err, tt.expectedErr, tt.description)
				return
			}
			if err == nil && result != nil {
				// Verify correlation matrix properties
				n := len(result)
				if n != len(tt.cov) {
					t.Errorf("Result matrix size = %d, want %d", n, len(tt.cov))
				}
				// Diagonal should be 1.0
				for i := 0; i < n; i++ {
					if math.Abs(result[i][i]-1.0) > 0.0001 {
						t.Errorf("Correlation[%d][%d] = %v, want 1.0", i, i, result[i][i])
					}
					// Matrix should be symmetric
					for j := 0; j < n; j++ {
						if math.Abs(result[i][j]-result[j][i]) > 0.0001 {
							t.Errorf("Matrix not symmetric: [%d][%d]=%v, [%d][%d]=%v", i, j, result[i][j], j, i, result[j][i])
						}
						// Values should be between -1 and 1
						if result[i][j] < -1.0 || result[i][j] > 1.0 {
							t.Errorf("Correlation[%d][%d] = %v, want between -1 and 1", i, j, result[i][j])
						}
					}
				}
			}
		})
	}
}

func TestCorrelationToDistance(t *testing.T) {
	tests := []struct {
		name        string
		corrMatrix  [][]float64
		description string
	}{
		{
			name:        "empty matrix",
			corrMatrix:  [][]float64{},
			description: "Empty matrix should return empty",
		},
		{
			name:        "perfect correlation",
			corrMatrix:  [][]float64{{1.0, 1.0}, {1.0, 1.0}},
			description: "Perfect correlation should give zero distance",
		},
		{
			name:        "perfect negative correlation",
			corrMatrix:  [][]float64{{1.0, -1.0}, {-1.0, 1.0}},
			description: "Perfect negative correlation should give maximum distance",
		},
		{
			name:        "valid 2x2 matrix",
			corrMatrix:  [][]float64{{1.0, 0.5}, {0.5, 1.0}},
			description: "Valid correlation matrix",
		},
		{
			name:        "3x3 matrix",
			corrMatrix:  [][]float64{{1.0, 0.5, 0.3}, {0.5, 1.0, 0.4}, {0.3, 0.4, 1.0}},
			description: "Valid 3x3 correlation matrix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CorrelationToDistance(tt.corrMatrix)
			n := len(tt.corrMatrix)
			if len(result) != n {
				t.Errorf("Distance matrix size = %d, want %d", len(result), n)
				return
			}
			if n == 0 {
				return
			}
			// Verify distance matrix properties
			for i := 0; i < n; i++ {
				if len(result[i]) != n {
					t.Errorf("Row %d size = %d, want %d", i, len(result[i]), n)
					return
				}
				// Diagonal should be 0 (distance to self)
				if math.Abs(result[i][i]) > 0.0001 {
					t.Errorf("Distance[%d][%d] = %v, want 0.0", i, i, result[i][i])
				}
				// Matrix should be symmetric
				for j := 0; j < n; j++ {
					if math.Abs(result[i][j]-result[j][i]) > 0.0001 {
						t.Errorf("Matrix not symmetric: [%d][%d]=%v, [%d][%d]=%v", i, j, result[i][j], j, i, result[j][i])
					}
					// Distance should be >= 0
					if result[i][j] < -0.0001 {
						t.Errorf("Distance[%d][%d] = %v, want >= 0", i, j, result[i][j])
					}
				}
			}
		})
	}
}

func TestInverseVarianceWeights(t *testing.T) {
	tests := []struct {
		name        string
		variances   []float64
		checkSum    bool
		description string
	}{
		{
			name:        "empty variances",
			variances:   []float64{},
			description: "Empty variances should return empty weights",
		},
		{
			name:        "single variance",
			variances:   []float64{2.0},
			checkSum:    true,
			description: "Single variance should give weight 1.0",
		},
		{
			name:        "two variances",
			variances:   []float64{1.0, 4.0},
			checkSum:    true,
			description: "Lower variance should get higher weight",
		},
		{
			name:        "zero variance",
			variances:   []float64{0.0, 1.0, 2.0},
			checkSum:    true,
			description: "Zero variance should get zero weight",
		},
		{
			name:        "all zero variances",
			variances:   []float64{0.0, 0.0, 0.0},
			checkSum:    true,
			description: "All zero variances should get equal weights",
		},
		{
			name:        "multiple variances",
			variances:   []float64{1.0, 2.0, 4.0, 8.0},
			checkSum:    true,
			description: "Inverse variance weighting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InverseVarianceWeights(tt.variances)
			if len(result) != len(tt.variances) {
				t.Errorf("Weights length = %d, want %d", len(result), len(tt.variances))
				return
			}
			if len(result) == 0 {
				return
			}
			// Weights should sum to 1.0
			var sum float64
			for _, w := range result {
				sum += w
			}
			if tt.checkSum && math.Abs(sum-1.0) > 0.0001 {
				t.Errorf("Weights sum = %v, want 1.0", sum)
			}
			// All weights should be >= 0
			for i, w := range result {
				if w < -0.0001 {
					t.Errorf("Weight[%d] = %v, want >= 0", i, w)
				}
			}
		})
	}
}
