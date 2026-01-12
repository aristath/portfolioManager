package alphavantage

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// =============================================================================
// Alpha Intelligence Endpoints
// =============================================================================

// GetTopGainersLosers returns top gainers, losers, and most active stocks.
func (c *Client) GetTopGainersLosers() (*MarketMovers, error) {
	cacheKey := buildCacheKey("TOP_GAINERS_LOSERS", nil)

	if cached, ok := c.getFromCache(cacheKey); ok {
		return cached.(*MarketMovers), nil
	}

	body, err := c.doRequest("TOP_GAINERS_LOSERS", nil)
	if err != nil {
		return nil, err
	}

	movers, err := parseMarketMovers(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse market movers: %w", err)
	}

	// Cache for 15 minutes as this is real-time data
	c.setCache(cacheKey, movers, c.cacheTTL.PriceData)
	return movers, nil
}

// GetEarningsCallTranscript returns the transcript for an earnings call.
func (c *Client) GetEarningsCallTranscript(symbol string, year, quarter int) (*Transcript, error) {
	params := map[string]string{
		"symbol":  symbol,
		"year":    strconv.Itoa(year),
		"quarter": strconv.Itoa(quarter),
	}

	cacheKey := buildCacheKey("EARNINGS_CALL_TRANSCRIPT", params)

	if cached, ok := c.getFromCache(cacheKey); ok {
		return cached.(*Transcript), nil
	}

	body, err := c.doRequest("EARNINGS_CALL_TRANSCRIPT", params)
	if err != nil {
		return nil, err
	}

	transcript, err := parseTranscript(body, symbol, year, quarter)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transcript: %w", err)
	}

	c.setCache(cacheKey, transcript, c.cacheTTL.Fundamentals)
	return transcript, nil
}

// GetInsiderTransactions returns insider trading data for a symbol.
func (c *Client) GetInsiderTransactions(symbol string) ([]InsiderTransaction, error) {
	params := map[string]string{
		"symbol": symbol,
	}

	cacheKey := buildCacheKey("INSIDER_TRANSACTIONS", params)

	if cached, ok := c.getFromCache(cacheKey); ok {
		return cached.([]InsiderTransaction), nil
	}

	body, err := c.doRequest("INSIDER_TRANSACTIONS", params)
	if err != nil {
		return nil, err
	}

	transactions, err := parseInsiderTransactions(body, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to parse insider transactions: %w", err)
	}

	c.setCache(cacheKey, transactions, c.cacheTTL.Fundamentals)
	return transactions, nil
}

// GetAnalyticsFixedWindow returns analytics for a fixed time window.
func (c *Client) GetAnalyticsFixedWindow(symbols []string, startDate, endDate string) ([]AnalyticsWindow, error) {
	params := map[string]string{
		"SYMBOLS":      strings.Join(symbols, ","),
		"RANGE":        startDate + " " + endDate,
		"CALCULATIONS": "MEAN,MIN,MAX,STDDEV",
	}

	cacheKey := buildCacheKey("ANALYTICS_FIXED_WINDOW", params)

	if cached, ok := c.getFromCache(cacheKey); ok {
		return cached.([]AnalyticsWindow), nil
	}

	body, err := c.doRequest("ANALYTICS_FIXED_WINDOW", params)
	if err != nil {
		return nil, err
	}

	analytics, err := parseAnalytics(body, symbols)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fixed window analytics: %w", err)
	}

	c.setCache(cacheKey, analytics, c.cacheTTL.TechnicalIndicators)
	return analytics, nil
}

// GetAnalyticsSlidingWindow returns analytics for a sliding time window.
func (c *Client) GetAnalyticsSlidingWindow(symbols []string, windowSize int) ([]AnalyticsWindow, error) {
	params := map[string]string{
		"SYMBOLS":      strings.Join(symbols, ","),
		"WINDOW_SIZE":  strconv.Itoa(windowSize),
		"CALCULATIONS": "MEAN,MIN,MAX,STDDEV",
	}

	cacheKey := buildCacheKey("ANALYTICS_SLIDING_WINDOW", params)

	if cached, ok := c.getFromCache(cacheKey); ok {
		return cached.([]AnalyticsWindow), nil
	}

	body, err := c.doRequest("ANALYTICS_SLIDING_WINDOW", params)
	if err != nil {
		return nil, err
	}

	analytics, err := parseAnalytics(body, symbols)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sliding window analytics: %w", err)
	}

	c.setCache(cacheKey, analytics, c.cacheTTL.TechnicalIndicators)
	return analytics, nil
}

