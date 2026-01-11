// Package services provides core business services shared across multiple modules.
//
// Shared services handle cross-cutting business functionality that doesn't belong
// to a specific domain but is essential for the system to function.
//
// See services/README.md for architecture documentation and usage patterns.
package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/rs/zerolog"
)

// ConversionStep represents a single step in a currency conversion path
type ConversionStep struct {
	FromCurrency string
	ToCurrency   string
	Symbol       string
	Action       string // "BUY" or "SELL"
}

// ExchangeRate holds exchange rate information
type ExchangeRate struct {
	FromCurrency string
	ToCurrency   string
	Rate         float64
	Bid          float64
	Ask          float64
	Symbol       string
}

// cacheEntry holds a cached exchange rate with expiration time
type cacheEntry struct {
	rate      float64
	expiresAt time.Time
}

const (
	// cacheTTL is how long FX rates are cached in memory (5 minutes)
	// FX rates change relatively slowly, so 5 minutes is a good balance
	// between freshness and API call reduction
	cacheTTL = 5 * time.Minute
)

// CurrencyExchangeService handles currency conversions via broker FX pairs
//
// Supports direct conversions between EUR, USD, HKD, and GBP.
// For pairs without direct instruments (GBP<->HKD), routes via EUR.
//
// Uses in-memory caching to reduce API calls and prevent rate limiting.
//
// Faithful translation from Python: app/shared/services/currency_exchange_service.py
type CurrencyExchangeService struct {
	brokerClient domain.BrokerClient
	log          zerolog.Logger
	cache        map[string]cacheEntry // key: "FROM:TO"
	cacheMu      sync.RWMutex          // Protects cache map
}

// DirectPairs contains direct currency pairs available on broker
// Format: (from_currency, to_currency) -> (symbol, action)
var DirectPairs = map[string]struct {
	Symbol string
	Action string
}{
	// EUR <-> USD (ITS_MONEY market)
	"EUR:USD": {"EURUSD_T0.ITS", "BUY"},  // Fixed: was SELL
	"USD:EUR": {"EURUSD_T0.ITS", "SELL"}, // Fixed: was BUY
	// EUR <-> GBP (ITS_MONEY market)
	"EUR:GBP": {"EURGBP_T0.ITS", "BUY"},  // Fixed: was SELL
	"GBP:EUR": {"EURGBP_T0.ITS", "SELL"}, // Fixed: was BUY
	// GBP <-> USD (ITS_MONEY market)
	"GBP:USD": {"GBPUSD_T0.ITS", "BUY"},  // Fixed: was SELL
	"USD:GBP": {"GBPUSD_T0.ITS", "SELL"}, // Fixed: was BUY
	// HKD <-> EUR (MONEY market, EXANTE)
	"EUR:HKD": {"HKD/EUR", "BUY"},
	"HKD:EUR": {"HKD/EUR", "SELL"},
	// HKD <-> USD (MONEY market, EXANTE)
	"USD:HKD": {"HKD/USD", "BUY"},
	"HKD:USD": {"HKD/USD", "SELL"},
}


// NewCurrencyExchangeService creates a new currency exchange service
func NewCurrencyExchangeService(brokerClient domain.BrokerClient, log zerolog.Logger) *CurrencyExchangeService {
	return &CurrencyExchangeService{
		brokerClient: brokerClient,
		log:          log.With().Str("service", "currency_exchange").Logger(),
		cache:        make(map[string]cacheEntry),
	}
}

// getCacheKey generates a cache key from currency pair
func (s *CurrencyExchangeService) getCacheKey(fromCurrency, toCurrency string) string {
	return fromCurrency + ":" + toCurrency
}

// getFromCache retrieves a rate from cache if valid (not expired)
func (s *CurrencyExchangeService) getFromCache(fromCurrency, toCurrency string) (float64, bool) {
	key := s.getCacheKey(fromCurrency, toCurrency)

	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	// Handle nil cache (e.g., in tests that create struct directly)
	if s.cache == nil {
		return 0, false
	}

	entry, exists := s.cache[key]
	if !exists {
		return 0, false
	}

	// Check if cache entry is expired
	if time.Now().After(entry.expiresAt) {
		return 0, false
	}

	return entry.rate, true
}

