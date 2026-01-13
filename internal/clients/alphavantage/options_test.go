package alphavantage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionsParseOptionsChain(t *testing.T) {
	jsonData := `{
		"meta": {
			"symbol": "IBM",
			"date": "2024-01-15"
		},
		"data": [
			{
				"contractID": "IBM240119C00150000",
				"symbol": "IBM",
				"expiration": "2024-01-19",
				"strike": "150.00",
				"type": "call",
				"last": "35.50",
				"mark": "35.40",
				"bid": "35.20",
				"bid_size": "10",
				"ask": "35.60",
				"ask_size": "15",
				"volume": "500",
				"open_interest": "1500",
				"implied_volatility": "0.25",
				"delta": "0.85",
				"gamma": "0.02",
				"theta": "-0.05",
				"vega": "0.15"
			},
			{
				"contractID": "IBM240119P00150000",
				"symbol": "IBM",
				"expiration": "2024-01-19",
				"strike": "150.00",
				"type": "put",
				"last": "0.50",
				"mark": "0.55",
				"bid": "0.45",
				"bid_size": "20",
				"ask": "0.65",
				"ask_size": "25",
				"volume": "800",
				"open_interest": "2500",
				"implied_volatility": "0.28",
				"delta": "-0.15",
				"gamma": "0.02",
				"theta": "-0.03",
				"vega": "0.10"
			}
		]
	}`

	chain, err := parseOptionsChain([]byte(jsonData), "IBM", "2024-01-15")
	require.NoError(t, err)

	assert.Equal(t, "IBM", chain.Symbol)
	assert.Equal(t, 2024, chain.Date.Year())
	assert.Equal(t, 1, int(chain.Date.Month()))
	assert.Equal(t, 15, chain.Date.Day())

	require.Len(t, chain.Calls, 1)
	assert.Equal(t, "IBM240119C00150000", chain.Calls[0].ContractID)
	assert.Equal(t, 150.0, chain.Calls[0].Strike)
	assert.Equal(t, 35.5, chain.Calls[0].Last)
	assert.Equal(t, 35.2, chain.Calls[0].Bid)
	assert.Equal(t, 35.6, chain.Calls[0].Ask)
	assert.Equal(t, int64(500), chain.Calls[0].Volume)
	assert.Equal(t, int64(1500), chain.Calls[0].OpenInterest)
	assert.Equal(t, 0.25, chain.Calls[0].ImpliedVol)
	assert.Equal(t, 0.85, chain.Calls[0].Delta)

	require.Len(t, chain.Puts, 1)
	assert.Equal(t, "IBM240119P00150000", chain.Puts[0].ContractID)
	assert.Equal(t, 150.0, chain.Puts[0].Strike)
	assert.Equal(t, 0.5, chain.Puts[0].Last)
	assert.Equal(t, -0.15, chain.Puts[0].Delta)
}

func TestOptionsParseOptionsChain_OnlyCalls(t *testing.T) {
	jsonData := `{
		"meta": {
			"symbol": "AAPL",
			"date": "2024-01-15"
		},
		"data": [
			{
				"contractID": "AAPL240119C00180000",
				"symbol": "AAPL",
				"expiration": "2024-01-19",
				"strike": "180.00",
				"type": "call",
				"last": "8.50",
				"mark": "8.45",
				"bid": "8.30",
				"ask": "8.60",
				"volume": "1000",
				"open_interest": "5000",
				"implied_volatility": "0.20",
				"delta": "0.65"
			}
		]
	}`

	chain, err := parseOptionsChain([]byte(jsonData), "AAPL", "2024-01-15")
	require.NoError(t, err)

	require.Len(t, chain.Calls, 1)
	assert.Empty(t, chain.Puts)
}

func TestOptionsParseOptionsChain_OnlyPuts(t *testing.T) {
	jsonData := `{
		"meta": {
			"symbol": "AAPL",
			"date": "2024-01-15"
		},
		"data": [
			{
				"contractID": "AAPL240119P00180000",
				"symbol": "AAPL",
				"expiration": "2024-01-19",
				"strike": "180.00",
				"type": "put",
				"last": "2.50",
				"mark": "2.55",
				"bid": "2.40",
				"ask": "2.70",
				"volume": "2000",
				"open_interest": "8000",
				"implied_volatility": "0.22",
				"delta": "-0.35"
			}
		]
	}`

	chain, err := parseOptionsChain([]byte(jsonData), "AAPL", "2024-01-15")
	require.NoError(t, err)

	assert.Empty(t, chain.Calls)
	require.Len(t, chain.Puts, 1)
}

func TestOptionsParseOptionsChain_InvalidJSON(t *testing.T) {
	_, err := parseOptionsChain([]byte("not json"), "IBM", "2024-01-15")
	assert.Error(t, err)
}

func TestOptionsParseOptionsChain_EmptyData(t *testing.T) {
	jsonData := `{
		"meta": {
			"symbol": "XYZ",
			"date": "2024-01-15"
		},
		"data": []
	}`

	chain, err := parseOptionsChain([]byte(jsonData), "XYZ", "2024-01-15")
	require.NoError(t, err)

	assert.Equal(t, "XYZ", chain.Symbol)
	assert.Empty(t, chain.Calls)
	assert.Empty(t, chain.Puts)
}

// TestClientOptionsMethods verifies all options methods exist.
func TestClientOptionsMethods(t *testing.T) {
	client := newTestClient("test-key")

	var _ func(string, string) (*OptionsChain, error) = client.GetHistoricalOptions
}
