package quantum

import (
	"math"
	"testing"
)

func TestQuantumProbabilityCalculator_CalculateEnergyLevel(t *testing.T) {
	calc := NewQuantumProbabilityCalculator()

	tests := []struct {
		name      string
		rawEnergy float64
		expected  float64
	}{
		{"Negative energy (π)", -3.5, -math.Pi},
		{"Zero energy", 0.0, 0.0},
		{"Positive energy (π)", 3.5, math.Pi},
		{"Small negative (-π/2)", -2.0, -math.Pi / 2.0},
		{"Small positive (π/2)", 2.0, math.Pi / 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateEnergyLevel(tt.rawEnergy)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("CalculateEnergyLevel() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestQuantumProbabilityCalculator_NormalizeState(t *testing.T) {
	calc := NewQuantumProbabilityCalculator()

	tests := []struct {
		name      string
		pValue    float64
		pBubble   float64
		wantSum   float64
		wantRatio float64
	}{
		{"Equal probabilities", 0.5, 0.5, 1.0, 1.0},
		{"Value dominant", 0.8, 0.2, 1.0, 4.0},
		{"Bubble dominant", 0.2, 0.8, 1.0, 0.25},
		{"Zero values", 0.0, 0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normValue, normBubble := calc.NormalizeState(tt.pValue, tt.pBubble)
			sum := normValue + normBubble
			if math.Abs(sum-tt.wantSum) > 0.001 {
				t.Errorf("NormalizeState() sum = %v, want %v", sum, tt.wantSum)
			}
			ratio := normValue / normBubble
			if math.Abs(ratio-tt.wantRatio) > 0.01 {
				t.Errorf("NormalizeState() ratio = %v, want %v", ratio, tt.wantRatio)
			}
		})
	}
}

func TestQuantumProbabilityCalculator_CalculateInterference(t *testing.T) {
	calc := NewQuantumProbabilityCalculator()

	tests := []struct {
		name      string
		p1        float64
		p2        float64
		energy1   float64
		energy2   float64
		wantRange [2]float64 // [min, max] expected range
	}{
		{"Same energy (constructive)", 0.5, 0.5, 0.0, 0.0, [2]float64{0.9, 1.1}},
		{"Opposite energy", 0.5, 0.5, -math.Pi, math.Pi, [2]float64{0.9, 1.1}}, // cos(2π) = 1
		{"Different probabilities", 0.8, 0.2, 0.0, math.Pi / 2.0, [2]float64{-0.5, 0.5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateInterference(tt.p1, tt.p2, tt.energy1, tt.energy2)
			if result < tt.wantRange[0] || result > tt.wantRange[1] {
				t.Errorf("CalculateInterference() = %v, want in range %v", result, tt.wantRange)
			}
		})
	}
}

func TestQuantumProbabilityCalculator_CalculateBubbleProbability(t *testing.T) {
	calc := NewQuantumProbabilityCalculator()

	tests := []struct {
		name           string
		cagr           float64
		sharpe         float64
		sortino        float64
		volatility     float64
		fundamentals   float64
		regimeScore    float64
		wantRange      [2]float64
		wantHighBubble bool
	}{
		{
			name:           "High CAGR, poor risk (bubble)",
			cagr:           0.20, // 20%
			sharpe:         0.3,  // Poor
			sortino:        0.3,  // Poor
			volatility:     0.45, // High
			fundamentals:   0.5,  // Low
			regimeScore:    0.0,
			wantRange:      [2]float64{0.6, 1.0},
			wantHighBubble: true,
		},
		{
			name:           "High CAGR, good risk (not bubble)",
			cagr:           0.18, // 18%
			sharpe:         1.5,  // Good
			sortino:        1.5,  // Good
			volatility:     0.25, // Moderate
			fundamentals:   0.8,  // High
			regimeScore:    0.0,
			wantRange:      [2]float64{0.0, 0.5},
			wantHighBubble: false,
		},
		{
			name:           "Low CAGR (not bubble)",
			cagr:           0.10, // 10%
			sharpe:         0.5,
			sortino:        0.5,
			volatility:     0.30,
			fundamentals:   0.7,
			regimeScore:    0.0,
			wantRange:      [2]float64{0.0, 0.4},
			wantHighBubble: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateBubbleProbability(
				tt.cagr,
				tt.sharpe,
				tt.sortino,
				tt.volatility,
				tt.fundamentals,
				tt.regimeScore,
				nil,
			)

			if result < tt.wantRange[0] || result > tt.wantRange[1] {
				t.Errorf("CalculateBubbleProbability() = %v, want in range %v", result, tt.wantRange)
			}

			isHighBubble := result > 0.7
			if isHighBubble != tt.wantHighBubble {
				t.Errorf("CalculateBubbleProbability() high bubble = %v, want %v", isHighBubble, tt.wantHighBubble)
			}
		})
	}
}

func TestQuantumProbabilityCalculator_CalculateValueTrapProbability(t *testing.T) {
	calc := NewQuantumProbabilityCalculator()

	tests := []struct {
		name         string
		peVsMarket   float64
		fundamentals float64
		longTerm     float64
		momentum     float64
		volatility   float64
		regimeScore  float64
		wantRange    [2]float64
		wantTrap     bool
	}{
		{
			name:         "Cheap with poor fundamentals (trap)",
			peVsMarket:   -0.30, // 30% cheaper
			fundamentals: 0.4,   // Poor
			longTerm:     0.3,   // Poor
			momentum:     -0.1,  // Negative
			volatility:   0.40,  // High
			regimeScore:  0.0,
			wantRange:    [2]float64{0.6, 1.0},
			wantTrap:     true,
		},
		{
			name:         "Cheap with good fundamentals (value)",
			peVsMarket:   -0.25, // 25% cheaper
			fundamentals: 0.8,   // Good
			longTerm:     0.8,   // Good
			momentum:     0.1,   // Positive
			volatility:   0.20,  // Low
			regimeScore:  0.0,
			wantRange:    [2]float64{0.0, 0.4},
			wantTrap:     false,
		},
		{
			name:         "Not cheap (not trap)",
			peVsMarket:   -0.10, // Only 10% cheaper
			fundamentals: 0.5,
			longTerm:     0.5,
			momentum:     0.0,
			volatility:   0.30,
			regimeScore:  0.0,
			wantRange:    [2]float64{0.0, 0.3},
			wantTrap:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateValueTrapProbability(
				tt.peVsMarket,
				tt.fundamentals,
				tt.longTerm,
				tt.momentum,
				tt.volatility,
				tt.regimeScore,
			)

			if result < tt.wantRange[0] || result > tt.wantRange[1] {
				t.Errorf("CalculateValueTrapProbability() = %v, want in range %v", result, tt.wantRange)
			}

			isTrap := result > 0.7
			if isTrap != tt.wantTrap {
				t.Errorf("CalculateValueTrapProbability() trap = %v, want %v", isTrap, tt.wantTrap)
			}
		})
	}
}

