package alphavantage

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/aristath/sentinel/internal/clientdata"
)

// =============================================================================
// Cryptocurrency Endpoints
// =============================================================================

// GetCryptoExchangeRate returns real-time exchange rate for a cryptocurrency.
// Crypto data uses symbol:market as cache key (not ISIN).
func (c *Client) GetCryptoExchangeRate(fromCurrency, toCurrency string) (*ExchangeRate, error) {
	// Use crypto pair as cache key
	cacheKey := "CRYPTO:" + fromCurrency + ":" + toCurrency

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
		return nil, fmt.Errorf("failed to parse crypto exchange rate: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, rate, clientdata.TTLExchangeRate); err != nil {
			c.log.Warn().Err(err).Str("pair", cacheKey).Msg("Failed to cache crypto exchange rate")
		}
	}

	return rate, nil
}

// GetCryptoDaily returns daily cryptocurrency data.
// Crypto data uses symbol:market as cache key (not ISIN).
func (c *Client) GetCryptoDaily(symbol, market string) ([]CryptoPrice, error) {
	// Use crypto symbol + market + function as cache key
	cacheKey := "CRYPTO:" + symbol + ":" + market + ":DAILY"

	// Check cache (using current_prices table)
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var prices []CryptoPrice
			if err := json.Unmarshal(data, &prices); err == nil {
				return prices, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{
		"symbol": symbol,
		"market": market,
	}

	body, err := c.doRequest("DIGITAL_CURRENCY_DAILY", params)
	if err != nil {
		return nil, err
	}

	prices, err := parseCryptoTimeSeries(body, "Time Series (Digital Currency Daily)", market)
	if err != nil {
		return nil, fmt.Errorf("failed to parse crypto daily: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, prices, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to cache crypto daily")
		}
	}

	return prices, nil
}

// GetCryptoWeekly returns weekly cryptocurrency data.
func (c *Client) GetCryptoWeekly(symbol, market string) ([]CryptoPrice, error) {
	cacheKey := "CRYPTO:" + symbol + ":" + market + ":WEEKLY"

	// Check cache
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var prices []CryptoPrice
			if err := json.Unmarshal(data, &prices); err == nil {
				return prices, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{
		"symbol": symbol,
		"market": market,
	}

	body, err := c.doRequest("DIGITAL_CURRENCY_WEEKLY", params)
	if err != nil {
		return nil, err
	}

	prices, err := parseCryptoTimeSeries(body, "Time Series (Digital Currency Weekly)", market)
	if err != nil {
		return nil, fmt.Errorf("failed to parse crypto weekly: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, prices, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to cache crypto weekly")
		}
	}

	return prices, nil
}

// GetCryptoMonthly returns monthly cryptocurrency data.
func (c *Client) GetCryptoMonthly(symbol, market string) ([]CryptoPrice, error) {
	cacheKey := "CRYPTO:" + symbol + ":" + market + ":MONTHLY"

	// Check cache
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var prices []CryptoPrice
			if err := json.Unmarshal(data, &prices); err == nil {
				return prices, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{
		"symbol": symbol,
		"market": market,
	}

	body, err := c.doRequest("DIGITAL_CURRENCY_MONTHLY", params)
	if err != nil {
		return nil, err
	}

	prices, err := parseCryptoTimeSeries(body, "Time Series (Digital Currency Monthly)", market)
	if err != nil {
		return nil, fmt.Errorf("failed to parse crypto monthly: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, prices, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to cache crypto monthly")
		}
	}

	return prices, nil
}

// =============================================================================
// Parsing Functions
// =============================================================================

func parseCryptoTimeSeries(body []byte, timeSeriesKey, market string) ([]CryptoPrice, error) {
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

	prices := make([]CryptoPrice, 0, len(timeSeries))
	for dateStr, data := range timeSeries {
		date := parseDate(dateStr)

		// Keys are formatted with the market currency, e.g., "1a. open (USD)"
		openKey := fmt.Sprintf("1a. open (%s)", market)
		highKey := fmt.Sprintf("2a. high (%s)", market)
		lowKey := fmt.Sprintf("3a. low (%s)", market)
		closeKey := fmt.Sprintf("4a. close (%s)", market)

		// Fallback to numbered keys if market-specific keys don't exist
		open := parseFloat64(data[openKey])
		if open == 0 {
			open = parseFloat64(data["1. open"])
		}

		high := parseFloat64(data[highKey])
		if high == 0 {
			high = parseFloat64(data["2. high"])
		}

		low := parseFloat64(data[lowKey])
		if low == 0 {
			low = parseFloat64(data["3. low"])
		}

		close := parseFloat64(data[closeKey])
		if close == 0 {
			close = parseFloat64(data["4. close"])
		}

		prices = append(prices, CryptoPrice{
			Date:      date,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    parseFloat64(data["5. volume"]),
			MarketCap: parseInt64(data["6. market cap (USD)"]),
		})
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Date.After(prices[j].Date)
	})

	return prices, nil
}
