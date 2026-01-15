package universe

import (
	"fmt"
	"time"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/rs/zerolog"
)

// SyncService handles data synchronization for securities
// Uses Tradernet as the single source for price data
type SyncService struct {
	securityRepo    *SecurityRepository
	historicalSync  *HistoricalSyncService
	scoreCalculator ScoreCalculator
	brokerClient    domain.BrokerClient
	setupService    *SecuritySetupService
	db              DBExecutor
	log             zerolog.Logger
}

// NewSyncService creates a new sync service
func NewSyncService(
	securityRepo *SecurityRepository,
	historicalSync *HistoricalSyncService,
	scoreCalculator ScoreCalculator,
	brokerClient domain.BrokerClient,
	setupService *SecuritySetupService,
	db DBExecutor,
	log zerolog.Logger,
) *SyncService {
	return &SyncService{
		securityRepo:    securityRepo,
		historicalSync:  historicalSync,
		scoreCalculator: scoreCalculator,
		brokerClient:    brokerClient,
		setupService:    setupService,
		db:              db,
		log:             log.With().Str("service", "sync").Logger(),
	}
}

// SetScoreCalculator sets the score calculator (for deferred wiring)
func (s *SyncService) SetScoreCalculator(calculator ScoreCalculator) {
	s.scoreCalculator = calculator
}

// SyncThresholdHours is how old last_synced must be to require processing (24 hours)
const SyncThresholdHours = 24

// SyncSecuritiesData runs the securities data sync for all securities needing sync
// Faithful translation from Python: app/jobs/securities_data_sync.py -> run_securities_data_sync()
//
// This is the main entry point called by the scheduler every hour.
// It processes securities that haven't been synced in 24 hours.
func (s *SyncService) SyncSecuritiesData() (int, int, error) {
	s.log.Info().Msg("Starting securities data sync")

	securities, err := s.getSecuritiesNeedingSync()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get securities needing sync: %w", err)
	}

	if len(securities) == 0 {
		s.log.Info().Msg("All securities are up to date, no processing needed")
		return 0, 0, nil
	}

	s.log.Info().Int("count", len(securities)).Msg("Processing securities needing sync")

	processed := 0
	errors := 0

	for _, security := range securities {
		err := s.processSingleSecurity(security.Symbol)
		if err != nil {
			s.log.Error().Err(err).Str("symbol", security.Symbol).Msg("Pipeline failed for security")
			errors++
		} else {
			processed++
		}
	}

	s.log.Info().
		Int("processed", processed).
		Int("errors", errors).
		Msg("Securities data sync complete")

	return processed, errors, nil
}

// RefreshSingleSecurity force refreshes a single security's data
// Faithful translation from Python: app/jobs/securities_data_sync.py -> refresh_single_security()
//
// This bypasses the last_synced check and immediately processes the security.
// Used by the API endpoint for manual refreshes.
func (s *SyncService) RefreshSingleSecurity(symbol string) error {
	s.log.Info().Str("symbol", symbol).Msg("Force refreshing data for security")

	// Run the full pipeline for this security
	err := s.processSingleSecurity(symbol)
	if err != nil {
		return fmt.Errorf("force refresh failed: %w", err)
	}

	s.log.Info().Str("symbol", symbol).Msg("Force refresh complete")
	return nil
}

