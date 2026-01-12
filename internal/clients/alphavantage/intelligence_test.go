package alphavantage

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntelligenceParseMarketMovers(t *testing.T) {
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
			},
			{
				"ticker": "GOOGL",
				"price": "140.25",
				"change_amount": "4.25",
				"change_percentage": "3.12%",
				"volume": "30000000"
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
		"most_actively_traded": [
			{
				"ticker": "TSLA",
				"price": "220.50",
				"change_amount": "2.50",
				"change_percentage": "1.15%",
				"volume": "100000000"
			}
		]
	}`

	movers, err := parseMarketMovers([]byte(jsonData))
	require.NoError(t, err)

	require.Len(t, movers.TopGainers, 2)
	assert.Equal(t, "AAPL", movers.TopGainers[0].Ticker)
	assert.Equal(t, 185.5, movers.TopGainers[0].Price)
	assert.Equal(t, 5.5, movers.TopGainers[0].ChangeAmount)
	assert.Equal(t, 3.05, movers.TopGainers[0].ChangePercent)
	assert.Equal(t, int64(50000000), movers.TopGainers[0].Volume)

	require.Len(t, movers.TopLosers, 1)
	assert.Equal(t, "MSFT", movers.TopLosers[0].Ticker)
	assert.Equal(t, -10.0, movers.TopLosers[0].ChangeAmount)
	assert.Equal(t, -2.56, movers.TopLosers[0].ChangePercent)

	require.Len(t, movers.MostActive, 1)
	assert.Equal(t, "TSLA", movers.MostActive[0].Ticker)
}

func TestIntelligenceParseTranscript(t *testing.T) {
	jsonData := `{
		"symbol": "IBM",
		"quarter": 4,
		"year": 2023,
		"transcript": "Good afternoon everyone. Welcome to IBM's fourth quarter earnings call..."
	}`

	transcript, err := parseTranscript([]byte(jsonData), "IBM", 2023, 4)
	require.NoError(t, err)

	assert.Equal(t, "IBM", transcript.Symbol)
	assert.Equal(t, 4, transcript.Quarter)
	assert.Equal(t, 2023, transcript.Year)
	assert.Contains(t, transcript.Transcript, "fourth quarter earnings call")
}

func TestIntelligenceParseInsiderTransactions(t *testing.T) {
	jsonData := `{
		"data": [
			{
				"filing_date": "2024-01-12",
				"transaction_date": "2024-01-10",
				"owner_name": "John Smith",
				"owner_title": "CEO",
				"transaction_type": "P-Purchase",
				"acquisition_or_disposition": "A",
				"securities_transacted": "10000",
				"share_price": "150.00",
				"shares_owned_following": "500000"
			},
			{
				"filing_date": "2024-01-10",
				"transaction_date": "2024-01-08",
				"owner_name": "Jane Doe",
				"owner_title": "CFO",
				"transaction_type": "S-Sale",
				"acquisition_or_disposition": "D",
				"securities_transacted": "5000",
				"share_price": "148.50",
				"shares_owned_following": "100000"
			}
		]
	}`

	transactions, err := parseInsiderTransactions([]byte(jsonData), "IBM")
	require.NoError(t, err)
	require.Len(t, transactions, 2)

	assert.Equal(t, "John Smith", transactions[0].OwnerName)
	assert.Equal(t, "CEO", transactions[0].OwnerTitle)
	assert.Equal(t, "P-Purchase", transactions[0].TransactionType)
	assert.Equal(t, int64(10000), transactions[0].SharesTraded)

	assert.Equal(t, "Jane Doe", transactions[1].OwnerName)
	assert.Equal(t, "S-Sale", transactions[1].TransactionType)
}

func TestIntelligenceParseAnalytics(t *testing.T) {
	// Analytics response is keyed by symbol with MEAN/MAX/MIN/STDDEV fields
	jsonData := `{
		"IBM": {
			"MEAN": 185.50,
			"MAX": 190.00,
			"MIN": 180.00,
			"STDDEV": 0.15,
			"MEAN_VOLUME": 5000000,
			"TOTAL_VOLUME": 75000000
		},
		"MSFT": {
			"MEAN": 375.00,
			"MAX": 385.00,
			"MIN": 365.00,
			"STDDEV": 0.12,
			"MEAN_VOLUME": 8000000,
			"TOTAL_VOLUME": 120000000
		}
	}`

	analytics, err := parseAnalytics([]byte(jsonData), []string{"IBM", "MSFT"})
	require.NoError(t, err)
	require.Len(t, analytics, 2)

	// Find IBM in results (order not guaranteed)
	var ibm, msft *AnalyticsWindow
	for i := range analytics {
		if analytics[i].Symbol == "IBM" {
			ibm = &analytics[i]
		}
		if analytics[i].Symbol == "MSFT" {
			msft = &analytics[i]
		}
	}

	require.NotNil(t, ibm)
	assert.Equal(t, 185.50, ibm.AveragePrice)
	assert.Equal(t, 190.00, ibm.HighPrice)
	assert.Equal(t, 180.00, ibm.LowPrice)

	require.NotNil(t, msft)
	assert.Equal(t, 375.00, msft.AveragePrice)
}

func TestIntelligenceParseMarketMovers_InvalidJSON(t *testing.T) {
	_, err := parseMarketMovers([]byte("not json"))
	assert.Error(t, err)
}

func TestIntelligenceParseTranscript_Empty(t *testing.T) {
	jsonData := `{
		"symbol": "XYZ",
		"quarter": 1,
		"year": 2024,
		"transcript": ""
	}`

	transcript, err := parseTranscript([]byte(jsonData), "XYZ", 2024, 1)
	require.NoError(t, err)
	assert.Equal(t, "", transcript.Transcript)
}

func TestIntelligenceParseInsiderTransactions_InvalidJSON(t *testing.T) {
	_, err := parseInsiderTransactions([]byte("not json"), "IBM")
	assert.Error(t, err)
}

func TestIntelligenceParseMarketMovers_Empty(t *testing.T) {
	jsonData := `{
		"metadata": {},
		"top_gainers": [],
		"top_losers": [],
		"most_actively_traded": []
	}`

	movers, err := parseMarketMovers([]byte(jsonData))
	require.NoError(t, err)
	assert.Empty(t, movers.TopGainers)
	assert.Empty(t, movers.TopLosers)
	assert.Empty(t, movers.MostActive)
}

// TestClientIntelligenceMethods verifies all intelligence methods exist.
func TestClientIntelligenceMethods(t *testing.T) {
	client := NewClient("test-key", zerolog.Nop())

	var _ func() (*MarketMovers, error) = client.GetTopGainersLosers
	var _ func(string, int, int) (*Transcript, error) = client.GetEarningsCallTranscript
	var _ func(string) ([]InsiderTransaction, error) = client.GetInsiderTransactions
	var _ func([]string, string, string) ([]AnalyticsWindow, error) = client.GetAnalyticsFixedWindow
	var _ func([]string, int) ([]AnalyticsWindow, error) = client.GetAnalyticsSlidingWindow
}