// =============================================================================
// Parsing Functions
// =============================================================================

func parseMarketMovers(body []byte) (*MarketMovers, error) {
	var response struct {
		Metadata   map[string]string   `json:"metadata"`
		TopGainers []map[string]string `json:"top_gainers"`
		TopLosers  []map[string]string `json:"top_losers"`
		MostActive []map[string]string `json:"most_actively_traded"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	parseMovers := func(data []map[string]string) []MarketMover {
		movers := make([]MarketMover, 0, len(data))
		for _, m := range data {
			changePercent := strings.TrimSuffix(m["change_percentage"], "%")
			movers = append(movers, MarketMover{
				Ticker:        m["ticker"],
				Price:         parseFloat64(m["price"]),
				ChangeAmount:  parseFloat64(m["change_amount"]),
				ChangePercent: parseFloat64(changePercent),
				Volume:        parseInt64(m["volume"]),
			})
		}
		return movers
	}

	lastUpdated := parseDateTime(response.Metadata["last_updated"])

	return &MarketMovers{
		LastUpdated: lastUpdated,
		TopGainers:  parseMovers(response.TopGainers),
		TopLosers:   parseMovers(response.TopLosers),
		MostActive:  parseMovers(response.MostActive),
	}, nil
}

func parseTranscript(body []byte, symbol string, year, quarter int) (*Transcript, error) {
	var response struct {
		Symbol     string `json:"symbol"`
		Quarter    int    `json:"quarter"`
		Year       int    `json:"year"`
		Transcript string `json:"transcript"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		// The transcript might be returned as plain text
		return &Transcript{
			Symbol:     symbol,
			Quarter:    quarter,
			Year:       year,
			Transcript: string(body),
		}, nil
	}

	return &Transcript{
		Symbol:     response.Symbol,
		Quarter:    response.Quarter,
		Year:       response.Year,
		Transcript: response.Transcript,
	}, nil
}

func parseInsiderTransactions(body []byte, symbol string) ([]InsiderTransaction, error) {
	var response struct {
		Data []map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	transactions := make([]InsiderTransaction, 0, len(response.Data))
	for _, t := range response.Data {
		getString := func(key string) string {
			if v, ok := t[key].(string); ok {
				return v
			}
			return ""
		}

		getFloat64Ptr := func(key string) *float64 {
			switch v := t[key].(type) {
			case float64:
				return &v
			case string:
				return parseFloat64Ptr(v)
			}
			return nil
		}

		getInt64 := func(key string) int64 {
			switch v := t[key].(type) {
			case float64:
				return int64(v)
			case string:
				return parseInt64(v)
			}
			return 0
		}

		transactions = append(transactions, InsiderTransaction{
			Symbol:                 symbol,
			FilingDate:             parseDate(getString("filing_date")),
			TransactionDate:        parseDate(getString("transaction_date")),
			OwnerName:              getString("owner_name"),
			OwnerTitle:             getString("owner_title"),
			TransactionType:        getString("transaction_type"),
			AcquisitionDisposition: getString("acquisition_or_disposition"),
			SharesTraded:           getInt64("securities_transacted"),
			Price:                  getFloat64Ptr("share_price"),
			SharesOwned:            getInt64("shares_owned_following"),
		})
	}

	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].FilingDate.After(transactions[j].FilingDate)
	})

	return transactions, nil
}

func parseAnalytics(body []byte, symbols []string) ([]AnalyticsWindow, error) {
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	analytics := make([]AnalyticsWindow, 0, len(symbols))

	// The response structure varies by endpoint, so we try to parse generically
	for _, symbol := range symbols {
		if data, ok := response[symbol].(map[string]interface{}); ok {
			getFloat := func(key string) float64 {
				if v, ok := data[key].(float64); ok {
					return v
				}
				return 0
			}

			getInt := func(key string) int64 {
				if v, ok := data[key].(float64); ok {
					return int64(v)
				}
				return 0
			}

			analytics = append(analytics, AnalyticsWindow{
				Symbol:             symbol,
				AveragePrice:       getFloat("MEAN"),
				HighPrice:          getFloat("MAX"),
				LowPrice:           getFloat("MIN"),
				Volatility:         getFloat("STDDEV"),
				AverageVolume:      getInt("MEAN_VOLUME"),
				TotalVolume:        getInt("TOTAL_VOLUME"),
				PriceChange:        getFloat("PRICE_CHANGE"),
				PriceChangePercent: getFloat("PRICE_CHANGE_PERCENT"),
			})
		}
	}

	return analytics, nil
}
