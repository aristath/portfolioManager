package scheduler

import (
	"github.com/rs/zerolog"
)

// MetadataService defines the interface for metadata synchronization
type MetadataService interface {
	GetAllActiveISINs() []string
	SyncMetadataBatch(isins []string) (int, error)
}

// MetadataSyncJob runs batch metadata sync at 3 AM daily.
// This job uses the batch API to sync all active securities in a single request,
// eliminating 429 rate limit errors from individual sync requests.
//
// Schedule: "0 3 * * *" (3 AM daily)
// Fallback: If batch sync fails, work processor handles individual retries
type MetadataSyncJob struct {
	JobBase
	metadataService MetadataService
	log             zerolog.Logger
}

// NewMetadataSyncJob creates a new metadata batch sync job
func NewMetadataSyncJob(metadataService MetadataService, log zerolog.Logger) *MetadataSyncJob {
	return &MetadataSyncJob{
		metadataService: metadataService,
		log:             log.With().Str("job", "metadata_batch_sync").Logger(),
	}
}

// Run executes the batch metadata sync
func (j *MetadataSyncJob) Run() error {
	j.log.Info().Msg("Starting daily metadata batch sync (3 AM)")

	// Get all active ISINs (excludes indices)
	isins := j.metadataService.GetAllActiveISINs()
	if len(isins) == 0 {
		j.log.Warn().Msg("No active ISINs found for batch sync")
		return nil
	}

	j.log.Debug().
		Int("count", len(isins)).
		Msg("Attempting batch metadata sync")

	// Attempt batch sync
	successCount, err := j.metadataService.SyncMetadataBatch(isins)
	if err != nil {
		// Log warning but don't propagate error - work processor will handle individual retries
		j.log.Warn().
			Err(err).
			Int("isins", len(isins)).
			Msg("Batch sync failed, work processor will handle individual retries")
		return nil
	}

	j.log.Info().
		Int("total", len(isins)).
		Int("success", successCount).
		Msg("Daily batch metadata sync completed successfully")

	return nil
}

// Name returns the job name
func (j *MetadataSyncJob) Name() string {
	return "metadata_batch_sync"
}
