package alphavantage

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/aristath/sentinel/internal/clientdata"
)

// =============================================================================
// Time Series Endpoints
// =============================================================================

// GetDailyPrices returns daily OHLCV data for a symbol.
func (c *Client) GetDailyPrices(symbol string, full bool) ([]DailyPrice, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	outputSize := "compact"
	if full {
		outputSize = "full"
	}

	// Use composite key: isin:function:outputsize
	cacheKey := isin + ":TIME_SERIES_DAILY:" + outputSize

	// Check cache (using current_prices table for time series data)
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var prices []DailyPrice
			if err := json.Unmarshal(data, &prices); err == nil {
				return prices, nil
			}
		}
	}

	// Fetch from API
	body, err := c.doRequest("TIME_SERIES_DAILY", map[string]string{
		"symbol":     symbol,
		"outputsize": outputSize,
	})
	if err != nil {
		return nil, err
	}

	prices, err := parseDailyTimeSeries(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse daily prices: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, prices, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache daily prices")
		}
	}

	return prices, nil
}

// GetDailyAdjustedPrices returns daily adjusted OHLCV data including dividends and splits.
func (c *Client) GetDailyAdjustedPrices(symbol string, full bool) ([]AdjustedPrice, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	outputSize := "compact"
	if full {
		outputSize = "full"
	}

	// Use composite key
	cacheKey := isin + ":TIME_SERIES_DAILY_ADJUSTED:" + outputSize

	// Check cache
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var prices []AdjustedPrice
			if err := json.Unmarshal(data, &prices); err == nil {
				return prices, nil
			}
		}
	}

	// Fetch from API
	body, err := c.doRequest("TIME_SERIES_DAILY_ADJUSTED", map[string]string{
		"symbol":     symbol,
		"outputsize": outputSize,
	})
	if err != nil {
		return nil, err
	}

	prices, err := parseAdjustedTimeSeries(body, "Time Series (Daily)")
	if err != nil {
		return nil, fmt.Errorf("failed to parse adjusted daily prices: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, prices, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache adjusted daily prices")
		}
	}

	return prices, nil
}

// GetWeeklyPrices returns weekly OHLCV data for a symbol.
func (c *Client) GetWeeklyPrices(symbol string) ([]DailyPrice, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	cacheKey := isin + ":TIME_SERIES_WEEKLY"

	// Check cache
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var prices []DailyPrice
			if err := json.Unmarshal(data, &prices); err == nil {
				return prices, nil
			}
		}
	}

	body, err := c.doRequest("TIME_SERIES_WEEKLY", map[string]string{"symbol": symbol})
	if err != nil {
		return nil, err
	}

	prices, err := parseWeeklyMonthlyTimeSeries(body, "Weekly Time Series")
	if err != nil {
		return nil, fmt.Errorf("failed to parse weekly prices: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, prices, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache weekly prices")
		}
	}

	return prices, nil
}

// GetWeeklyAdjustedPrices returns weekly adjusted OHLCV data.
func (c *Client) GetWeeklyAdjustedPrices(symbol string) ([]AdjustedPrice, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	cacheKey := isin + ":TIME_SERIES_WEEKLY_ADJUSTED"

	// Check cache
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var prices []AdjustedPrice
			if err := json.Unmarshal(data, &prices); err == nil {
				return prices, nil
			}
		}
	}

	body, err := c.doRequest("TIME_SERIES_WEEKLY_ADJUSTED", map[string]string{"symbol": symbol})
	if err != nil {
		return nil, err
	}

	prices, err := parseAdjustedTimeSeries(body, "Weekly Adjusted Time Series")
	if err != nil {
		return nil, fmt.Errorf("failed to parse adjusted weekly prices: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, prices, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache adjusted weekly prices")
		}
	}

	return prices, nil
}

// GetMonthlyPrices returns monthly OHLCV data for a symbol.
func (c *Client) GetMonthlyPrices(symbol string) ([]DailyPrice, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	cacheKey := isin + ":TIME_SERIES_MONTHLY"

	// Check cache
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var prices []DailyPrice
			if err := json.Unmarshal(data, &prices); err == nil {
				return prices, nil
			}
		}
	}

	body, err := c.doRequest("TIME_SERIES_MONTHLY", map[string]string{"symbol": symbol})
	if err != nil {
		return nil, err
	}

	prices, err := parseWeeklyMonthlyTimeSeries(body, "Monthly Time Series")
	if err != nil {
		return nil, fmt.Errorf("failed to parse monthly prices: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, prices, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache monthly prices")
		}
	}

	return prices, nil
}

// GetMonthlyAdjustedPrices returns monthly adjusted OHLCV data.
func (c *Client) GetMonthlyAdjustedPrices(symbol string) ([]AdjustedPrice, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	cacheKey := isin + ":TIME_SERIES_MONTHLY_ADJUSTED"

	// Check cache
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var prices []AdjustedPrice
			if err := json.Unmarshal(data, &prices); err == nil {
				return prices, nil
			}
		}
	}

	body, err := c.doRequest("TIME_SERIES_MONTHLY_ADJUSTED", map[string]string{"symbol": symbol})
	if err != nil {
		return nil, err
	}

	prices, err := parseAdjustedTimeSeries(body, "Monthly Adjusted Time Series")
	if err != nil {
		return nil, fmt.Errorf("failed to parse adjusted monthly prices: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, prices, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache adjusted monthly prices")
		}
	}

	return prices, nil
}

