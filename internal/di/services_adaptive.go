/**
 * Package di provides dependency injection for adaptive market service initialization.
 *
 * Step 14: Initialize Adaptive Market Services
 * Adaptive market services handle market regime detection and adaptive behavior.
 */
package di

import (
	"github.com/aristath/sentinel/internal/market_regime"
	"github.com/aristath/sentinel/internal/modules/adaptation"
	symbolic_regression "github.com/aristath/sentinel/internal/modules/symbolic_regression"
	"github.com/rs/zerolog"
)

// initializeAdaptiveMarketServices initializes adaptive market-related services.
func initializeAdaptiveMarketServices(container *Container, log zerolog.Logger) error {
	// Market index service for market-wide regime detection
	// Manages market indices (SPY, QQQ, etc.) for regime detection
	// Use TradingSecurityProviderAdapter for ISIN lookups
	marketIndexSecurityProvider := NewTradingSecurityProviderAdapter(container.SecurityRepo)
	container.MarketIndexService = market_regime.NewMarketIndexService(
		marketIndexSecurityProvider,
		container.HistoryDBClient,
		container.BrokerClient,
		log,
	)

	// Index repository for per-region market indices
	// Stores market index configuration (which indices to track per region)
	container.IndexRepository = market_regime.NewIndexRepository(container.ConfigDB.Conn(), log)

	// Index sync service - ensures indices exist in both config DB and universe DB
	// Syncs index definitions to both databases (idempotent operation)
	container.IndexSyncService = market_regime.NewIndexSyncService(
		container.SecurityRepo,
		container.OverrideRepo,
		container.ConfigDB.Conn(),
		log,
	)

	// Sync known indices to both databases (idempotent - safe to run on every startup)
	// This ensures indices are in market_indices (config) AND securities (universe) tables
	// Market regime detection needs indices in both places
	if err := container.IndexSyncService.SyncAll(); err != nil {
		log.Warn().Err(err).Msg("Failed to sync market indices to databases (will use fallback)")
		// Don't fail startup - fallback to hardcoded indices will work
	}

	// Sync historical prices for indices (needed for regime calculation)
	// This fetches price data from broker API for all PRICE indices
	// First run: fetches 10 years of data; subsequent runs: fetches 1 year of updates
	// Regime detection requires historical price data to calculate moving averages
	if container.HistoricalSyncService != nil {
		if err := container.IndexSyncService.SyncHistoricalPricesForIndices(container.HistoricalSyncService); err != nil {
			log.Warn().Err(err).Msg("Failed to sync historical prices for indices (regime calculation may be limited)")
			// Don't fail startup - regime detection will fall back to neutral scores
		}
	}

	// Regime persistence for smoothing and history
	// Stores regime history and provides smoothing to prevent regime oscillation
	container.RegimePersistence = market_regime.NewRegimePersistence(container.ConfigDB.Conn(), log)

	// Market regime detector
	// Detects market regime (bull, bear, sideways) based on index moving averages
	container.RegimeDetector = market_regime.NewMarketRegimeDetector(log)
	container.RegimeDetector.SetMarketIndexService(container.MarketIndexService)
	container.RegimeDetector.SetRegimePersistence(container.RegimePersistence)

	// Adaptive market service
	// Implements Adaptive Market Hypothesis - adjusts behavior based on market regime
	container.AdaptiveMarketService = adaptation.NewAdaptiveMarketService(
		container.RegimeDetector,
		nil, // performanceTracker - optional
		nil, // weightsCalculator - optional
		nil, // repository - optional
		log,
	)

	// Regime score provider adapter
	// Provides current regime score (0-1) for adaptive services
	container.RegimeScoreProvider = market_regime.NewRegimeScoreProviderAdapter(container.RegimePersistence)

	// Wire up adaptive services to integration points
	// OptimizerService uses adaptive service for regime-aware optimization
	container.OptimizerService.SetAdaptiveService(container.AdaptiveMarketService)
	container.OptimizerService.SetRegimeScoreProvider(container.RegimeScoreProvider)
	log.Info().Msg("Adaptive service wired to OptimizerService")

	// TagAssigner: adaptive quality gates
	// Quality gate thresholds adjust based on market regime
	// Create adapter to bridge type mismatch
	tagAssignerAdapter := &qualityGatesAdapter{service: container.AdaptiveMarketService}
	container.TagAssigner.SetAdaptiveService(tagAssignerAdapter)
	container.TagAssigner.SetRegimeScoreProvider(container.RegimeScoreProvider)
	log.Info().Msg("Adaptive service wired to TagAssigner")

	// SecurityScorer: adaptive weights and per-region regime scores
	// Scoring weights adjust based on market regime
	// AdaptiveMarketService implements scorers.AdaptiveWeightsProvider interface directly
	container.SecurityScorer.SetAdaptiveService(container.AdaptiveMarketService)
	container.SecurityScorer.SetRegimeScoreProvider(container.RegimeScoreProvider)

	// Wire formula storage for discovered scoring formulas
	// SecurityScorer can use discovered formulas for improved scoring accuracy
	// FormulaStorage is stateless (wraps DB), so sharing instance is optional but cleaner
	// ReturnsCalculator already creates its own, but we wire shared instance to SecurityScorer here
	if container.PortfolioDB != nil {
		formulaStorage := symbolic_regression.NewFormulaStorage(container.PortfolioDB.Conn(), log)
		container.SecurityScorer.SetFormulaStorage(formulaStorage)
		log.Info().Msg("Formula storage wired to SecurityScorer")
	}

	log.Info().Msg("Adaptive service and regime score provider wired to SecurityScorer")

	return nil
}
