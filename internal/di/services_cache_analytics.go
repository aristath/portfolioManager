/**
 * Package di provides dependency injection for calculation cache and analytics initialization.
 *
 * Step 11: Initialize Calculation Cache and Analytics
 * Calculation cache stores expensive computation results (risk models, optimizer results).
 */
package di

import (
	"github.com/aristath/sentinel/internal/modules/analytics"
	"github.com/aristath/sentinel/internal/modules/calculations"
	"github.com/rs/zerolog"
)

// initializeCacheAndAnalytics initializes calculation cache and analytics services.
func initializeCacheAndAnalytics(container *Container, log zerolog.Logger) error {
	// Calculation cache (for technical indicators and optimizer results)
	// Caches expensive computation results (risk models, HRP allocations, MV allocations)
	// Reduces computation time for repeated calculations
	container.CalculationCache = calculations.NewCache(container.CalculationsDB.Conn())

	// Wire cache into RiskBuilder for optimizer caching
	// Risk models (covariance matrices) are expensive to compute - cache them
	container.RiskBuilder.SetCache(container.CalculationCache)

	// Wire cache into OptimizerService for HRP and MV caching
	// Optimizer results (HRP allocations, MV allocations) are cached
	container.OptimizerService.SetCache(container.CalculationCache)

	// Factor Exposure Tracker
	// Tracks portfolio exposure to various risk factors (sector, geography, etc.)
	container.FactorExposureTracker = analytics.NewFactorExposureTracker(log)

	return nil
}
