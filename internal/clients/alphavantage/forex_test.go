package alphavantage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForexParseExchangeRate(t *testing.T) {
	jsonData := `{
		"Realtime Currency Exchange Rate": {
			"1. From_Currency Code": "USD",
			"2. From_Currency Name": "United States Dollar",
			"3. To_Currency Code": "EUR",
			"4. To_Currency Name": "Euro",
			"5. Exchange Rate": "0.9250",
			"6. Last Refreshed": "2024-01-15 14:30:00",
			"7. Time Zone": "UTC",
			"8. Bid Price": "0.9248",
			"9. Ask Price": "0.9252"
		}
	}`

	rate, err := parseExchangeRate([]byte(jsonData))
	require.NoError(t, err)

	assert.Equal(t, "USD", rate.FromCurrency)
	assert.Equal(t, "United States Dollar", rate.FromCurrencyName)
	assert.Equal(t, "EUR", rate.ToCurrency)
	assert.Equal(t, "Euro", rate.ToCurrencyName)
	assert.Equal(t, 0.925, rate.ExchangeRate)
	assert.Equal(t, 0.9248, rate.BidPrice)
	assert.Equal(t, 0.9252, rate.AskPrice)
}

func TestForexParseFXTimeSeries(t *testing.T) {
	jsonData := `{
		"Meta Data": {
			"1. Information": "Forex Daily Prices",
			"2. From Symbol": "EUR",
			"3. To Symbol": "USD"
		},
		"Time Series FX (Daily)": {
			"2024-01-15": {
				"1. open": "1.0950",
				"2. high": "1.0980",
				"3. low": "1.0920",
				"4. close": "1.0965"
			},
			"2024-01-12": {
				"1. open": "1.0930",
				"2. high": "1.0960",
				"3. low": "1.0900",
				"4. close": "1.0950"
			}
		}
	}`

	prices, err := parseFXTimeSeries([]byte(jsonData), "Time Series FX (Daily)")
	require.NoError(t, err)
	require.Len(t, prices, 2)

	// Verify sorting (newest first)
	assert.Equal(t, 15, prices[0].Date.Day())
	assert.Equal(t, 1.095, prices[0].Open)
	assert.Equal(t, 1.098, prices[0].High)
	assert.Equal(t, 1.092, prices[0].Low)
	assert.Equal(t, 1.0965, prices[0].Close)
}

func TestForexParseFXTimeSeries_Weekly(t *testing.T) {
	jsonData := `{
		"Meta Data": {},
		"Time Series FX (Weekly)": {
			"2024-01-12": {
				"1. open": "1.0900",
				"2. high": "1.1000",
				"3. low": "1.0850",
				"4. close": "1.0965"
			}
		}
	}`

	prices, err := parseFXTimeSeries([]byte(jsonData), "Time Series FX (Weekly)")
	require.NoError(t, err)
	require.Len(t, prices, 1)

	assert.Equal(t, 1.09, prices[0].Open)
	assert.Equal(t, 1.10, prices[0].High)
}

func TestForexParseFXTimeSeries_Monthly(t *testing.T) {
	jsonData := `{
		"Meta Data": {},
		"Time Series FX (Monthly)": {
			"2024-01-31": {
				"1. open": "1.0800",
				"2. high": "1.1100",
				"3. low": "1.0700",
				"4. close": "1.0965"
			}
		}
	}`

	prices, err := parseFXTimeSeries([]byte(jsonData), "Time Series FX (Monthly)")
	require.NoError(t, err)
	require.Len(t, prices, 1)

	assert.Equal(t, 1.08, prices[0].Open)
}

func TestForexParseExchangeRate_InvalidJSON(t *testing.T) {
	_, err := parseExchangeRate([]byte("not json"))
	assert.Error(t, err)
}

func TestForexParseFXTimeSeries_InvalidJSON(t *testing.T) {
	_, err := parseFXTimeSeries([]byte("not json"), "Time Series FX (Daily)")
	assert.Error(t, err)
}

func TestForexParseFXTimeSeries_MissingKey(t *testing.T) {
	jsonData := `{
		"Meta Data": {}
	}`

	_, err := parseFXTimeSeries([]byte(jsonData), "Time Series FX (Daily)")
	assert.Error(t, err)
}

func TestForexDateSorting(t *testing.T) {
	jsonData := `{
		"Meta Data": {},
		"Time Series FX (Daily)": {
			"2024-01-01": {"1. open": "1.00", "2. high": "1.00", "3. low": "1.00", "4. close": "1.00"},
			"2024-01-15": {"1. open": "1.00", "2. high": "1.00", "3. low": "1.00", "4. close": "1.00"},
			"2024-01-10": {"1. open": "1.00", "2. high": "1.00", "3. low": "1.00", "4. close": "1.00"}
		}
	}`

	prices, err := parseFXTimeSeries([]byte(jsonData), "Time Series FX (Daily)")
	require.NoError(t, err)
	require.Len(t, prices, 3)

	// Should be sorted newest first
	assert.Equal(t, 15, prices[0].Date.Day())
	assert.Equal(t, 10, prices[1].Date.Day())
	assert.Equal(t, 1, prices[2].Date.Day())
}

// TestClientForexMethods verifies all forex methods exist.
func TestClientForexMethods(t *testing.T) {
	client := newTestClient("test-key")

	var _ func(string, string) (*ExchangeRate, error) = client.GetExchangeRate
	var _ func(string, string, bool) ([]FXPrice, error) = client.GetFXDaily
	var _ func(string, string) ([]FXPrice, error) = client.GetFXWeekly
	var _ func(string, string) ([]FXPrice, error) = client.GetFXMonthly
}
