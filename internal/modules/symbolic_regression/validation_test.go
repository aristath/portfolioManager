package symbolic_regression

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalkForwardValidation(t *testing.T) {
	// Create mock training examples
	trainExamples := []TrainingExample{
		{
			Inputs: TrainingInputs{
				CAGR:        0.10,
				TotalScore:  0.75,
				RegimeScore: 0.3,
			},
			TargetReturn: 0.10,
		},
		{
			Inputs: TrainingInputs{
				CAGR:        0.12,
				TotalScore:  0.80,
				RegimeScore: 0.4,
			},
			TargetReturn: 0.12,
		},
		{
			Inputs: TrainingInputs{
				CAGR:        0.08,
				TotalScore:  0.70,
				RegimeScore: 0.2,
			},
			TargetReturn: 0.08,
		},
	}

	testExamples := []TrainingExample{
		{
			Inputs: TrainingInputs{
				CAGR:        0.11,
				TotalScore:  0.78,
				RegimeScore: 0.35,
			},
			TargetReturn: 0.11,
		},
		{
			Inputs: TrainingInputs{
				CAGR:        0.09,
				TotalScore:  0.72,
				RegimeScore: 0.25,
			},
			TargetReturn: 0.09,
		},
	}

	// Create a simple formula: cagr (identity)
	formula := &Node{
		Type:     NodeTypeVariable,
		Variable: "cagr",
	}

	// Run validation
	metrics := WalkForwardValidation(formula, trainExamples, testExamples, FitnessTypeMAE)

	require.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics["train_mae"], 0.0)
	assert.GreaterOrEqual(t, metrics["test_mae"], 0.0)
	assert.GreaterOrEqual(t, metrics["train_rmse"], 0.0)
	assert.GreaterOrEqual(t, metrics["test_rmse"], 0.0)
}

func TestCompareFormulas(t *testing.T) {
	examples := []TrainingExample{
		{
			Inputs: TrainingInputs{
				CAGR:       0.10,
				TotalScore: 0.75,
			},
			TargetReturn: 0.10,
		},
		{
			Inputs: TrainingInputs{
				CAGR:       0.12,
				TotalScore: 0.80,
			},
			TargetReturn: 0.12,
		},
	}

	// Formula 1: cagr (simple)
	formula1 := &Node{
		Type:     NodeTypeVariable,
		Variable: "cagr",
	}

	// Formula 2: cagr * 0.5 + total_score * 0.5 (weighted)
	formula2 := &Node{
		Type: NodeTypeOperation,
		Op:   OpAdd,
		Left: &Node{
			Type: NodeTypeOperation,
			Op:   OpMultiply,
			Left: &Node{
				Type:     NodeTypeVariable,
				Variable: "cagr",
			},
			Right: &Node{
				Type:  NodeTypeConstant,
				Value: 0.5,
			},
		},
		Right: &Node{
			Type: NodeTypeOperation,
			Op:   OpMultiply,
			Left: &Node{
				Type:     NodeTypeVariable,
				Variable: "total_score",
			},
			Right: &Node{
				Type:  NodeTypeConstant,
				Value: 0.5,
			},
		},
	}

	comparison := CompareFormulas(formula1, formula2, examples, FitnessTypeMAE)

	require.NotNil(t, comparison)
	assert.GreaterOrEqual(t, comparison["formula1_mae"], 0.0)
	assert.GreaterOrEqual(t, comparison["formula2_mae"], 0.0)
	assert.Contains(t, comparison, "improvement_pct")
}
