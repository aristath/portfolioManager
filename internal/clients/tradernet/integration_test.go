//go:build integration
// +build integration

package tradernet

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_UserInfo tests UserInfo() with real Tradernet API
// Requires TRADERNET_API_KEY and TRADERNET_API_SECRET environment variables
func TestIntegration_UserInfo(t *testing.T) {
	apiKey := os.Getenv("TRADERNET_API_KEY")
	apiSecret := os.Getenv("TRADERNET_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: TRADERNET_API_KEY and TRADERNET_API_SECRET not set")
	}

	log := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	client := NewClient(apiKey, apiSecret, log)

	result, err := client.HealthCheck()
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Connected, "Should be connected to Tradernet API")
}

// TestIntegration_GetPortfolio tests GetPortfolio() with real Tradernet API
func TestIntegration_GetPortfolio(t *testing.T) {
	apiKey := os.Getenv("TRADERNET_API_KEY")
	apiSecret := os.Getenv("TRADERNET_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: TRADERNET_API_KEY and TRADERNET_API_SECRET not set")
	}

	log := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	client := NewClient(apiKey, apiSecret, log)

	positions, err := client.GetPortfolio()
	require.NoError(t, err)
	assert.NotNil(t, positions)
	// Positions can be empty, but should not be nil
}

// TestIntegration_GetCashBalances tests GetCashBalances() with real Tradernet API
func TestIntegration_GetCashBalances(t *testing.T) {
	apiKey := os.Getenv("TRADERNET_API_KEY")
	apiSecret := os.Getenv("TRADERNET_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: TRADERNET_API_KEY and TRADERNET_API_SECRET not set")
	}

	log := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	client := NewClient(apiKey, apiSecret, log)

	balances, err := client.GetCashBalances()
	require.NoError(t, err)
	assert.NotNil(t, balances)
	// Balances can be empty, but should not be nil
}

// TestIntegration_GetPendingOrders tests GetPendingOrders() with real Tradernet API
func TestIntegration_GetPendingOrders(t *testing.T) {
	apiKey := os.Getenv("TRADERNET_API_KEY")
	apiSecret := os.Getenv("TRADERNET_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: TRADERNET_API_KEY and TRADERNET_API_SECRET not set")
	}

	log := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	client := NewClient(apiKey, apiSecret, log)

	orders, err := client.GetPendingOrders()
	require.NoError(t, err)
	assert.NotNil(t, orders)
	// Orders can be empty, but should not be nil
}

// TestIntegration_FindSymbol tests FindSymbol() with real Tradernet API
func TestIntegration_FindSymbol(t *testing.T) {
	apiKey := os.Getenv("TRADERNET_API_KEY")
	apiSecret := os.Getenv("TRADERNET_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: TRADERNET_API_KEY and TRADERNET_API_SECRET not set")
	}

	log := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	client := NewClient(apiKey, apiSecret, log)

	securities, err := client.FindSymbol("AAPL", nil)
	require.NoError(t, err)
	assert.NotNil(t, securities)
	// Should find at least one result for AAPL
	assert.Greater(t, len(securities), 0, "Should find at least one security for AAPL")
}

// TestIntegration_GetQuote tests GetQuote() with real Tradernet API
func TestIntegration_GetQuote(t *testing.T) {
	apiKey := os.Getenv("TRADERNET_API_KEY")
	apiSecret := os.Getenv("TRADERNET_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: TRADERNET_API_KEY and TRADERNET_API_SECRET not set")
	}

	log := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	client := NewClient(apiKey, apiSecret, log)

	quote, err := client.GetQuote("AAPL.US")
	require.NoError(t, err)
	assert.NotNil(t, quote)
	assert.Equal(t, "AAPL.US", quote.Symbol)
	assert.Greater(t, quote.Price, 0.0, "Price should be greater than 0")
}

// TestIntegration_GetExecutedTrades tests GetExecutedTrades() with real Tradernet API
func TestIntegration_GetExecutedTrades(t *testing.T) {
	apiKey := os.Getenv("TRADERNET_API_KEY")
	apiSecret := os.Getenv("TRADERNET_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: TRADERNET_API_KEY and TRADERNET_API_SECRET not set")
	}

	log := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	client := NewClient(apiKey, apiSecret, log)

	trades, err := client.GetExecutedTrades(10)
	require.NoError(t, err)
	assert.NotNil(t, trades)
	// Trades can be empty, but should not be nil
}

// TestIntegration_GetAllCashFlows tests GetAllCashFlows() with real Tradernet API
func TestIntegration_GetAllCashFlows(t *testing.T) {
	apiKey := os.Getenv("TRADERNET_API_KEY")
	apiSecret := os.Getenv("TRADERNET_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: TRADERNET_API_KEY and TRADERNET_API_SECRET not set")
	}

	log := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	client := NewClient(apiKey, apiSecret, log)

	transactions, err := client.GetAllCashFlows(100)
	require.NoError(t, err)
	assert.NotNil(t, transactions)
	// Transactions can be empty, but should not be nil
}
