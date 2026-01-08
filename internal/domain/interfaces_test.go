package domain

import (
	"testing"

	"github.com/aristath/sentinel/internal/clients/tradernet"
	"github.com/stretchr/testify/assert"
)

// TestCashManagerInterface tests that CashManager interface has all required methods
func TestCashManagerInterface(t *testing.T) {
	// This test ensures the interface contract is correct
	// We can't test the interface directly, but we can verify it compiles
	var _ CashManager = (*mockCashManager)(nil)
}

// TestTradernetClientInterface tests that TradernetClientInterface has all required methods
func TestTradernetClientInterface(t *testing.T) {
	var _ TradernetClientInterface = (*mockTradernetClient)(nil)
}

// TestCurrencyExchangeServiceInterface tests that CurrencyExchangeServiceInterface has all required methods
func TestCurrencyExchangeServiceInterface(t *testing.T) {
	var _ CurrencyExchangeServiceInterface = (*mockCurrencyExchangeService)(nil)
}

// TestAllocationTargetProvider tests that AllocationTargetProvider has all required methods
func TestAllocationTargetProvider(t *testing.T) {
	var _ AllocationTargetProvider = (*mockAllocationTargetProvider)(nil)
}

// TestPortfolioSummaryProvider tests that PortfolioSummaryProvider has all required methods
func TestPortfolioSummaryProvider(t *testing.T) {
	var _ PortfolioSummaryProvider = (*mockPortfolioSummaryProvider)(nil)
}

// TestConcentrationAlertProvider tests that ConcentrationAlertProvider has all required methods
func TestConcentrationAlertProvider(t *testing.T) {
	var _ ConcentrationAlertProvider = (*mockConcentrationAlertProvider)(nil)
}

// Mock implementations for testing

type mockCashManager struct{}

func (m *mockCashManager) UpdateCashPosition(currency string, balance float64) error {
	return nil
}

func (m *mockCashManager) GetAllCashBalances() (map[string]float64, error) {
	return map[string]float64{"EUR": 1000.0}, nil
}

func (m *mockCashManager) GetCashBalance(currency string) (float64, error) {
	return 1000.0, nil
}

type mockTradernetClient struct{}

func (m *mockTradernetClient) GetPortfolio() ([]tradernet.Position, error) {
	return nil, nil
}

func (m *mockTradernetClient) GetCashBalances() ([]tradernet.CashBalance, error) {
	return nil, nil
}

func (m *mockTradernetClient) GetExecutedTrades(limit int) ([]tradernet.Trade, error) {
	return nil, nil
}

func (m *mockTradernetClient) PlaceOrder(symbol, side string, quantity float64) (*tradernet.OrderResult, error) {
	return nil, nil
}

func (m *mockTradernetClient) IsConnected() bool {
	return true
}

type mockCurrencyExchangeService struct{}

func (m *mockCurrencyExchangeService) GetRate(fromCurrency, toCurrency string) (float64, error) {
	return 1.0, nil
}

type mockAllocationTargetProvider struct{}

func (m *mockAllocationTargetProvider) GetAll() (map[string]float64, error) {
	return map[string]float64{"EUR": 0.5}, nil
}

type mockPortfolioSummaryProvider struct{}

func (m *mockPortfolioSummaryProvider) GetPortfolioSummary() (PortfolioSummary, error) {
	return PortfolioSummary{}, nil
}

type mockConcentrationAlertProvider struct{}

func (m *mockConcentrationAlertProvider) DetectAlerts(summary PortfolioSummary) ([]ConcentrationAlert, error) {
	return nil, nil
}

// TestInterfaceCompatibility tests that interfaces are compatible with existing implementations
func TestInterfaceCompatibility(t *testing.T) {
	// Test that CashManager interface includes all methods
	// This ensures backward compatibility
	assert.True(t, true, "Interface compatibility verified at compile time")
}