// processSingleSecurity processes a single security through the full data pipeline
// Faithful translation from Python: app/jobs/securities_data_sync.py -> _process_single_security()
//
// Steps:
// 1. Sync historical prices from Tradernet
// 2. Detect and update country/exchange (from Tradernet metadata)
// 3. Detect and update industry (from Tradernet metadata)
// 4. Refresh security score
// 5. Update last_synced timestamp
func (s *SyncService) processSingleSecurity(symbol string) error {
	s.log.Info().Str("symbol", symbol).Msg("Processing security")

	// Step 1: Sync historical prices
	if s.historicalSync != nil {
		err := s.historicalSync.SyncHistoricalPrices(symbol)
		if err != nil {
			return fmt.Errorf("failed to sync historical prices: %w", err)
		}
	}

	// Step 2: Refresh score
	security, err := s.securityRepo.GetBySymbol(symbol)
	if err != nil {
		return fmt.Errorf("failed to get security: %w", err)
	}
	if security == nil {
		return fmt.Errorf("security not found: %s", symbol)
	}
	if security.ISIN == "" {
		return fmt.Errorf("security missing ISIN: %s", symbol)
	}

	if s.scoreCalculator != nil {
		// ScoreCalculator accepts symbol (looks up ISIN internally)
		// Client symbols no longer needed - all data comes from internal sources
		err = s.scoreCalculator.CalculateAndSaveScore(
			symbol,
			security.Geography,
			security.Industry,
		)
		if err != nil {
			s.log.Warn().Err(err).Str("symbol", symbol).Str("isin", security.ISIN).Msg("Failed to refresh score")
			// Continue - not fatal
		}
	}

	// Step 5: Mark as synced (using ISIN)
	err = s.updateLastSynced(security.ISIN)
	if err != nil {
		return fmt.Errorf("failed to update last_synced: %w", err)
	}

	s.log.Info().Str("symbol", symbol).Msg("Pipeline complete for security")
	return nil
}

// getSecuritiesNeedingSync gets all active securities that need to be synced
// Faithful translation from Python: app/jobs/securities_data_sync.py -> _get_securities_needing_sync()
//
// A security needs sync if:
// - last_synced is NULL (never synced)
// - last_synced is older than SYNC_THRESHOLD_HOURS
func (s *SyncService) getSecuritiesNeedingSync() ([]Security, error) {
	allSecurities, err := s.securityRepo.GetAllActive()
	if err != nil {
		return nil, fmt.Errorf("failed to get active securities: %w", err)
	}

	thresholdUnix := time.Now().Add(-SyncThresholdHours * time.Hour).Unix()

	var securitiesNeedingSync []Security
	for _, security := range allSecurities {
		if security.LastSynced == nil {
			// Never synced
			securitiesNeedingSync = append(securitiesNeedingSync, security)
		} else if *security.LastSynced < thresholdUnix {
			// Synced more than threshold hours ago - compare Unix timestamps directly
			securitiesNeedingSync = append(securitiesNeedingSync, security)
		}
	}

	return securitiesNeedingSync, nil
}

// updateLastSynced updates the last_synced timestamp for a security
// After migration: accepts ISIN as primary identifier
func (s *SyncService) updateLastSynced(isin string) error {
	now := time.Now().Unix()

	err := s.securityRepo.Update(isin, map[string]interface{}{
		"last_synced": now,
	})
	if err != nil {
		return fmt.Errorf("failed to update last_synced: %w", err)
	}

	return nil
}

// SyncAllPrices syncs current prices for all active securities
// Faithful translation from Python: app/jobs/daily_sync.py -> sync_prices()
//
// This gets current quotes from Tradernet and updates position prices.
func (s *SyncService) SyncAllPrices() (int, error) {
	return s.SyncAllPricesWithReporter(nil)
}

