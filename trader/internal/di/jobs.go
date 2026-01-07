// Package di provides dependency injection for scheduler jobs.
package di

import (
	"fmt"

	"github.com/aristath/portfolioManager/internal/config"
	"github.com/aristath/portfolioManager/internal/database"
	"github.com/aristath/portfolioManager/internal/modules/cleanup"
	"github.com/aristath/portfolioManager/internal/modules/display"
	"github.com/aristath/portfolioManager/internal/modules/symbolic_regression"
	"github.com/aristath/portfolioManager/internal/reliability"
	"github.com/aristath/portfolioManager/internal/scheduler"
	"github.com/rs/zerolog"
)

// RegisterJobs registers all background jobs with the scheduler
// Returns JobInstances for manual triggering via API
func RegisterJobs(container *Container, cfg *config.Config, sched *scheduler.Scheduler, displayManager *display.StateManager, log zerolog.Logger) (*JobInstances, error) {
	if container == nil {
		return nil, fmt.Errorf("container cannot be nil")
	}

	instances := &JobInstances{}

	// ==========================================
	// Job 1: Health Check (daily at 4:00 AM)
	// ==========================================
	healthCheck := scheduler.NewHealthCheckJob(scheduler.HealthCheckConfig{
		Log:         log,
		DataDir:     cfg.DataDir,
		UniverseDB:  container.UniverseDB,
		ConfigDB:    container.ConfigDB,
		LedgerDB:    container.LedgerDB,
		PortfolioDB: container.PortfolioDB,
		AgentsDB:    container.AgentsDB,
		HistoryDB:   container.HistoryDB,
		CacheDB:     container.CacheDB,
	})
	if err := sched.AddJob("0 0 4 * * *", healthCheck); err != nil {
		return nil, fmt.Errorf("failed to register health_check job: %w", err)
	}
	instances.HealthCheck = healthCheck

	// ==========================================
	// Job 2: Sync Cycle (every 5 minutes)
	// ==========================================
	// Create balance adapter to enable negative balance checks
	balanceAdapter := scheduler.NewBalanceAdapter(container.CashManager, log)
	syncCycle := scheduler.NewSyncCycleJob(scheduler.SyncCycleConfig{
		Log:                 log,
		PortfolioService:    container.PortfolioService,
		CashFlowsService:    container.CashFlowsService,
		TradingService:      container.TradingService,
		UniverseService:     container.UniverseService,
		BalanceService:      balanceAdapter,
		DisplayManager:      displayManager,
		EmergencyRebalance:  container.EmergencyRebalance,
		UpdateDisplayTicker: container.UpdateDisplayTicker,
	})
	if err := sched.AddJob("0 */5 * * * *", syncCycle); err != nil {
		return nil, fmt.Errorf("failed to register sync_cycle job: %w", err)
	}
	instances.SyncCycle = syncCycle

	// ==========================================
	// Job 3: Dividend Reinvestment (daily at 10:00 AM)
	// ==========================================
	dividendReinvest := scheduler.NewDividendReinvestmentJob(scheduler.DividendReinvestmentConfig{
		Log:                   log,
		DividendRepo:          container.DividendRepo,
		SecurityRepo:          container.SecurityRepo,
		ScoreRepo:             container.ScoreRepo,
		PortfolioService:      container.PortfolioService,
		TradingService:        container.TradingService,
		TradeExecutionService: container.TradeExecutionService,
		TradernetClient:       container.TradernetClient,
		YahooClient:           container.YahooClient,
	})
	if err := sched.AddJob("0 0 10 * * *", dividendReinvest); err != nil {
		return nil, fmt.Errorf("failed to register dividend_reinvestment job: %w", err)
	}
	instances.DividendReinvest = dividendReinvest

	// ==========================================
	// Job 7: Planner Batch (every 15 minutes)
	// ==========================================
	plannerBatch := scheduler.NewPlannerBatchJob(scheduler.PlannerBatchConfig{
		Log:                    log,
		PositionRepo:           container.PositionRepo,
		SecurityRepo:           container.SecurityRepo,
		AllocRepo:              container.AllocRepo,
		CashManager:            container.CashManager,
		TradernetClient:        container.TradernetClient,
		YahooClient:            container.YahooClient,
		OptimizerService:       container.OptimizerService,
		OpportunitiesService:   container.OpportunitiesService,
		SequencesService:       container.SequencesService,
		EvaluationService:      container.EvaluationService,
		PlannerService:         container.PlannerService,
		ConfigRepo:             container.PlannerConfigRepo,
		RecommendationRepo:     container.RecommendationRepo,
		PortfolioDB:            container.PortfolioDB.Conn(),
		ConfigDB:               container.ConfigDB.Conn(),
		ScoreRepo:              container.ScoreRepo,
		MinPlanningIntervalMin: 15, // Minimum 15 minutes between planning cycles
	})
	if err := sched.AddJob("0 */15 * * * *", plannerBatch); err != nil {
		return nil, fmt.Errorf("failed to register planner_batch job: %w", err)
	}
	instances.PlannerBatch = plannerBatch

	// ==========================================
	// Job 8: Event-Based Trading (every 5 minutes)
	// ==========================================
	eventBasedTrading := scheduler.NewEventBasedTradingJob(scheduler.EventBasedTradingConfig{
		Log:                     log,
		RecommendationRepo:      container.RecommendationRepo,
		TradingService:          container.TradingService,
		MinExecutionIntervalMin: 30, // Minimum 30 minutes between trade executions
	})
	if err := sched.AddJob("0 */5 * * * *", eventBasedTrading); err != nil {
		return nil, fmt.Errorf("failed to register event_based_trading job: %w", err)
	}
	instances.EventBasedTrading = eventBasedTrading

	// ==========================================
	// RELIABILITY JOBS
	// ==========================================

	// Job 9: History Cleanup (daily at midnight)
	historyCleanup := cleanup.NewHistoryCleanupJob(
		container.HistoryDB,
		container.PortfolioDB,
		container.UniverseDB,
		log,
	)
	if err := sched.AddJob("0 0 0 * * *", historyCleanup); err != nil {
		return nil, fmt.Errorf("failed to register history_cleanup job: %w", err)
	}
	instances.HistoryCleanup = historyCleanup

	// Job 10: Hourly Backup (every hour at :00)
	hourlyBackup := reliability.NewHourlyBackupJob(container.BackupService)
	if err := sched.AddJob("0 0 * * * *", hourlyBackup); err != nil {
		return nil, fmt.Errorf("failed to register hourly_backup job: %w", err)
	}
	instances.HourlyBackup = hourlyBackup

	// Job 11: Daily Backup (daily at 1:00 AM, before maintenance)
	dailyBackup := reliability.NewDailyBackupJob(container.BackupService)
	if err := sched.AddJob("0 0 1 * * *", dailyBackup); err != nil {
		return nil, fmt.Errorf("failed to register daily_backup job: %w", err)
	}
	instances.DailyBackup = dailyBackup

	// Job 12: Daily Maintenance (daily at 2:00 AM)
	dataDir := cfg.DataDir
	backupDir := dataDir + "/backups"
	databases := map[string]*database.DB{
		"universe":  container.UniverseDB,
		"config":    container.ConfigDB,
		"ledger":    container.LedgerDB,
		"portfolio": container.PortfolioDB,
		"agents":    container.AgentsDB,
		"history":   container.HistoryDB,
		"cache":     container.CacheDB,
	}
	dailyMaintenance := reliability.NewDailyMaintenanceJob(databases, container.HealthServices, backupDir, log)
	if err := sched.AddJob("0 0 2 * * *", dailyMaintenance); err != nil {
		return nil, fmt.Errorf("failed to register daily_maintenance job: %w", err)
	}
	instances.DailyMaintenance = dailyMaintenance

	// ==========================================
	// ADAPTIVE MARKET HYPOTHESIS (AMH) SYSTEM
	// ==========================================

	// Adaptive Market Check Job (daily at 6:00 AM, after market data sync)
	adaptiveMarketJob := scheduler.NewAdaptiveMarketJob(scheduler.AdaptiveMarketJobConfig{
		Log:                 log,
		RegimeDetector:      container.RegimeDetector,
		RegimePersistence:   container.RegimePersistence,
		AdaptiveService:     container.AdaptiveMarketService,
		AdaptationThreshold: 0.1,                       // 10% change threshold
		ConfigDB:            container.ConfigDB.Conn(), // Database for storing adaptive parameters
	})
	if err := sched.AddJob("0 0 6 * * *", adaptiveMarketJob); err != nil {
		return nil, fmt.Errorf("failed to register adaptive_market_check job: %w", err)
	}
	instances.AdaptiveMarketJob = adaptiveMarketJob
	log.Info().Msg("Adaptive market check job registered (daily at 6:00 AM)")

	// ==========================================
	// Tag Update Jobs (multiple frequency tiers)
	// ==========================================
	tagUpdateJob := scheduler.NewTagUpdateJob(scheduler.TagUpdateConfig{
		Log:          log,
		SecurityRepo: container.SecurityRepo,
		ScoreRepo:    container.ScoreRepo,
		TagAssigner:  container.TagAssigner,
		YahooClient:  container.YahooClient,
		HistoryDB:    container.HistoryDBClient,
		PortfolioDB:  container.PortfolioDB.Conn(),
		PositionRepo: container.PositionRepo,
	})

	// Very Dynamic: Every 10 minutes (price/technical tags)
	if err := sched.AddJob("0 */10 * * * *", tagUpdateJob); err != nil {
		return nil, fmt.Errorf("failed to register tag_update (10min) job: %w", err)
	}

	// Dynamic: Every hour (opportunity/risk tags)
	if err := sched.AddJob("0 0 * * * *", tagUpdateJob); err != nil {
		return nil, fmt.Errorf("failed to register tag_update (hourly) job: %w", err)
	}

	// Stable: Daily at 3:00 AM (quality/characteristic tags)
	if err := sched.AddJob("0 0 3 * * *", tagUpdateJob); err != nil {
		return nil, fmt.Errorf("failed to register tag_update (daily) job: %w", err)
	}

	// Very Stable: Weekly on Sunday at 3:00 AM (long-term tags)
	if err := sched.AddJob("0 0 3 * * 0", tagUpdateJob); err != nil {
		return nil, fmt.Errorf("failed to register tag_update (weekly) job: %w", err)
	}
	instances.TagUpdate = tagUpdateJob

	// ==========================================
	// Additional Reliability Jobs
	// ==========================================

	// Job 13: Weekly Backup (Sunday at 1:00 AM)
	weeklyBackup := reliability.NewWeeklyBackupJob(container.BackupService)
	if err := sched.AddJob("0 0 1 * * 0", weeklyBackup); err != nil {
		return nil, fmt.Errorf("failed to register weekly_backup job: %w", err)
	}
	instances.WeeklyBackup = weeklyBackup

	// Job 14: Weekly Maintenance (Sunday at 3:30 AM)
	weeklyMaintenance := reliability.NewWeeklyMaintenanceJob(databases, container.HistoryDB, log)
	if err := sched.AddJob("0 30 3 * * 0", weeklyMaintenance); err != nil {
		return nil, fmt.Errorf("failed to register weekly_maintenance job: %w", err)
	}
	instances.WeeklyMaintenance = weeklyMaintenance

	// Job 15: Monthly Backup (1st day at 1:00 AM)
	monthlyBackup := reliability.NewMonthlyBackupJob(container.BackupService)
	if err := sched.AddJob("0 0 1 1 * *", monthlyBackup); err != nil {
		return nil, fmt.Errorf("failed to register monthly_backup job: %w", err)
	}
	instances.MonthlyBackup = monthlyBackup

	// Job 16: Monthly Maintenance (1st day at 4:00 AM)
	monthlyMaintenance := reliability.NewMonthlyMaintenanceJob(databases, container.HealthServices, container.AgentsDB, backupDir, log)
	if err := sched.AddJob("0 0 4 1 * *", monthlyMaintenance); err != nil {
		return nil, fmt.Errorf("failed to register monthly_maintenance job: %w", err)
	}
	instances.MonthlyMaintenance = monthlyMaintenance

	// ==========================================
	// Job 17: Formula Discovery (1st day of each month at 5:00 AM)
	// ==========================================
	// Initialize symbolic regression components
	formulaStorage := symbolic_regression.NewFormulaStorage(container.ConfigDB.Conn(), log)
	dataPrep := symbolic_regression.NewDataPrep(
		container.HistoryDB.Conn(),
		container.PortfolioDB.Conn(),
		container.ConfigDB.Conn(),
		container.UniverseDB.Conn(),
		log,
	)
	discoveryService := symbolic_regression.NewDiscoveryService(dataPrep, formulaStorage, log)
	formulaScheduler := symbolic_regression.NewScheduler(discoveryService, dataPrep, formulaStorage, log)
	formulaDiscovery := scheduler.NewFormulaDiscoveryJob(scheduler.FormulaDiscoveryConfig{
		Scheduler:      formulaScheduler,
		Log:            log,
		IntervalMonths: 1, // Monthly
		ForwardMonths:  6, // 6-month forward returns
	})
	if err := sched.AddJob("0 0 5 1 * *", formulaDiscovery); err != nil {
		return nil, fmt.Errorf("failed to register formula_discovery job: %w", err)
	}
	instances.FormulaDiscovery = formulaDiscovery

	log.Info().Int("jobs", 17).Msg("Background jobs registered successfully")

	return instances, nil
}
