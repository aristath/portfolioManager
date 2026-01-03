package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aristath/arduino-trader/internal/clients/tradernet"
	"github.com/aristath/arduino-trader/internal/clients/yahoo"
	"github.com/aristath/arduino-trader/internal/config"
	"github.com/aristath/arduino-trader/internal/database"
	"github.com/aristath/arduino-trader/internal/modules/allocation"
	"github.com/aristath/arduino-trader/internal/modules/cash_flows"
	"github.com/aristath/arduino-trader/internal/modules/display"
	"github.com/aristath/arduino-trader/internal/modules/dividends"
	"github.com/aristath/arduino-trader/internal/modules/opportunities"
	"github.com/aristath/arduino-trader/internal/modules/planning"
	planningconfig "github.com/aristath/arduino-trader/internal/modules/planning/config"
	planningevaluation "github.com/aristath/arduino-trader/internal/modules/planning/evaluation"
	planningplanner "github.com/aristath/arduino-trader/internal/modules/planning/planner"
	planningrepo "github.com/aristath/arduino-trader/internal/modules/planning/repository"
	"github.com/aristath/arduino-trader/internal/modules/portfolio"
	"github.com/aristath/arduino-trader/internal/modules/satellites"
	"github.com/aristath/arduino-trader/internal/modules/sequences"
	"github.com/aristath/arduino-trader/internal/modules/trading"
	"github.com/aristath/arduino-trader/internal/modules/universe"
	"github.com/aristath/arduino-trader/internal/scheduler"
	"github.com/aristath/arduino-trader/internal/server"
	"github.com/aristath/arduino-trader/internal/services"
	"github.com/aristath/arduino-trader/pkg/logger"
	"github.com/rs/zerolog"
)

