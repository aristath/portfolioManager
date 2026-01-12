package alphavantage

import (
	"encoding/json"
	"fmt"
	"sort"
)

// =============================================================================
// Forex Endpoints
// =============================================================================

// GetExchangeRate returns real-time exchange rate between two currencies.
func (c *Client) GetExchangeRate(fromCurrency, toCurrency string) (*ExchangeRate, error) {
	params := map[string]string{
		"from_currency": fromCurrency,
		"to_currency":   toCurrency,
	}

	cacheKey := buildCacheKey("CURRENCY_EXCHANGE_RATE", params)

	if cached, ok := c.getFromCache(cacheKey); ok {
		return cached.(*ExchangeRate), nil
	}

	body, err := c.doRequest("CURRENCY_EXCHANGE_RATE", params)
	if err != nil {
		return nil, err
	}

	rate, err := parseExchangeRate(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse exchange rate: %w", err)
	}

	c.setCache(cacheKey, rate, c.cacheTTL.ExchangeRates)
	return rate, nil
}

// GetFXDaily returns daily forex data for a currency pair.
func (c *Client) GetFXDaily(fromCurrency, toCurrency string, full bool) ([]FXPrice, error) {
	outputSize := "compact"
	if full {
		outputSize = "full"
	}

	params := map[string]string{
		"from_symbol": fromCurrency,
		"to_symbol":   toCurrency,
		"outputsize":  outputSize,
	}

	cacheKey := buildCacheKey("FX_DAILY", params)

	if cached, ok := c.getFromCache(cacheKey); ok {
		return cached.([]FXPrice), nil
	}

	body, err := c.doRequest("FX_DAILY", params)
	if err != nil {
		return nil, err
	}

	prices, err := parseFXTimeSeries(body, "Time Series FX (Daily)")
	if err != nil {
		return nil, fmt.Errorf("failed to parse FX daily: %w", err)
	}

	c.setCache(cacheKey, prices, c.cacheTTL.ExchangeRates)
	return prices, nil
}

// GetFXWeekly returns weekly forex data for a currency pair.
func (c *Client) GetFXWeekly(fromCurrency, toCurrency string) ([]FXPrice, error) {
	params := map[string]string{
		"from_symbol": fromCurrency,
		"to_symbol":   toCurrency,
	}

	cacheKey := buildCacheKey("FX_WEEKLY", params)

	if cached, ok := c.getFromCache(cacheKey); ok {
		return cached.([]FXPrice), nil
	}

	body, err := c.doRequest("FX_WEEKLY", params)
	if err != nil {
		return nil, err
	}

	prices, err := parseFXTimeSeries(body, "Time Series FX (Weekly)")
	if err != nil {
		return nil, fmt.Errorf("failed to parse FX weekly: %w", err)
	}

	c.setCache(cacheKey, prices, c.cacheTTL.ExchangeRates)
	return prices, nil
}

// GetFXMonthly returns monthly forex data for a currency pair.
func (c *Client) GetFXMonthly(fromCurrency, toCurrency string) ([]FXPrice, error) {
	params := map[string]string{
		"from_symbol": fromCurrency,
		"to_symbol":   toCurrency,
	}

	cacheKey := buildCacheKey("FX_MONTHLY", params)

	if cached, ok := c.getFromCache(cacheKey); ok {
		return cached.([]FXPrice), nil
	}

	body, err := c.doRequest("FX_MONTHLY", params)
	if err != nil {
		return nil, err
	}

	prices, err := parseFXTimeSeries(body, "Time Series FX (Monthly)")
	if err != nil {
		return nil, fmt.Errorf("failed to parse FX monthly: %w", err)
	}

	c.setCache(cacheKey, prices, c.cacheTTL.ExchangeRates)
	return prices, nil
}

// =============================================================================
// Parsing Functions
// =============================================================================

func parseExchangeRate(body []byte) (*ExchangeRate, error) {
	var response struct {
		Rate map[string]string `json:"Realtime Currency Exchange Rate"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if response.Rate == nil {
		return nil, fmt.Errorf("no exchange rate data in response")
	}

	data := response.Rate
	return &ExchangeRate{
		FromCurrency:     data["1. From_Currency Code"],
		FromCurrencyName: data["2. From_Currency Name"],
		ToCurrency:       data["3. To_Currency Code"],
		ToCurrencyName:   data["4. To_Currency Name"],
		ExchangeRate:     parseFloat64(data["5. Exchange Rate"]),
		LastRefreshed:    parseDateTime(data["6. Last Refreshed"]),
		Timezone:         data["7. Time Zone"],
		BidPrice:         parseFloat64(data["8. Bid Price"]),
		AskPrice:         parseFloat64(data["9. Ask Price"]),
	}, nil
}

func parseFXTimeSeries(body []byte, timeSeriesKey string) ([]FXPrice, error) {
	var rawResponse map[string]json.RawMessage
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, err
	}

	timeSeriesData, ok := rawResponse[timeSeriesKey]
	if !ok {
		return nil, fmt.Errorf("no %s data in response", timeSeriesKey)
	}

	var timeSeries map[string]map[string]string
	if err := json.Unmarshal(timeSeriesData, &timeSeries); err != nil {
		return nil, err
	}

	prices := make([]FXPrice, 0, len(timeSeries))
	for dateStr, data := range timeSeries {
		date := parseDate(dateStr)
		prices = append(prices, FXPrice{
			Date:  date,
			Open:  parseFloat64(data["1. open"]),
			High:  parseFloat64(data["2. high"]),
			Low:   parseFloat64(data["3. low"]),
			Close: parseFloat64(data["4. close"]),
		})
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Date.After(prices[j].Date)
	})

	return prices, nil
}
