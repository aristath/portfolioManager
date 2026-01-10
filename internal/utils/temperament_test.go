package utils

import (
	"math"
	"testing"
)

func TestGetTemperamentValue_Linear(t *testing.T) {
	params := TemperamentParams{
		Type:       "risk_tolerance",
		Min:        0.3,
		Max:        0.7,
		Progression: "linear",
		Base:       0.5,
	}

	// Test value = 0.0 (should return Min)
	result := GetTemperamentValue(params)
	params.Value = 0.0
	if math.Abs(result-0.3) > 0.01 {
		t.Errorf("Expected ~0.3 for value=0.0, got %f", result)
	}

	// Test value = 0.5 (should return Base)
	params.Value = 0.5
	result = GetTemperamentValue(params)
	if math.Abs(result-0.5) > 0.01 {
		t.Errorf("Expected ~0.5 for value=0.5, got %f", result)
	}

	// Test value = 1.0 (should return Max)
	params.Value = 1.0
	result = GetTemperamentValue(params)
	if math.Abs(result-0.7) > 0.01 {
		t.Errorf("Expected ~0.7 for value=1.0, got %f", result)
	}
}

func TestGetTemperamentValue_LinearReverse(t *testing.T) {
	params := TemperamentParams{
		Type:       "risk_tolerance",
		Value:      0.0,
		Min:        0.3,
		Max:        0.7,
		Progression: "linear-reverse",
		Base:       0.5,
	}

	// Test value = 0.0 (reverse: should return Max)
	result := GetTemperamentValue(params)
	if math.Abs(result-0.7) > 0.01 {
		t.Errorf("Expected ~0.7 for value=0.0 (reverse), got %f", result)
	}

	// Test value = 0.5 (should return Base)
	params.Value = 0.5
	result = GetTemperamentValue(params)
	if math.Abs(result-0.5) > 0.01 {
		t.Errorf("Expected ~0.5 for value=0.5, got %f", result)
	}

	// Test value = 1.0 (reverse: should return Min)
	params.Value = 1.0
	result = GetTemperamentValue(params)
	if math.Abs(result-0.3) > 0.01 {
		t.Errorf("Expected ~0.3 for value=1.0 (reverse), got %f", result)
	}
}

func TestGetTemperamentValue_Exponential(t *testing.T) {
	params := TemperamentParams{
		Type:       "risk_tolerance",
		Min:        0.3,
		Max:        0.7,
		Progression: "exponential",
		Base:       0.5,
	}

	// Test value = 0.0 (should return Min)
	params.Value = 0.0
	result := GetTemperamentValue(params)
	if math.Abs(result-0.3) > 0.01 {
		t.Errorf("Expected ~0.3 for value=0.0, got %f", result)
	}

	// Test value = 0.5 (should return Base)
	params.Value = 0.5
	result = GetTemperamentValue(params)
	if math.Abs(result-0.5) > 0.05 { // Allow slightly more tolerance for non-linear curves
		t.Errorf("Expected ~0.5 for value=0.5, got %f", result)
	}

	// Test value = 1.0 (should return Max)
	params.Value = 1.0
	result = GetTemperamentValue(params)
	if math.Abs(result-0.7) > 0.01 {
		t.Errorf("Expected ~0.7 for value=1.0, got %f", result)
	}
}

func TestGetTemperamentValue_Sigmoid(t *testing.T) {
	params := TemperamentParams{
		Type:       "risk_tolerance",
		Min:        0.3,
		Max:        0.7,
		Progression: "sigmoid",
		Base:       0.5,
	}

	// Test value = 0.5 (should return Base exactly)
	params.Value = 0.5
	result := GetTemperamentValue(params)
	if math.Abs(result-0.5) > 0.01 {
		t.Errorf("Expected ~0.5 for value=0.5 (sigmoid), got %f", result)
	}

	// Test value = 0.0 (should be closer to Min)
	params.Value = 0.0
	result = GetTemperamentValue(params)
	if result < 0.3 || result > 0.5 {
		t.Errorf("Expected result between 0.3 and 0.5 for value=0.0, got %f", result)
	}

	// Test value = 1.0 (should be closer to Max)
	params.Value = 1.0
	result = GetTemperamentValue(params)
	if result < 0.5 || result > 0.7 {
		t.Errorf("Expected result between 0.5 and 0.7 for value=1.0, got %f", result)
	}
}

func TestGetTemperamentValue_Bounds(t *testing.T) {
	params := TemperamentParams{
		Type:       "risk_tolerance",
		Min:        0.2,
		Max:        0.8,
		Progression: "linear",
		Base:       0.5,
	}

	// Test clamping: value < 0
	params.Value = -0.5
	result := GetTemperamentValue(params)
	if result < 0.2 || result > 0.8 {
		t.Errorf("Result %f should be clamped to [0.2, 0.8]", result)
	}

	// Test clamping: value > 1
	params.Value = 1.5
	result = GetTemperamentValue(params)
	if result < 0.2 || result > 0.8 {
		t.Errorf("Result %f should be clamped to [0.2, 0.8]", result)
	}
}

func TestGetTemperamentValue_AllProgressions(t *testing.T) {
	progressions := []string{
		"linear",
		"linear-reverse",
		"exponential",
		"exponential-reverse",
		"logarithmic",
		"logarithmic-reverse",
		"sigmoid",
		"sigmoid-reverse",
	}

	params := TemperamentParams{
		Type:  "risk_tolerance",
		Min:   0.3,
		Max:   0.7,
		Base:  0.5,
		Value: 0.5,
	}

	for _, prog := range progressions {
		params.Progression = prog
		result := GetTemperamentValue(params)

		// At value=0.5, result should be close to Base (0.5)
		// Allow tolerance of 0.1 for non-linear curves that may have slight deviations
		if math.Abs(result-0.5) > 0.1 {
			t.Errorf("Progression %s: Expected ~0.5 for value=0.5, got %f", prog, result)
		}

		// Result should always be within bounds
		if result < 0.3 || result > 0.7 {
			t.Errorf("Progression %s: Result %f should be in [0.3, 0.7]", prog, result)
		}
	}
}

func TestGetTemperamentValue_EdgeCases(t *testing.T) {
	params := TemperamentParams{
		Type:       "risk_tolerance",
		Min:        0.0,
		Max:        1.0,
		Progression: "linear",
		Base:       0.5,
	}

	// Test with Base not in center
	params.Min = 0.3
	params.Max = 0.9
	params.Base = 0.6 // Base closer to min
	params.Value = 0.5
	result := GetTemperamentValue(params)
	if math.Abs(result-0.6) > 0.05 {
		t.Errorf("Expected ~0.6 for Base=0.6, got %f", result)
	}

	// Test with Base = Min
	params.Base = 0.3
	result = GetTemperamentValue(params)
	if result < 0.3 || result > 0.9 {
		t.Errorf("Result %f should be in [0.3, 0.9]", result)
	}
}
