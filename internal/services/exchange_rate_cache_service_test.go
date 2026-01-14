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

func TestExchangeRateCacheService_GetRate_HardcodedFallback(t *testing.T) {
	// No Tradernet, no DB cache - should use hardcoded
	log := zerolog.Nop()
	service := NewExchangeRateCacheService(nil, nil, nil, log)

	rate, err := service.GetRate("EUR", "USD")

	assert.NoError(t, err)
	assert.InDelta(t, 1.10, rate, 0.001)
}

func TestExchangeRateCacheService_GetRate_AllCurrencyPairs(t *testing.T) {
	log := zerolog.Nop()
	service := NewExchangeRateCacheService(nil, nil, nil, log)

	testCases := []struct {
		from, to string
	}{
		{"EUR", "USD"},
		{"USD", "EUR"},
		{"EUR", "GBP"},
		{"GBP", "EUR"},
		{"EUR", "HKD"},
		{"HKD", "EUR"},
		{"USD", "GBP"},
		{"GBP", "USD"},
		{"USD", "HKD"},
		{"HKD", "USD"},
		{"GBP", "HKD"},
		{"HKD", "GBP"},
	}

	for _, tc := range testCases {
		t.Run(tc.from+"_to_"+tc.to, func(t *testing.T) {
			rate, err := service.GetRate(tc.from, tc.to)
			assert.NoError(t, err)
			assert.Greater(t, rate, 0.0)
		})
	}
}

func TestExchangeRateCacheService_GetRate_UnsupportedPair(t *testing.T) {
	log := zerolog.Nop()
	service := NewExchangeRateCacheService(nil, nil, nil, log)

	// CNY is not in our hardcoded rates
	rate, err := service.GetRate("EUR", "CNY")

	assert.Error(t, err)
	assert.Equal(t, 0.0, rate)
}

// ============================================================================
// Hardcoded Rate Tests
// ============================================================================

func TestExchangeRateCacheService_HardcodedRates_Consistency(t *testing.T) {
	log := zerolog.Nop()
	service := NewExchangeRateCacheService(nil, nil, nil, log)

	// Test that inverse rates are approximately consistent
	eurUsd, _ := service.GetRate("EUR", "USD")
	usdEur, _ := service.GetRate("USD", "EUR")

	// EUR→USD and USD→EUR should be inverses (with some tolerance)
	product := eurUsd * usdEur
	assert.InDelta(t, 1.0, product, 0.05)
}

// ============================================================================
// SyncRates Tests
// ============================================================================

func TestExchangeRateCacheService_SyncRates_PartialSuccess(t *testing.T) {
	mockExchange := &MockCurrencyExchangeService{
		rates: map[string]float64{
			"EUR:USD": 1.10,
			"EUR:GBP": 0.85,
			// Other pairs will fallback to hardcoded
		},
	}
	log := zerolog.Nop()
	service := NewExchangeRateCacheService(mockExchange, nil, nil, log)

	err := service.SyncRates()

	// Should succeed with partial rates (hardcoded fallback)
	assert.NoError(t, err)
}
