// Package di provides dependency injection for scheduler jobs.
package di

import (
	"fmt"

	"github.com/aristath/portfolioManager/internal/config"
	"github.com/aristath/portfolioManager/internal/database"
	"github.com/aristath/portfolioManager/internal/modules/cleanup"
	"github.com/aristath/portfolioManager/internal/modules/display"
	"github.com/aristath/portfolioManager/internal/modules/symbolic_regression"
	"github.com/aristath/portfolioManager/internal/queue"
	"github.com/aristath/portfolioManager/internal/reliability"
	"github.com/aristath/portfolioManager/internal/scheduler"
	"github.com/rs/zerolog"
)

// RegisterJobs registers all jobs with the queue system
// Returns JobInstances for manual triggering via API
func RegisterJobs(container *Container, cfg *config.Config, displayManager *display.StateManager, log zerolog.Logger) (*JobInstances, error) {
	if container == nil {
		return nil, fmt.Errorf("container cannot be nil")
	}

	instances := &JobInstances{}

	// ==========================================
	// Register Event Listeners
	// ==========================================
	queue.RegisterListeners(container.EventBus, container.QueueManager, container.JobRegistry, log)

	// ==========================================
	// Job 1: Health Check
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
	container.JobRegistry.Register(queue.JobTypeHealthCheck, queue.JobToHandler(healthCheck))
	instances.HealthCheck = healthCheck

	// ==========================================
	// Job 2: Sync Cycle
	// ==========================================
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
		EventManager:        container.EventManager,
	})
	container.JobRegistry.Register(queue.JobTypeSyncCycle, queue.JobToHandler(syncCycle))
	instances.SyncCycle = syncCycle

	// ==========================================
	// Job 3: Dividend Reinvestment
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
	container.JobRegistry.Register(queue.JobTypeDividendReinvest, queue.JobToHandler(dividendReinvest))
	instances.DividendReinvest = dividendReinvest

	// ==========================================
	// Job 4: Planner Batch
	// ==========================================
	plannerBatch := scheduler.NewPlannerBatchJob(scheduler.PlannerBatchConfig{
		Log:                  log,
		PositionRepo:         container.PositionRepo,
		SecurityRepo:         container.SecurityRepo,
		AllocRepo:            container.AllocRepo,
		CashManager:          container.CashManager,
		TradernetClient:      container.TradernetClient,
		YahooClient:          container.YahooClient,
		OptimizerService:     container.OptimizerService,
		OpportunitiesService: container.OpportunitiesService,
		SequencesService:     container.SequencesService,
		EvaluationService:    container.EvaluationService,
		PlannerService:       container.PlannerService,
		ConfigRepo:           container.PlannerConfigRepo,
		RecommendationRepo:   container.RecommendationRepo,
		PortfolioDB:          container.PortfolioDB.Conn(),
		ConfigDB:             container.ConfigDB.Conn(),
		ScoreRepo:            container.ScoreRepo,
		EventManager:         container.EventManager,
	})
	container.JobRegistry.Register(queue.JobTypePlannerBatch, queue.JobToHandler(plannerBatch))
	instances.PlannerBatch = plannerBatch

	// ==========================================
	// Job 5: Event-Based Trading
	// ==========================================
	eventBasedTrading := scheduler.NewEventBasedTradingJob(scheduler.EventBasedTradingConfig{
		Log:                log,
		RecommendationRepo: container.RecommendationRepo,
		TradingService:     container.TradingService,
		EventManager:       container.EventManager,
	})
	container.JobRegistry.Register(queue.JobTypeEventBasedTrading, queue.JobToHandler(eventBasedTrading))
	instances.EventBasedTrading = eventBasedTrading

	// ==========================================
	// Job 6: Tag Update
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
	container.JobRegistry.Register(queue.JobTypeTagUpdate, queue.JobToHandler(tagUpdateJob))
	instances.TagUpdate = tagUpdateJob

	// ==========================================
	// RELIABILITY JOBS
	// ==========================================

	// Job 7: History Cleanup
	historyCleanup := cleanup.NewHistoryCleanupJob(
		container.HistoryDB,
		container.PortfolioDB,
		container.UniverseDB,
		log,
	)
	container.JobRegistry.Register(queue.JobTypeHistoryCleanup, queue.JobToHandler(historyCleanup))
	instances.HistoryCleanup = historyCleanup

	// Job 8: Hourly Backup
	hourlyBackup := reliability.NewHourlyBackupJob(container.BackupService)
	container.JobRegistry.Register(queue.JobTypeHourlyBackup, queue.JobToHandler(hourlyBackup))
	instances.HourlyBackup = hourlyBackup

	// Job 9: Daily Backup
	dailyBackup := reliability.NewDailyBackupJob(container.BackupService)
	container.JobRegistry.Register(queue.JobTypeDailyBackup, queue.JobToHandler(dailyBackup))
	instances.DailyBackup = dailyBackup

	// Job 10: Daily Maintenance
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
	container.JobRegistry.Register(queue.JobTypeDailyMaintenance, queue.JobToHandler(dailyMaintenance))
	instances.DailyMaintenance = dailyMaintenance

	// Job 11: Weekly Backup
	weeklyBackup := reliability.NewWeeklyBackupJob(container.BackupService)
	container.JobRegistry.Register(queue.JobTypeWeeklyBackup, queue.JobToHandler(weeklyBackup))
	instances.WeeklyBackup = weeklyBackup

	// Job 12: Weekly Maintenance
	weeklyMaintenance := reliability.NewWeeklyMaintenanceJob(databases, container.HistoryDB, log)
	container.JobRegistry.Register(queue.JobTypeWeeklyMaintenance, queue.JobToHandler(weeklyMaintenance))
	instances.WeeklyMaintenance = weeklyMaintenance

	// Job 13: Monthly Backup
	monthlyBackup := reliability.NewMonthlyBackupJob(container.BackupService)
	container.JobRegistry.Register(queue.JobTypeMonthlyBackup, queue.JobToHandler(monthlyBackup))
	instances.MonthlyBackup = monthlyBackup

	// Job 14: Monthly Maintenance
	monthlyMaintenance := reliability.NewMonthlyMaintenanceJob(databases, container.HealthServices, container.AgentsDB, backupDir, log)
	container.JobRegistry.Register(queue.JobTypeMonthlyMaintenance, queue.JobToHandler(monthlyMaintenance))
	instances.MonthlyMaintenance = monthlyMaintenance

	// ==========================================
	// ADAPTIVE MARKET HYPOTHESIS (AMH) SYSTEM
	// ==========================================

	// Job 15: Adaptive Market Check
	adaptiveMarketJob := scheduler.NewAdaptiveMarketJob(scheduler.AdaptiveMarketJobConfig{
		Log:                 log,
		RegimeDetector:      container.RegimeDetector,
		RegimePersistence:   container.RegimePersistence,
		AdaptiveService:     container.AdaptiveMarketService,
		AdaptationThreshold: 0.1,
		ConfigDB:            container.ConfigDB.Conn(),
	})
	container.JobRegistry.Register(queue.JobTypeAdaptiveMarket, queue.JobToHandler(adaptiveMarketJob))
	instances.AdaptiveMarketJob = adaptiveMarketJob

	// ==========================================
	// Job 16: Formula Discovery
	// ==========================================
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
		IntervalMonths: 1,
		ForwardMonths:  6,
	})
	container.JobRegistry.Register(queue.JobTypeFormulaDiscovery, queue.JobToHandler(formulaDiscovery))
	instances.FormulaDiscovery = formulaDiscovery

	// ==========================================
	// Start Queue System
	// ==========================================
	container.WorkerPool.Start()
	container.TimeScheduler.Start()

	log.Info().Int("jobs", 16).Msg("Jobs registered with queue system")

	return instances, nil
}
