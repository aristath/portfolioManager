package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockBrokerClientForTest is a simple mock for testing interface specification
type mockBrokerClientForTest struct {
	portfolio      []BrokerPosition
	cashBalances   []BrokerCashBalance
	trades         []BrokerTrade
	pendingOrders  []BrokerPendingOrder
	cashFlows      []BrokerCashFlow
	quote          *BrokerQuote
	securities     []BrokerSecurityInfo
	cashMovements  *BrokerCashMovement
	healthResult   *BrokerHealthResult
	orderResult    *BrokerOrderResult
	connected      bool
	credentialsSet bool
	returnError    bool
}

// GetPortfolio implements BrokerClient
func (m *mockBrokerClientForTest) GetPortfolio() ([]BrokerPosition, error) {
	if m.returnError {
		return nil, errors.New("mock error")
	}
	return m.portfolio, nil
}

// GetCashBalances implements BrokerClient
func (m *mockBrokerClientForTest) GetCashBalances() ([]BrokerCashBalance, error) {
	if m.returnError {
		return nil, errors.New("mock error")
	}
	return m.cashBalances, nil
}

// GetExecutedTrades implements BrokerClient
func (m *mockBrokerClientForTest) GetExecutedTrades(limit int) ([]BrokerTrade, error) {
	if m.returnError {
		return nil, errors.New("mock error")
	}
	return m.trades, nil
}

// PlaceOrder implements BrokerClient
func (m *mockBrokerClientForTest) PlaceOrder(symbol, side string, quantity float64) (*BrokerOrderResult, error) {
	if m.returnError {
		return nil, errors.New("mock error")
	}
	return m.orderResult, nil
}

// IsConnected implements BrokerClient
func (m *mockBrokerClientForTest) IsConnected() bool {
	return m.connected
}

// GetAllCashFlows implements BrokerClient
func (m *mockBrokerClientForTest) GetAllCashFlows(limit int) ([]BrokerCashFlow, error) {
	if m.returnError {
		return nil, errors.New("mock error")
	}
	return m.cashFlows, nil
}

// FindSymbol implements BrokerClient
func (m *mockBrokerClientForTest) FindSymbol(symbol string, exchange *string) ([]BrokerSecurityInfo, error) {
	if m.returnError {
		return nil, errors.New("mock error")
	}
	return m.securities, nil
}

// GetQuote implements BrokerClient
func (m *mockBrokerClientForTest) GetQuote(symbol string) (*BrokerQuote, error) {
	if m.returnError {
		return nil, errors.New("mock error")
	}
	return m.quote, nil
}

// GetPendingOrders implements BrokerClient
func (m *mockBrokerClientForTest) GetPendingOrders() ([]BrokerPendingOrder, error) {
	if m.returnError {
		return nil, errors.New("mock error")
	}
	return m.pendingOrders, nil
}

// HealthCheck implements BrokerClient
func (m *mockBrokerClientForTest) HealthCheck() (*BrokerHealthResult, error) {
	if m.returnError {
		return nil, errors.New("mock error")
	}
	return m.healthResult, nil
}

// GetCashMovements implements BrokerClient
func (m *mockBrokerClientForTest) GetCashMovements() (*BrokerCashMovement, error) {
	if m.returnError {
		return nil, errors.New("mock error")
	}
	return m.cashMovements, nil
}

// SetCredentials implements BrokerClient
func (m *mockBrokerClientForTest) SetCredentials(apiKey, apiSecret string) {
	m.credentialsSet = true
}

// Compile-time check that mockBrokerClientForTest implements BrokerClient
var _ BrokerClient = (*mockBrokerClientForTest)(nil)

// TestBrokerClientInterface_GetPortfolio tests GetPortfolio method spec
func TestBrokerClientInterface_GetPortfolio(t *testing.T) {
	mock := &mockBrokerClientForTest{
		portfolio: []BrokerPosition{
			{Symbol: "AAPL", Quantity: 10},
		},
	}

	positions, err := mock.GetPortfolio()
	assert.NoError(t, err)
	assert.Len(t, positions, 1)
	assert.Equal(t, "AAPL", positions[0].Symbol)
}

// TestBrokerClientInterface_GetCashBalances tests GetCashBalances method spec
func TestBrokerClientInterface_GetCashBalances(t *testing.T) {
	mock := &mockBrokerClientForTest{
		cashBalances: []BrokerCashBalance{
			{Currency: "EUR", Amount: 1000},
		},
	}

	balances, err := mock.GetCashBalances()
	assert.NoError(t, err)
	assert.Len(t, balances, 1)
	assert.Equal(t, "EUR", balances[0].Currency)
}

// TestBrokerClientInterface_PlaceOrder tests PlaceOrder method spec
func TestBrokerClientInterface_PlaceOrder(t *testing.T) {
	mock := &mockBrokerClientForTest{
		orderResult: &BrokerOrderResult{
			OrderID: "order-123",
			Symbol:  "MSFT",
			Side:    "BUY",
		},
	}

	result, err := mock.PlaceOrder("MSFT", "BUY", 5.0)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "order-123", result.OrderID)
}

