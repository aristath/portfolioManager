package universe

import (
	"fmt"
	"time"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/rs/zerolog"
)

// HistoricalSyncService handles synchronization of historical price data from Tradernet
// Refactored from Yahoo to Tradernet as the single source of truth
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
// Used by HistoricalSyncService to enable testing with mocks
type HistoryDBInterface interface {
	HasMonthlyData(isin string) (bool, error)
	SyncHistoricalPrices(isin string, prices []DailyPrice) error
}

// NewHistoricalSyncService creates a new historical sync service
// Uses Tradernet (brokerClient) as the single source of truth for historical data
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

// SyncHistoricalPrices synchronizes historical price data for a security
// Uses Tradernet (getHloc API) as the single source of truth
//
// Workflow:
// 1. Get security metadata from database
// 2. Check if monthly_prices has data (determines date range)
// 3. Fetch from Tradernet (10y initial seed, 1y ongoing updates)
// 4. Insert/replace daily_prices in transaction
// 5. Aggregate to monthly_prices
// 6. Rate limit delay
func (s *HistoricalSyncService) SyncHistoricalPrices(symbol string) error {
	s.log.Info().Str("symbol", symbol).Msg("Starting historical price sync (Tradernet)")

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
	var dateFrom time.Time
	now := time.Now()
	if hasMonthly {
		dateFrom = now.AddDate(-1, 0, 0) // 1 year
		s.log.Debug().Str("symbol", symbol).Msg("Monthly data exists, fetching 1-year update")
	} else {
		dateFrom = now.AddDate(-10, 0, 0) // 10 years initial seed
		s.log.Info().Str("symbol", symbol).Msg("No monthly data found, performing 10-year initial seed")
	}

	// Check broker client availability
	if s.brokerClient == nil {
		return fmt.Errorf("broker client not available")
	}

	// Fetch historical prices from Tradernet (primary source)
	// Use security's Tradernet symbol, daily timeframe (86400 seconds)
	tradernetSymbol := security.Symbol
	ohlcData, err := s.brokerClient.GetHistoricalPrices(tradernetSymbol, dateFrom.Unix(), now.Unix(), 86400)
	if err != nil {
		return fmt.Errorf("failed to fetch historical prices from Tradernet: %w", err)
	}

	if len(ohlcData) == 0 {
		s.log.Warn().Str("symbol", symbol).Msg("No price data from Tradernet")
		return nil
	}

	s.log.Info().
		Str("symbol", symbol).
		Int("count", len(ohlcData)).
		Msg("Fetched historical prices from Tradernet")

	// Convert Tradernet BrokerOHLCV to HistoryDB DailyPrice format
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

	// Write to history database (transaction, daily + monthly aggregation)
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
		Msg("Historical price sync complete (Tradernet)")

	return nil
}
