package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aristath/sentinel/internal/config"
	"github.com/aristath/sentinel/internal/deployment"
	"github.com/aristath/sentinel/internal/di"
	"github.com/aristath/sentinel/internal/modules/display"
	"github.com/aristath/sentinel/internal/server"
	"github.com/aristath/sentinel/pkg/logger"
)

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	// Load configuration first to get log level
	cfg, err := config.Load()
	if err != nil {
		// Use fallback logger if config fails
		fallbackLog := logger.New(logger.Config{
			Level:  "info",
			Pretty: true,
		})
		fallbackLog.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize logger with config level
	log := logger.New(logger.Config{
		Level:  cfg.LogLevel,
		Pretty: true,
	})

	log.Info().Msg("Starting Sentinel")

	// Display manager (state holder for LED display) - must be initialized before server.New()
	displayManager := display.NewStateManager(log)
	log.Info().Msg("Display manager initialized")

	// Initialize deployment manager BEFORE DI wiring so it can be passed to job registration
	var deploymentManager *deployment.Manager
	if cfg.Deployment != nil && cfg.Deployment.Enabled {
		deployConfig := cfg.Deployment.ToDeploymentConfig(cfg.GitHubToken)
		version := getEnv("VERSION", "dev")
		deploymentManager = deployment.NewManager(deployConfig, version, log)
		log.Info().Msg("Deployment manager initialized")
	}

	// Wire all dependencies using DI container
	// This replaces the massive registerJobs function and all manual wiring
	// Pass deployment manager so it can be registered as a job
	container, jobs, err := di.Wire(cfg, log, displayManager, deploymentManager)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to wire dependencies")
	}

	// Cleanup databases on exit
	defer container.UniverseDB.Close()
	defer container.ConfigDB.Close()
	defer container.LedgerDB.Close()
	defer container.PortfolioDB.Close()
	defer container.AgentsDB.Close()
	defer container.HistoryDB.Close()
	defer container.CacheDB.Close()

	// Update config from settings DB (credentials, etc.)
	// Note: SettingsRepo is now in container, but we need to update config before server creation
	// So we access it directly from container
	if err := cfg.UpdateFromSettings(container.SettingsRepo); err != nil {
		log.Warn().Err(err).Msg("Failed to update config from settings DB, using environment variables")
	}

	// Warn if credentials are loaded from .env (deprecated)
	if cfg.TradernetAPIKey != "" || cfg.TradernetAPISecret != "" {
		// Check if credentials came from env vars (not settings DB)
		apiKeyFromDB, _ := container.SettingsRepo.Get("tradernet_api_key")
		apiSecretFromDB, _ := container.SettingsRepo.Get("tradernet_api_secret")
		usingEnvVars := (apiKeyFromDB == nil || *apiKeyFromDB == "") && cfg.TradernetAPIKey != "" ||
			(apiSecretFromDB == nil || *apiSecretFromDB == "") && cfg.TradernetAPISecret != ""
		if usingEnvVars {
			log.Warn().Msg("Tradernet credentials loaded from .env file - this is deprecated. Please configure credentials via Settings UI (Credentials tab) or API. The .env file will no longer be required in future versions.")
		}
	}

	// Create deployment handlers if deployment is enabled
	var deploymentHandlers *server.DeploymentHandlers
	if deploymentManager != nil {
		deploymentHandlers = server.NewDeploymentHandlers(deploymentManager, log)
	}

	// Initialize HTTP server
	// Pass container to server so it can use all services
	srv := server.New(server.Config{
		Port:               cfg.Port,
		Log:                log,
		UniverseDB:         container.UniverseDB,
		ConfigDB:           container.ConfigDB,
		LedgerDB:           container.LedgerDB,
		PortfolioDB:        container.PortfolioDB,
		AgentsDB:           container.AgentsDB,
		HistoryDB:          container.HistoryDB,
		CacheDB:            container.CacheDB,
		Config:             cfg,
		DevMode:            cfg.DevMode,
		DisplayManager:     displayManager,
		DeploymentHandlers: deploymentHandlers,
		Container:          container, // Pass container for handlers to use
	})

	// Wire up jobs for manual triggering via API
	srv.SetJobs(
		jobs.HealthCheck,
		jobs.SyncCycle,
		jobs.DividendReinvest,
		jobs.PlannerBatch,
		jobs.EventBasedTrading,
		jobs.TagUpdate,
		// Individual sync jobs
		jobs.SyncTrades,
		jobs.SyncCashFlows,
		jobs.SyncPortfolio,
		jobs.SyncPrices,
		jobs.CheckNegativeBalances,
		jobs.UpdateDisplayTicker,
		// Individual planning jobs
		jobs.GeneratePortfolioHash,
		jobs.GetOptimizerWeights,
		jobs.BuildOpportunityContext,
		jobs.CreateTradePlan,
		jobs.StoreRecommendations,
		// Individual dividend jobs
		jobs.GetUnreinvestedDividends,
		jobs.GroupDividendsBySymbol,
		jobs.CheckDividendYields,
		jobs.CreateDividendRecommendations,
		jobs.SetPendingBonuses,
		jobs.ExecuteDividendTrades,
		// Individual health check jobs
		jobs.CheckCoreDatabases,
		jobs.CheckHistoryDatabases,
		jobs.CheckWALCheckpoints,
	)

	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	log.Info().Int("port", cfg.Port).Msg("Server started successfully")

	// Start LED status monitors
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start service heartbeat monitor (LED3)
	serviceMonitor := display.NewServiceMonitor("sentinel", displayManager, log)
	go serviceMonitor.MonitorService(ctx)
	log.Info().Msg("Service heartbeat monitor started (LED3)")

	// Start planner action monitor (LED4)
	// Use recommendation repo from container
	plannerMonitor := display.NewPlannerMonitor(container.RecommendationRepo, displayManager, log)
	go plannerMonitor.MonitorPlannerActions(ctx)
	log.Info().Msg("Planner action monitor started (LED4)")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Cancel context to stop monitors
	cancel()
	log.Info().Msg("Stopping LED monitors...")

	log.Info().Msg("Shutting down server...")

	// Stop queue system components
	if container.TimeScheduler != nil {
		container.TimeScheduler.Stop()
		log.Info().Msg("Time scheduler stopped")
	}
	if container.WorkerPool != nil {
		container.WorkerPool.Stop()
		log.Info().Msg("Worker pool stopped")
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server stopped")
}

// Note: registerJobs function has been moved to internal/di/jobs.go
// JobInstances type has been moved to internal/di/types.go
// All dependency wiring is now handled by di.Wire()
// The entire registerJobs function (842 lines) has been extracted to:
//   - internal/di/databases.go (database initialization)
//   - internal/di/repositories.go (repository creation)
//   - internal/di/services.go (service creation - single source of truth)
//   - internal/di/jobs.go (job registration)
//   - internal/di/wire.go (main orchestration)
