package services

import (
	"fmt"
	"time"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/rs/zerolog"
)

// SettingsServiceInterface defines the contract for settings operations
type SettingsServiceInterface interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) (bool, error)
}

// ExchangeRateCacheService provides cached exchange rates with fallback
// Uses two-tier fallback: Tradernet API → Cached rates from history.db
type ExchangeRateCacheService struct {
	currencyExchangeService domain.CurrencyExchangeServiceInterface // Tradernet
	historyDB               universe.HistoryDBInterface             // Cache storage
	settingsService         SettingsServiceInterface                // Staleness config
	log                     zerolog.Logger
}

// NewExchangeRateCacheService creates a new exchange rate cache service
func NewExchangeRateCacheService(
	currencyExchangeService domain.CurrencyExchangeServiceInterface,
	historyDB universe.HistoryDBInterface,
	settingsService SettingsServiceInterface,
	log zerolog.Logger,
) *ExchangeRateCacheService {
	return &ExchangeRateCacheService{
		currencyExchangeService: currencyExchangeService,
		historyDB:               historyDB,
		settingsService:         settingsService,
		log:                     log.With().Str("service", "exchange_rate_cache").Logger(),
	}
}

// GetRate returns exchange rate with 2-tier fallback:
// 1. Try Tradernet (primary - uses broker's FX instruments)
// 2. Try cached rate from DB
func (s *ExchangeRateCacheService) GetRate(fromCurrency, toCurrency string) (float64, error) {
	if fromCurrency == toCurrency {
		return 1.0, nil
	}

	// Tier 1: Try Tradernet
	if s.currencyExchangeService != nil {
		rate, err := s.currencyExchangeService.GetRate(fromCurrency, toCurrency)
		if err == nil && rate > 0 {
			s.log.Debug().
				Str("from", fromCurrency).
				Str("to", toCurrency).
				Float64("rate", rate).
				Str("source", "tradernet").
				Msg("Got rate from Tradernet")

			// Cache the fresh rate
			if s.historyDB != nil {
				if err := s.historyDB.UpsertExchangeRate(fromCurrency, toCurrency, rate); err != nil {
					s.log.Warn().Err(err).Msg("Failed to cache rate")
				}
			}

			return rate, nil
		}
		s.log.Warn().Err(err).Msg("Tradernet rate fetch failed, trying cache")
	}

	// Tier 2: Try cached rate from DB
	rate, err := s.GetCachedRate(fromCurrency, toCurrency)
	if err == nil && rate > 0 {
		s.log.Warn().
			Str("from", fromCurrency).
			Str("to", toCurrency).
			Float64("rate", rate).
			Str("source", "cache").
			Msg("Using cached rate (API failed)")
		return rate, nil
	}

	return 0, fmt.Errorf("no rate available for %s/%s", fromCurrency, toCurrency)
}

// GetCachedRate fetches rate from database only
// Checks staleness and logs warning if rate is old
func (s *ExchangeRateCacheService) GetCachedRate(fromCurrency, toCurrency string) (float64, error) {
	if s.historyDB == nil {
		return 0, fmt.Errorf("history DB not available")
	}

	er, err := s.historyDB.GetLatestExchangeRate(fromCurrency, toCurrency)
	if err != nil {
		return 0, err
	}
	if er == nil {
		return 0, fmt.Errorf("no cached rate found")
	}

	// Check staleness
	maxAgeHours := 48.0
	if s.settingsService != nil {
		if val, err := s.settingsService.Get("max_exchange_rate_age_hours"); err == nil {
			if floatVal, ok := val.(float64); ok {
				maxAgeHours = floatVal
			}
		}
	}

	age := time.Since(er.Date)
	if age > time.Duration(maxAgeHours)*time.Hour {
		s.log.Warn().
			Str("from", fromCurrency).
			Str("to", toCurrency).
			Dur("age", age).
			Msg("Cached rate is stale but using anyway")
	}

	return er.Rate, nil
}

// SyncRates fetches and caches rates for all currency pairs
// Returns error only if ALL rate fetches fail
// Partial success is OK - logged as warnings
func (s *ExchangeRateCacheService) SyncRates() error {
	currencies := []string{"EUR", "USD", "GBP", "HKD"}

	errorCount := 0
	successCount := 0

	for _, from := range currencies {
		for _, to := range currencies {
			if from == to {
				continue
			}

			rate, err := s.GetRate(from, to)
			if err != nil {
				s.log.Error().
					Err(err).
					Str("from", from).
					Str("to", to).
					Msg("Failed to get rate")
				errorCount++
				continue
			}

			// GetRate already caches the rate in DB
			s.log.Debug().
				Str("from", from).
				Str("to", to).
				Float64("rate", rate).
				Msg("Synced exchange rate")

			successCount++
		}
	}

	s.log.Info().
		Int("success", successCount).
		Int("errors", errorCount).
		Msg("Exchange rate sync completed")

	if successCount == 0 {
		return fmt.Errorf("all rate fetches failed")
	}

	return nil // Partial success OK
}