// storeInCache stores a rate in cache with expiration time
func (s *CurrencyExchangeService) storeInCache(fromCurrency, toCurrency string, rate float64) {
	key := s.getCacheKey(fromCurrency, toCurrency)

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	// Initialize cache if nil (e.g., in tests that create struct directly)
	if s.cache == nil {
		s.cache = make(map[string]cacheEntry)
	}

	s.cache[key] = cacheEntry{
		rate:      rate,
		expiresAt: time.Now().Add(cacheTTL),
	}
}

// GetConversionPath returns the conversion path between two currencies
func (s *CurrencyExchangeService) GetConversionPath(fromCurrency, toCurrency string) ([]ConversionStep, error) {
	if fromCurrency == toCurrency {
		return []ConversionStep{}, nil
	}

	// Check for direct pair
	pairKey := fromCurrency + ":" + toCurrency
	if pair, ok := DirectPairs[pairKey]; ok {
		return []ConversionStep{
			{
				FromCurrency: fromCurrency,
				ToCurrency:   toCurrency,
				Symbol:       pair.Symbol,
				Action:       pair.Action,
			},
		}, nil
	}

	// GBP <-> HKD requires routing via EUR
	if (fromCurrency == "GBP" && toCurrency == "HKD") || (fromCurrency == "HKD" && toCurrency == "GBP") {
		steps := []ConversionStep{}

		// Step 1: from_currency -> EUR
		step1Key := fromCurrency + ":EUR"
		if pair1, ok := DirectPairs[step1Key]; ok {
			steps = append(steps, ConversionStep{
				FromCurrency: fromCurrency,
				ToCurrency:   "EUR",
				Symbol:       pair1.Symbol,
				Action:       pair1.Action,
			})
		}

		// Step 2: EUR -> to_currency
		step2Key := "EUR:" + toCurrency
		if pair2, ok := DirectPairs[step2Key]; ok {
			steps = append(steps, ConversionStep{
				FromCurrency: "EUR",
				ToCurrency:   toCurrency,
				Symbol:       pair2.Symbol,
				Action:       pair2.Action,
			})
		}

		if len(steps) == 2 {
			return steps, nil
		}
	}

	return nil, fmt.Errorf("no conversion path from %s to %s", fromCurrency, toCurrency)
}

// GetRate returns the current exchange rate between two currencies
//
// Returns how many units of toCurrency per 1 fromCurrency
//
// Uses in-memory cache to reduce API calls. Cache entries expire after 5 minutes.
func (s *CurrencyExchangeService) GetRate(fromCurrency, toCurrency string) (float64, error) {
	if fromCurrency == toCurrency {
		return 1.0, nil
	}

	// Check cache first
	if cachedRate, ok := s.getFromCache(fromCurrency, toCurrency); ok {
		s.log.Debug().
			Str("from", fromCurrency).
			Str("to", toCurrency).
			Float64("rate", cachedRate).
			Msg("Using cached FX rate")
		return cachedRate, nil
	}

	if !s.brokerClient.IsConnected() {
		return 0, fmt.Errorf("tradernet not connected")
	}

	// Check if direct pair exists
	_, hasDirectPair := DirectPairs[fromCurrency+":"+toCurrency]
	_, hasInversePair := DirectPairs[toCurrency+":"+fromCurrency]

	if !hasDirectPair && !hasInversePair {
		// Try to get rate via path (multi-step conversion)
		return s.getRateViaPath(fromCurrency, toCurrency)
	}

	// Get FX rates using fromCurrency as base
	rates, err := s.brokerClient.GetFXRates(fromCurrency, []string{toCurrency})
	if err != nil {
		s.log.Warn().Err(err).Str("from", fromCurrency).Str("to", toCurrency).Msg("Failed to get FX rates")
		return 0, fmt.Errorf("failed to get FX rates: %w", err)
	}

	rate, ok := rates[toCurrency]
	if !ok {
		return 0, fmt.Errorf("rate not found for %s to %s", fromCurrency, toCurrency)
	}

	if rate <= 0 {
		return 0, fmt.Errorf("invalid rate: %f", rate)
	}

	// Store in cache for future requests
	s.storeInCache(fromCurrency, toCurrency, rate)

	return rate, nil
}

