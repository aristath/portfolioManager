/**
 * Package di provides dependency injection for service implementations.
 *
 * This package initializes all business logic services in the correct dependency order.
 * Services are the SINGLE SOURCE OF TRUTH for all service creation - all services
 * must be created here to ensure proper dependency injection and initialization order.
 *
 * Initialization Sequence:
 * 1. Clients (broker, market status WebSocket)
 * 2. Basic Services (currency exchange, market hours, event system)
 * 3. Cash Manager (cash-as-balances architecture)
 * 4. Trading Services (trade safety, trading, trade execution)
 * 5. Universe Services (historical sync, symbol resolver, security setup)
 * 6. Portfolio Service (portfolio management)
 * 7. Cash Flows Services (dividends, deposits)
 * 8. Remaining Universe Services (sync, universe, tag assigner)
 * 9. Planning Services (opportunities, sequences, evaluation, planner)
 * 10. Optimization Services (risk builder, constraints, returns, Kelly, CVaR, BL, optimizer)
 * 11. Calculation Cache and Analytics
 * 12. Rebalancing Services
 * 13. Ticker and Display Services
 * 14. Adaptive Market Services (market regime detection)
 * 15. Reliability Services (backup, health checks, R2)
 * 16. Concentration Alert Service
 * 17. Quantum Calculator
 * 18. Callbacks (for jobs)
 */
package di

import (
	"fmt"

	"github.com/aristath/sentinel/internal/config"
	"github.com/aristath/sentinel/internal/modules/display"
	"github.com/rs/zerolog"
)

/**
 * InitializeServices creates all services and stores them in the container.
 *
 * This is the SINGLE SOURCE OF TRUTH for all service creation.
 * Services are created in dependency order to ensure all dependencies exist
 * before they are needed.
 *
 * The initialization is organized into logical steps, each handled by a separate
 * initialization function for better maintainability and debuggability.
 *
 * @param container - Container to store service instances (must not be nil)
 * @param cfg - Application configuration (with settings loaded from database)
 * @param displayManager - LED display state manager (can be nil in tests)
 * @param log - Structured logger instance
 * @returns error - Error if service initialization fails
 */
func InitializeServices(container *Container, cfg *config.Config, displayManager *display.StateManager, log zerolog.Logger) error {
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}

	// Step 1: Initialize Clients
	if err := initializeClients(container, cfg, displayManager, log); err != nil {
		return fmt.Errorf("failed to initialize clients: %w", err)
	}

	// Step 2: Initialize Basic Services
	if err := initializeBasicServices(container, log); err != nil {
		return fmt.Errorf("failed to initialize basic services: %w", err)
	}

	// Step 3: Initialize Cash Manager
	cashManager := initializeCashManager(container, log)

	// Step 4: Initialize Trading Services
	if err := initializeTradingServices(container, cashManager, log); err != nil {
		return fmt.Errorf("failed to initialize trading services: %w", err)
	}

	// Step 5: Initialize Universe Services
	if err := initializeUniverseServices(container, log); err != nil {
		return fmt.Errorf("failed to initialize universe services: %w", err)
	}

	// Step 6: Initialize Portfolio Service
	if err := initializePortfolioService(container, cashManager, log); err != nil {
		return fmt.Errorf("failed to initialize portfolio service: %w", err)
	}

	// Step 7: Initialize Cash Flows Services
	if err := initializeCashFlowsServices(container, cashManager, displayManager, log); err != nil {
		return fmt.Errorf("failed to initialize cash flows services: %w", err)
	}

	// Step 8: Initialize Remaining Universe Services
	if err := initializeRemainingUniverseServices(container, log); err != nil {
		return fmt.Errorf("failed to initialize remaining universe services: %w", err)
	}

	// Step 9: Initialize Planning Services
	if err := initializePlanningServices(container, log); err != nil {
		return fmt.Errorf("failed to initialize planning services: %w", err)
	}

	// Step 10: Initialize Optimization Services
	if err := initializeOptimizationServices(container, log); err != nil {
		return fmt.Errorf("failed to initialize optimization services: %w", err)
	}

	// Step 11: Initialize Calculation Cache and Analytics
	if err := initializeCacheAndAnalytics(container, log); err != nil {
		return fmt.Errorf("failed to initialize cache and analytics: %w", err)
	}

	// Step 12: Initialize Rebalancing Services
	if err := initializeRebalancingServices(container, cashManager, log); err != nil {
		return fmt.Errorf("failed to initialize rebalancing services: %w", err)
	}

	// Step 13: Initialize Ticker and Display Services
	if err := initializeDisplayServices(container, cashManager, displayManager, log); err != nil {
		return fmt.Errorf("failed to initialize display services: %w", err)
	}

	// Step 14: Initialize Adaptive Market Services
	if err := initializeAdaptiveMarketServices(container, log); err != nil {
		return fmt.Errorf("failed to initialize adaptive market services: %w", err)
	}

	// Step 15: Initialize Reliability Services
	if err := initializeReliabilityServices(container, cfg, log); err != nil {
		return fmt.Errorf("failed to initialize reliability services: %w", err)
	}

	// Step 16: Initialize Concentration Alert Service
	if err := initializeConcentrationAlertService(container, log); err != nil {
		return fmt.Errorf("failed to initialize concentration alert service: %w", err)
	}

	// Step 17: Initialize Quantum Calculator
	initializeQuantumCalculator(container)

	// Step 18: Initialize Callbacks
	initializeCallbacks(container, displayManager, log)

	log.Info().Msg("All services initialized")

	return nil
}