// TestBrokerClientInterface_GetExecutedTrades tests GetExecutedTrades method spec
func TestBrokerClientInterface_GetExecutedTrades(t *testing.T) {
	mock := &mockBrokerClientForTest{
		trades: []BrokerTrade{
			{OrderID: "trade-1", Symbol: "TSLA"},
		},
	}

	trades, err := mock.GetExecutedTrades(100)
	assert.NoError(t, err)
	assert.Len(t, trades, 1)
	assert.Equal(t, "TSLA", trades[0].Symbol)
}

// TestBrokerClientInterface_IsConnected tests IsConnected method spec
func TestBrokerClientInterface_IsConnected(t *testing.T) {
	mockConnected := &mockBrokerClientForTest{connected: true}
	assert.True(t, mockConnected.IsConnected())

	mockDisconnected := &mockBrokerClientForTest{connected: false}
	assert.False(t, mockDisconnected.IsConnected())
}

// TestBrokerClientInterface_GetAllCashFlows tests GetAllCashFlows method spec
func TestBrokerClientInterface_GetAllCashFlows(t *testing.T) {
	mock := &mockBrokerClientForTest{
		cashFlows: []BrokerCashFlow{
			{ID: "cf-1", Type: "deposit"},
		},
	}

	flows, err := mock.GetAllCashFlows(1000)
	assert.NoError(t, err)
	assert.Len(t, flows, 1)
	assert.Equal(t, "deposit", flows[0].Type)
}

// TestBrokerClientInterface_FindSymbol tests FindSymbol method spec
func TestBrokerClientInterface_FindSymbol(t *testing.T) {
	name := "Apple Inc."
	mock := &mockBrokerClientForTest{
		securities: []BrokerSecurityInfo{
			{Symbol: "AAPL", Name: &name},
		},
	}

	securities, err := mock.FindSymbol("AAPL", nil)
	assert.NoError(t, err)
	assert.Len(t, securities, 1)
	assert.Equal(t, "AAPL", securities[0].Symbol)
}

// TestBrokerClientInterface_GetQuote tests GetQuote method spec
func TestBrokerClientInterface_GetQuote(t *testing.T) {
	mock := &mockBrokerClientForTest{
		quote: &BrokerQuote{
			Symbol: "GOOGL",
			Price:  140.50,
		},
	}

	quote, err := mock.GetQuote("GOOGL")
	assert.NoError(t, err)
	assert.NotNil(t, quote)
	assert.Equal(t, "GOOGL", quote.Symbol)
	assert.Equal(t, 140.50, quote.Price)
}

// TestBrokerClientInterface_GetPendingOrders tests GetPendingOrders method spec
func TestBrokerClientInterface_GetPendingOrders(t *testing.T) {
	mock := &mockBrokerClientForTest{
		pendingOrders: []BrokerPendingOrder{
			{OrderID: "pending-1", Symbol: "AMZN"},
		},
	}

	orders, err := mock.GetPendingOrders()
	assert.NoError(t, err)
	assert.Len(t, orders, 1)
	assert.Equal(t, "AMZN", orders[0].Symbol)
}

// TestBrokerClientInterface_HealthCheck tests HealthCheck method spec
func TestBrokerClientInterface_HealthCheck(t *testing.T) {
	mock := &mockBrokerClientForTest{
		healthResult: &BrokerHealthResult{
			Connected: true,
			Timestamp: "2025-01-08T12:00:00Z",
		},
	}

	result, err := mock.HealthCheck()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Connected)
}

// TestBrokerClientInterface_GetCashMovements tests GetCashMovements method spec
func TestBrokerClientInterface_GetCashMovements(t *testing.T) {
	mock := &mockBrokerClientForTest{
		cashMovements: &BrokerCashMovement{
			TotalWithdrawals: 5000.0,
			Note:             "test",
		},
	}

	movements, err := mock.GetCashMovements()
	assert.NoError(t, err)
	assert.NotNil(t, movements)
	assert.Equal(t, 5000.0, movements.TotalWithdrawals)
}

// TestBrokerClientInterface_SetCredentials tests SetCredentials method spec
func TestBrokerClientInterface_SetCredentials(t *testing.T) {
	mock := &mockBrokerClientForTest{}
	assert.False(t, mock.credentialsSet)

	mock.SetCredentials("test-key", "test-secret")
	assert.True(t, mock.credentialsSet)
}

// TestBrokerClientInterface_ErrorHandling tests error propagation
func TestBrokerClientInterface_ErrorHandling(t *testing.T) {
	mock := &mockBrokerClientForTest{returnError: true}

	_, err := mock.GetPortfolio()
	assert.Error(t, err)

	_, err = mock.GetCashBalances()
	assert.Error(t, err)

	_, err = mock.PlaceOrder("TEST", "BUY", 1.0)
	assert.Error(t, err)

	_, err = mock.GetExecutedTrades(100)
	assert.Error(t, err)

	_, err = mock.GetAllCashFlows(100)
	assert.Error(t, err)

	_, err = mock.FindSymbol("TEST", nil)
	assert.Error(t, err)

	_, err = mock.GetQuote("TEST")
	assert.Error(t, err)

	_, err = mock.GetPendingOrders()
	assert.Error(t, err)

	_, err = mock.HealthCheck()
	assert.Error(t, err)

	_, err = mock.GetCashMovements()
	assert.Error(t, err)
}