// getRateViaPath gets exchange rate via conversion path
func (s *CurrencyExchangeService) getRateViaPath(fromCurrency, toCurrency string) (float64, error) {
	path, err := s.GetConversionPath(fromCurrency, toCurrency)
	if err != nil {
		return 0, err
	}

	if len(path) == 1 {
		// Single step: use GetRate to benefit from caching
		return s.GetRate(path[0].FromCurrency, path[0].ToCurrency)
	} else if len(path) == 2 {
		// Multi-step: multiply rates (both calls to GetRate will use cache)
		rate1, err := s.GetRate(path[0].FromCurrency, path[0].ToCurrency)
		if err != nil {
			return 0, err
		}
		rate2, err := s.GetRate(path[1].FromCurrency, path[1].ToCurrency)
		if err != nil {
			return 0, err
		}
		return rate1 * rate2, nil
	}

	return 0, fmt.Errorf("no conversion path found")
}

// Exchange executes a currency exchange
func (s *CurrencyExchangeService) Exchange(fromCurrency, toCurrency string, amount float64) error {
	if !s.validateExchangeRequest(fromCurrency, toCurrency, amount) {
		return fmt.Errorf("invalid exchange request")
	}

	path, err := s.GetConversionPath(fromCurrency, toCurrency)
	if err != nil {
		return err
	}

	if len(path) == 0 {
		return fmt.Errorf("no conversion path")
	} else if len(path) == 1 {
		return s.executeStep(path[0], amount)
	} else {
		return s.executeMultiStepConversion(path, amount)
	}
}

// validateExchangeRequest validates exchange request parameters
func (s *CurrencyExchangeService) validateExchangeRequest(fromCurrency, toCurrency string, amount float64) bool {
	if fromCurrency == toCurrency {
		s.log.Warn().Str("currency", fromCurrency).Msg("Same currency exchange requested")
		return false
	}

	if amount <= 0 {
		s.log.Error().Float64("amount", amount).Msg("Invalid exchange amount")
		return false
	}

	if !s.brokerClient.IsConnected() {
		s.log.Error().Msg("Tradernet not connected for exchange")
		return false
	}

	return true
}

// executeMultiStepConversion executes multi-step currency conversion
func (s *CurrencyExchangeService) executeMultiStepConversion(path []ConversionStep, amount float64) error {
	currentAmount := amount

	for _, step := range path {
		if err := s.executeStep(step, currentAmount); err != nil {
			s.log.Error().
				Err(err).
				Str("from", step.FromCurrency).
				Str("to", step.ToCurrency).
				Msg("Failed at conversion step")
			return err
		}

		// Update amount for next step
		rate, err := s.GetRate(step.FromCurrency, step.ToCurrency)
		if err == nil {
			currentAmount = currentAmount * rate
		}
	}

	return nil
}

// executeStep executes a single conversion step
func (s *CurrencyExchangeService) executeStep(step ConversionStep, amount float64) error {
	s.log.Info().
		Str("action", step.Action).
		Str("symbol", step.Symbol).
		Float64("amount", amount).
		Str("from", step.FromCurrency).
		Str("to", step.ToCurrency).
		Msg("Executing FX conversion")

	// Currency exchange uses market orders (limitPrice = 0.0)
	// FX pairs are highly liquid with tight spreads, no limit needed
	_, err := s.brokerClient.PlaceOrder(step.Symbol, step.Action, amount, 0.0)
	return err
}

