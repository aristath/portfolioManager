package optimization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiScaleShortWeight_ExtremesAndNeutral(t *testing.T) {
	// Bearish -> emphasize short-term risk.
	wShortBear := multiScaleShortWeight(-1.0)
	assert.Greater(t, wShortBear, 0.85)
	assert.LessOrEqual(t, wShortBear, 1.0)

	// Neutral -> balanced.
	wShortNeutral := multiScaleShortWeight(0.0)
	assert.InDelta(t, 0.5, wShortNeutral, 1e-12)

	// Bullish -> emphasize long-term risk (short weight low).
	wShortBull := multiScaleShortWeight(1.0)
	assert.Less(t, wShortBull, 0.15)
	assert.GreaterOrEqual(t, wShortBull, 0.0)
}

func TestHRPLinkageForRegime_Thresholds(t *testing.T) {
	assert.Equal(t, hrpLinkageComplete, hrpLinkageForRegime(-0.8))
	assert.Equal(t, hrpLinkageSingle, hrpLinkageForRegime(0.0))
	assert.Equal(t, hrpLinkageAverage, hrpLinkageForRegime(0.8))
}
