package services

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Mock Dependencies
// ============================================================================

// MockCurrencyExchangeService for testing
type MockCurrencyExchangeService struct {
	rates       map[string]float64
	shouldError bool
}

func (m *MockCurrencyExchangeService) GetRate(fromCurrency, toCurrency string) (float64, error) {
	if m.shouldError {
		return 0, assert.AnError
	}
	key := fromCurrency + ":" + toCurrency
	if rate, ok := m.rates[key]; ok {
		return rate, nil
	}
	return 0, assert.AnError
}

func (m *MockCurrencyExchangeService) EnsureBalance(currency string, minAmount float64, sourceCurrency string) (bool, error) {
	return true, nil
}

// MockHistoryDB for testing
type MockHistoryDB struct {
	rates       map[string]float64
	shouldError bool
}

// MockExchangeRate matches what HistoryDB.GetLatestExchangeRate returns
type MockExchangeRate struct {
	Rate float64
}

func (m *MockHistoryDB) UpsertExchangeRate(from, to string, rate float64) error {
	if m.shouldError {
		return assert.AnError
	}
	if m.rates == nil {
		m.rates = make(map[string]float64)
	}
	m.rates[from+":"+to] = rate
	return nil
}

// MockSettingsService for testing
type MockSettingsService struct {
	settings map[string]interface{}
}

func (m *MockSettingsService) Get(key string) (interface{}, error) {
	if m.settings == nil {
		return nil, assert.AnError
	}
	if val, ok := m.settings[key]; ok {
		return val, nil
	}
	return nil, assert.AnError
}

func (m *MockSettingsService) Set(key string, value interface{}) (bool, error) {
	if m.settings == nil {
		m.settings = make(map[string]interface{})
	}
	m.settings[key] = value
	return true, nil
}

// ============================================================================
// ExchangeRateCacheService.GetRate Tests
// ============================================================================

func TestExchangeRateCacheService_GetRate_SameCurrency(t *testing.T) {
	log := zerolog.Nop()
	service := NewExchangeRateCacheService(nil, nil, nil, log)

	rate, err := service.GetRate("EUR", "EUR")

	assert.NoError(t, err)
	assert.Equal(t, 1.0, rate)
}

func TestExchangeRateCacheService_GetRate_FromTradernet(t *testing.T) {
	mockExchange := &MockCurrencyExchangeService{
		rates: map[string]float64{
			"EUR:USD": 1.10,
		},
	}
	log := zerolog.Nop()
	service := NewExchangeRateCacheService(mockExchange, nil, nil, log)

	rate, err := service.GetRate("EUR", "USD")

	assert.NoError(t, err)
	assert.InDelta(t, 1.10, rate, 0.001)
}

func TestExchangeRateCacheService_GetRate_NoSourcesAvailable(t *testing.T) {
	// No Tradernet, no DB cache - should return error (no hardcoded fallback)
	log := zerolog.Nop()
	service := NewExchangeRateCacheService(nil, nil, nil, log)

	rate, err := service.GetRate("EUR", "USD")

	assert.Error(t, err)
	assert.Equal(t, 0.0, rate)
	assert.Contains(t, err.Error(), "no rate available")
}

func TestExchangeRateCacheService_GetRate_UnsupportedPair(t *testing.T) {
	// Mock exchange returns error for unsupported pair
	mockExchange := &MockCurrencyExchangeService{
		rates: map[string]float64{}, // no rates defined
	}
	log := zerolog.Nop()
	service := NewExchangeRateCacheService(mockExchange, nil, nil, log)

	rate, err := service.GetRate("EUR", "CNY")

	assert.Error(t, err)
	assert.Equal(t, 0.0, rate)
}

// ============================================================================
// SyncRates Tests
// ============================================================================

func TestExchangeRateCacheService_SyncRates_AllSuccess(t *testing.T) {
	mockExchange := &MockCurrencyExchangeService{
		rates: map[string]float64{
			"EUR:USD": 1.10,
			"EUR:GBP": 0.85,
			"EUR:HKD": 8.50,
			"USD:EUR": 0.91,
			"USD:GBP": 0.77,
			"USD:HKD": 7.80,
			"GBP:EUR": 1.18,
			"GBP:USD": 1.30,
			"GBP:HKD": 10.00,
			"HKD:EUR": 0.12,
			"HKD:USD": 0.13,
			"HKD:GBP": 0.10,
		},
	}
	log := zerolog.Nop()
	service := NewExchangeRateCacheService(mockExchange, nil, nil, log)

	err := service.SyncRates()

	assert.NoError(t, err)
}

func TestExchangeRateCacheService_SyncRates_AllFail(t *testing.T) {
	// No mock exchange, no DB - all rates will fail
	log := zerolog.Nop()
	service := NewExchangeRateCacheService(nil, nil, nil, log)

	err := service.SyncRates()

	// Should return error when all rate fetches fail
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all rate fetches failed")
}
