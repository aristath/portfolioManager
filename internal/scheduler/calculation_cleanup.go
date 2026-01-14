package scheduler

import (
	"github.com/aristath/sentinel/internal/modules/calculations"
	"github.com/rs/zerolog"
)

// CalculationCleanupJob removes expired entries from the calculation cache.
// This runs daily to clean up stale technical_cache and optimizer_cache entries.
type CalculationCleanupJob struct {
	JobBase
	cache *calculations.Cache
	log   zerolog.Logger
}

// NewCalculationCleanupJob creates a new calculation cleanup job
func NewCalculationCleanupJob(cache *calculations.Cache, log zerolog.Logger) *CalculationCleanupJob {
	return &CalculationCleanupJob{
		cache: cache,
		log:   log.With().Str("job", "calculation_cleanup").Logger(),
	}
}

// Name returns the job name
func (j *CalculationCleanupJob) Name() string {
	return "calculation_cleanup"
}

// Run executes the cleanup
func (j *CalculationCleanupJob) Run() error {
	count, err := j.cache.Cleanup()
	if err != nil {
		j.log.Error().Err(err).Msg("Cleanup failed")
		return err
	}
	if count > 0 {
		j.log.Info().Int64("removed", count).Msg("Cleaned expired cache entries")
	} else {
		j.log.Debug().Msg("No expired cache entries to clean")
	}
	return nil
}