// EnsureBalance ensures we have at least minAmount in the target currency
//
// If insufficient balance, converts from sourceCurrency.
func (s *CurrencyExchangeService) EnsureBalance(currency string, minAmount float64, sourceCurrency string) (bool, error) {
	if currency == sourceCurrency {
		return true, nil
	}

	if !s.brokerClient.IsConnected() {
		return false, fmt.Errorf("tradernet not connected")
	}

	// Get balances
	currentBalance, sourceBalance, err := s.getBalances(currency, sourceCurrency)
	if err != nil {
		return false, err
	}

	// Block conversion if source balance is negative
	if sourceBalance < 0 {
		s.log.Error().
			Str("source_currency", sourceCurrency).
			Float64("source_balance", sourceBalance).
			Msg("Cannot ensure balance: source currency has negative balance")
		return false, fmt.Errorf("source currency %s has negative balance: %.2f", sourceCurrency, sourceBalance)
	}

	if currentBalance >= minAmount {
		s.log.Info().
			Str("currency", currency).
			Float64("balance", currentBalance).
			Float64("min_amount", minAmount).
			Msg("Sufficient balance")
		return true, nil
	}

	needed := minAmount - currentBalance
	return s.convertForBalance(currency, sourceCurrency, needed, sourceBalance)
}

// getBalances returns current and source currency balances
func (s *CurrencyExchangeService) getBalances(currency, sourceCurrency string) (float64, float64, error) {
	balances, err := s.brokerClient.GetCashBalances()
	if err != nil {
		return 0, 0, err
	}

	var currentBalance, sourceBalance float64

	for _, bal := range balances {
		if bal.Currency == currency {
			currentBalance = bal.Amount
			if currentBalance < 0 {
				s.log.Warn().
					Str("currency", currency).
					Float64("balance", currentBalance).
					Msg("Negative balance detected")
			}
		} else if bal.Currency == sourceCurrency {
			sourceBalance = bal.Amount
			if sourceBalance < 0 {
				s.log.Warn().
					Str("currency", sourceCurrency).
					Float64("balance", sourceBalance).
					Msg("Negative balance detected")
			}
		}
	}

	return currentBalance, sourceBalance, nil
}

// convertForBalance converts source currency to target currency to meet balance requirement
func (s *CurrencyExchangeService) convertForBalance(currency, sourceCurrency string, needed, sourceBalance float64) (bool, error) {
	// Safety check: block conversion if source balance is negative
	if sourceBalance < 0 {
		s.log.Error().
			Str("source_currency", sourceCurrency).
			Float64("source_balance", sourceBalance).
			Msg("Cannot convert: source balance is negative")
		return false, fmt.Errorf("source balance is negative")
	}

	// Add 2% buffer
	neededWithBuffer := needed * 1.02

	rate, err := s.GetRate(sourceCurrency, currency)
	if err != nil {
		s.log.Error().Err(err).Msgf("Could not get rate for %s/%s", sourceCurrency, currency)
		return false, err
	}

	sourceAmountNeeded := neededWithBuffer / rate

	if sourceBalance < sourceAmountNeeded {
		s.log.Warn().
			Str("source_currency", sourceCurrency).
			Float64("need", sourceAmountNeeded).
			Float64("have", sourceBalance).
			Msg("Insufficient source currency to convert")
		return false, fmt.Errorf("insufficient %s to convert", sourceCurrency)
	}

	s.log.Info().
		Float64("amount", sourceAmountNeeded).
		Str("from", sourceCurrency).
		Str("to", currency).
		Float64("needed", needed).
		Msg("Converting currency")

	if err := s.Exchange(sourceCurrency, currency, sourceAmountNeeded); err != nil {
		s.log.Error().Err(err).Msgf("Failed to convert %s to %s", sourceCurrency, currency)
		return false, err
	}

	s.log.Info().Msg("Currency exchange completed")
	return true, nil
}

// GetAvailableCurrencies returns list of currencies that can be exchanged
func (s *CurrencyExchangeService) GetAvailableCurrencies() []string {
	currencies := make(map[string]bool)
	for key := range DirectPairs {
		// Split "FROM:TO" into currencies
		from := key[:3]
		to := key[4:]
		currencies[from] = true
		currencies[to] = true
	}

	result := make([]string, 0, len(currencies))
	for curr := range currencies {
		result = append(result, curr)
	}
	return result
}
