// Package di provides dependency injection wiring and initialization.
package di

import (
	"fmt"
	"os"

	"github.com/aristath/sentinel/internal/config"
	"github.com/aristath/sentinel/internal/deployment"
	"github.com/aristath/sentinel/internal/modules/display"
	"github.com/rs/zerolog"
)

// Wire initializes all dependencies and returns a fully configured container
// This is the main entry point for dependency injection
// Order of operations:
// 1. Initialize databases
// 2. Initialize repositories
// 3. Load settings from database into config
// 4. Create deployment manager (with settings-loaded config)
// 5. Initialize services (with settings-loaded config)
// 6. Initialize work processor (with deployment manager available)
func Wire(cfg *config.Config, log zerolog.Logger, displayManager *display.StateManager) (*Container, error) {
	// Step 1: Initialize databases
	container, err := InitializeDatabases(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize databases: %w", err)
	}

	// Step 2: Initialize repositories
	if err := InitializeRepositories(container, log); err != nil {
		// Cleanup databases on error
		container.UniverseDB.Close()
		container.ConfigDB.Close()
		container.LedgerDB.Close()
		container.PortfolioDB.Close()
		container.HistoryDB.Close()
		container.CacheDB.Close()
		container.ClientDataDB.Close()
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	// Step 3: Load settings from database into config
	// This must happen BEFORE creating services and deployment manager
	// so they have access to credentials and configuration from the database
	if err := cfg.UpdateFromSettings(container.SettingsRepo); err != nil {
		log.Warn().Err(err).Msg("Failed to load settings from database, using environment variables")
	}

	// Step 4: Create deployment manager with settings-loaded config
	// This must happen BEFORE initializing work processor so the deployment
	// work type has access to the manager
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
	// Services can now use credentials and settings from the database
	if err := InitializeServices(container, cfg, displayManager, log); err != nil {
		// Cleanup on error
		container.UniverseDB.Close()
		container.ConfigDB.Close()
		container.LedgerDB.Close()
		container.PortfolioDB.Close()
		container.HistoryDB.Close()
		container.CacheDB.Close()
		container.ClientDataDB.Close()
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Step 6: Initialize work processor
	// Work types can now access deployment manager via container
	workComponents, err := InitializeWork(container, log)
	if err != nil {
		// Cleanup on error
		container.UniverseDB.Close()
		container.ConfigDB.Close()
		container.LedgerDB.Close()
		container.PortfolioDB.Close()
		container.HistoryDB.Close()
		container.CacheDB.Close()
		container.ClientDataDB.Close()
		return nil, fmt.Errorf("failed to initialize work processor: %w", err)
	}
	container.WorkComponents = workComponents

	log.Info().Msg("Dependency injection wiring completed successfully")

	return container, nil
}
