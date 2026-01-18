/**
 * Package di provides dependency injection for quantum calculator initialization.
 *
 * Step 17: Initialize Quantum Calculator
 * Quantum calculator provides quantum probability calculations for bubble/trap detection.
 */
package di

import (
	"github.com/aristath/sentinel/internal/modules/quantum"
)

// initializeQuantumCalculator initializes the quantum calculator.
func initializeQuantumCalculator(container *Container) {
	container.QuantumCalculator = quantum.NewQuantumProbabilityCalculator()
}
