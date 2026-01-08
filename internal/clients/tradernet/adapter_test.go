package tradernet

import (
	"errors"
	"testing"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTradernetBrokerAdapter tests adapter creation
func TestNewTradernetBrokerAdapter(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	adapter := NewTradernetBrokerAdapter("test-key", "test-secret", log)

	assert.NotNil(t, adapter)
	assert.NotNil(t, adapter.client)
}

// TestTradernetBrokerAdapter_GetPortfolio tests GetPortfolio transformation
func TestTradernetBrokerAdapter_GetPortfolio(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	t.Run("success", func(t *testing.T) {
		mockSDK := &mockSDKClient{
			accountSummaryResult: map[string]interface{}{
				"result": map[string]interface{}{
					"ps": map[string]interface{}{
						"pos": []interface{}{
							map[string]interface{}{
								"i":            "AAPL",
								"q":            10.0,
								"bal_price_a":  150.0,
								"mkt_price":    155.0,
								"profit_close": 50.0,
								"curr":         "USD",
							},
						},
					},
				},
			},
		}

		client := NewClientWithSDK(mockSDK, log)
		adapter := &TradernetBrokerAdapter{client: client}

		positions, err := adapter.GetPortfolio()
		require.NoError(t, err)
		assert.Len(t, positions, 1)
		assert.Equal(t, "AAPL", positions[0].Symbol)
		assert.Equal(t, 10.0, positions[0].Quantity)
		assert.Equal(t, 150.0, positions[0].AvgPrice)
	})

	t.Run("sdk error", func(t *testing.T) {
		mockSDK := &mockSDKClient{
			accountSummaryError: errors.New("sdk error"),
		}

		client := NewClientWithSDK(mockSDK, log)
		adapter := &TradernetBrokerAdapter{client: client}

		positions, err := adapter.GetPortfolio()
		assert.Error(t, err)
		assert.Nil(t, positions)
	})
}

// TestTradernetBrokerAdapter_GetCashBalances tests GetCashBalances transformation
func TestTradernetBrokerAdapter_GetCashBalances(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	t.Run("success", func(t *testing.T) {
		mockSDK := &mockSDKClient{
			accountSummaryResult: map[string]interface{}{
				"result": map[string]interface{}{
					"ps": map[string]interface{}{
						"acc": []interface{}{
							map[string]interface{}{
								"curr": "EUR",
								"s":    1000.0,
							},
						},
					},
				},
			},
		}

		client := NewClientWithSDK(mockSDK, log)
		adapter := &TradernetBrokerAdapter{client: client}

		balances, err := adapter.GetCashBalances()
		require.NoError(t, err)
		assert.Len(t, balances, 1)
		assert.Equal(t, "EUR", balances[0].Currency)
		assert.Equal(t, 1000.0, balances[0].Amount)
	})

	t.Run("sdk error", func(t *testing.T) {
		mockSDK := &mockSDKClient{
			accountSummaryError: errors.New("sdk error"),
		}

		client := NewClientWithSDK(mockSDK, log)
		adapter := &TradernetBrokerAdapter{client: client}

		balances, err := adapter.GetCashBalances()
		assert.Error(t, err)
		assert.Nil(t, balances)
	})
}

// TestTradernetBrokerAdapter_PlaceOrder tests PlaceOrder transformation
func TestTradernetBrokerAdapter_PlaceOrder(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	t.Run("buy success", func(t *testing.T) {
		mockSDK := &mockSDKClient{
			buyResult: map[string]interface{}{
				"id":    "order-123",
				"price": 150.50,
			},
		}

		client := NewClientWithSDK(mockSDK, log)
		adapter := &TradernetBrokerAdapter{client: client}

		result, err := adapter.PlaceOrder("AAPL", "BUY", 5.0)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "order-123", result.OrderID)
		assert.Equal(t, "AAPL", result.Symbol)
		assert.Equal(t, "BUY", result.Side)
	})

	t.Run("sell success", func(t *testing.T) {
		mockSDK := &mockSDKClient{
			sellResult: map[string]interface{}{
				"id":    "order-456",
				"price": 320.75,
			},
		}

		client := NewClientWithSDK(mockSDK, log)
		adapter := &TradernetBrokerAdapter{client: client}

		result, err := adapter.PlaceOrder("MSFT", "SELL", 3.0)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "order-456", result.OrderID)
		assert.Equal(t, "MSFT", result.Symbol)
		assert.Equal(t, "SELL", result.Side)
	})

	t.Run("sdk error", func(t *testing.T) {
		mockSDK := &mockSDKClient{
			buyError: errors.New("sdk error"),
		}

		client := NewClientWithSDK(mockSDK, log)
		adapter := &TradernetBrokerAdapter{client: client}

		result, err := adapter.PlaceOrder("AAPL", "BUY", 5.0)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestTradernetBrokerAdapter_GetExecutedTrades tests GetExecutedTrades transformation
func TestTradernetBrokerAdapter_GetExecutedTrades(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockSDK := &mockSDKClient{
		getTradesHistoryResult: map[string]interface{}{
			"result": []interface{}{
				map[string]interface{}{
					"id":          "trade-1",
					"i":           "TSLA",
					"side":        "BUY",
					"q":           2.0,
					"p":           250.0,
					"executed_at": "2025-01-08T10:00:00Z",
				},
			},
		},
	}

	client := NewClientWithSDK(mockSDK, log)
	adapter := &TradernetBrokerAdapter{client: client}

	trades, err := adapter.GetExecutedTrades(100)
	require.NoError(t, err)
	assert.Len(t, trades, 1)
	assert.Equal(t, "trade-1", trades[0].OrderID)
	assert.Equal(t, "TSLA", trades[0].Symbol)
	assert.Equal(t, "BUY", trades[0].Side)
}

// TestTradernetBrokerAdapter_IsConnected tests IsConnected
func TestTradernetBrokerAdapter_IsConnected(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	t.Run("connected", func(t *testing.T) {
		mockSDK := &mockSDKClient{
			userInfoResult: map[string]interface{}{
				"result": map[string]interface{}{
					"id": 123,
				},
			},
		}

		client := NewClientWithSDK(mockSDK, log)
		adapter := &TradernetBrokerAdapter{client: client}

		assert.True(t, adapter.IsConnected())
	})

	t.Run("disconnected", func(t *testing.T) {
		mockSDK := &mockSDKClient{
			userInfoError: errors.New("connection error"),
		}

		client := NewClientWithSDK(mockSDK, log)
		adapter := &TradernetBrokerAdapter{client: client}

		assert.False(t, adapter.IsConnected())
	})
}

// TestTradernetBrokerAdapter_GetAllCashFlows tests GetAllCashFlows transformation
func TestTradernetBrokerAdapter_GetAllCashFlows(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockSDK := &mockSDKClient{
		getClientCpsHistoryResult: map[string]interface{}{
			"result": []interface{}{
				map[string]interface{}{
					"id":          "cf-1",
					"transaction": "tx-1",
					"type":        "deposit",
					"sm":          1000.0,
					"c":           "EUR",
					"dt":          "2025-01-08",
					"description": "Monthly deposit",
				},
			},
		},
	}

	client := NewClientWithSDK(mockSDK, log)
	adapter := &TradernetBrokerAdapter{client: client}

	flows, err := adapter.GetAllCashFlows(1000)
	require.NoError(t, err)
	assert.Len(t, flows, 1)
	assert.Equal(t, "cf-1", flows[0].ID)
	assert.Equal(t, "deposit", flows[0].Type)
}

// TestTradernetBrokerAdapter_FindSymbol tests FindSymbol transformation
func TestTradernetBrokerAdapter_FindSymbol(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockSDK := &mockSDKClient{
		findSymbolResult: map[string]interface{}{
			"result": []interface{}{
				map[string]interface{}{
					"t":       "AAPL",
					"nm":      "Apple Inc.",
					"isin":    "US0378331005",
					"x_curr":  "USD",
					"mkt":     "NASDAQ",
					"codesub": "XNAS",
				},
			},
		},
	}

	client := NewClientWithSDK(mockSDK, log)
	adapter := &TradernetBrokerAdapter{client: client}

	securities, err := adapter.FindSymbol("AAPL", nil)
	require.NoError(t, err)
	assert.Len(t, securities, 1)
	assert.Equal(t, "AAPL", securities[0].Symbol)
	assert.NotNil(t, securities[0].Name)
	assert.Equal(t, "Apple Inc.", *securities[0].Name)
}

// TestTradernetBrokerAdapter_GetQuote tests GetQuote transformation
func TestTradernetBrokerAdapter_GetQuote(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockSDK := &mockSDKClient{
		getQuotesResult: map[string]interface{}{
			"result": map[string]interface{}{
				"GOOGL": map[string]interface{}{
					"p":  140.50,
					"ch": 2.5,
					"cp": 1.8,
					"v":  1000000,
					"t":  "2025-01-08T15:30:00Z",
				},
			},
		},
	}

	client := NewClientWithSDK(mockSDK, log)
	adapter := &TradernetBrokerAdapter{client: client}

	quote, err := adapter.GetQuote("GOOGL")
	require.NoError(t, err)
	assert.NotNil(t, quote)
	assert.Equal(t, "GOOGL", quote.Symbol)
	assert.Equal(t, 140.50, quote.Price)
}

// TestTradernetBrokerAdapter_GetPendingOrders tests GetPendingOrders transformation
func TestTradernetBrokerAdapter_GetPendingOrders(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockSDK := &mockSDKClient{
		getPlacedResult: map[string]interface{}{
			"result": []interface{}{
				map[string]interface{}{
					"id":   "order-789",
					"i":    "AMZN",
					"d":    "BUY",
					"q":    3.0,
					"p":    175.0,
					"curr": "USD",
				},
			},
		},
	}

	client := NewClientWithSDK(mockSDK, log)
	adapter := &TradernetBrokerAdapter{client: client}

	orders, err := adapter.GetPendingOrders()
	require.NoError(t, err)
	assert.Len(t, orders, 1)
	assert.Equal(t, "order-789", orders[0].OrderID)
	assert.Equal(t, "AMZN", orders[0].Symbol)
}

// TestTradernetBrokerAdapter_HealthCheck tests HealthCheck
func TestTradernetBrokerAdapter_HealthCheck(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	t.Run("healthy", func(t *testing.T) {
		mockSDK := &mockSDKClient{
			userInfoResult: map[string]interface{}{
				"result": map[string]interface{}{
					"id": 123,
				},
			},
		}

		client := NewClientWithSDK(mockSDK, log)
		adapter := &TradernetBrokerAdapter{client: client}

		result, err := adapter.HealthCheck()
		require.NoError(t, err)
		assert.True(t, result.Connected)
	})

	t.Run("unhealthy", func(t *testing.T) {
		mockSDK := &mockSDKClient{
			userInfoError: errors.New("connection error"),
		}

		client := NewClientWithSDK(mockSDK, log)
		adapter := &TradernetBrokerAdapter{client: client}

		result, err := adapter.HealthCheck()
		require.NoError(t, err) // HealthCheck doesn't return error, just status
		assert.False(t, result.Connected)
	})
}

// TestTradernetBrokerAdapter_GetCashMovements tests GetCashMovements transformation
func TestTradernetBrokerAdapter_GetCashMovements(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockSDK := &mockSDKClient{
		getClientCpsHistoryResult: map[string]interface{}{
			"result": []interface{}{
				map[string]interface{}{
					"sm": 2000.0,
				},
				map[string]interface{}{
					"sm": 3000.0,
				},
			},
		},
	}

	client := NewClientWithSDK(mockSDK, log)
	adapter := &TradernetBrokerAdapter{client: client}

	movements, err := adapter.GetCashMovements()
	require.NoError(t, err)
	assert.NotNil(t, movements)
}

// TestTradernetBrokerAdapter_SetCredentials tests SetCredentials
func TestTradernetBrokerAdapter_SetCredentials(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	adapter := NewTradernetBrokerAdapter("old-key", "old-secret", log)

	// SetCredentials should not panic
	adapter.SetCredentials("new-key", "new-secret")
}

// Compile-time check that TradernetBrokerAdapter implements domain.BrokerClient
var _ domain.BrokerClient = (*TradernetBrokerAdapter)(nil)
