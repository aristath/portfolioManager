package scheduler

import (
	"fmt"

	"github.com/rs/zerolog"
)

// SyncPricesJob syncs prices for all securities
type SyncPricesJob struct {
	log             zerolog.Logger
	universeService UniverseServiceInterface
}

// SyncPricesConfig holds configuration for sync prices job
type SyncPricesConfig struct {
	Log             zerolog.Logger
	UniverseService UniverseServiceInterface
}

// NewSyncPricesJob creates a new sync prices job
func NewSyncPricesJob(cfg SyncPricesConfig) *SyncPricesJob {
	return &SyncPricesJob{
		log:             cfg.Log.With().Str("job", "sync_prices").Logger(),
		universeService: cfg.UniverseService,
	}
}

// Name returns the job name
func (j *SyncPricesJob) Name() string {
	return "sync_prices"
}

// Run executes the sync prices job
func (j *SyncPricesJob) Run() error {
	j.log.Debug().Msg("Syncing prices for all securities")

	if j.universeService == nil {
		j.log.Warn().Msg("Universe service not available, skipping price sync")
		return nil
	}

	if err := j.universeService.SyncPrices(); err != nil {
		j.log.Error().Err(err).Msg("Price sync failed")
		return fmt.Errorf("sync prices failed: %w", err)
	}

	j.log.Debug().Msg("Price sync completed")
	return nil
}
