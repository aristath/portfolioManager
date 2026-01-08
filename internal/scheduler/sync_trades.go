package scheduler

import (
	"fmt"

	"github.com/rs/zerolog"
)

// SyncTradesJob syncs trades from Tradernet
type SyncTradesJob struct {
	log            zerolog.Logger
	tradingService TradingServiceInterface
}

// SyncTradesConfig holds configuration for sync trades job
type SyncTradesConfig struct {
	Log            zerolog.Logger
	TradingService TradingServiceInterface
}

// NewSyncTradesJob creates a new sync trades job
func NewSyncTradesJob(cfg SyncTradesConfig) *SyncTradesJob {
	return &SyncTradesJob{
		log:            cfg.Log.With().Str("job", "sync_trades").Logger(),
		tradingService: cfg.TradingService,
	}
}

// Name returns the job name
func (j *SyncTradesJob) Name() string {
	return "sync_trades"
}

// Run executes the sync trades job
func (j *SyncTradesJob) Run() error {
	j.log.Info().Msg("Starting sync trades")

	if j.tradingService == nil {
		j.log.Warn().Msg("Trading service not available, skipping")
		return nil // Non-critical, don't fail
	}

	if err := j.tradingService.SyncFromTradernet(); err != nil {
		j.log.Error().Err(err).Msg("Failed to sync trades")
		return fmt.Errorf("sync trades failed: %w", err)
	}

	j.log.Info().Msg("Sync trades completed successfully")
	return nil
}
