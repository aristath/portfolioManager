package scheduler

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

// HealthCheckJob orchestrates individual health check jobs
// Runs every 6 hours to ensure database health
type HealthCheckJob struct {
	log                      zerolog.Logger
	checkCoreDatabasesJob    Job
	checkHistoryDatabasesJob Job
	checkWALCheckpointsJob   Job
}

// HealthCheckConfig holds configuration for health check job
type HealthCheckConfig struct {
	Log                      zerolog.Logger
	CheckCoreDatabasesJob    Job
	CheckHistoryDatabasesJob Job
	CheckWALCheckpointsJob   Job
}

// NewHealthCheckJob creates a new health check job
func NewHealthCheckJob(cfg HealthCheckConfig) *HealthCheckJob {
	return &HealthCheckJob{
		log:                      cfg.Log.With().Str("job", "health_check").Logger(),
		checkCoreDatabasesJob:    cfg.CheckCoreDatabasesJob,
		checkHistoryDatabasesJob: cfg.CheckHistoryDatabasesJob,
		checkWALCheckpointsJob:   cfg.CheckWALCheckpointsJob,
	}
}

// Name returns the job name
func (j *HealthCheckJob) Name() string {
	return "health_check"
}

// Run executes the health check by orchestrating individual health check jobs
// Note: Concurrent execution is prevented by the scheduler's SkipIfStillRunning wrapper
func (j *HealthCheckJob) Run() error {
	j.log.Info().Msg("Starting database health check")
	startTime := time.Now()

	// Step 1: Check core database integrity
	if j.checkCoreDatabasesJob != nil {
		if err := j.checkCoreDatabasesJob.Run(); err != nil {
			j.log.Error().Err(err).Msg("Core database integrity check failed")
			return err
		}
	} else {
		return fmt.Errorf("check core databases job not available")
	}

	// Step 2: Check history databases
	if j.checkHistoryDatabasesJob != nil {
		if err := j.checkHistoryDatabasesJob.Run(); err != nil {
			j.log.Error().Err(err).Msg("History database check failed (non-critical)")
			// Continue - history check is non-critical
		}
	}

	// Step 3: Check WAL checkpoints
	if j.checkWALCheckpointsJob != nil {
		if err := j.checkWALCheckpointsJob.Run(); err != nil {
			j.log.Error().Err(err).Msg("WAL checkpoint check failed (non-critical)")
			// Continue - WAL check is non-critical
		}
	}

	duration := time.Since(startTime)
	j.log.Info().
		Dur("duration", duration).
		Msg("Health check completed successfully")

	return nil
}
