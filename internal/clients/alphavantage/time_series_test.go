package alphavantage

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTimeSeriesTestServer creates a test server for time series endpoints.
func setupTimeSeriesTestServer(t *testing.T, responses map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		function := r.URL.Query().Get("function")
		response, ok := responses[function]
		if !ok {
			http.Error(w, "Unknown function", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
}

func TestGetDailyPrices(t *testing.T) {
	responses := map[string]string{
		"TIME_SERIES_DAILY": `{
			"Meta Data": {
				"1. Information": "Daily Prices (open, high, low, close) and Volumes",
				"2. Symbol": "IBM",
				"3. Last Refreshed": "2024-01-15",
				"4. Output Size": "Compact",
				"5. Time Zone": "US/Eastern"
			},
			"Time Series (Daily)": {
				"2024-01-15": {
					"1. open": "185.00",
					"2. high": "186.50",
					"3. low": "184.50",
					"4. close": "186.20",
					"5. volume": "3456789"
				},
				"2024-01-12": {
					"1. open": "184.50",
					"2. high": "185.50",
					"3. low": "184.00",
					"4. close": "185.00",
					"5. volume": "3214567"
				},
				"2024-01-11": {
					"1. open": "183.00",
					"2. high": "185.00",
					"3. low": "182.50",
					"4. close": "184.50",
					"5. volume": "2987654"
				}
			}
		}`,
	}

	server := setupTimeSeriesTestServer(t, responses)
	defer server.Close()

	// Test parsing (direct function call)
	prices, err := parseDailyTimeSeries([]byte(responses["TIME_SERIES_DAILY"]))
	require.NoError(t, err)
	require.Len(t, prices, 3)

	// Verify sorting (newest first)
	assert.Equal(t, 2024, prices[0].Date.Year())
	assert.Equal(t, time.January, prices[0].Date.Month())
	assert.Equal(t, 15, prices[0].Date.Day())

	// Verify data parsing
	assert.Equal(t, 185.00, prices[0].Open)
	assert.Equal(t, 186.50, prices[0].High)
	assert.Equal(t, 184.50, prices[0].Low)
	assert.Equal(t, 186.20, prices[0].Close)
	assert.Equal(t, int64(3456789), prices[0].Volume)
}

func TestGetDailyAdjustedPrices(t *testing.T) {
	jsonData := `{
		"Meta Data": {
			"1. Information": "Daily Time Series with Splits and Dividend Events",
			"2. Symbol": "IBM"
		},
		"Time Series (Daily)": {
			"2024-01-15": {
				"1. open": "185.00",
				"2. high": "186.50",
				"3. low": "184.50",
				"4. close": "186.20",
				"5. adjusted close": "186.20",
				"6. volume": "3456789",
				"7. dividend amount": "0.00",
				"8. split coefficient": "1.0"
			}
		}
	}`

	prices, err := parseAdjustedTimeSeries([]byte(jsonData), "Time Series (Daily)")
	require.NoError(t, err)
	require.Len(t, prices, 1)

	assert.Equal(t, 185.00, prices[0].Open)
	assert.Equal(t, 186.20, prices[0].AdjustedClose)
	assert.Equal(t, 0.00, prices[0].DividendAmount)
	assert.Equal(t, 1.0, prices[0].SplitCoefficient)
}

func TestGetWeeklyPrices(t *testing.T) {
	jsonData := `{
		"Meta Data": {
			"1. Information": "Weekly Prices",
			"2. Symbol": "IBM"
		},
		"Weekly Time Series": {
			"2024-01-12": {
				"1. open": "180.00",
				"2. high": "187.00",
				"3. low": "179.00",
				"4. close": "186.20",
				"5. volume": "15000000"
			},
			"2024-01-05": {
				"1. open": "175.00",
				"2. high": "181.00",
				"3. low": "174.00",
				"4. close": "180.00",
				"5. volume": "12000000"
			}
		}
	}`

	prices, err := parseWeeklyMonthlyTimeSeries([]byte(jsonData), "Weekly Time Series")
	require.NoError(t, err)
	require.Len(t, prices, 2)

	// Verify newest first
	assert.Equal(t, 12, prices[0].Date.Day())
	assert.Equal(t, 186.20, prices[0].Close)
}

func TestGetMonthlyPrices(t *testing.T) {
	jsonData := `{
		"Meta Data": {
			"1. Information": "Monthly Prices",
			"2. Symbol": "IBM"
		},
		"Monthly Time Series": {
			"2024-01-31": {
				"1. open": "175.00",
				"2. high": "190.00",
				"3. low": "170.00",
				"4. close": "186.20",
				"5. volume": "50000000"
			},
			"2023-12-29": {
				"1. open": "165.00",
				"2. high": "180.00",
				"3. low": "160.00",
				"4. close": "175.00",
				"5. volume": "45000000"
			}
		}
	}`

	prices, err := parseWeeklyMonthlyTimeSeries([]byte(jsonData), "Monthly Time Series")
	require.NoError(t, err)
	require.Len(t, prices, 2)

	assert.Equal(t, time.January, prices[0].Date.Month())
	assert.Equal(t, int64(50000000), prices[0].Volume)
}

func TestGetGlobalQuote(t *testing.T) {
	jsonData := `{
		"Global Quote": {
			"01. symbol": "IBM",
			"02. open": "185.0000",
			"03. high": "186.5000",
			"04. low": "184.5000",
			"05. price": "186.2000",
			"06. volume": "3456789",
			"07. latest trading day": "2024-01-15",
			"08. previous close": "185.0000",
			"09. change": "1.2000",
			"10. change percent": "0.6486%"
		}
	}`

	quote, err := parseGlobalQuote([]byte(jsonData))
	require.NoError(t, err)

	assert.Equal(t, "IBM", quote.Symbol)
	assert.Equal(t, 185.00, quote.Open)
	assert.Equal(t, 186.50, quote.High)
	assert.Equal(t, 184.50, quote.Low)
	assert.Equal(t, 186.20, quote.Price)
	assert.Equal(t, int64(3456789), quote.Volume)
	assert.Equal(t, 2024, quote.LatestTradingDay.Year())
	assert.Equal(t, 185.00, quote.PreviousClose)
	assert.Equal(t, 1.20, quote.Change)
	assert.Equal(t, 0.6486, quote.ChangePercent)
}

func TestSearchSymbol(t *testing.T) {
	jsonData := `{
		"bestMatches": [
			{
				"1. symbol": "IBM",
				"2. name": "International Business Machines Corp",
				"3. type": "Equity",
				"4. region": "United States",
				"5. marketOpen": "09:30",
				"6. marketClose": "16:00",
				"7. timezone": "UTC-05",
				"8. currency": "USD",
				"9. matchScore": "1.0000"
			},
			{
				"1. symbol": "IBM.LON",
				"2. name": "International Business Machines Corp",
				"3. type": "Equity",
				"4. region": "United Kingdom",
				"5. marketOpen": "08:00",
				"6. marketClose": "16:30",
				"7. timezone": "UTC+00",
				"8. currency": "GBP",
				"9. matchScore": "0.8000"
			}
		]
	}`

	matches, err := parseSymbolSearch([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, matches, 2)

	assert.Equal(t, "IBM", matches[0].Symbol)
	assert.Equal(t, "International Business Machines Corp", matches[0].Name)
	assert.Equal(t, "Equity", matches[0].Type)
	assert.Equal(t, "United States", matches[0].Region)
	assert.Equal(t, "09:30", matches[0].MarketOpen)
	assert.Equal(t, "16:00", matches[0].MarketClose)
	assert.Equal(t, "USD", matches[0].Currency)
	assert.Equal(t, "1.0000", matches[0].MatchScore)

	assert.Equal(t, "IBM.LON", matches[1].Symbol)
	assert.Equal(t, "GBP", matches[1].Currency)
	assert.Equal(t, "0.8000", matches[1].MatchScore)
}

func TestParseTimeSeriesEmpty(t *testing.T) {
	// Test handling of empty response - should return error because nil check
	jsonData := `{
		"Meta Data": {
			"1. Information": "Daily Prices",
			"2. Symbol": "XYZ"
		},
		"Time Series (Daily)": {}
	}`

	prices, err := parseDailyTimeSeries([]byte(jsonData))
	// Empty map is still valid, just no data
	require.NoError(t, err)
	assert.Empty(t, prices)
}

func TestParseTimeSeriesInvalidJSON(t *testing.T) {
	_, err := parseDailyTimeSeries([]byte("not valid json"))
	assert.Error(t, err)
}

func TestParseGlobalQuoteEmpty(t *testing.T) {
	jsonData := `{
		"Global Quote": {}
	}`

	_, err := parseGlobalQuote([]byte(jsonData))
	// Empty quote returns error
	assert.Error(t, err)
}

func TestParseSymbolSearchEmpty(t *testing.T) {
	jsonData := `{
		"bestMatches": []
	}`

	matches, err := parseSymbolSearch([]byte(jsonData))
	require.NoError(t, err)
	assert.Empty(t, matches)
}

func TestTimeSeriesDateSorting(t *testing.T) {
	// Verify dates are sorted newest first
	jsonData := `{
		"Meta Data": {},
		"Time Series (Daily)": {
			"2024-01-01": {"1. open": "100", "2. high": "100", "3. low": "100", "4. close": "100", "5. volume": "1000"},
			"2024-01-15": {"1. open": "100", "2. high": "100", "3. low": "100", "4. close": "100", "5. volume": "1000"},
			"2024-01-10": {"1. open": "100", "2. high": "100", "3. low": "100", "4. close": "100", "5. volume": "1000"}
		}
	}`

	prices, err := parseDailyTimeSeries([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, prices, 3)

	// Should be sorted: 2024-01-15, 2024-01-10, 2024-01-01
	assert.Equal(t, 15, prices[0].Date.Day())
	assert.Equal(t, 10, prices[1].Date.Day())
	assert.Equal(t, 1, prices[2].Date.Day())
}

func TestParseTimeSeriesMissingKey(t *testing.T) {
	jsonData := `{
		"Meta Data": {},
		"Some Other Key": {}
	}`

	_, err := parseDailyTimeSeries([]byte(jsonData))
	assert.Error(t, err)
}

func TestParseWeeklyMonthlyMissingKey(t *testing.T) {
	jsonData := `{
		"Meta Data": {}
	}`

	_, err := parseWeeklyMonthlyTimeSeries([]byte(jsonData), "Weekly Time Series")
	assert.Error(t, err)
}

func TestParseAdjustedTimeSeriesMissingKey(t *testing.T) {
	jsonData := `{
		"Meta Data": {}
	}`

	_, err := parseAdjustedTimeSeries([]byte(jsonData), "Time Series (Daily)")
	assert.Error(t, err)
}

// TestClientMethods tests that client methods exist and have correct signatures.
func TestClientTimeSeriesMethods(t *testing.T) {
	client := newTestClient("test-key")

	// Verify methods exist (compile-time check through type assertions)
	var _ func(string, bool) ([]DailyPrice, error) = client.GetDailyPrices
	var _ func(string, bool) ([]AdjustedPrice, error) = client.GetDailyAdjustedPrices
	var _ func(string) ([]DailyPrice, error) = client.GetWeeklyPrices
	var _ func(string) ([]AdjustedPrice, error) = client.GetWeeklyAdjustedPrices
	var _ func(string) ([]DailyPrice, error) = client.GetMonthlyPrices
	var _ func(string) ([]AdjustedPrice, error) = client.GetMonthlyAdjustedPrices
	var _ func(string) (*GlobalQuote, error) = client.GetGlobalQuote
	var _ func(string) ([]SymbolMatch, error) = client.SearchSymbol
}

// BenchmarkParseDailyTimeSeries benchmarks time series parsing.
func BenchmarkParseDailyTimeSeries(b *testing.B) {
	jsonData := []byte(`{
		"Meta Data": {"1. Information": "Daily Prices", "2. Symbol": "IBM"},
		"Time Series (Daily)": {
			"2024-01-15": {"1. open": "185.00", "2. high": "186.50", "3. low": "184.50", "4. close": "186.20", "5. volume": "3456789"},
			"2024-01-14": {"1. open": "184.50", "2. high": "185.50", "3. low": "184.00", "4. close": "185.00", "5. volume": "3214567"},
			"2024-01-13": {"1. open": "183.00", "2. high": "185.00", "3. low": "182.50", "4. close": "184.50", "5. volume": "2987654"}
		}
	}`)

	for i := 0; i < b.N; i++ {
		_, _ = parseDailyTimeSeries(jsonData)
	}
}
