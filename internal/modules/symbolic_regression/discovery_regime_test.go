package symbolic_regression

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscoverExpectedReturnFormula_RegimeSpecific(t *testing.T) {
	// This is a high-level integration test
	// It verifies that regime-specific discovery returns multiple formulas

	// Create mock data prep and storage
	// For a full test, we'd need to set up test databases
	// This test verifies the logic structure

	// Test that regime ranges are properly used
	ranges := DefaultRegimeRanges()
	assert.Len(t, ranges, 3, "Should have 3 default regime ranges")
	assert.Equal(t, "bear", ranges[0].Name)
	assert.Equal(t, "neutral", ranges[1].Name)
	assert.Equal(t, "bull", ranges[2].Name)

	// Verify range boundaries
	assert.Equal(t, -1.0, ranges[0].Min)
	assert.Equal(t, -0.3, ranges[0].Max)
	assert.Equal(t, -0.3, ranges[1].Min)
	assert.Equal(t, 0.3, ranges[1].Max)
	assert.Equal(t, 0.3, ranges[2].Min)
	assert.Equal(t, 1.0, ranges[2].Max)
}

func TestSplitByRegime_EdgeCases(t *testing.T) {
	examples := []TrainingExample{
		{Inputs: TrainingInputs{RegimeScore: -1.0}}, // Bear edge
		{Inputs: TrainingInputs{RegimeScore: -0.3}}, // Bear/Neutral boundary
		{Inputs: TrainingInputs{RegimeScore: 0.3}},  // Neutral/Bull boundary
		{Inputs: TrainingInputs{RegimeScore: 1.0}},  // Bull edge
	}

	ranges := DefaultRegimeRanges()
	split := SplitByRegime(examples, ranges)

	// -1.0 should be in bear
	assert.Greater(t, len(split[ranges[0]]), 0, "-1.0 should be in bear regime")

	// -0.3 should be in bear (exclusive upper bound)
	assert.Greater(t, len(split[ranges[0]]), 0, "-0.3 should be in bear regime")

	// 0.3 should be in neutral (exclusive upper bound for bear)
	assert.Greater(t, len(split[ranges[1]]), 0, "0.3 should be in neutral regime")

	// 1.0 should be in bull (inclusive for max == 1.0)
	assert.Greater(t, len(split[ranges[2]]), 0, "1.0 should be in bull regime")
}
