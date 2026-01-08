package scheduler

import (
	"fmt"

	"github.com/rs/zerolog"
)

// SyncCashFlowsJob syncs cash flows from Tradernet
type SyncCashFlowsJob struct {
	log              zerolog.Logger
	cashFlowsService CashFlowsServiceInterface
}

// SyncCashFlowsConfig holds configuration for sync cash flows job
type SyncCashFlowsConfig struct {
	Log              zerolog.Logger
	CashFlowsService CashFlowsServiceInterface
}

// NewSyncCashFlowsJob creates a new sync cash flows job
func NewSyncCashFlowsJob(cfg SyncCashFlowsConfig) *SyncCashFlowsJob {
	return &SyncCashFlowsJob{
		log:              cfg.Log.With().Str("job", "sync_cash_flows").Logger(),
		cashFlowsService: cfg.CashFlowsService,
	}
}

// Name returns the job name
func (j *SyncCashFlowsJob) Name() string {
	return "sync_cash_flows"
}

// Run executes the sync cash flows job
func (j *SyncCashFlowsJob) Run() error {
	j.log.Info().Msg("Starting sync cash flows")

	if j.cashFlowsService == nil {
		j.log.Warn().Msg("Cash flows service not available, skipping")
		return nil // Non-critical, don't fail
	}

	if err := j.cashFlowsService.SyncFromTradernet(); err != nil {
		j.log.Error().Err(err).Msg("Failed to sync cash flows")
		return fmt.Errorf("sync cash flows failed: %w", err)
	}

	j.log.Info().Msg("Sync cash flows completed successfully")
	return nil
}
