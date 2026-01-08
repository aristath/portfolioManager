package symbolic_regression

import (
	"math"
	"math/rand"
	"sort"
)

// FormulaWithFitness represents a formula with its fitness score
type FormulaWithFitness struct {
	Formula    *Node
	Fitness    float64
	Complexity int
}

// EvolutionConfig holds configuration for genetic algorithm evolution
type EvolutionConfig struct {
	PopulationSize   int
	MaxGenerations   int
	MaxDepth         int
	MaxNodes         int
	MutationRate     float64
	CrossoverRate    float64
	TournamentSize   int
	ElitismCount     int
	FitnessType      FitnessType
	ComplexityWeight float64 // Weight for complexity penalty (0 = no penalty)
}

// RunEvolution runs the full genetic programming evolution process
func RunEvolution(
	variables []string,
	examples []TrainingExample,
	config EvolutionConfig,
) *FormulaWithFitness {
	// Initialize population
	population := make([]*FormulaWithFitness, config.PopulationSize)
	for i := 0; i < config.PopulationSize; i++ {
		formula := RandomFormula(variables, config.MaxDepth, config.MaxNodes)
		fitness := CalculateFitness(formula, examples, config.FitnessType)
		complexity := CalculateComplexity(formula)

		// Apply complexity penalty
		adjustedFitness := fitness + config.ComplexityWeight*float64(complexity)

		population[i] = &FormulaWithFitness{
			Formula:    formula,
			Fitness:    adjustedFitness,
			Complexity: complexity,
		}
	}

	// Sort by fitness (lower is better)
	sort.Slice(population, func(i, j int) bool {
		return population[i].Fitness < population[j].Fitness
	})

	// Evolve for max generations
	for generation := 0; generation < config.MaxGenerations; generation++ {
		population = EvolveGeneration(
			population,
			examples,
			config.FitnessType,
			variables,
			config.MutationRate,
			config.CrossoverRate,
			config.TournamentSize,
			config.ElitismCount,
		)

		// Recalculate fitness with complexity penalty
		for _, ind := range population {
			ind.Fitness = CalculateFitness(ind.Formula, examples, config.FitnessType) +
				config.ComplexityWeight*float64(ind.Complexity)
		}

		// Sort by fitness
		sort.Slice(population, func(i, j int) bool {
			return population[i].Fitness < population[j].Fitness
		})
	}

	// Return best individual
	return population[0]
}

// EvolveGeneration evolves one generation of the population
func EvolveGeneration(
	population []*FormulaWithFitness,
	examples []TrainingExample,
	fitnessType FitnessType,
	variables []string,
	mutationRate float64,
	crossoverRate float64,
	tournamentSize int,
	elitismCount int,
) []*FormulaWithFitness {
	newPopulation := make([]*FormulaWithFitness, len(population))

	// Elitism: keep best individuals
	elite := SelectElite(population, elitismCount)
	for i := 0; i < elitismCount && i < len(population); i++ {
		newPopulation[i] = &FormulaWithFitness{
			Formula:    elite[i].Formula.Copy(),
			Fitness:    elite[i].Fitness,
			Complexity: elite[i].Complexity,
		}
	}

	// Generate rest of population through selection, crossover, and mutation
	for i := elitismCount; i < len(population); i++ {
		var child *Node

		if rand.Float64() < crossoverRate {
			// Crossover
			parent1 := TournamentSelection(population, tournamentSize, 1)[0]
			parent2 := TournamentSelection(population, tournamentSize, 1)[0]

			child1, child2 := Crossover(parent1.Formula, parent2.Formula)
			if rand.Float64() < 0.5 {
				child = child1
			} else {
				child = child2
			}
		} else {
			// Clone parent
			parent := TournamentSelection(population, tournamentSize, 1)[0]
			child = parent.Formula.Copy()
		}

		// Mutation
		child = Mutate(child, variables, mutationRate)

		// Calculate fitness
		fitness := CalculateFitness(child, examples, fitnessType)
		complexity := CalculateComplexity(child)

		newPopulation[i] = &FormulaWithFitness{
			Formula:    child,
			Fitness:    fitness,
			Complexity: complexity,
		}
	}

	return newPopulation
}

// TournamentSelection performs tournament selection
func TournamentSelection(
	population []*FormulaWithFitness,
	tournamentSize int,
	numSelections int,
) []*FormulaWithFitness {
	selected := make([]*FormulaWithFitness, numSelections)

	for i := 0; i < numSelections; i++ {
		// Select tournament participants
		tournament := make([]*FormulaWithFitness, tournamentSize)
		for j := 0; j < tournamentSize; j++ {
			tournament[j] = population[rand.Intn(len(population))]
		}

		// Select best from tournament
		best := tournament[0]
		for _, participant := range tournament[1:] {
			if participant.Fitness < best.Fitness {
				best = participant
			}
		}

		selected[i] = best
	}

	return selected
}

// SelectElite selects the best N individuals from the population
func SelectElite(population []*FormulaWithFitness, count int) []*FormulaWithFitness {
	if count > len(population) {
		count = len(population)
	}

	// Create a copy to avoid modifying original
	sorted := make([]*FormulaWithFitness, len(population))
	copy(sorted, population)

	// Sort by fitness (lower is better)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Fitness < sorted[j].Fitness
	})

	// Return top N
	elite := make([]*FormulaWithFitness, count)
	for i := 0; i < count; i++ {
		elite[i] = sorted[i]
	}

	return elite
}

// CalculateAdjustedFitness calculates fitness with complexity penalty
func CalculateAdjustedFitness(
	formula *Node,
	examples []TrainingExample,
	fitnessType FitnessType,
	complexityWeight float64,
) float64 {
	fitness := CalculateFitness(formula, examples, fitnessType)
	complexity := CalculateComplexity(formula)
	return fitness + complexityWeight*float64(complexity)
}

// GetBestFormula returns the best formula from a population
func GetBestFormula(population []*FormulaWithFitness) *FormulaWithFitness {
	if len(population) == 0 {
		return nil
	}

	best := population[0]
	for _, ind := range population[1:] {
		if ind.Fitness < best.Fitness {
			best = ind
		}
	}

	return best
}

// GetAverageFitness calculates average fitness of population
func GetAverageFitness(population []*FormulaWithFitness) float64 {
	if len(population) == 0 {
		return math.MaxFloat64
	}

	sum := 0.0
	for _, ind := range population {
		sum += ind.Fitness
	}

	return sum / float64(len(population))
}

// GetDiversity calculates diversity of population (standard deviation of fitness)
func GetDiversity(population []*FormulaWithFitness) float64 {
	if len(population) < 2 {
		return 0.0
	}

	avg := GetAverageFitness(population)

	sumSqDiff := 0.0
	for _, ind := range population {
		diff := ind.Fitness - avg
		sumSqDiff += diff * diff
	}

	return math.Sqrt(sumSqDiff / float64(len(population)))
}
