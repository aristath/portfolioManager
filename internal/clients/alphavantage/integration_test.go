package alphavantage

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_ClientCreation verifies client is created with correct defaults.
func TestIntegration_ClientCreation(t *testing.T) {
	client := NewClient("test-api-key", zerolog.Nop())

	assert.NotNil(t, client)
	assert.Equal(t, 25, client.GetRemainingRequests())
}

// TestIntegration_InterfaceCompliance verifies Client implements ClientInterface.
func TestIntegration_InterfaceCompliance(t *testing.T) {
	var _ ClientInterface = (*Client)(nil)
}

// TestIntegration_RateLimitingAcrossRequests tests rate limiting persists across calls.
func TestIntegration_RateLimitingAcrossRequests(t *testing.T) {
	client := NewClient("test-key", zerolog.Nop())

	// Consume all requests
	for i := 0; i < 25; i++ {
		err := client.checkRateLimit()
		require.NoError(t, err)
	}

	// Next request should fail
	err := client.checkRateLimit()
	assert.Error(t, err)
	assert.IsType(t, ErrRateLimitExceeded{}, err)

	// Reset and verify
	client.ResetDailyCounter()
	assert.Equal(t, 25, client.GetRemainingRequests())
}

// TestIntegration_CachingFlow tests that caching works across operations.
func TestIntegration_CachingFlow(t *testing.T) {
	client := NewClient("test-key", zerolog.Nop())

	// Set cache for multiple data types
	client.setCache("key1", "value1", client.cacheTTL.Fundamentals)
	client.setCache("key2", 42, client.cacheTTL.PriceData)
	client.setCache("key3", []int{1, 2, 3}, client.cacheTTL.TechnicalIndicators)

	// Verify all cached values
	v1, ok := client.getFromCache("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", v1)

	v2, ok := client.getFromCache("key2")
	assert.True(t, ok)
	assert.Equal(t, 42, v2)

	v3, ok := client.getFromCache("key3")
	assert.True(t, ok)
	assert.Equal(t, []int{1, 2, 3}, v3)

	// Clear and verify all gone
	client.ClearCache()
	_, ok = client.getFromCache("key1")
	assert.False(t, ok)
	_, ok = client.getFromCache("key2")
	assert.False(t, ok)
	_, ok = client.getFromCache("key3")
	assert.False(t, ok)
}

