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
	"github.com/aristath/arduino-trader/internal/modules/portfolio"
	"github.com/aristath/arduino-trader/internal/modules/satellites"
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

	// Initialize databases (Python uses multiple SQLite databases)
	configDB, err := database.New(cfg.DatabasePath) // config.db
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize config database")
	}
	defer configDB.Close()

	// state.db - positions, scores
	stateDB, err := database.New("../data/state.db")
	if err != nil {
		// Note: log.Fatal calls os.Exit, preventing deferred cleanup
		// In practice, OS cleans up file handles on exit, so risk is low
		//nolint:gocritic // exitAfterDefer: Accepted tradeoff for simpler code
		log.Fatal().Err(err).Msg("Failed to initialize state database")
	}
	defer stateDB.Close()

	// snapshots.db - portfolio snapshots
	snapshotsDB, err := database.New("../data/snapshots.db")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize snapshots database")
	}
	defer snapshotsDB.Close()

	// ledger.db - trades (append-only ledger)
	ledgerDB, err := database.New("../data/ledger.db")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize ledger database")
	}
	defer ledgerDB.Close()

	// dividends.db - dividend records with DRIP tracking
	dividendsDB, err := database.New("../data/dividends.db")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize dividends database")
	}
	defer dividendsDB.Close()

	// satellites.db - bucket management and satellite accounts
	satellitesDB, err := database.New("../data/satellites.db")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize satellites database")
	}
	defer satellitesDB.Close()

	// Run migrations
	if err := configDB.Migrate(); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	// Initialize scheduler
	sched := scheduler.New(log)
	sched.Start()
	defer sched.Stop()

	// Register background jobs
	if err := registerJobs(sched, configDB, stateDB, snapshotsDB, ledgerDB, dividendsDB, satellitesDB, cfg, log); err != nil {
		log.Fatal().Err(err).Msg("Failed to register jobs")
	}

	// Initialize HTTP server
	srv := server.New(server.Config{
		Port:         cfg.Port,
		Log:          log,
		ConfigDB:     configDB,
		StateDB:      stateDB,
		SnapshotsDB:  snapshotsDB,
		LedgerDB:     ledgerDB,
		DividendsDB:  dividendsDB,
		SatellitesDB: satellitesDB,
		Config:       cfg,
		DevMode:      cfg.DevMode,
	})

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

func registerJobs(sched *scheduler.Scheduler, configDB, stateDB, snapshotsDB, ledgerDB, dividendsDB, satellitesDB *database.DB, cfg *config.Config, log zerolog.Logger) error {
	// Initialize required repositories and services for jobs

	// Repositories
	positionRepo := portfolio.NewPositionRepository(stateDB.Conn(), configDB.Conn(), log)
	securityRepo := universe.NewSecurityRepository(configDB.Conn(), log)
	scoreRepo := universe.NewScoreRepository(stateDB.Conn(), log)
	dividendRepo := dividends.NewDividendRepository(dividendsDB.Conn(), log)
	bucketRepo := satellites.NewBucketRepository(satellitesDB.Conn(), log)
	balanceRepo := satellites.NewBalanceRepository(satellitesDB.Conn(), log)

	// Clients
	tradernetClient := tradernet.NewClient(cfg.TradernetServiceURL, log)
	yahooClient := yahoo.NewClient(log)

	// Currency exchange service
	currencyExchangeService := services.NewCurrencyExchangeService(tradernetClient, log)

	// Satellite services
	balanceService := satellites.NewBalanceService(balanceRepo, bucketRepo, log)
	bucketService := satellites.NewBucketService(bucketRepo, balanceRepo, currencyExchangeService, log)
	metaAllocator := satellites.NewMetaAllocator(bucketService, balanceService, balanceRepo, log)
	reconciliationService := satellites.NewReconciliationService(balanceRepo, bucketRepo, log)

	// Trading and portfolio services
	tradeRepo := portfolio.NewTradeRepository(ledgerDB.Conn(), log)
	tradingService := trading.NewTradingService(trading.NewTradeRepository(ledgerDB.Conn(), log), tradernetClient, log)

	portfolioRepo := portfolio.NewPortfolioRepository(snapshotsDB.Conn(), log)
	allocRepo := allocation.NewRepository(configDB.Conn(), log)
	turnoverTracker := portfolio.NewTurnoverTracker(ledgerDB.Conn(), snapshotsDB.Conn(), log)
	attributionCalc := portfolio.NewAttributionCalculator(tradeRepo, configDB.Conn(), cfg.HistoryPath, log)
	portfolioService := portfolio.NewPortfolioService(portfolioRepo, positionRepo, allocRepo, turnoverTracker, attributionCalc, configDB.Conn(), log)

	// Cash flows service
	cashFlowsService := cash_flows.NewCashFlowsService(log)

	// Universe service (simplified for jobs)
	universeService := universe.NewUniverseService(log)

	// Display manager (can be nil for now if not available)
	var displayManager *display.StateManager // TODO: Initialize if display is enabled

	// Market hours service
	marketHours := scheduler.NewMarketHoursService(log)

	// Register Job 1: Health Check (daily at 4:00 AM)
	healthCheck := scheduler.NewHealthCheckJob(scheduler.HealthCheckConfig{
		Log:         log,
		DataDir:     "../data",
		ConfigDB:    configDB,
		StateDB:     stateDB,
		SnapshotsDB: snapshotsDB,
		LedgerDB:    ledgerDB,
		DividendsDB: dividendsDB,
		HistoryPath: cfg.HistoryPath,
	})
	if err := sched.AddJob("0 0 4 * * *", healthCheck); err != nil {
		return fmt.Errorf("failed to register health_check job: %w", err)
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
		return fmt.Errorf("failed to register sync_cycle job: %w", err)
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
		return fmt.Errorf("failed to register dividend_reinvestment job: %w", err)
	}

	// Register Job 4: Satellite Maintenance (daily at 11:00 AM)
	satelliteMaintenance := scheduler.NewSatelliteMaintenanceJob(log, bucketService, positionRepo)
	if err := sched.AddJob("0 0 11 * * *", satelliteMaintenance); err != nil {
		return fmt.Errorf("failed to register satellite_maintenance job: %w", err)
	}

	// Register Job 5: Satellite Reconciliation (daily at 11:30 PM)
	satelliteReconciliation := scheduler.NewSatelliteReconciliationJob(log, tradernetClient, reconciliationService)
	if err := sched.AddJob("0 30 23 * * *", satelliteReconciliation); err != nil {
		return fmt.Errorf("failed to register satellite_reconciliation job: %w", err)
	}

	// Register Job 6: Satellite Evaluation (weekly on Sunday at 3:00 AM)
	satelliteEvaluation := scheduler.NewSatelliteEvaluationJob(log, metaAllocator)
	if err := sched.AddJob("0 0 3 * * 0", satelliteEvaluation); err != nil {
		return fmt.Errorf("failed to register satellite_evaluation job: %w", err)
	}

	log.Info().Int("jobs", 6).Msg("Background jobs registered successfully")

	return nil
}
