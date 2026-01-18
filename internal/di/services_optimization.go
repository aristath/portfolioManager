/**
 * Package di provides dependency injection for optimization service initialization.
 *
 * Step 10: Initialize Optimization Services
 * Optimization services handle portfolio optimization (HRP, Mean-Variance, Black-Litterman).
 */
package di

import (
	"github.com/aristath/sentinel/internal/modules/optimization"
	"github.com/rs/zerolog"
)

// initializeOptimizationServices initializes optimization-related services.
// Note: ReturnsCalc must already be initialized (done in planning services).
func initializeOptimizationServices(container *Container, log zerolog.Logger) error {
	// Constraints manager
	// Manages portfolio constraints (allocation limits, concentration limits, etc.)
	container.ConstraintsMgr = optimization.NewConstraintsManager(log)

	// Note: ReturnsCalc already initialized above (before OpportunityContextBuilder) for unified expected returns

	// Kelly Position Sizer
	// Calculates optimal position sizes using Kelly Criterion
	// Default parameters will be overridden by temperament settings
	container.KellySizer = optimization.NewKellyPositionSizer(
		0.02,  // riskFreeRate: 2%
		0.5,   // fixedFractional: 0.5 (half-Kelly) - default, will be overridden by temperament
		0.005, // minPositionSize: 0.5% - default, will be overridden by temperament
		0.20,  // maxPositionSize: 20% - default, will be overridden by temperament
		container.ReturnsCalc,
		container.RiskBuilder,
		container.RegimeDetector,
	)
	// Wire settings service for temperament-aware Kelly parameters
	// Kelly sizing adjusts based on user's risk tolerance and aggression
	kellySettingsAdapterInstance := &kellySettingsAdapter{service: container.SettingsService}
	container.KellySizer.SetSettingsService(kellySettingsAdapterInstance)

	// CVaR Calculator
	// Calculates Conditional Value at Risk (expected shortfall)
	container.CVaRCalculator = optimization.NewCVaRCalculator(
		container.RiskBuilder,
		container.RegimeDetector,
		log,
	)

	// View Generator (for Black-Litterman)
	// Generates market views for Black-Litterman optimization
	container.ViewGenerator = optimization.NewViewGenerator(log)

	// Black-Litterman Optimizer
	// Implements Black-Litterman portfolio optimization (combines market equilibrium with views)
	container.BlackLittermanOptimizer = optimization.NewBlackLittermanOptimizer(
		container.ViewGenerator,
		container.RiskBuilder,
		log,
	)

	// Optimizer service
	// Main portfolio optimization service (HRP, Mean-Variance, Black-Litterman)
	container.OptimizerService = optimization.NewOptimizerService(
		container.ConstraintsMgr,
		container.ReturnsCalc,
		container.RiskBuilder,
		log,
	)

	// Wire Kelly Sizer into OptimizerService
	// Optimizer uses Kelly sizing for position size recommendations
	container.OptimizerService.SetKellySizer(container.KellySizer)

	// Wire CVaR Calculator into OptimizerService
	// Optimizer uses CVaR for risk-adjusted optimization
	container.OptimizerService.SetCVaRCalculator(container.CVaRCalculator)

	// Wire Settings Service into OptimizerService (for CVaR threshold configuration)
	// CVaR thresholds adjust based on user's risk tolerance
	container.OptimizerService.SetSettingsService(container.SettingsService)

	// Wire Black-Litterman Optimizer into OptimizerService
	// Optimizer can use Black-Litterman for view-based optimization
	container.OptimizerService.SetBlackLittermanOptimizer(container.BlackLittermanOptimizer)

	return nil
}
