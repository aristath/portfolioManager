package alphavantage

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/aristath/sentinel/internal/clientdata"
)

// =============================================================================
// Forex Endpoints
// =============================================================================

// GetExchangeRate returns real-time exchange rate between two currencies.
// Forex data uses currency pair as cache key (not ISIN).
func (c *Client) GetExchangeRate(fromCurrency, toCurrency string) (*ExchangeRate, error) {
	// Use currency pair as cache key (e.g., "EUR:USD")
	cacheKey := fromCurrency + ":" + toCurrency

	// Check cache (using exchangerate table)
	table := "exchangerate"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var rate ExchangeRate
			if err := json.Unmarshal(data, &rate); err == nil {
				return &rate, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{
		"from_currency": fromCurrency,
		"to_currency":   toCurrency,
	}

	body, err := c.doRequest("CURRENCY_EXCHANGE_RATE", params)
	if err != nil {
		return nil, err
	}

	rate, err := parseExchangeRate(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse exchange rate: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, rate, clientdata.TTLExchangeRate); err != nil {
			c.log.Warn().Err(err).Str("pair", cacheKey).Msg("Failed to cache exchange rate")
		}
	}

	return rate, nil
}

// GetFXDaily returns daily forex data for a currency pair.
func (c *Client) GetFXDaily(fromCurrency, toCurrency string, full bool) ([]FXPrice, error) {
	outputSize := "compact"
	if full {
		outputSize = "full"
	}

	// Use currency pair + function + outputsize as cache key
	cacheKey := fromCurrency + ":" + toCurrency + ":FX_DAILY:" + outputSize

	// Check cache (using exchangerate table)
	table := "exchangerate"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var prices []FXPrice
			if err := json.Unmarshal(data, &prices); err == nil {
				return prices, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{
		"from_symbol": fromCurrency,
		"to_symbol":   toCurrency,
		"outputsize":  outputSize,
	}

	body, err := c.doRequest("FX_DAILY", params)
	if err != nil {
		return nil, err
	}

	prices, err := parseFXTimeSeries(body, "Time Series FX (Daily)")
	if err != nil {
		return nil, fmt.Errorf("failed to parse FX daily: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, prices, clientdata.TTLExchangeRate); err != nil {
			c.log.Warn().Err(err).Str("pair", fromCurrency+":"+toCurrency).Msg("Failed to cache FX daily")
		}
	}

	return prices, nil
}

// GetFXWeekly returns weekly forex data for a currency pair.
func (c *Client) GetFXWeekly(fromCurrency, toCurrency string) ([]FXPrice, error) {
	// Use currency pair + function as cache key
	cacheKey := fromCurrency + ":" + toCurrency + ":FX_WEEKLY"

	// Check cache
	table := "exchangerate"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var prices []FXPrice
			if err := json.Unmarshal(data, &prices); err == nil {
				return prices, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{
		"from_symbol": fromCurrency,
		"to_symbol":   toCurrency,
	}

	body, err := c.doRequest("FX_WEEKLY", params)
	if err != nil {
		return nil, err
	}

	prices, err := parseFXTimeSeries(body, "Time Series FX (Weekly)")
	if err != nil {
		return nil, fmt.Errorf("failed to parse FX weekly: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, prices, clientdata.TTLExchangeRate); err != nil {
			c.log.Warn().Err(err).Str("pair", fromCurrency+":"+toCurrency).Msg("Failed to cache FX weekly")
		}
	}

	return prices, nil
}

// GetFXMonthly returns monthly forex data for a currency pair.
func (c *Client) GetFXMonthly(fromCurrency, toCurrency string) ([]FXPrice, error) {
	// Use currency pair + function as cache key
	cacheKey := fromCurrency + ":" + toCurrency + ":FX_MONTHLY"

	// Check cache
	table := "exchangerate"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var prices []FXPrice
			if err := json.Unmarshal(data, &prices); err == nil {
				return prices, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{
		"from_symbol": fromCurrency,
		"to_symbol":   toCurrency,
	}

	body, err := c.doRequest("FX_MONTHLY", params)
	if err != nil {
		return nil, err
	}

	prices, err := parseFXTimeSeries(body, "Time Series FX (Monthly)")
	if err != nil {
		return nil, fmt.Errorf("failed to parse FX monthly: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, prices, clientdata.TTLExchangeRate); err != nil {
			c.log.Warn().Err(err).Str("pair", fromCurrency+":"+toCurrency).Msg("Failed to cache FX monthly")
		}
	}

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
