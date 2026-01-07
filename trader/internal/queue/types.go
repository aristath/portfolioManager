package queue

import "time"

// JobType represents the type of job
type JobType string

const (
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
