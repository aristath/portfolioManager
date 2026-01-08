package scheduler

import (
	"fmt"

	"github.com/rs/zerolog"
)

// SyncPortfolioJob syncs portfolio positions from Tradernet
// CRITICAL - errors are returned and stop execution
type SyncPortfolioJob struct {
	log              zerolog.Logger
	portfolioService PortfolioServiceInterface
}

// SyncPortfolioConfig holds configuration for sync portfolio job
type SyncPortfolioConfig struct {
	Log              zerolog.Logger
	PortfolioService PortfolioServiceInterface
}

// NewSyncPortfolioJob creates a new sync portfolio job
func NewSyncPortfolioJob(cfg SyncPortfolioConfig) *SyncPortfolioJob {
	return &SyncPortfolioJob{
		log:              cfg.Log.With().Str("job", "sync_portfolio").Logger(),
		portfolioService: cfg.PortfolioService,
	}
}

// Name returns the job name
func (j *SyncPortfolioJob) Name() string {
	return "sync_portfolio"
}

// Run executes the sync portfolio job
// CRITICAL - errors are returned and stop execution
func (j *SyncPortfolioJob) Run() error {
	j.log.Info().Msg("Starting sync portfolio (CRITICAL)")

	if j.portfolioService == nil {
		return fmt.Errorf("portfolio service not available")
	}

	if err := j.portfolioService.SyncFromTradernet(); err != nil {
		j.log.Error().Err(err).Msg("CRITICAL: Portfolio sync failed")
		return fmt.Errorf("sync portfolio failed: %w", err)
	}

	j.log.Info().Msg("Portfolio sync completed successfully")
	return nil
}
