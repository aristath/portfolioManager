package universe

import (
	"fmt"
	"time"

	"github.com/aristath/sentinel/internal/clients/yahoo"
	"github.com/rs/zerolog"
)

// HistoricalSyncService handles synchronization of historical price data from Yahoo Finance
// Faithful translation from Python: app/jobs/securities_data_sync.py -> _sync_historical_for_symbol()
type HistoricalSyncService struct {
	yahooClient    yahoo.FullClientInterface
	securityRepo   *SecurityRepository
	historyDB      *HistoryDB
	rateLimitDelay time.Duration // External API rate limit delay
	log            zerolog.Logger
}

// NewHistoricalSyncService creates a new historical sync service
func NewHistoricalSyncService(
	yahooClient yahoo.FullClientInterface,
	securityRepo *SecurityRepository,
	historyDB *HistoryDB,
	rateLimitDelay time.Duration,
	log zerolog.Logger,
) *HistoricalSyncService {
	return &HistoricalSyncService{
		yahooClient:    yahooClient,
		securityRepo:   securityRepo,
		historyDB:      historyDB,
		rateLimitDelay: rateLimitDelay,
		log:            log.With().Str("service", "historical_sync").Logger(),
	}
}

// SyncHistoricalPrices synchronizes historical price data for a security
// Faithful translation from Python: app/jobs/securities_data_sync.py -> _sync_historical_for_symbol()
//
// Workflow:
// 1. Get security's yahoo_symbol from database
// 2. Check if monthly_prices has data (determines period)
// 3. Fetch from Yahoo Finance (10y initial seed, 1y ongoing updates)
// 4. Insert/replace daily_prices in transaction
// 5. Aggregate to monthly_prices
// 6. Rate limit delay
func (s *HistoricalSyncService) SyncHistoricalPrices(symbol string) error {
	s.log.Info().Str("symbol", symbol).Msg("Starting historical price sync")

	// Get security metadata
	security, err := s.securityRepo.GetBySymbol(symbol)
	if err != nil {
		return fmt.Errorf("failed to get security: %w", err)
	}
	if security == nil {
		return fmt.Errorf("security not found: %s", symbol)
	}

	// Extract ISIN - required for history database operations
	if security.ISIN == "" {
		return fmt.Errorf("security %s has no ISIN, cannot sync historical prices", symbol)
	}
	isin := security.ISIN

	// Check if we have monthly data (indicates initial seeding was done)
	hasMonthly, err := s.historyDB.HasMonthlyData(isin)
	if err != nil {
		s.log.Warn().Err(err).Str("symbol", symbol).Str("isin", isin).Msg("Failed to check monthly data, assuming no data")
		hasMonthly = false
	}

	// Initial seed: 10 years for CAGR calculations
	// Ongoing updates: 1 year for daily charts
	period := "1y"
	if !hasMonthly {
		period = "10y"
		s.log.Info().Str("symbol", symbol).Msg("No monthly data found, performing 10-year initial seed")
	}

	// Fetch historical prices from Yahoo Finance
	yahooSymbolPtr := &security.YahooSymbol
	if security.YahooSymbol == "" {
		yahooSymbolPtr = nil
	}

	// If yahoo_symbol is an ISIN, try to look it up and update the database
	if yahooSymbolPtr != nil && IsISIN(*yahooSymbolPtr) {
		s.log.Info().
			Str("symbol", symbol).
			Str("isin", *yahooSymbolPtr).
			Msg("yahoo_symbol is an ISIN, attempting to look up ticker symbol")

		ticker, err := s.yahooClient.LookupTickerFromISIN(*yahooSymbolPtr)
		if err != nil {
			s.log.Warn().
				Err(err).
				Str("symbol", symbol).
				Str("isin", *yahooSymbolPtr).
				Msg("Failed to look up ticker from ISIN, falling back to Tradernet conversion")
			// Fall through to use Tradernet conversion
			yahooSymbolPtr = nil
		} else {
			// Update the database with the found ticker symbol
			// Lookup ISIN from symbol
			security, err := s.securityRepo.GetBySymbol(symbol)
			if err != nil || security == nil || security.ISIN == "" {
				s.log.Warn().Str("symbol", symbol).Msg("Failed to lookup ISIN, skipping update")
			} else {
				err = s.securityRepo.Update(security.ISIN, map[string]interface{}{
					"yahoo_symbol": ticker,
				})
				if err != nil {
					s.log.Warn().
						Err(err).
						Str("symbol", symbol).
						Str("ticker", ticker).
						Msg("Failed to update yahoo_symbol in database, but will use it for this request")
				} else {
					s.log.Info().
						Str("symbol", symbol).
						Str("isin", security.ISIN).
						Str("ticker", ticker).
						Msg("Updated yahoo_symbol from ISIN to ticker symbol")
				}
			}
			// Use the found ticker symbol
			yahooSymbolPtr = &ticker
		}
	}

	// Use security's Tradernet symbol for API call (not the parameter, which might be different)
	tradernetSymbol := security.Symbol
	ohlcData, err := s.yahooClient.GetHistoricalPrices(tradernetSymbol, yahooSymbolPtr, period)
	if err != nil {
		return fmt.Errorf("failed to fetch historical prices from Yahoo: %w", err)
	}

	if len(ohlcData) == 0 {
		s.log.Warn().Str("symbol", symbol).Msg("No price data from Yahoo Finance")
		return nil
	}

	s.log.Info().
		Str("symbol", symbol).
		Str("period", period).
		Int("count", len(ohlcData)).
		Msg("Fetched historical prices from Yahoo Finance")

	// Convert Yahoo HistoricalPrice to HistoryDB DailyPrice format
	dailyPrices := make([]DailyPrice, len(ohlcData))
	for i, yPrice := range ohlcData {
		volume := yPrice.Volume
		dailyPrices[i] = DailyPrice{
			Date:   yPrice.Date.Format("2006-01-02"),
			Open:   yPrice.Open,
			High:   yPrice.High,
			Low:    yPrice.Low,
			Close:  yPrice.Close,
			Volume: &volume,
		}
	}

	// Write to history database (transaction, daily + monthly aggregation)
	// Use ISIN instead of Tradernet symbol
	err = s.historyDB.SyncHistoricalPrices(isin, dailyPrices)
	if err != nil {
		return fmt.Errorf("failed to sync historical prices to database: %w", err)
	}

	// Rate limit delay to avoid overwhelming Yahoo Finance
	if s.rateLimitDelay > 0 {
		s.log.Debug().
			Str("symbol", symbol).
			Dur("delay", s.rateLimitDelay).
			Msg("Rate limit delay")
		time.Sleep(s.rateLimitDelay)
	}

	s.log.Info().
		Str("symbol", symbol).
		Str("isin", isin).
		Int("count", len(dailyPrices)).
		Msg("Historical price sync complete")

	return nil
}