func TestQuantumProbabilityCalculator_BornRule(t *testing.T) {
	calc := NewQuantumProbabilityCalculator()

	tests := []struct {
		name      string
		amplitude complex128
		want      float64
	}{
		{"Real amplitude 1", complex(1.0, 0.0), 1.0},
		{"Real amplitude 0.5", complex(0.5, 0.0), 0.25},
		{"Complex amplitude", complex(0.707, 0.707), 1.0}, // |0.707+0.707i|² = 1.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.BornRule(tt.amplitude)
			if math.Abs(result-tt.want) > 0.01 {
				t.Errorf("BornRule() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestQuantumProbabilityCalculator_CalculateAdaptiveInterferenceWeight(t *testing.T) {
	calc := NewQuantumProbabilityCalculator()

	tests := []struct {
		name        string
		regimeScore float64
		want        float64
	}{
		{"Bull market", 0.6, 0.4},
		{"Bear market", -0.6, 0.2},
		{"Sideways", 0.0, 0.3},
		{"Neutral", 0.2, 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateAdaptiveInterferenceWeight(tt.regimeScore)
			if math.Abs(result-tt.want) > 0.01 {
				t.Errorf("CalculateAdaptiveInterferenceWeight() = %v, want %v", result, tt.want)
			}
		})
	}
}

func BenchmarkCalculateBubbleProbability(b *testing.B) {
	calc := NewQuantumProbabilityCalculator()
	cagr := 0.18
	sharpe := 0.5
	sortino := 0.5
	volatility := 0.35
	fundamentals := 0.7
	regimeScore := 0.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calc.CalculateBubbleProbability(cagr, sharpe, sortino, volatility, fundamentals, regimeScore, nil)
	}
}

func BenchmarkCalculateValueTrapProbability(b *testing.B) {
	calc := NewQuantumProbabilityCalculator()
	peVsMarket := -0.25
	fundamentals := 0.6
	longTerm := 0.6
	momentum := 0.0
	volatility := 0.30
	regimeScore := 0.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calc.CalculateValueTrapProbability(peVsMarket, fundamentals, longTerm, momentum, volatility, regimeScore)
	}
}

func TestQuantumProbabilityCalculator_SetTimeParameter(t *testing.T) {
	calc := NewQuantumProbabilityCalculator()

	// Default time parameter should be 1.0
	calc.SetTimeParameter(2.0)
	// Verify it affects interference calculation
	result1 := calc.CalculateInterference(0.5, 0.5, 0.0, math.Pi/2.0)

	calc.SetTimeParameter(0.5)
	result2 := calc.CalculateInterference(0.5, 0.5, 0.0, math.Pi/2.0)

	// Different time parameters should produce different interference values
	if math.Abs(result1-result2) < 0.01 {
		t.Errorf("SetTimeParameter() should affect interference calculation, got same results: %v", result1)
	}
}

