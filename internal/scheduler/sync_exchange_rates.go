package scheduler

import (
	"github.com/aristath/sentinel/internal/services"
	"github.com/rs/zerolog"
)

// SyncExchangeRatesJob syncs exchange rates from external APIs
// Non-critical - errors logged but don't block sync_cycle
type SyncExchangeRatesJob struct {
	log                      zerolog.Logger
	exchangeRateCacheService *services.ExchangeRateCacheService
}

// SyncExchangeRatesConfig holds configuration for sync exchange rates job
type SyncExchangeRatesConfig struct {
	Log                      zerolog.Logger
	ExchangeRateCacheService *services.ExchangeRateCacheService
}

// NewSyncExchangeRatesJob creates a new sync exchange rates job
func NewSyncExchangeRatesJob(cfg SyncExchangeRatesConfig) *SyncExchangeRatesJob {
	return &SyncExchangeRatesJob{
		log:                      cfg.Log.With().Str("job", "sync_exchange_rates").Logger(),
		exchangeRateCacheService: cfg.ExchangeRateCacheService,
	}
}

// Name returns the job name
func (j *SyncExchangeRatesJob) Name() string {
	return "sync_exchange_rates"
}

// Run executes the sync exchange rates job
// Non-critical - errors logged but don't propagate
func (j *SyncExchangeRatesJob) Run() error {
	j.log.Info().Msg("Starting exchange rate sync")

	if j.exchangeRateCacheService == nil {
		j.log.Error().Msg("Exchange rate cache service not available")
		return nil // Don't propagate error
	}

	if err := j.exchangeRateCacheService.SyncRates(); err != nil {
		j.log.Warn().Err(err).Msg("Exchange rate sync failed (non-critical)")
		return nil // Don't propagate error
	}

	j.log.Info().Msg("Exchange rate sync completed successfully")
	return nil
}
