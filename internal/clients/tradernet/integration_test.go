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

// TestIntegration_GetSecurityMetadataBatch tests batch metadata sync with real Tradernet API
// This test verifies that batch requests work correctly and avoid 429 rate limit errors
func TestIntegration_GetSecurityMetadataBatch(t *testing.T) {
	apiKey := os.Getenv("TRADERNET_API_KEY")
	apiSecret := os.Getenv("TRADERNET_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: TRADERNET_API_KEY and TRADERNET_API_SECRET not set")
	}

	log := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	client := NewClient(apiKey, apiSecret, log)

	// Test with a reasonable subset of securities (10 securities)
	symbols := []string{
		"AAPL.US",
		"MSFT.US",
		"GOOGL.US",
		"AMZN.US",
		"TSLA.US",
		"NVDA.US",
		"META.US",
		"BRK.B.US",
		"JPM.US",
		"V.US",
	}

	result, err := client.GetSecurityMetadataBatch(symbols)
	require.NoError(t, err, "Batch request should not return error")
	assert.NotNil(t, result, "Result should not be nil")

	// Verify response structure
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "Result should be a map")

	securities, ok := resultMap["securities"].([]interface{})
	require.True(t, ok, "Result should contain securities array")
	assert.Greater(t, len(securities), 0, "Should return at least some securities")

	// Verify we got data for the requested symbols
	t.Logf("Batch API returned %d securities for %d symbols", len(securities), len(symbols))
}

// TestIntegration_GetSecurityMetadataBatch_LargeBatch tests batch request with larger set
// This simulates real-world usage with 40 securities (typical portfolio size)
func TestIntegration_GetSecurityMetadataBatch_LargeBatch(t *testing.T) {
	apiKey := os.Getenv("TRADERNET_API_KEY")
	apiSecret := os.Getenv("TRADERNET_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: TRADERNET_API_KEY and TRADERNET_API_SECRET not set")
	}

	log := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	client := NewClient(apiKey, apiSecret, log)

	// Test with 40 securities (real-world scenario)
	symbols := []string{
		"AAPL.US", "MSFT.US", "GOOGL.US", "AMZN.US", "TSLA.US",
		"NVDA.US", "META.US", "BRK.B.US", "JPM.US", "V.US",
		"UNH.US", "JNJ.US", "WMT.US", "PG.US", "MA.US",
		"HD.US", "CVX.US", "LLY.US", "ABBV.US", "MRK.US",
		"PEP.US", "KO.US", "COST.US", "AVGO.US", "ADBE.US",
		"CSCO.US", "ACN.US", "NFLX.US", "TMO.US", "NKE.US",
		"DIS.US", "ABT.US", "CRM.US", "VZ.US", "INTC.US",
		"AMD.US", "TXN.US", "QCOM.US", "CMCSA.US", "PFE.US",
	}

	result, err := client.GetSecurityMetadataBatch(symbols)
	require.NoError(t, err, "Batch request with 40 securities should not return error")
	assert.NotNil(t, result, "Result should not be nil")

	// Verify response structure
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "Result should be a map")

	securities, ok := resultMap["securities"].([]interface{})
	require.True(t, ok, "Result should contain securities array")
	assert.Greater(t, len(securities), 0, "Should return securities")

	t.Logf("Large batch API returned %d securities for %d symbols", len(securities), len(symbols))
}

// TestIntegration_GetSecurityMetadataBatch_NoRateLimitErrors tests that batch requests
// respect rate limits and don't trigger 429 errors
func TestIntegration_GetSecurityMetadataBatch_NoRateLimitErrors(t *testing.T) {
	apiKey := os.Getenv("TRADERNET_API_KEY")
	apiSecret := os.Getenv("TRADERNET_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: TRADERNET_API_KEY and TRADERNET_API_SECRET not set")
	}

	log := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	client := NewClient(apiKey, apiSecret, log)

	// Test batch requests in succession (3 times)
	// This verifies we don't hit rate limits with batch API
	symbols := []string{
		"AAPL.US", "MSFT.US", "GOOGL.US", "AMZN.US", "TSLA.US",
		"NVDA.US", "META.US", "BRK.B.US", "JPM.US", "V.US",
	}

	for i := 0; i < 3; i++ {
		result, err := client.GetSecurityMetadataBatch(symbols)
		require.NoError(t, err, "Batch request %d should not return error (no 429)", i+1)
		assert.NotNil(t, result, "Result %d should not be nil", i+1)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok, "Result %d should be a map", i+1)

		securities, ok := resultMap["securities"].([]interface{})
		require.True(t, ok, "Result %d should contain securities array", i+1)
		assert.Greater(t, len(securities), 0, "Batch %d should return securities", i+1)

		t.Logf("Batch request %d: returned %d securities", i+1, len(securities))

		// Rate limit is 1.5s per request - batch API should handle this internally
		// No manual sleep needed between batch calls
	}
}