func TestQuantumProbabilityCalculator_CalculateMultimodalCorrection(t *testing.T) {
	calc := NewQuantumProbabilityCalculator()

	tests := []struct {
		name        string
		volatility  float64
		kurtosis    *float64
		wantRange   [2]float64
		description string
	}{
		{
			name:        "low volatility, no kurtosis",
			volatility:  0.1,
			kurtosis:    nil,
			wantRange:   [2]float64{0.0, 0.02},
			description: "Should be small correction",
		},
		{
			name:        "high volatility, no kurtosis",
			volatility:  0.5,
			kurtosis:    nil,
			wantRange:   [2]float64{0.04, 0.06},
			description: "Should be larger correction",
		},
		{
			name:        "low volatility, high kurtosis",
			volatility:  0.1,
			kurtosis:    floatPtr(10.0),
			wantRange:   [2]float64{0.03, 0.05},
			description: "Kurtosis should increase correction",
		},
		{
			name:        "high volatility, high kurtosis",
			volatility:  0.5,
			kurtosis:    floatPtr(10.0),
			wantRange:   [2]float64{0.15, 0.21}, // Should be capped at 0.2
			description: "Should be capped at 0.2",
		},
		{
			name:        "negative kurtosis normalized to 0",
			volatility:  0.3,
			kurtosis:    floatPtr(-5.0),
			wantRange:   [2]float64{0.02, 0.04},
			description: "Negative kurtosis should be normalized",
		},
		{
			name:        "very high kurtosis normalized",
			volatility:  0.3,
			kurtosis:    floatPtr(20.0),
			wantRange:   [2]float64{0.12, 0.14}, // 0.1 * 0.3 * (1.0 + 10.0/3.0) = 0.13
			description: "Very high kurtosis should be normalized to 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateMultimodalCorrection(tt.volatility, tt.kurtosis)
			if result < tt.wantRange[0] || result > tt.wantRange[1] {
				t.Errorf("CalculateMultimodalCorrection() = %v, want in range %v - %s", result, tt.wantRange, tt.description)
			}
			// Verify it's capped at 0.2
			if result > 0.2 {
				t.Errorf("CalculateMultimodalCorrection() = %v, should be capped at 0.2", result)
			}
		})
	}
}

func TestQuantumProbabilityCalculator_CalculateQuantumAmplitude(t *testing.T) {
	calc := NewQuantumProbabilityCalculator()

	tests := []struct {
		name        string
		probability float64
		energy      float64
		wantReal    float64
		wantImag    float64
		tolerance   float64
	}{
		{
			name:        "probability 1.0, energy 0",
			probability: 1.0,
			energy:      0.0,
			wantReal:    1.0,
			wantImag:    0.0,
			tolerance:   0.01,
		},
		{
			name:        "probability 0.25, energy 0",
			probability: 0.25,
			energy:      0.0,
			wantReal:    0.5, // sqrt(0.25) = 0.5
			wantImag:    0.0,
			tolerance:   0.01,
		},
		{
			name:        "probability 0.5, energy π/2",
			probability: 0.5,
			energy:      math.Pi / 2.0,
			wantReal:    0.0,   // cos(π/2) = 0
			wantImag:    0.707, // sin(π/2) = 1, sqrt(0.5) * 1 ≈ 0.707
			tolerance:   0.01,
		},
		{
			name:        "probability out of range (clamped)",
			probability: 1.5, // > 1.0, should be clamped
			energy:      0.0,
			wantReal:    1.0, // sqrt(1.0) = 1.0
			wantImag:    0.0,
			tolerance:   0.01,
		},
		{
			name:        "negative probability (clamped)",
			probability: -0.5, // < 0.0, should be clamped
			energy:      0.0,
			wantReal:    0.0, // sqrt(0.0) = 0.0
			wantImag:    0.0,
			tolerance:   0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateQuantumAmplitude(tt.probability, tt.energy)
			realDiff := math.Abs(real(result) - tt.wantReal)
			imagDiff := math.Abs(imag(result) - tt.wantImag)

			if realDiff > tt.tolerance {
				t.Errorf("CalculateQuantumAmplitude() real part = %v, want %v (diff: %v)", real(result), tt.wantReal, realDiff)
			}
			if imagDiff > tt.tolerance {
				t.Errorf("CalculateQuantumAmplitude() imag part = %v, want %v (diff: %v)", imag(result), tt.wantImag, imagDiff)
			}
		})
	}
}

// Helper function to create float pointer
func floatPtr(f float64) *float64 {
	return &f
}