func main() {
	// Initialize logger
	log := logger.New(logger.Config{
		Level:  "info",
		Pretty: true,
	})

	log.Info().Msg("Starting Arduino Trader")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize databases - NEW 8-database architecture
	// Architecture: universe, config, ledger, portfolio, satellites, agents, history, cache

	// 1. universe.db - Investment universe (securities, groups)
	universeDB, err := database.New(database.Config{
		Path:    "../data/universe.db",
		Profile: database.ProfileStandard,
		Name:    "universe",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize universe database")
	}
	defer universeDB.Close()

	// 2. config.db - Application configuration (REDUCED: settings, allocation targets)
	configDB, err := database.New(database.Config{
		Path:    cfg.DatabasePath,
		Profile: database.ProfileStandard,
		Name:    "config",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize config database")
	}
	defer configDB.Close()

	// 3. ledger.db - Immutable financial audit trail (EXPANDED: trades, cash flows, dividends)
	ledgerDB, err := database.New(database.Config{
		Path:    "../data/ledger.db",
		Profile: database.ProfileLedger, // Maximum safety for immutable audit trail
		Name:    "ledger",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize ledger database")
	}
	defer ledgerDB.Close()

	// 4. portfolio.db - Current portfolio state (positions, scores, metrics, snapshots)
	portfolioDB, err := database.New(database.Config{
		Path:    "../data/portfolio.db",
		Profile: database.ProfileStandard,
		Name:    "portfolio",
	})
	if err != nil {
		//nolint:gocritic // exitAfterDefer: Accepted tradeoff for simpler code
		log.Fatal().Err(err).Msg("Failed to initialize portfolio database")
	}
	defer portfolioDB.Close()

	// 5. satellites.db - Multi-bucket portfolio system (UPDATED: added agent_id)
	satellitesDB, err := database.New(database.Config{
		Path:    "../data/satellites.db",
		Profile: database.ProfileStandard,
		Name:    "satellites",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize satellites database")
	}
	defer satellitesDB.Close()

	// 6. agents.db - Strategy management (TOML configs, sequences, evaluations)
	agentsDB, err := database.New(database.Config{
		Path:    "../data/agents.db",
		Profile: database.ProfileStandard,
		Name:    "agents",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize agents database")
	}
	defer agentsDB.Close()

	// 7. history.db - Historical time-series data (prices, rates, cleanup tracking)
	historyDB, err := database.New(database.Config{
		Path:    "../data/history.db",
		Profile: database.ProfileStandard,
		Name:    "history",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize history database")
	}
	defer historyDB.Close()

	// 8. cache.db - Ephemeral operational data (recommendations, cache)
	cacheDB, err := database.New(database.Config{
		Path:    "../data/cache.db",
		Profile: database.ProfileCache, // Maximum speed for ephemeral data
		Name:    "cache",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize cache database")
	}
	defer cacheDB.Close()

	// Run migrations on all databases
	for _, db := range []*database.DB{universeDB, configDB, ledgerDB, portfolioDB, satellitesDB, agentsDB, historyDB, cacheDB} {
		if err := db.Migrate(); err != nil {
			log.Fatal().Err(err).Str("database", db.Name()).Msg("Failed to run migrations")
		}
	}

	// Initialize scheduler
	sched := scheduler.New(log)
	sched.Start()
	defer sched.Stop()

	// Register background jobs
	jobs, err := registerJobs(sched, universeDB, configDB, ledgerDB, portfolioDB, satellitesDB, agentsDB, historyDB, cacheDB, cfg, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to register jobs")
	}

	// Initialize HTTP server
	srv := server.New(server.Config{
		Port:         cfg.Port,
		Log:          log,
		UniverseDB:   universeDB,
		ConfigDB:     configDB,
		LedgerDB:     ledgerDB,
		PortfolioDB:  portfolioDB,
		SatellitesDB: satellitesDB,
		AgentsDB:     agentsDB,
		HistoryDB:    historyDB,
		CacheDB:      cacheDB,
		Config:       cfg,
		DevMode:      cfg.DevMode,
		Scheduler:    sched,
	})

	// Wire up jobs for manual triggering via API
	srv.SetJobs(
		jobs.HealthCheck,
		jobs.SyncCycle,
		jobs.DividendReinvest,
		jobs.SatelliteMaintenance,
		jobs.SatelliteReconciliation,
		jobs.SatelliteEvaluation,
		jobs.PlannerBatch,
		jobs.EventBasedTrading,
	)

	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	log.Info().Int("port", cfg.Port).Msg("Server started successfully")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server stopped")
}

// JobInstances holds references to all registered jobs for manual triggering
type JobInstances struct {
	HealthCheck             scheduler.Job
	SyncCycle               scheduler.Job
	DividendReinvest        scheduler.Job
	SatelliteMaintenance    scheduler.Job
	SatelliteReconciliation scheduler.Job
	SatelliteEvaluation     scheduler.Job
	PlannerBatch            scheduler.Job
	EventBasedTrading       scheduler.Job
}

func registerJobs(sched *scheduler.Scheduler, universeDB, configDB, ledgerDB, portfolioDB, satellitesDB, agentsDB, historyDB, cacheDB *database.DB, cfg *config.Config, log zerolog.Logger) (*JobInstances, error) {
	// Initialize required repositories and services for jobs

	// Repositories - NEW 8-database architecture
	positionRepo := portfolio.NewPositionRepository(portfolioDB.Conn(), universeDB.Conn(), log)
	securityRepo := universe.NewSecurityRepository(universeDB.Conn(), log)
	scoreRepo := universe.NewScoreRepository(portfolioDB.Conn(), log)
	dividendRepo := dividends.NewDividendRepository(ledgerDB.Conn(), log)
	bucketRepo := satellites.NewBucketRepository(satellitesDB.Conn(), log)
	balanceRepo := satellites.NewBalanceRepository(satellitesDB.Conn(), log)

	// Clients
	tradernetClient := tradernet.NewClient(cfg.TradernetServiceURL, log)
	yahooClient := yahoo.NewClient(log)

	// Currency exchange service
	currencyExchangeService := services.NewCurrencyExchangeService(tradernetClient, log)

	// Market hours service
	marketHours := scheduler.NewMarketHoursService(log)

	// Satellite services
	balanceService := satellites.NewBalanceService(balanceRepo, bucketRepo, log)
	bucketService := satellites.NewBucketService(bucketRepo, balanceRepo, currencyExchangeService, log)
	metaAllocator := satellites.NewMetaAllocator(bucketService, balanceService, balanceRepo, log)
	reconciliationService := satellites.NewReconciliationService(balanceRepo, bucketRepo, log)

	// Trading and portfolio services
	tradeRepo := portfolio.NewTradeRepository(ledgerDB.Conn(), log)

	// Trade safety service with all validation layers
	tradeSafetyService := trading.NewTradeSafetyService(
		trading.NewTradeRepository(ledgerDB.Conn(), log),
		positionRepo,
		securityRepo,
		nil, // settingsService - will use defaults
		marketHours,
		log,
	)

	tradingService := trading.NewTradingService(
		trading.NewTradeRepository(ledgerDB.Conn(), log),
		tradernetClient,
		tradeSafetyService,
		log,
	)

	portfolioRepo := portfolio.NewPortfolioRepository(portfolioDB.Conn(), log)
	allocRepo := allocation.NewRepository(configDB.Conn(), log)
	turnoverTracker := portfolio.NewTurnoverTracker(ledgerDB.Conn(), portfolioDB.Conn(), log)
	attributionCalc := portfolio.NewAttributionCalculator(tradeRepo, historyDB.Conn(), cfg.HistoryPath, log)
	portfolioService := portfolio.NewPortfolioService(portfolioRepo, positionRepo, allocRepo, turnoverTracker, attributionCalc, universeDB.Conn(), log)

	// Cash flows service
	cashFlowsService := cash_flows.NewCashFlowsService(log)

	// Universe service (simplified for jobs)
	universeService := universe.NewUniverseService(log)

	// Display manager (can be nil for now if not available)
	var displayManager *display.StateManager // TODO: Initialize if display is enabled

	// Register Job 1: Health Check (daily at 4:00 AM)
	healthCheck := scheduler.NewHealthCheckJob(scheduler.HealthCheckConfig{
		Log:          log,
		DataDir:      "../data",
		UniverseDB:   universeDB,
		ConfigDB:     configDB,
		LedgerDB:     ledgerDB,
		PortfolioDB:  portfolioDB,
		SatellitesDB: satellitesDB,
		AgentsDB:     agentsDB,
		HistoryDB:    historyDB,
		CacheDB:      cacheDB,
		HistoryPath:  cfg.HistoryPath,
	})
	if err := sched.AddJob("0 0 4 * * *", healthCheck); err != nil {
		return nil, fmt.Errorf("failed to register health_check job: %w", err)
	}

	// Register Job 2: Sync Cycle (every 5 minutes)
	syncCycle := scheduler.NewSyncCycleJob(scheduler.SyncCycleConfig{
		Log:                 log,
		PortfolioService:    portfolioService,
		CashFlowsService:    cashFlowsService,
		TradingService:      tradingService,
		UniverseService:     universeService,
		DisplayManager:      displayManager,
		MarketHours:         marketHours,
		EmergencyRebalance:  nil, // TODO: Wire up emergency rebalance callback
		UpdateDisplayTicker: nil, // TODO: Wire up display ticker callback
	})
	if err := sched.AddJob("0 */5 * * * *", syncCycle); err != nil {
		return nil, fmt.Errorf("failed to register sync_cycle job: %w", err)
	}

	// Register Job 3: Dividend Reinvestment (daily at 10:00 AM)
	dividendReinvest := scheduler.NewDividendReinvestmentJob(scheduler.DividendReinvestmentConfig{
		Log:              log,
		DividendRepo:     dividendRepo,
		SecurityRepo:     securityRepo,
		ScoreRepo:        scoreRepo,
		PortfolioService: portfolioService,
		TradingService:   tradingService,
		TradernetClient:  tradernetClient,
		YahooClient:      yahooClient,
	})
	if err := sched.AddJob("0 0 10 * * *", dividendReinvest); err != nil {
		return nil, fmt.Errorf("failed to register dividend_reinvestment job: %w", err)
	}

	// Register Job 4: Satellite Maintenance (daily at 11:00 AM)
	satelliteMaintenance := scheduler.NewSatelliteMaintenanceJob(log, bucketService, positionRepo)
	if err := sched.AddJob("0 0 11 * * *", satelliteMaintenance); err != nil {
		return nil, fmt.Errorf("failed to register satellite_maintenance job: %w", err)
	}

	// Register Job 5: Satellite Reconciliation (daily at 11:30 PM)
	satelliteReconciliation := scheduler.NewSatelliteReconciliationJob(log, tradernetClient, reconciliationService)
	if err := sched.AddJob("0 30 23 * * *", satelliteReconciliation); err != nil {
		return nil, fmt.Errorf("failed to register satellite_reconciliation job: %w", err)
	}

	// Register Job 6: Satellite Evaluation (weekly on Sunday at 3:00 AM)
	satelliteEvaluation := scheduler.NewSatelliteEvaluationJob(log, metaAllocator)
	if err := sched.AddJob("0 0 3 * * 0", satelliteEvaluation); err != nil {
		return nil, fmt.Errorf("failed to register satellite_evaluation job: %w", err)
	}

	// Planning module repositories and services
	recommendationRepo := planning.NewRecommendationRepository(cacheDB.Conn(), log)
	configLoader := planningconfig.NewLoader(log)
	plannerConfigRepo := planningrepo.NewConfigRepository(agentsDB, configLoader, log)
	opportunitiesService := opportunities.NewService(log)
	sequencesService := sequences.NewService(log)
	evaluationService := planningevaluation.NewService(4, log) // 4 workers
	plannerService := planningplanner.NewPlanner(opportunitiesService, sequencesService, evaluationService, log)

	// Register Job 7: Planner Batch (every 15 minutes)
	plannerBatch := scheduler.NewPlannerBatchJob(scheduler.PlannerBatchConfig{
		Log:                    log,
		PositionRepo:           positionRepo,
		SecurityRepo:           securityRepo,
		AllocRepo:              allocRepo,
		TradernetClient:        tradernetClient,
		OpportunitiesService:   opportunitiesService,
		SequencesService:       sequencesService,
		EvaluationService:      evaluationService,
		PlannerService:         plannerService,
		ConfigRepo:             plannerConfigRepo,
		RecommendationRepo:     recommendationRepo,
		MinPlanningIntervalMin: 15, // Minimum 15 minutes between planning cycles
	})
	if err := sched.AddJob("0 */15 * * * *", plannerBatch); err != nil {
		return nil, fmt.Errorf("failed to register planner_batch job: %w", err)
	}

	// Register Job 8: Event-Based Trading (every 5 minutes)
	eventBasedTrading := scheduler.NewEventBasedTradingJob(scheduler.EventBasedTradingConfig{
		Log:                     log,
		RecommendationRepo:      recommendationRepo,
		TradingService:          tradingService,
		MinExecutionIntervalMin: 30, // Minimum 30 minutes between trade executions
	})
	if err := sched.AddJob("0 */5 * * * *", eventBasedTrading); err != nil {
		return nil, fmt.Errorf("failed to register event_based_trading job: %w", err)
	}

	log.Info().Int("jobs", 8).Msg("Background jobs registered successfully")

	return &JobInstances{
		HealthCheck:             healthCheck,
		SyncCycle:               syncCycle,
		DividendReinvest:        dividendReinvest,
		SatelliteMaintenance:    satelliteMaintenance,
		SatelliteReconciliation: satelliteReconciliation,
		SatelliteEvaluation:     satelliteEvaluation,
		PlannerBatch:            plannerBatch,
		EventBasedTrading:       eventBasedTrading,
	}, nil
}
