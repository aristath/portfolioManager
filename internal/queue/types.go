package queue

import "time"

// JobType represents the type of job
type JobType string

const (
	// Original composite jobs (kept for backward compatibility)
	JobTypePlannerBatch       JobType = "planner_batch"
	JobTypeEventBasedTrading  JobType = "event_based_trading"
	JobTypeTagUpdate          JobType = "tag_update"
	JobTypeSyncCycle          JobType = "sync_cycle"
	JobTypeDividendReinvest   JobType = "dividend_reinvestment"
	JobTypeHealthCheck        JobType = "health_check"
	JobTypeHourlyBackup       JobType = "hourly_backup"
	JobTypeDailyBackup        JobType = "daily_backup"
	JobTypeDailyMaintenance   JobType = "daily_maintenance"
	JobTypeWeeklyBackup       JobType = "weekly_backup"
	JobTypeWeeklyMaintenance  JobType = "weekly_maintenance"
	JobTypeMonthlyBackup      JobType = "monthly_backup"
	JobTypeMonthlyMaintenance JobType = "monthly_maintenance"
	JobTypeFormulaDiscovery   JobType = "formula_discovery"
	JobTypeAdaptiveMarket     JobType = "adaptive_market_check"
	JobTypeHistoryCleanup     JobType = "history_cleanup"

	// Sync jobs - individual responsibilities split from sync_cycle
	JobTypeSyncTrades            JobType = "sync_trades"
	JobTypeSyncCashFlows         JobType = "sync_cash_flows"
	JobTypeSyncPortfolio         JobType = "sync_portfolio"
	JobTypeCheckNegativeBalances JobType = "check_negative_balances"
	JobTypeSyncPrices            JobType = "sync_prices"
	JobTypeUpdateDisplayTicker   JobType = "update_display_ticker"

	// Planning jobs - individual responsibilities split from planner_batch
	JobTypeGeneratePortfolioHash   JobType = "generate_portfolio_hash"
	JobTypeGetOptimizerWeights     JobType = "get_optimizer_weights"
	JobTypeBuildOpportunityContext JobType = "build_opportunity_context"
	JobTypeIdentifyOpportunities   JobType = "identify_opportunities"
	JobTypeGenerateSequences       JobType = "generate_sequences"
	JobTypeEvaluateSequences       JobType = "evaluate_sequences"
	JobTypeCreateTradePlan         JobType = "create_trade_plan"
	JobTypeStoreRecommendations    JobType = "store_recommendations"

	// Dividend jobs - individual responsibilities split from dividend_reinvestment
	JobTypeGetUnreinvestedDividends      JobType = "get_unreinvested_dividends"
	JobTypeGroupDividendsBySymbol        JobType = "group_dividends_by_symbol"
	JobTypeCheckDividendYields           JobType = "check_dividend_yields"
	JobTypeCreateDividendRecommendations JobType = "create_dividend_recommendations"
	JobTypeSetPendingBonuses             JobType = "set_pending_bonuses"
	JobTypeExecuteDividendTrades         JobType = "execute_dividend_trades"

	// Health check jobs - individual responsibilities split from health_check
	JobTypeCheckCoreDatabases    JobType = "check_core_databases"
	JobTypeCheckHistoryDatabases JobType = "check_history_databases"
	JobTypeCheckWALCheckpoints   JobType = "check_wal_checkpoints"
)

// Priority represents job priority
type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

// Job represents a queued job
type Job struct {
	ID          string
	Type        JobType
	Priority    Priority
	Payload     map[string]interface{}
	CreatedAt   time.Time
	AvailableAt time.Time
	Retries     int
	MaxRetries  int
}

// Queue interface for job queue operations
type Queue interface {
	Enqueue(job *Job) error
	Dequeue() (*Job, error)
	Size() int
}