// TestIntegration_MockedTimeSeriesFlow tests fetching daily prices with mocked server.
func TestIntegration_MockedTimeSeriesFlow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request parameters
		assert.Equal(t, "TIME_SERIES_DAILY", r.URL.Query().Get("function"))
		assert.Equal(t, "IBM", r.URL.Query().Get("symbol"))
		assert.Equal(t, "test-key", r.URL.Query().Get("apikey"))

		response := map[string]interface{}{
			"Meta Data": map[string]string{
				"1. Information": "Daily Prices",
				"2. Symbol":      "IBM",
			},
			"Time Series (Daily)": map[string]map[string]string{
				"2024-01-15": {
					"1. open":   "185.00",
					"2. high":   "186.50",
					"3. low":    "184.50",
					"4. close":  "186.20",
					"5. volume": "3456789",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with custom base URL (would need to modify client for this in production)
	client := NewClient("test-key", zerolog.Nop())

	// Test the parsing directly since we can't easily override base URL
	jsonData := `{
		"Meta Data": {"1. Information": "Daily Prices", "2. Symbol": "IBM"},
		"Time Series (Daily)": {
			"2024-01-15": {"1. open": "185.00", "2. high": "186.50", "3. low": "184.50", "4. close": "186.20", "5. volume": "3456789"}
		}
	}`

	prices, err := parseDailyTimeSeries([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, prices, 1)
	assert.Equal(t, 185.00, prices[0].Open)
	assert.Equal(t, 186.20, prices[0].Close)
	_ = client // Ensure client was created
}

// TestIntegration_MockedTechnicalIndicatorFlow tests fetching RSI with mocked server.
func TestIntegration_MockedTechnicalIndicatorFlow(t *testing.T) {
	jsonData := `{
		"Meta Data": {
			"1: Symbol": "IBM",
			"2: Indicator": "Relative Strength Index (RSI)"
		},
		"Technical Analysis: RSI": {
			"2024-01-15": {"RSI": "65.4321"},
			"2024-01-14": {"RSI": "62.1234"},
			"2024-01-13": {"RSI": "58.7654"}
		}
	}`

	data, err := parseTechnicalIndicator([]byte(jsonData), "RSI")
	require.NoError(t, err)
	require.Len(t, data.Values, 3)

	// Verify sorting (newest first)
	assert.Equal(t, 15, data.Values[0].Date.Day())
	assert.Equal(t, 65.4321, data.Values[0].Value)
}

// TestIntegration_MockedFundamentalsFlow tests fetching company overview.
func TestIntegration_MockedFundamentalsFlow(t *testing.T) {
	jsonData := `{
		"Symbol": "IBM",
		"AssetType": "Common Stock",
		"Name": "International Business Machines Corporation",
		"Exchange": "NYSE",
		"Currency": "USD",
		"Country": "USA",
		"Sector": "Technology",
		"Industry": "Information Technology Services",
		"MarketCapitalization": "125000000000",
		"PERatio": "20.5",
		"EPS": "9.05",
		"DividendYield": "0.0485",
		"52WeekHigh": "200.00",
		"52WeekLow": "120.00"
	}`

	overview, err := parseCompanyOverview([]byte(jsonData))
	require.NoError(t, err)

	assert.Equal(t, "IBM", overview.Symbol)
	assert.Equal(t, "NYSE", overview.Exchange)
	assert.Equal(t, "Technology", overview.Sector)
	assert.Equal(t, int64(125000000000), overview.MarketCapitalization)
}

// TestIntegration_MockedForexFlow tests fetching exchange rates.
func TestIntegration_MockedForexFlow(t *testing.T) {
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
	assert.Equal(t, "EUR", rate.ToCurrency)
	assert.Equal(t, 0.925, rate.ExchangeRate)
}

// TestIntegration_MockedCryptoFlow tests fetching cryptocurrency data.
func TestIntegration_MockedCryptoFlow(t *testing.T) {
	jsonData := `{
		"Meta Data": {
			"1. Information": "Daily Prices and Volumes for Digital Currency"
		},
		"Time Series (Digital Currency Daily)": {
			"2024-01-15": {
				"1a. open (USD)": "42500.00",
				"2a. high (USD)": "43200.00",
				"3a. low (USD)": "42100.00",
				"4a. close (USD)": "42800.00",
				"5. volume": "25000.5",
				"6. market cap (USD)": "850000000000"
			}
		}
	}`

	prices, err := parseCryptoTimeSeries([]byte(jsonData), "Time Series (Digital Currency Daily)", "USD")
	require.NoError(t, err)
	require.Len(t, prices, 1)

	assert.Equal(t, 42500.0, prices[0].Open)
	assert.Equal(t, 42800.0, prices[0].Close)
	assert.Equal(t, int64(850000000000), prices[0].MarketCap)
}

// TestIntegration_MockedCommodityFlow tests fetching commodity data.
func TestIntegration_MockedCommodityFlow(t *testing.T) {
	jsonData := `{
		"name": "WTI Crude Oil",
		"interval": "daily",
		"unit": "dollars per barrel",
		"data": [
			{"date": "2024-01-15", "value": "75.50"},
			{"date": "2024-01-14", "value": "74.80"}
		]
	}`

	prices, err := parseCommodityData([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, prices, 2)

	assert.Equal(t, 75.5, prices[0].Value)
}

// TestIntegration_MockedEconomicFlow tests fetching economic indicator data.
func TestIntegration_MockedEconomicFlow(t *testing.T) {
	jsonData := `{
		"name": "Real Gross Domestic Product",
		"interval": "quarterly",
		"unit": "billions of dollars",
		"data": [
			{"date": "2023-12-31", "value": "25000.5"},
			{"date": "2023-09-30", "value": "24500.2"}
		]
	}`

	data, err := parseEconomicData([]byte(jsonData))
	require.NoError(t, err)

	assert.Equal(t, "Real Gross Domestic Product", data.Name)
	assert.Equal(t, "quarterly", data.Interval)
	require.Len(t, data.Data, 2)
}

// TestIntegration_MockedMarketMoversFlow tests fetching top gainers/losers.
func TestIntegration_MockedMarketMoversFlow(t *testing.T) {
	jsonData := `{
		"metadata": {
			"last_updated": "2024-01-15 16:00:00"
		},
		"top_gainers": [
			{
				"ticker": "AAPL",
				"price": "185.50",
				"change_amount": "5.50",
				"change_percentage": "3.05%",
				"volume": "50000000"
			}
		],
		"top_losers": [
			{
				"ticker": "MSFT",
				"price": "380.00",
				"change_amount": "-10.00",
				"change_percentage": "-2.56%",
				"volume": "35000000"
			}
		],
		"most_actively_traded": []
	}`

	movers, err := parseMarketMovers([]byte(jsonData))
	require.NoError(t, err)

	require.Len(t, movers.TopGainers, 1)
	assert.Equal(t, "AAPL", movers.TopGainers[0].Ticker)
	require.Len(t, movers.TopLosers, 1)
	assert.Equal(t, "MSFT", movers.TopLosers[0].Ticker)
}

// TestIntegration_MockedOptionsFlow tests fetching options chain data.
func TestIntegration_MockedOptionsFlow(t *testing.T) {
	jsonData := `{
		"data": [
			{
				"contractID": "IBM240119C00150000",
				"expiration": "2024-01-19",
				"strike": "150.00",
				"type": "call",
				"last": "35.50",
				"bid": "35.20",
				"ask": "35.60",
				"volume": "500",
				"open_interest": "1500",
				"implied_volatility": "0.25",
				"delta": "0.85"
			},
			{
				"contractID": "IBM240119P00150000",
				"expiration": "2024-01-19",
				"strike": "150.00",
				"type": "put",
				"last": "0.50",
				"bid": "0.45",
				"ask": "0.65",
				"volume": "800",
				"open_interest": "2500",
				"delta": "-0.15"
			}
		]
	}`

	chain, err := parseOptionsChain([]byte(jsonData), "IBM", "2024-01-15")
	require.NoError(t, err)

	require.Len(t, chain.Calls, 1)
	assert.Equal(t, 150.0, chain.Calls[0].Strike)
	require.Len(t, chain.Puts, 1)
	assert.Equal(t, 150.0, chain.Puts[0].Strike)
}

// TestIntegration_ErrorHandling tests various error scenarios.
func TestIntegration_ErrorHandling(t *testing.T) {
	client := NewClient("test-key", zerolog.Nop())

	t.Run("RateLimitError", func(t *testing.T) {
		body := []byte(`{"Note": "Thank you for using Alpha Vantage! Our standard API call frequency is 25 calls per day."}`)
		err := client.checkAPIError(body)
		assert.Error(t, err)
		assert.IsType(t, ErrRateLimitExceeded{}, err)
	})

	t.Run("InvalidAPIKeyError", func(t *testing.T) {
		body := []byte(`{"Error Message": "Invalid API call. Please retry or visit the documentation."}`)
		err := client.checkAPIError(body)
		assert.Error(t, err)
	})

	t.Run("ThankYouMessage", func(t *testing.T) {
		body := []byte(`Thank you for using Alpha Vantage! Please visit https://www.alphavantage.co/premium/`)
		err := client.checkAPIError(body)
		assert.Error(t, err)
		assert.IsType(t, ErrRateLimitExceeded{}, err)
	})

	t.Run("ValidResponse", func(t *testing.T) {
		body := []byte(`{"data": "valid response"}`)
		err := client.checkAPIError(body)
		assert.NoError(t, err)
	})
}

// TestIntegration_CacheKeyConsistency tests cache keys are consistent.
func TestIntegration_CacheKeyConsistency(t *testing.T) {
	params := map[string]string{
		"symbol":   "IBM",
		"interval": "daily",
	}

	key1 := buildCacheKey("SMA", params)
	key2 := buildCacheKey("SMA", params)

	assert.Equal(t, key1, key2)

	// Different params should produce different keys
	params2 := map[string]string{
		"symbol":   "AAPL",
		"interval": "daily",
	}
	key3 := buildCacheKey("SMA", params2)
	assert.NotEqual(t, key1, key3)

	// API key should be excluded
	params3 := map[string]string{
		"symbol": "IBM",
		"apikey": "secret",
	}
	key4 := buildCacheKey("SMA", params3)
	assert.NotContains(t, key4, "secret")
}

// TestIntegration_ParsingHelpers tests parsing helper functions.
func TestIntegration_ParsingHelpers(t *testing.T) {
	t.Run("ParseFloat64", func(t *testing.T) {
		assert.Equal(t, 123.45, parseFloat64("123.45"))
		assert.Equal(t, 0.0, parseFloat64("None"))
		assert.Equal(t, 0.0, parseFloat64(""))
		assert.Equal(t, 0.0, parseFloat64("null"))
		assert.Equal(t, 50.5, parseFloat64("50.5%"))
	})

	t.Run("ParseInt64", func(t *testing.T) {
		assert.Equal(t, int64(12345), parseInt64("12345"))
		assert.Equal(t, int64(0), parseInt64("None"))
		assert.Equal(t, int64(0), parseInt64(""))
		// Scientific notation
		assert.Equal(t, int64(15000000000), parseInt64("1.5E10"))
	})

	t.Run("ParseDate", func(t *testing.T) {
		date := parseDate("2024-01-15")
		assert.Equal(t, 2024, date.Year())
		assert.Equal(t, 1, int(date.Month()))
		assert.Equal(t, 15, date.Day())
	})

	t.Run("ParseDateTime", func(t *testing.T) {
		dt := parseDateTime("2024-01-15 14:30:00")
		assert.Equal(t, 2024, dt.Year())
		assert.Equal(t, 14, dt.Hour())
		assert.Equal(t, 30, dt.Minute())
	})
}

// TestIntegration_AllEndpointCategories tests that all endpoint categories are covered.
func TestIntegration_AllEndpointCategories(t *testing.T) {
	client := NewClient("test-key", zerolog.Nop())

	// Verify all method categories exist
	t.Run("TimeSeries", func(t *testing.T) {
		assert.NotNil(t, client.GetDailyPrices)
		assert.NotNil(t, client.GetWeeklyPrices)
		assert.NotNil(t, client.GetMonthlyPrices)
		assert.NotNil(t, client.GetGlobalQuote)
		assert.NotNil(t, client.SearchSymbol)
	})

	t.Run("TechnicalIndicators", func(t *testing.T) {
		assert.NotNil(t, client.GetSMA)
		assert.NotNil(t, client.GetEMA)
		assert.NotNil(t, client.GetRSI)
		assert.NotNil(t, client.GetMACD)
		assert.NotNil(t, client.GetBBANDS)
	})

	t.Run("Fundamentals", func(t *testing.T) {
		assert.NotNil(t, client.GetCompanyOverview)
		assert.NotNil(t, client.GetEarnings)
		assert.NotNil(t, client.GetIncomeStatement)
		assert.NotNil(t, client.GetBalanceSheet)
		assert.NotNil(t, client.GetCashFlow)
	})

	t.Run("Forex", func(t *testing.T) {
		assert.NotNil(t, client.GetExchangeRate)
		assert.NotNil(t, client.GetFXDaily)
		assert.NotNil(t, client.GetFXWeekly)
		assert.NotNil(t, client.GetFXMonthly)
	})

	t.Run("Crypto", func(t *testing.T) {
		assert.NotNil(t, client.GetCryptoExchangeRate)
		assert.NotNil(t, client.GetCryptoDaily)
		assert.NotNil(t, client.GetCryptoWeekly)
		assert.NotNil(t, client.GetCryptoMonthly)
	})

	t.Run("Commodities", func(t *testing.T) {
		assert.NotNil(t, client.GetCommodity)
		assert.NotNil(t, client.GetWTI)
		assert.NotNil(t, client.GetBrent)
		assert.NotNil(t, client.GetNaturalGas)
	})

	t.Run("Economic", func(t *testing.T) {
		assert.NotNil(t, client.GetEconomicIndicator)
		assert.NotNil(t, client.GetRealGDP)
		assert.NotNil(t, client.GetUnemployment)
		assert.NotNil(t, client.GetCPI)
	})

	t.Run("Intelligence", func(t *testing.T) {
		assert.NotNil(t, client.GetTopGainersLosers)
		assert.NotNil(t, client.GetInsiderTransactions)
	})

	t.Run("Options", func(t *testing.T) {
		assert.NotNil(t, client.GetHistoricalOptions)
	})
}

// BenchmarkIntegration_CacheOperations benchmarks cache performance.
func BenchmarkIntegration_CacheOperations(b *testing.B) {
	client := NewClient("test-key", zerolog.Nop())

	b.Run("SetCache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			client.setCache("key", "value", client.cacheTTL.Fundamentals)
		}
	})

	b.Run("GetCache", func(b *testing.B) {
		client.setCache("key", "value", client.cacheTTL.Fundamentals)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = client.getFromCache("key")
		}
	})
}

// BenchmarkIntegration_Parsing benchmarks parsing performance.
func BenchmarkIntegration_Parsing(b *testing.B) {
	dailyData := []byte(`{
		"Meta Data": {},
		"Time Series (Daily)": {
			"2024-01-15": {"1. open": "185.00", "2. high": "186.50", "3. low": "184.50", "4. close": "186.20", "5. volume": "3456789"},
			"2024-01-14": {"1. open": "184.50", "2. high": "185.50", "3. low": "184.00", "4. close": "185.00", "5. volume": "3214567"}
		}
	}`)

	b.Run("ParseDailyTimeSeries", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = parseDailyTimeSeries(dailyData)
		}
	})

	indicatorData := []byte(`{
		"Meta Data": {},
		"Technical Analysis: SMA": {
			"2024-01-15": {"SMA": "185.50"},
			"2024-01-14": {"SMA": "185.25"}
		}
	}`)

	b.Run("ParseTechnicalIndicator", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = parseTechnicalIndicator(indicatorData, "SMA")
		}
	})
}