// SyncAllPricesWithReporter syncs current prices using Tradernet as the sole source
func (s *SyncService) SyncAllPricesWithReporter(reporter ProgressReporter) (int, error) {
	s.log.Info().Msg("Starting price sync for all active securities (Tradernet primary)")

	const totalSteps = 3

	// Step 1: Get all active securities
	if reporter != nil {
		reporter.Report(1, totalSteps, "Getting securities")
	}

	securities, err := s.securityRepo.GetAllActive()
	if err != nil {
		return 0, fmt.Errorf("failed to get active securities: %w", err)
	}

	if len(securities) == 0 {
		s.log.Info().Msg("No securities to sync prices for")
		return 0, nil
	}

	// Build symbol list and mappings
	symbols := make([]string, 0, len(securities))
	symbolToSecurity := make(map[string]Security)
	for _, security := range securities {
		symbols = append(symbols, security.Symbol)
		symbolToSecurity[security.Symbol] = security
	}

	// Step 2: Fetch quotes from Tradernet (primary source)
	if reporter != nil {
		reporter.Report(2, totalSteps, "Fetching prices from Tradernet")
	}

	if s.brokerClient == nil {
		return 0, fmt.Errorf("broker client not available")
	}

	quotes, err := s.brokerClient.GetQuotes(symbols)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch quotes from Tradernet: %w", err)
	}

	// Step 3: Validate and update prices
	if reporter != nil {
		reporter.Report(3, totalSteps, "Updating positions")
	}

	updated := 0
	now := time.Now().Unix()

	for symbol, quote := range quotes {
		if quote == nil || quote.Price <= 0 {
			s.log.Debug().Str("symbol", symbol).Msg("No price data from Tradernet")
			continue
		}

		security, hasSecurity := symbolToSecurity[symbol]
		if !hasSecurity {
			s.log.Warn().Str("symbol", symbol).Msg("Security not found in mapping")
			continue
		}

		if security.ISIN == "" {
			s.log.Warn().Str("symbol", symbol).Msg("No ISIN found for symbol")
			continue
		}

		// Use Tradernet price directly (single data source)
		validatedPrice := quote.Price

		// Update positions table (using ISIN as PRIMARY KEY)
		result, err := s.db.Exec(`
			UPDATE positions
			SET current_price = ?,
				market_value_eur = quantity * ? / currency_rate,
				last_updated = ?
			WHERE isin = ?
		`, validatedPrice, validatedPrice, now, security.ISIN)

		if err != nil {
			s.log.Error().Err(err).Str("symbol", symbol).Str("isin", security.ISIN).Msg("Failed to update position price")
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			updated++
		}
	}

	s.log.Info().
		Int("total", len(securities)).
		Int("updated", updated).
		Msg("Price sync complete (Tradernet primary)")

	return updated, nil
}

// SyncAllHistoricalData syncs historical price data for all active securities
// Faithful translation from Python: app/jobs/historical_data_sync.py -> sync_historical_data()
//
// This syncs historical prices for all securities (not just those needing sync).
func (s *SyncService) SyncAllHistoricalData() (int, int, error) {
	s.log.Info().Msg("Starting historical data sync for all securities")

	securities, err := s.securityRepo.GetAllActive()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get active securities: %w", err)
	}

	if len(securities) == 0 {
		s.log.Info().Msg("No securities to sync")
		return 0, 0, nil
	}

	s.log.Info().Int("count", len(securities)).Msg("Syncing historical data for all securities")

	processed := 0
	errors := 0

	for _, security := range securities {
		if s.historicalSync != nil {
			err := s.historicalSync.SyncHistoricalPrices(security.Symbol)
			if err != nil {
				s.log.Error().Err(err).Str("symbol", security.Symbol).Msg("Failed to sync historical prices")
				errors++
			} else {
				processed++
			}
		}
	}

	s.log.Info().
		Int("processed", processed).
		Int("errors", errors).
		Msg("Historical data sync complete")

	return processed, errors, nil
}

