package universe

import (
	"fmt"
	"time"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/rs/zerolog"
)

// HistoricalSyncService handles synchronization of historical price data.
// Uses Tradernet as the single source for historical price data.
// Stores raw data without validation - filtering happens on read via HistoryDB.
type HistoricalSyncService struct {
	brokerClient   domain.BrokerClient
	securityRepo   SecurityLookupInterface
	historyDB      HistoryDBInterface
	rateLimitDelay time.Duration // API rate limit delay
	log            zerolog.Logger
}

// SecurityLookupInterface defines minimal security lookup for HistoricalSyncService
// Used by HistoricalSyncService to enable testing with mocks
type SecurityLookupInterface interface {
	GetBySymbol(symbol string) (*Security, error)
}

// HistoryDBInterface defines the contract for history database operations
// Used by services that need access to filtered, cached price data
type HistoryDBInterface interface {
	// Price read operations (filtered and cached)
	GetDailyPrices(isin string, limit int) ([]DailyPrice, error)
	GetRecentPrices(isin string, days int) ([]DailyPrice, error)
	GetMonthlyPrices(isin string, limit int) ([]MonthlyPrice, error)
	HasMonthlyData(isin string) (bool, error)

	// Price write operations (raw data, invalidates cache)
	SyncHistoricalPrices(isin string, prices []DailyPrice) error
	DeletePricesForSecurity(isin string) error

	// Exchange rate operations
	UpsertExchangeRate(fromCurrency, toCurrency string, rate float64) error
	GetLatestExchangeRate(fromCurrency, toCurrency string) (*ExchangeRate, error)

	// Cache management
	InvalidateCache(isin string)
	InvalidateAllCaches()
}

// NewHistoricalSyncService creates a new historical sync service.
// Uses Tradernet (via broker client) as the single source of truth for historical prices.
// No validation on write - filtering happens on read in HistoryDB.
func NewHistoricalSyncService(
	brokerClient domain.BrokerClient,
	securityRepo SecurityLookupInterface,
	historyDB HistoryDBInterface,
	rateLimitDelay time.Duration,
	log zerolog.Logger,
) *HistoricalSyncService {
	return &HistoricalSyncService{
		brokerClient:   brokerClient,
		securityRepo:   securityRepo,
		historyDB:      historyDB,
		rateLimitDelay: rateLimitDelay,
		log:            log.With().Str("service", "historical_sync").Logger(),
	}
}

// SyncHistoricalPrices synchronizes historical price data for a security.
// Uses Tradernet as the single source of truth for historical prices.
// Stores raw data without validation - filtering happens on read.
//
// Workflow:
// 1. Get security metadata from database
// 2. Check if monthly_prices has data (determines date range)
// 3. Fetch historical data from Tradernet
// 4. Store raw data to history database
// 5. Rate limit delay
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
	var years int
	if hasMonthly {
		years = 1
		s.log.Debug().Str("symbol", symbol).Msg("Monthly data exists, fetching 1-year update")
	} else {
		years = 10
		s.log.Info().Str("symbol", symbol).Msg("No monthly data found, performing 10-year initial seed")
	}

	// Fetch historical prices from Tradernet
	if s.brokerClient == nil {
		return fmt.Errorf("broker client not available")
	}

	now := time.Now()
	dateFrom := now.AddDate(-years, 0, 0)

	ohlcData, err := s.brokerClient.GetHistoricalPrices(security.Symbol, dateFrom.Unix(), now.Unix(), 86400)
	if err != nil {
		return fmt.Errorf("failed to fetch historical prices from Tradernet: %w", err)
	}

	// Convert to DailyPrice format
	dailyPrices := make([]DailyPrice, len(ohlcData))
	for i, ohlc := range ohlcData {
		volume := ohlc.Volume
		dailyPrices[i] = DailyPrice{
			Date:   time.Unix(ohlc.Timestamp, 0).Format("2006-01-02"),
			Open:   ohlc.Open,
			High:   ohlc.High,
			Low:    ohlc.Low,
			Close:  ohlc.Close,
			Volume: &volume,
		}
	}

	if len(dailyPrices) == 0 {
		s.log.Warn().Str("symbol", symbol).Msg("No price data returned")
		return nil
	}

	s.log.Info().
		Str("symbol", symbol).
		Int("count", len(dailyPrices)).
		Msg("Fetched historical prices")

	// Write raw data to history database (no validation - filtering happens on read)
	err = s.historyDB.SyncHistoricalPrices(isin, dailyPrices)
	if err != nil {
		return fmt.Errorf("failed to sync historical prices to database: %w", err)
	}

	// Rate limit delay to avoid overwhelming the API
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
