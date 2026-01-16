package scheduler

import (
	"time"

	"github.com/aristath/sentinel/internal/modules/symbolic_regression"
	"github.com/rs/zerolog"
)

// FormulaDiscoveryJob handles periodic formula re-discovery
// Runs monthly/quarterly to update formulas as market conditions change
type FormulaDiscoveryJob struct {
	JobBase
	scheduler      *symbolic_regression.Scheduler
	log            zerolog.Logger
	intervalMonths int // How often to run (1 = monthly, 3 = quarterly)
	forwardMonths  int // Forward return horizon (6 or 12)
}

// FormulaDiscoveryConfig holds configuration for formula discovery job
type FormulaDiscoveryConfig struct {
	Scheduler      *symbolic_regression.Scheduler
	Log            zerolog.Logger
	IntervalMonths int // 1 = monthly, 3 = quarterly
	ForwardMonths  int // 6 or 12
}

// NewFormulaDiscoveryJob creates a new formula discovery job
func NewFormulaDiscoveryJob(cfg FormulaDiscoveryConfig) *FormulaDiscoveryJob {
	return &FormulaDiscoveryJob{
		scheduler:      cfg.Scheduler,
		log:            cfg.Log.With().Str("job", "formula_discovery").Logger(),
		intervalMonths: cfg.IntervalMonths,
		forwardMonths:  cfg.ForwardMonths,
	}
}

// Name returns the job name
func (j *FormulaDiscoveryJob) Name() string {
	return "formula_discovery"
}

// Run executes the formula discovery for all formula types
func (j *FormulaDiscoveryJob) Run() error {
	j.log.Info().Msg("Starting periodic formula discovery")
	startTime := time.Now()

	// Run all scheduled discoveries
	err := j.scheduler.RunAllScheduledDiscoveries(j.intervalMonths, j.forwardMonths)
	if err != nil {
		j.log.Error().Err(err).Msg("Formula discovery failed")
		return err
	}

	duration := time.Since(startTime)
	j.log.Info().
		Dur("duration", duration).
		Int("interval_months", j.intervalMonths).
		Int("forward_months", j.forwardMonths).
		Msg("Periodic formula discovery completed successfully")

	return nil
}
