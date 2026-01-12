package alphavantage

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCryptoParseCryptoTimeSeries(t *testing.T) {
	jsonData := `{
		"Meta Data": {
			"1. Information": "Daily Prices and Volumes for Digital Currency",
			"2. Digital Currency Code": "BTC",
			"3. Digital Currency Name": "Bitcoin",
			"4. Market Code": "USD",
			"5. Market Name": "United States Dollar"
		},
		"Time Series (Digital Currency Daily)": {
			"2024-01-15": {
				"1a. open (USD)": "42500.00",
				"1b. open (USD)": "42500.00",
				"2a. high (USD)": "43200.00",
				"2b. high (USD)": "43200.00",
				"3a. low (USD)": "42100.00",
				"3b. low (USD)": "42100.00",
				"4a. close (USD)": "42800.00",
				"4b. close (USD)": "42800.00",
				"5. volume": "25000.5",
				"6. market cap (USD)": "850000000000"
			},
			"2024-01-14": {
				"1a. open (USD)": "42000.00",
				"1b. open (USD)": "42000.00",
				"2a. high (USD)": "42600.00",
				"2b. high (USD)": "42600.00",
				"3a. low (USD)": "41800.00",
				"3b. low (USD)": "41800.00",
				"4a. close (USD)": "42500.00",
				"4b. close (USD)": "42500.00",
				"5. volume": "28000.2",
				"6. market cap (USD)": "840000000000"
			}
		}
	}`

	prices, err := parseCryptoTimeSeries([]byte(jsonData), "Time Series (Digital Currency Daily)", "USD")
	require.NoError(t, err)
	require.Len(t, prices, 2)

	// Verify sorting (newest first)
	assert.Equal(t, 15, prices[0].Date.Day())
	assert.Equal(t, 42500.0, prices[0].Open)
	assert.Equal(t, 43200.0, prices[0].High)
	assert.Equal(t, 42100.0, prices[0].Low)
	assert.Equal(t, 42800.0, prices[0].Close)
	assert.Equal(t, 25000.5, prices[0].Volume)
	assert.Equal(t, int64(850000000000), prices[0].MarketCap)
}

func TestCryptoParseCryptoTimeSeries_Weekly(t *testing.T) {
	jsonData := `{
		"Meta Data": {},
		"Time Series (Digital Currency Weekly)": {
			"2024-01-12": {
				"1a. open (USD)": "40000.00",
				"2a. high (USD)": "44000.00",
				"3a. low (USD)": "39000.00",
				"4a. close (USD)": "42800.00",
				"5. volume": "150000.0",
				"6. market cap (USD)": "850000000000"
			}
		}
	}`

	prices, err := parseCryptoTimeSeries([]byte(jsonData), "Time Series (Digital Currency Weekly)", "USD")
	require.NoError(t, err)
	require.Len(t, prices, 1)

	assert.Equal(t, 40000.0, prices[0].Open)
	assert.Equal(t, 44000.0, prices[0].High)
}

func TestCryptoParseCryptoTimeSeries_Monthly(t *testing.T) {
	jsonData := `{
		"Meta Data": {},
		"Time Series (Digital Currency Monthly)": {
			"2024-01-31": {
				"1a. open (USD)": "38000.00",
				"2a. high (USD)": "48000.00",
				"3a. low (USD)": "36000.00",
				"4a. close (USD)": "42800.00",
				"5. volume": "500000.0",
				"6. market cap (USD)": "850000000000"
			}
		}
	}`

	prices, err := parseCryptoTimeSeries([]byte(jsonData), "Time Series (Digital Currency Monthly)", "USD")
	require.NoError(t, err)
	require.Len(t, prices, 1)

	assert.Equal(t, 38000.0, prices[0].Open)
}

func TestCryptoParseCryptoTimeSeries_InvalidJSON(t *testing.T) {
	_, err := parseCryptoTimeSeries([]byte("not json"), "Time Series (Digital Currency Daily)", "USD")
	assert.Error(t, err)
}

func TestCryptoParseCryptoTimeSeries_MissingKey(t *testing.T) {
	jsonData := `{
		"Meta Data": {}
	}`

	_, err := parseCryptoTimeSeries([]byte(jsonData), "Time Series (Digital Currency Daily)", "USD")
	assert.Error(t, err)
}

func TestCryptoDateSorting(t *testing.T) {
	jsonData := `{
		"Meta Data": {},
		"Time Series (Digital Currency Daily)": {
			"2024-01-01": {"1a. open (USD)": "40000", "2a. high (USD)": "40000", "3a. low (USD)": "40000", "4a. close (USD)": "40000", "5. volume": "1000", "6. market cap (USD)": "800000000000"},
			"2024-01-15": {"1a. open (USD)": "42000", "2a. high (USD)": "42000", "3a. low (USD)": "42000", "4a. close (USD)": "42000", "5. volume": "1000", "6. market cap (USD)": "840000000000"},
			"2024-01-10": {"1a. open (USD)": "41000", "2a. high (USD)": "41000", "3a. low (USD)": "41000", "4a. close (USD)": "41000", "5. volume": "1000", "6. market cap (USD)": "820000000000"}
		}
	}`

	prices, err := parseCryptoTimeSeries([]byte(jsonData), "Time Series (Digital Currency Daily)", "USD")
	require.NoError(t, err)
	require.Len(t, prices, 3)

	// Should be sorted newest first
	assert.Equal(t, 15, prices[0].Date.Day())
	assert.Equal(t, 10, prices[1].Date.Day())
	assert.Equal(t, 1, prices[2].Date.Day())
}

// TestClientCryptoMethods verifies all crypto methods exist.
func TestClientCryptoMethods(t *testing.T) {
	client := NewClient("test-key", zerolog.Nop())

	var _ func(string, string) (*ExchangeRate, error) = client.GetCryptoExchangeRate
	var _ func(string, string) ([]CryptoPrice, error) = client.GetCryptoDaily
	var _ func(string, string) ([]CryptoPrice, error) = client.GetCryptoWeekly
	var _ func(string, string) ([]CryptoPrice, error) = client.GetCryptoMonthly
}
