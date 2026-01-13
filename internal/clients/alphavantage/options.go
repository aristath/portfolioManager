package alphavantage

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/aristath/sentinel/internal/clientdata"
)

// =============================================================================
// Options Endpoints
// =============================================================================

// GetHistoricalOptions returns historical options chain data.
func (c *Client) GetHistoricalOptions(symbol, date string) (*OptionsChain, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	// Use composite key: isin:date
	cacheKey := isin + ":OPTIONS:" + date

	// Check cache (using current_prices table)
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var chain OptionsChain
			if err := json.Unmarshal(data, &chain); err == nil {
				return &chain, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{
		"symbol": symbol,
		"date":   date,
	}

	body, err := c.doRequest("HISTORICAL_OPTIONS", params)
	if err != nil {
		return nil, err
	}

	chain, err := parseOptionsChain(body, symbol, date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse options chain: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, chain, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Str("date", date).Msg("Failed to cache options chain")
		}
	}

	return chain, nil
}

// =============================================================================
// Parsing Functions
// =============================================================================

func parseOptionsChain(body []byte, symbol, date string) (*OptionsChain, error) {
	var response struct {
		Data []map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	calls := make([]OptionContract, 0)
	puts := make([]OptionContract, 0)

	for _, opt := range response.Data {
		getString := func(key string) string {
			if v, ok := opt[key].(string); ok {
				return v
			}
			return ""
		}

		getFloat := func(key string) float64 {
			switch v := opt[key].(type) {
			case float64:
				return v
			case string:
				return parseFloat64(v)
			}
			return 0
		}

		getInt := func(key string) int64 {
			switch v := opt[key].(type) {
			case float64:
				return int64(v)
			case string:
				return parseInt64(v)
			}
			return 0
		}

		contract := OptionContract{
			ContractID:    getString("contractID"),
			Symbol:        symbol,
			Expiration:    parseDate(getString("expiration")),
			Strike:        getFloat("strike"),
			Type:          getString("type"),
			Last:          getFloat("last"),
			Mark:          getFloat("mark"),
			Bid:           getFloat("bid"),
			Ask:           getFloat("ask"),
			Change:        getFloat("change"),
			ChangePercent: getFloat("change_percent"),
			Volume:        getInt("volume"),
			OpenInterest:  getInt("open_interest"),
			ImpliedVol:    getFloat("implied_volatility"),
			Delta:         getFloat("delta"),
			Gamma:         getFloat("gamma"),
			Theta:         getFloat("theta"),
			Vega:          getFloat("vega"),
			Rho:           getFloat("rho"),
		}

		if contract.Type == "call" {
			calls = append(calls, contract)
		} else if contract.Type == "put" {
			puts = append(puts, contract)
		}
	}

	// Sort by strike price
	sort.Slice(calls, func(i, j int) bool {
		return calls[i].Strike < calls[j].Strike
	})
	sort.Slice(puts, func(i, j int) bool {
		return puts[i].Strike < puts[j].Strike
	})

	return &OptionsChain{
		Symbol: symbol,
		Date:   parseDate(date),
		Calls:  calls,
		Puts:   puts,
	}, nil
}
