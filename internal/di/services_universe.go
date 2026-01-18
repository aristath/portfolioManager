/**
 * Package di provides dependency injection for universe service initialization.
 *
 * Step 5: Initialize Universe Services
 * Universe services manage the investment universe (securities, historical data, symbol resolution).
 */
package di

import (
	"fmt"

	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/aristath/sentinel/internal/scheduler"
	"github.com/rs/zerolog"
)

// initializeUniverseServices initializes universe-related services.
func initializeUniverseServices(container *Container, log zerolog.Logger) error {
	// Historical sync service (uses Tradernet as primary source for historical data)
	// Fetches historical prices from Tradernet API and stores in history.db
	// Stores raw data - filtering happens on read via HistoryDB's PriceFilter
	container.HistoricalSyncService = universe.NewHistoricalSyncService(
		container.BrokerClient, // Tradernet is now single source of truth
		container.SecurityRepo,
		container.HistoryDBClient,
		log,
	)

	// Symbol resolver
	// Resolves security identifiers (ISIN, symbol) to security objects
	container.SymbolResolver = universe.NewSymbolResolver(
		container.BrokerClient,
		container.SecurityRepo,
		log,
	)

	// Security setup service (scoreCalculator will be set later)
	// Auto-adds missing securities when referenced in trades/positions
	// scoreCalculator will be wired later after SecurityScorer is created
	container.SetupService = universe.NewSecuritySetupService(
		container.SymbolResolver,
		container.SecurityRepo,
		container.BrokerClient,
		container.HistoricalSyncService,
		container.EventManager,
		nil, // scoreCalculator - will be set later
		log,
	)

	// Security deletion service
	// Handles security deletion with cleanup of related data (positions, scores, history)
	container.SecurityDeletionService = universe.NewSecurityDeletionService(
		container.SecurityRepo,
		container.PositionRepo,
		container.ScoreRepo,
		container.HistoryDBClient,
		container.BrokerClient,
		log,
	)

	// Metadata sync service (batch + individual)
	// Syncs security metadata from broker API (supports batch operations to avoid 429 rate limits)
	// Used by both the scheduled batch job (3 AM) and work processor (individual retries)
	container.MetadataSyncService = universe.NewMetadataSyncService(
		container.SecurityRepo,
		container.BrokerClient,
		log,
	)

	// Scheduler for time-based jobs (robfig/cron)
	// Manages cron-based job execution with proper concurrency control
	container.Scheduler = scheduler.New(log)

	// Daily batch metadata sync job (runs at 3 AM)
	// Uses MetadataSyncService to sync all security metadata in a single batch API call
	metadataSyncJob := scheduler.NewMetadataSyncJob(container.MetadataSyncService, log)
	if err := container.Scheduler.AddJob("0 0 3 * * *", metadataSyncJob); err != nil {
		return fmt.Errorf("failed to register metadata sync job: %w", err)
	}

	return nil
}
