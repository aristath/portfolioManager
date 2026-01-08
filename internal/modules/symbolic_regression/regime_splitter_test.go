package symbolic_regression

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitByRegime(t *testing.T) {
	examples := []TrainingExample{
		{Inputs: TrainingInputs{RegimeScore: -0.8}}, // Bear
		{Inputs: TrainingInputs{RegimeScore: -0.5}}, // Bear
		{Inputs: TrainingInputs{RegimeScore: 0.0}},  // Neutral
		{Inputs: TrainingInputs{RegimeScore: 0.2}},  // Neutral
		{Inputs: TrainingInputs{RegimeScore: 0.5}},  // Bull
		{Inputs: TrainingInputs{RegimeScore: 0.9}},  // Bull
		{Inputs: TrainingInputs{RegimeScore: 1.0}},  // Bull (edge case)
	}

	ranges := DefaultRegimeRanges()
	split := SplitByRegime(examples, ranges)

	// Check bear regime
	assert.Len(t, split[ranges[0]], 2, "Bear regime should have 2 examples")

	// Check neutral regime
	assert.Len(t, split[ranges[1]], 2, "Neutral regime should have 2 examples")

	// Check bull regime
	assert.Len(t, split[ranges[2]], 3, "Bull regime should have 3 examples (including edge case)")
}

func TestFilterByRegimeRange(t *testing.T) {
	examples := []TrainingExample{
		{Inputs: TrainingInputs{RegimeScore: -0.8}},
		{Inputs: TrainingInputs{RegimeScore: -0.5}},
		{Inputs: TrainingInputs{RegimeScore: 0.0}},
		{Inputs: TrainingInputs{RegimeScore: 0.5}},
		{Inputs: TrainingInputs{RegimeScore: 0.9}},
	}

	// Filter for bear regime
	bearExamples := FilterByRegimeRange(examples, -1.0, -0.3)
	assert.Len(t, bearExamples, 2, "Should filter 2 bear examples")

	// Filter for neutral regime
	neutralExamples := FilterByRegimeRange(examples, -0.3, 0.3)
	assert.Len(t, neutralExamples, 1, "Should filter 1 neutral example")

	// Filter for bull regime
	bullExamples := FilterByRegimeRange(examples, 0.3, 1.0)
	assert.Len(t, bullExamples, 2, "Should filter 2 bull examples")
}
