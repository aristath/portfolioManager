package testing

import (
	"errors"
	"testing"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMockBrokerClient tests mock creation
func TestNewMockBrokerClient(t *testing.T) {
	mock := NewMockBrokerClient()

	assert.NotNil(t, mock)
	assert.True(t, mock.IsConnected()) // Default is connected
	assert.Len(t, mock.portfolio, 0)
	assert.Len(t, mock.cashBalances, 0)
}

// TestMockBrokerClient_GetPortfolio tests GetPortfolio method
func TestMockBrokerClient_GetPortfolio(t *testing.T) {
	mock := NewMockBrokerClient()

	t.Run("success", func(t *testing.T) {
		mock.SetPortfolio([]domain.BrokerPosition{
			{Symbol: "AAPL", Quantity: 10},
		})

		positions, err := mock.GetPortfolio()
		require.NoError(t, err)
		assert.Len(t, positions, 1)
		assert.Equal(t, "AAPL", positions[0].Symbol)
	})

	t.Run("error", func(t *testing.T) {
		mock.SetError(errors.New("test error"))
		_, err := mock.GetPortfolio()
		assert.Error(t, err)
		mock.SetError(nil) // Reset
	})
}

// TestMockBrokerClient_GetCashBalances tests GetCashBalances method
func TestMockBrokerClient_GetCashBalances(t *testing.T) {
	mock := NewMockBrokerClient()

	mock.SetCashBalances([]domain.BrokerCashBalance{
		{Currency: "EUR", Amount: 1000},
		{Currency: "USD", Amount: 500},
	})

	balances, err := mock.GetCashBalances()
	require.NoError(t, err)
	assert.Len(t, balances, 2)
	assert.Equal(t, "EUR", balances[0].Currency)
}

// TestMockBrokerClient_PlaceOrder tests PlaceOrder method
func TestMockBrokerClient_PlaceOrder(t *testing.T) {
	mock := NewMockBrokerClient()

	t.Run("with custom result", func(t *testing.T) {
		mock.SetOrderResult(&domain.BrokerOrderResult{
			OrderID:  "custom-123",
			Symbol:   "MSFT",
			Side:     "BUY",
			Quantity: 5,
		})

		result, err := mock.PlaceOrder("MSFT", "BUY", 5)
		require.NoError(t, err)
		assert.Equal(t, "custom-123", result.OrderID)
	})

	t.Run("with default result", func(t *testing.T) {
		mock.SetOrderResult(nil) // Clear custom result
		result, err := mock.PlaceOrder("AAPL", "SELL", 3)
		require.NoError(t, err)
		assert.Equal(t, "mock_order_AAPL", result.OrderID)
		assert.Equal(t, "AAPL", result.Symbol)
		assert.Equal(t, "SELL", result.Side)
	})
}

// TestMockBrokerClient_GetExecutedTrades tests GetExecutedTrades method
func TestMockBrokerClient_GetExecutedTrades(t *testing.T) {
	mock := NewMockBrokerClient()

	mock.SetTrades([]domain.BrokerTrade{
		{OrderID: "trade-1"},
		{OrderID: "trade-2"},
		{OrderID: "trade-3"},
	})

	t.Run("with limit", func(t *testing.T) {
		trades, err := mock.GetExecutedTrades(2)
		require.NoError(t, err)
		assert.Len(t, trades, 2)
	})

	t.Run("without limit", func(t *testing.T) {
		trades, err := mock.GetExecutedTrades(0)
		require.NoError(t, err)
		assert.Len(t, trades, 3)
	})
}

// TestMockBrokerClient_IsConnected tests IsConnected method
func TestMockBrokerClient_IsConnected(t *testing.T) {
	mock := NewMockBrokerClient()

	assert.True(t, mock.IsConnected())

	mock.SetConnected(false)
	assert.False(t, mock.IsConnected())

	mock.SetConnected(true)
	assert.True(t, mock.IsConnected())
}

// TestMockBrokerClient_GetAllCashFlows tests GetAllCashFlows method
func TestMockBrokerClient_GetAllCashFlows(t *testing.T) {
	mock := NewMockBrokerClient()

	mock.SetCashFlows([]domain.BrokerCashFlow{
		{ID: "cf-1", Type: "deposit"},
		{ID: "cf-2", Type: "dividend"},
	})

	flows, err := mock.GetAllCashFlows(100)
	require.NoError(t, err)
	assert.Len(t, flows, 2)
	assert.Equal(t, "deposit", flows[0].Type)
}

// TestMockBrokerClient_FindSymbol tests FindSymbol method
func TestMockBrokerClient_FindSymbol(t *testing.T) {
	mock := NewMockBrokerClient()

	name := "Apple Inc."
	mock.SetSecurities([]domain.BrokerSecurityInfo{
		{Symbol: "AAPL", Name: &name},
	})

	securities, err := mock.FindSymbol("AAPL", nil)
	require.NoError(t, err)
	assert.Len(t, securities, 1)
	assert.Equal(t, "AAPL", securities[0].Symbol)
}

// TestMockBrokerClient_GetQuote tests GetQuote method
func TestMockBrokerClient_GetQuote(t *testing.T) {
	mock := NewMockBrokerClient()

	mock.SetQuote("GOOGL", &domain.BrokerQuote{
		Symbol: "GOOGL",
		Price:  140.50,
	})

	quote, err := mock.GetQuote("GOOGL")
	require.NoError(t, err)
	assert.NotNil(t, quote)
	assert.Equal(t, "GOOGL", quote.Symbol)
	assert.Equal(t, 140.50, quote.Price)
}

// TestMockBrokerClient_GetPendingOrders tests GetPendingOrders method
func TestMockBrokerClient_GetPendingOrders(t *testing.T) {
	mock := NewMockBrokerClient()

	mock.SetPendingOrders([]domain.BrokerPendingOrder{
		{OrderID: "pending-1", Symbol: "AMZN"},
	})

	orders, err := mock.GetPendingOrders()
	require.NoError(t, err)
	assert.Len(t, orders, 1)
	assert.Equal(t, "AMZN", orders[0].Symbol)
}

// TestMockBrokerClient_HealthCheck tests HealthCheck method
func TestMockBrokerClient_HealthCheck(t *testing.T) {
	mock := NewMockBrokerClient()

	t.Run("with custom result", func(t *testing.T) {
		mock.SetHealthResult(&domain.BrokerHealthResult{
			Connected: true,
			Timestamp: "2025-01-08T12:00:00Z",
		})

		result, err := mock.HealthCheck()
		require.NoError(t, err)
		assert.True(t, result.Connected)
		assert.Equal(t, "2025-01-08T12:00:00Z", result.Timestamp)
	})

	t.Run("with default result", func(t *testing.T) {
		mock.SetHealthResult(nil) // Clear custom result
		result, err := mock.HealthCheck()
		require.NoError(t, err)
		assert.True(t, result.Connected)
		assert.NotEmpty(t, result.Timestamp)
	})
}

// TestMockBrokerClient_GetCashMovements tests GetCashMovements method
func TestMockBrokerClient_GetCashMovements(t *testing.T) {
	mock := NewMockBrokerClient()

	mock.SetCashMovements(&domain.BrokerCashMovement{
		TotalWithdrawals: 5000,
		Note:             "test",
	})

	movements, err := mock.GetCashMovements()
	require.NoError(t, err)
	assert.NotNil(t, movements)
	assert.Equal(t, 5000.0, movements.TotalWithdrawals)
}

// TestMockBrokerClient_SetCredentials tests SetCredentials method
func TestMockBrokerClient_SetCredentials(t *testing.T) {
	mock := NewMockBrokerClient()

	assert.False(t, mock.credentialsSet)

	mock.SetCredentials("test-key", "test-secret")
	assert.True(t, mock.credentialsSet)
}

// TestMockBrokerClient_ThreadSafety tests thread-safe operations
func TestMockBrokerClient_ThreadSafety(t *testing.T) {
	mock := NewMockBrokerClient()

	// Test concurrent reads and writes
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			mock.SetPortfolio([]domain.BrokerPosition{{Symbol: "TEST"}})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			_, _ = mock.GetPortfolio()
		}
		done <- true
	}()

	<-done
	<-done
}

// TestMockBrokerClient_InterfaceCompliance tests interface implementation
func TestMockBrokerClient_InterfaceCompliance(t *testing.T) {
	var _ domain.BrokerClient = (*MockBrokerClient)(nil)
}