// GetGlobalQuote returns the latest price and volume information for a symbol.
func (c *Client) GetGlobalQuote(symbol string) (*GlobalQuote, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	// Check cache (using current_prices table)
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, isin); err == nil && data != nil {
			var quote GlobalQuote
			if err := json.Unmarshal(data, &quote); err == nil {
				return &quote, nil
			}
		}
	}

	// Fetch from API
	body, err := c.doRequest("GLOBAL_QUOTE", map[string]string{"symbol": symbol})
	if err != nil {
		return nil, err
	}

	quote, err := parseGlobalQuote(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse global quote: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, isin, quote, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache global quote")
		}
	}

	return quote, nil
}

// SearchSymbol searches for symbols matching the given keywords.
// Note: Symbol search doesn't map to a specific security, so we don't cache it.
func (c *Client) SearchSymbol(keywords string) ([]SymbolMatch, error) {
	// Symbol search doesn't require ISIN resolution and doesn't fit the ISIN-based cache model
	// Skip caching for now
	body, err := c.doRequest("SYMBOL_SEARCH", map[string]string{"keywords": keywords})
	if err != nil {
		return nil, err
	}

	matches, err := parseSymbolSearch(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse symbol search: %w", err)
	}

	// Note: Symbol search results are not cached persistently as they don't map to ISINs
	return matches, nil
}

// =============================================================================
// Parsing Functions
// =============================================================================

func parseDailyTimeSeries(body []byte) ([]DailyPrice, error) {
	var response struct {
		MetaData   map[string]string            `json:"Meta Data"`
		TimeSeries map[string]map[string]string `json:"Time Series (Daily)"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if response.TimeSeries == nil {
		return nil, fmt.Errorf("no time series data in response")
	}

	prices := make([]DailyPrice, 0, len(response.TimeSeries))
	for dateStr, data := range response.TimeSeries {
		date := parseDate(dateStr)
		prices = append(prices, DailyPrice{
			Date:   date,
			Open:   parseFloat64(data["1. open"]),
			High:   parseFloat64(data["2. high"]),
			Low:    parseFloat64(data["3. low"]),
			Close:  parseFloat64(data["4. close"]),
			Volume: parseInt64(data["5. volume"]),
		})
	}

	// Sort by date descending (newest first)
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Date.After(prices[j].Date)
	})

	return prices, nil
}

func parseWeeklyMonthlyTimeSeries(body []byte, timeSeriesKey string) ([]DailyPrice, error) {
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

	prices := make([]DailyPrice, 0, len(timeSeries))
	for dateStr, data := range timeSeries {
		date := parseDate(dateStr)
		prices = append(prices, DailyPrice{
			Date:   date,
			Open:   parseFloat64(data["1. open"]),
			High:   parseFloat64(data["2. high"]),
			Low:    parseFloat64(data["3. low"]),
			Close:  parseFloat64(data["4. close"]),
			Volume: parseInt64(data["5. volume"]),
		})
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Date.After(prices[j].Date)
	})

	return prices, nil
}

func parseAdjustedTimeSeries(body []byte, timeSeriesKey string) ([]AdjustedPrice, error) {
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

	prices := make([]AdjustedPrice, 0, len(timeSeries))
	for dateStr, data := range timeSeries {
		date := parseDate(dateStr)
		prices = append(prices, AdjustedPrice{
			Date:             date,
			Open:             parseFloat64(data["1. open"]),
			High:             parseFloat64(data["2. high"]),
			Low:              parseFloat64(data["3. low"]),
			Close:            parseFloat64(data["4. close"]),
			AdjustedClose:    parseFloat64(data["5. adjusted close"]),
			Volume:           parseInt64(data["6. volume"]),
			DividendAmount:   parseFloat64(data["7. dividend amount"]),
			SplitCoefficient: parseFloat64(data["8. split coefficient"]),
		})
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Date.After(prices[j].Date)
	})

	return prices, nil
}

func parseGlobalQuote(body []byte) (*GlobalQuote, error) {
	var response struct {
		GlobalQuote map[string]string `json:"Global Quote"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if len(response.GlobalQuote) == 0 {
		return nil, fmt.Errorf("no quote data in response")
	}

	data := response.GlobalQuote
	changePercentStr := data["10. change percent"]
	changePercentStr = strings.TrimSuffix(changePercentStr, "%")

	return &GlobalQuote{
		Symbol:           data["01. symbol"],
		Open:             parseFloat64(data["02. open"]),
		High:             parseFloat64(data["03. high"]),
		Low:              parseFloat64(data["04. low"]),
		Price:            parseFloat64(data["05. price"]),
		Volume:           parseInt64(data["06. volume"]),
		LatestTradingDay: parseDate(data["07. latest trading day"]),
		PreviousClose:    parseFloat64(data["08. previous close"]),
		Change:           parseFloat64(data["09. change"]),
		ChangePercent:    parseFloat64(changePercentStr),
	}, nil
}

func parseSymbolSearch(body []byte) ([]SymbolMatch, error) {
	var response struct {
		BestMatches []map[string]string `json:"bestMatches"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	matches := make([]SymbolMatch, 0, len(response.BestMatches))
	for _, m := range response.BestMatches {
		matches = append(matches, SymbolMatch{
			Symbol:      m["1. symbol"],
			Name:        m["2. name"],
			Type:        m["3. type"],
			Region:      m["4. region"],
			MarketOpen:  m["5. marketOpen"],
			MarketClose: m["6. marketClose"],
			Timezone:    m["7. timezone"],
			Currency:    m["8. currency"],
			MatchScore:  m["9. matchScore"],
		})
	}

	return matches, nil
}
