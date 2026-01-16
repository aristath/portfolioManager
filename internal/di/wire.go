/**
 * Package di provides dependency injection wiring and initialization.
 *
 * This package implements a clean architecture dependency injection container
 * that wires all application components in the correct dependency order.
 *
 * Architecture:
 * - Databases are initialized first (8-database architecture)
 * - Repositories are created with database connections
 * - Settings are loaded from database into config
 * - Services are created with repository dependencies
 * - Work processor is registered with all job types
 * - All dependencies are injected via constructor injection
 *
 * The container follows clean architecture principles:
 * - Domain layer is pure (no infrastructure dependencies)
 * - Dependency flows inward (handlers → services → repositories → domain)
 * - Repository pattern for all data access
 * - Constructor injection only
 */
package di

import (
	"fmt"
	"os"

	"github.com/aristath/sentinel/internal/config"
	"github.com/aristath/sentinel/internal/deployment"
	"github.com/aristath/sentinel/internal/modules/display"
	"github.com/rs/zerolog"
)

/**
 * Wire initializes all dependencies and returns a fully configured container.
 *
 * This is the main entry point for dependency injection. It orchestrates
 * the initialization of all application components in the correct order:
 *
 * 1. Initialize databases (8-database architecture)
 * 2. Initialize repositories (with database connections)
 * 3. Load settings from database into config (credentials, etc.)
 * 4. Create deployment manager (with settings-loaded config)
 * 5. Initialize services (with settings-loaded config and all dependencies)
 * 6. Initialize work processor (with deployment manager available)
 *
 * The function ensures proper cleanup on error by closing all databases
 * that were successfully initialized before the error occurred.
 *
 * @param cfg - Application configuration (from environment variables)
 * @param log - Structured logger instance
 * @param displayManager - LED display state manager (can be nil in tests)
 * @returns *Container - Fully configured dependency injection container
 * @returns error - Error if initialization fails at any step
 */
func Wire(cfg *config.Config, log zerolog.Logger, displayManager *display.StateManager) (*Container, error) {
	// Step 1: Initialize databases
	// Creates all 8 database connections and applies schemas
	container, err := InitializeDatabases(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize databases: %w", err)
	}

	// Step 2: Initialize repositories
	// Creates all repository instances with database connections
	if err := InitializeRepositories(container, log); err != nil {
		// Cleanup databases on error to prevent resource leaks
		container.UniverseDB.Close()
		container.ConfigDB.Close()
		container.LedgerDB.Close()
		container.PortfolioDB.Close()
		container.HistoryDB.Close()
		container.CacheDB.Close()
		container.ClientDataDB.Close()
		container.CalculationsDB.Close()
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	// Step 3: Load settings from database into config
	// This must happen BEFORE creating services and deployment manager
	// so they have access to credentials and configuration from the database.
	// Settings in the database take precedence over environment variables.
	if err := cfg.UpdateFromSettings(container.SettingsRepo); err != nil {
		log.Warn().Err(err).Msg("Failed to load settings from database, using environment variables")
	}

	// Step 4: Create deployment manager with settings-loaded config
	// This must happen BEFORE initializing work processor so the deployment
	// work type has access to the manager.
	// Deployment manager is optional - only created if deployment is enabled.
	if cfg.Deployment != nil && cfg.Deployment.Enabled {
		version := os.Getenv("VERSION")
		if version == "" {
			version = "dev"
		}
		deployConfig := cfg.Deployment.ToDeploymentConfig(cfg.GitHubToken)
		container.DeploymentManager = deployment.NewManager(deployConfig, version, log)
		log.Info().Msg("Deployment manager initialized")
	}

	// Step 5: Initialize services with settings-loaded config
	// Services can now use credentials and settings from the database.
	// This is the largest initialization step, creating all business logic services
	// in dependency order (see services.go for detailed initialization sequence).
	if err := InitializeServices(container, cfg, displayManager, log); err != nil {
		// Cleanup on error - close all databases
		container.UniverseDB.Close()
		container.ConfigDB.Close()
		container.LedgerDB.Close()
		container.PortfolioDB.Close()
		container.HistoryDB.Close()
		container.CacheDB.Close()
		container.ClientDataDB.Close()
		container.CalculationsDB.Close()
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Step 6: Initialize work processor
	// Work types can now access deployment manager via container.
	// The work processor manages all background jobs with event-driven execution,
	// dependency resolution, and market-aware scheduling.
	workComponents, err := InitializeWork(container, log)
	if err != nil {
		// Cleanup on error - close all databases
		container.UniverseDB.Close()
		container.ConfigDB.Close()
		container.LedgerDB.Close()
		container.PortfolioDB.Close()
		container.HistoryDB.Close()
		container.CacheDB.Close()
		container.ClientDataDB.Close()
		container.CalculationsDB.Close()
		return nil, fmt.Errorf("failed to initialize work processor: %w", err)
	}
	container.WorkComponents = workComponents

	log.Info().Msg("Dependency injection wiring completed successfully")

	return container, nil
}
