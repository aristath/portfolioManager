package di

import (
	"fmt"

	"github.com/aristath/portfolioManager/internal/config"
	"github.com/aristath/portfolioManager/internal/modules/display"
	"github.com/aristath/portfolioManager/internal/scheduler"
	"github.com/rs/zerolog"
)

// Wire initializes all dependencies and returns a fully configured container
// This is the main entry point for dependency injection
// Order of operations:
// 1. Initialize databases
// 2. Initialize repositories
// 3. Initialize services
// 4. Register jobs
func Wire(cfg *config.Config, log zerolog.Logger, sched *scheduler.Scheduler, displayManager *display.StateManager) (*Container, *JobInstances, error) {
	// Step 1: Initialize databases
	container, err := InitializeDatabases(cfg, log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize databases: %w", err)
	}

	// Step 2: Initialize repositories
	if err := InitializeRepositories(container, log); err != nil {
		// Cleanup databases on error
		container.UniverseDB.Close()
		container.ConfigDB.Close()
		container.LedgerDB.Close()
		container.PortfolioDB.Close()
		container.AgentsDB.Close()
		container.HistoryDB.Close()
		container.CacheDB.Close()
		return nil, nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	// Step 3: Initialize services
	if err := InitializeServices(container, cfg, displayManager, log); err != nil {
		// Cleanup on error
		container.UniverseDB.Close()
		container.ConfigDB.Close()
		container.LedgerDB.Close()
		container.PortfolioDB.Close()
		container.AgentsDB.Close()
		container.HistoryDB.Close()
		container.CacheDB.Close()
		return nil, nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Step 4: Register jobs
	jobs, err := RegisterJobs(container, cfg, sched, displayManager, log)
	if err != nil {
		// Cleanup on error
		container.UniverseDB.Close()
		container.ConfigDB.Close()
		container.LedgerDB.Close()
		container.PortfolioDB.Close()
		container.AgentsDB.Close()
		container.HistoryDB.Close()
		container.CacheDB.Close()
		return nil, nil, fmt.Errorf("failed to register jobs: %w", err)
	}

	log.Info().Msg("Dependency injection wiring completed successfully")

	return container, jobs, nil
}