// RebuildUniverseFromPortfolio rebuilds the universe from current portfolio positions
// Faithful translation from Python: app/modules/system/api/status.py -> rebuild_universe_from_portfolio()
//
// This gets all securities from the portfolio and adds any missing ones to the universe.
func (s *SyncService) RebuildUniverseFromPortfolio() (int, error) {
	s.log.Info().Msg("Rebuilding universe from portfolio")

	// Step 1: Check tradernet client availability
	if s.brokerClient == nil {
		return 0, fmt.Errorf("tradernet client not available")
	}

	// Step 2: Fetch current portfolio positions
	positions, err := s.brokerClient.GetPortfolio()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch portfolio: %w", err)
	}

	s.log.Info().Int("positions", len(positions)).Msg("Fetched portfolio positions")

	// Step 3: Identify missing securities
	missingSymbols := []string{}
	for _, pos := range positions {
		existing, err := s.securityRepo.GetBySymbol(pos.Symbol)
		if err != nil {
			s.log.Error().Err(err).Str("symbol", pos.Symbol).Msg("Failed to check security")
			continue
		}
		if existing == nil {
			missingSymbols = append(missingSymbols, pos.Symbol)
		}
	}

	if len(missingSymbols) == 0 {
		s.log.Info().Msg("All portfolio securities are already in universe")
		return 0, nil
	}

	// Step 4: Add missing securities using SecuritySetupService
	if s.setupService == nil {
		return 0, fmt.Errorf("setup service not available")
	}

	added := 0
	failed := 0

	for _, symbol := range missingSymbols {
		s.log.Info().Str("symbol", symbol).Msg("Adding missing security to universe")

		// Use AddSecurityByIdentifier (handles full data pipeline)
		// Note: User-configurable fields are set via security_overrides after creation
		security, err := s.setupService.AddSecurityByIdentifier(symbol)

		if err != nil {
			s.log.Error().Err(err).Str("symbol", symbol).Msg("Failed to add security")
			failed++
			continue
		}

		s.log.Info().
			Str("symbol", security.Symbol).
			Str("isin", security.ISIN).
			Msg("Successfully added security to universe")
		added++
	}

	s.log.Info().
		Int("added", added).
		Int("failed", failed).
		Int("total_missing", len(missingSymbols)).
		Msg("Universe rebuild complete")

	return added, nil
}

// SyncPricesForSymbols syncs prices for a filtered set of symbols
// Uses Tradernet as the sole source
// The symbolMap parameter (tradernet_symbol -> optional_override) is converted to a symbol list for Tradernet
func (s *SyncService) SyncPricesForSymbols(symbolMap map[string]*string) (int, error) {
	s.log.Info().Int("symbols", len(symbolMap)).Msg("Starting filtered price sync (Tradernet primary)")

	if len(symbolMap) == 0 {
		s.log.Info().Msg("No symbols to sync prices for")
		return 0, nil
	}

	if s.brokerClient == nil {
		return 0, fmt.Errorf("broker client not available")
	}

	// Build symbol list and mappings
	symbols := make([]string, 0, len(symbolMap))
	symbolToISIN := make(map[string]string)

	for symbol := range symbolMap {
		symbols = append(symbols, symbol)
		// Lookup ISIN from securities table (if securityRepo is available)
		if s.securityRepo != nil {
			security, err := s.securityRepo.GetBySymbol(symbol)
			if err == nil && security != nil && security.ISIN != "" {
				symbolToISIN[symbol] = security.ISIN
				continue
			}
		}
		// Fallback: If securityRepo is nil or lookup fails, use symbol as ISIN (for backward compatibility in tests)
		symbolToISIN[symbol] = symbol
	}

	// Fetch quotes from Tradernet (primary source)
	quotes, err := s.brokerClient.GetQuotes(symbols)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch quotes from Tradernet: %w", err)
	}

	// Update position prices in state.db (using ISIN)
	updated := 0
	now := time.Now().Unix()

	for symbol, quote := range quotes {
		if quote == nil || quote.Price <= 0 {
			s.log.Debug().Str("symbol", symbol).Msg("No price data from Tradernet")
			continue
		}

		// Lookup ISIN for this symbol
		isin, hasISIN := symbolToISIN[symbol]
		if !hasISIN || isin == "" {
			s.log.Warn().Str("symbol", symbol).Msg("No ISIN found for symbol, skipping position update")
			continue
		}

		// Use Tradernet price directly (single data source)
		validatedPrice := quote.Price

		// Update positions table (using ISIN as PRIMARY KEY)
		result, err := s.db.Exec(`
			UPDATE positions
			SET current_price = ?,
				market_value_eur = quantity * ? / currency_rate,
				last_updated = ?
			WHERE isin = ?
		`, validatedPrice, validatedPrice, now, isin)

		if err != nil {
			s.log.Error().Err(err).Str("symbol", symbol).Str("isin", isin).Msg("Failed to update position price")
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			updated++
		}
	}

	s.log.Info().
		Int("requested", len(symbolMap)).
		Int("updated", updated).
		Msg("Filtered price sync complete (Tradernet primary)")

	return updated, nil
}
