package optimization

import (
	"fmt"
	"math"
)

const (
	// Regime thresholds for dynamic clustering choice.
	hrpBearishThreshold = -0.40
	hrpBullishThreshold = 0.40

	// Controls how quickly multi-scale weighting transitions from short-term to long-term.
	multiScaleTanhK = 2.0
)

// multiScaleShortWeight returns the weight assigned to the short-horizon covariance matrix.
// It is a smooth, monotonic function of the regime score:
// - regimeScore -> -1: weight approaches 1 (favor short-term risk)
// - regimeScore ->  0: weight = 0.5
// - regimeScore -> +1: weight approaches 0 (favor long-term risk)
func multiScaleShortWeight(regimeScore float64) float64 {
	// Map regimeScore into [0,1] long-weight using tanh, then invert.
	x := math.Tanh(multiScaleTanhK * regimeScore) // in [-1,1]
	wLong := (x + 1.0) / 2.0                      // in [0,1]
	wShort := 1.0 - wLong
	if wShort < 0 {
		return 0
	}
	if wShort > 1 {
		return 1
	}
	return wShort
}

func hrpLinkageForRegime(regimeScore float64) hrpLinkage {
	if regimeScore <= hrpBearishThreshold {
		return hrpLinkageComplete
	}
	if regimeScore >= hrpBullishThreshold {
		return hrpLinkageAverage
	}
	return hrpLinkageSingle
}

func blendCovariances(covShort [][]float64, covLong [][]float64, wShort float64) ([][]float64, error) {
	if wShort < 0 || wShort > 1 || math.IsNaN(wShort) || math.IsInf(wShort, 0) {
		return nil, fmt.Errorf("invalid wShort: %v", wShort)
	}
	if len(covShort) == 0 || len(covLong) == 0 {
		return nil, fmt.Errorf("empty covariance input")
	}
	if len(covShort) != len(covLong) {
		return nil, fmt.Errorf("covariance dimension mismatch")
	}
	n := len(covShort)
	for i := 0; i < n; i++ {
		if len(covShort[i]) != n || len(covLong[i]) != n {
			return nil, fmt.Errorf("covariance matrix is not square")
		}
	}

	wLong := 1.0 - wShort
	out := make([][]float64, n)
	for i := 0; i < n; i++ {
		out[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			out[i][j] = wShort*covShort[i][j] + wLong*covLong[i][j]
		}
	}
	return out, nil
}
