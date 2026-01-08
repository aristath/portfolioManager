package symbolic_regression

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTournamentSelection(t *testing.T) {
	population := []*FormulaWithFitness{
		{Formula: &Node{Type: NodeTypeConstant, Value: 1.0}, Fitness: 0.1}, // Best
		{Formula: &Node{Type: NodeTypeConstant, Value: 2.0}, Fitness: 0.5},
		{Formula: &Node{Type: NodeTypeConstant, Value: 3.0}, Fitness: 0.9}, // Worst
		{Formula: &Node{Type: NodeTypeConstant, Value: 4.0}, Fitness: 0.3},
	}

	selected := TournamentSelection(population, 2, 10) // Tournament size 2, select 10

	require.Equal(t, 10, len(selected))

	// Best individual should be selected multiple times
	bestSelected := 0
	for _, s := range selected {
		if s.Fitness == 0.1 {
			bestSelected++
		}
	}
	assert.Greater(t, bestSelected, 0, "Best individual should be selected")
}

func TestEvolveGeneration(t *testing.T) {
	variables := []string{"cagr", "score"}

	// Create initial population
	population := make([]*FormulaWithFitness, 10)
	for i := 0; i < 10; i++ {
		population[i] = &FormulaWithFitness{
			Formula: RandomFormula(variables, 3, 5),
			Fitness: 1.0, // Will be recalculated
		}
	}

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

	// Evolve one generation
	newPopulation := EvolveGeneration(
		population,
		examples,
		FitnessTypeMAE,
		variables,
		0.1, // Mutation rate
		0.7, // Crossover rate
		2,   // Tournament size
		2,   // Elitism count
	)

	require.Equal(t, len(population), len(newPopulation))

	// Population should still be valid
	for _, ind := range newPopulation {
		assert.NotNil(t, ind.Formula)
		assert.GreaterOrEqual(t, ind.Fitness, 0.0)
	}
}

func TestRunEvolution(t *testing.T) {
	variables := []string{"cagr", "score", "regime"}

	// Create training examples
	examples := []TrainingExample{
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

	config := EvolutionConfig{
		PopulationSize:   20,
		MaxGenerations:   10,
		MaxDepth:         3,
		MaxNodes:         5,
		MutationRate:     0.1,
		CrossoverRate:    0.7,
		TournamentSize:   2,
		ElitismCount:     2,
		FitnessType:      FitnessTypeMAE,
		ComplexityWeight: 0.01, // Small complexity penalty
	}

	best := RunEvolution(variables, examples, config)

	require.NotNil(t, best)
	assert.NotNil(t, best.Formula)
	assert.GreaterOrEqual(t, best.Fitness, 0.0)
}

func TestElitism(t *testing.T) {
	population := []*FormulaWithFitness{
		{Formula: &Node{Type: NodeTypeConstant, Value: 1.0}, Fitness: 0.1}, // Best
		{Formula: &Node{Type: NodeTypeConstant, Value: 2.0}, Fitness: 0.5},
		{Formula: &Node{Type: NodeTypeConstant, Value: 3.0}, Fitness: 0.9}, // Worst
	}

	elite := SelectElite(population, 2)

	require.Equal(t, 2, len(elite))

	// Should select the best individuals
	assert.Equal(t, 0.1, elite[0].Fitness)
	assert.Equal(t, 0.5, elite[1].Fitness)
}
