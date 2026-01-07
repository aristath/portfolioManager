package scheduler

import (
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// Mock CashManager for testing
type mockCashManager struct {
	balances map[string]float64
	err      error
}

func (m *mockCashManager) GetAllCashBalances() (map[string]float64, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.balances, nil
}

func (m *mockCashManager) UpdateCashPosition(currency string, balance float64) error {
	if m.balances == nil {
		m.balances = make(map[string]float64)
	}
	m.balances[currency] = balance
	return nil
}

func (m *mockCashManager) GetCashBalance(currency string) (float64, error) {
	if balance, ok := m.balances[currency]; ok {
		return balance, nil
	}
	return 0.0, nil
}

func TestBalanceAdapter_GetAllCurrencies(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockCashManager := &mockCashManager{
		balances: map[string]float64{
			"EUR": 1000.0,
			"USD": 500.0,
			"GBP": 200.0,
		},
	}

	adapter := NewBalanceAdapter(mockCashManager, log)

	currencies, err := adapter.GetAllCurrencies()

	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"EUR", "USD", "GBP"}, currencies)
}

func TestBalanceAdapter_GetAllCurrencies_Empty(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockCashManager := &mockCashManager{
		balances: map[string]float64{},
	}

	adapter := NewBalanceAdapter(mockCashManager, log)

	currencies, err := adapter.GetAllCurrencies()

	assert.NoError(t, err)
	assert.Empty(t, currencies)
}

func TestBalanceAdapter_GetAllCurrencies_Error(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockCashManager := &mockCashManager{
		err: errors.New("database connection failed"),
	}

	adapter := NewBalanceAdapter(mockCashManager, log)

	currencies, err := adapter.GetAllCurrencies()

	assert.Error(t, err)
	assert.Nil(t, currencies)
	assert.Contains(t, err.Error(), "database connection failed")
}

func TestBalanceAdapter_GetTotalByCurrency(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockCashManager := &mockCashManager{
		balances: map[string]float64{
			"EUR": 1000.0,
			"USD": 500.0,
		},
	}

	adapter := NewBalanceAdapter(mockCashManager, log)

	total, err := adapter.GetTotalByCurrency("EUR")
	assert.NoError(t, err)
	assert.Equal(t, 1000.0, total)

	total, err = adapter.GetTotalByCurrency("USD")
	assert.NoError(t, err)
	assert.Equal(t, 500.0, total)
}

func TestBalanceAdapter_GetTotalByCurrency_NotFound(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockCashManager := &mockCashManager{
		balances: map[string]float64{
			"EUR": 1000.0,
			"USD": 500.0,
		},
	}

	adapter := NewBalanceAdapter(mockCashManager, log)

	// Non-existent currency returns 0.0 (not an error)
	total, err := adapter.GetTotalByCurrency("GBP")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, total)
}

func TestBalanceAdapter_GetTotalByCurrency_Error(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockCashManager := &mockCashManager{
		err: errors.New("database connection failed"),
	}

	adapter := NewBalanceAdapter(mockCashManager, log)

	total, err := adapter.GetTotalByCurrency("EUR")

	assert.Error(t, err)
	assert.Equal(t, 0.0, total)
	assert.Contains(t, err.Error(), "database connection failed")
}

func TestBalanceAdapter_GetTotalByCurrency_NegativeBalance(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockCashManager := &mockCashManager{
		balances: map[string]float64{
			"EUR": -100.0, // Negative balance
			"USD": 500.0,
		},
	}

	adapter := NewBalanceAdapter(mockCashManager, log)

	total, err := adapter.GetTotalByCurrency("EUR")
	assert.NoError(t, err)
	assert.Equal(t, -100.0, total) // Should return negative value as-is
}
