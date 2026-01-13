package alphavantage

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/aristath/sentinel/internal/clientdata"
)

// =============================================================================
// Commodity Endpoints
// =============================================================================

// GetCommodity returns data for a specific commodity.
// Commodity data uses commodity name + interval as cache key (not ISIN).
func (c *Client) GetCommodity(commodity, interval string) ([]CommodityPrice, error) {
	// Use commodity name + interval as cache key
	cacheKey := "COMMODITY:" + commodity
	if interval != "" {
		cacheKey = cacheKey + ":" + interval
	}

	// Check cache (using current_prices table)
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var prices []CommodityPrice
			if err := json.Unmarshal(data, &prices); err == nil {
				return prices, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{
		"interval": interval,
	}

	body, err := c.doRequest(commodity, params)
	if err != nil {
		return nil, err
	}

	prices, err := parseCommodityData(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", commodity, err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, prices, clientdata.TTLCommodity); err != nil {
			c.log.Warn().Err(err).Str("commodity", commodity).Msg("Failed to cache commodity data")
		}
	}

	return prices, nil
}

// GetWTI returns West Texas Intermediate crude oil prices.
func (c *Client) GetWTI(interval string) ([]CommodityPrice, error) {
	return c.GetCommodity("WTI", interval)
}

// GetBrent returns Brent crude oil prices.
func (c *Client) GetBrent(interval string) ([]CommodityPrice, error) {
	return c.GetCommodity("BRENT", interval)
}

// GetNaturalGas returns natural gas prices.
func (c *Client) GetNaturalGas(interval string) ([]CommodityPrice, error) {
	return c.GetCommodity("NATURAL_GAS", interval)
}

// GetCopper returns copper prices.
func (c *Client) GetCopper(interval string) ([]CommodityPrice, error) {
	return c.GetCommodity("COPPER", interval)
}

// GetAluminum returns aluminum prices.
func (c *Client) GetAluminum(interval string) ([]CommodityPrice, error) {
	return c.GetCommodity("ALUMINUM", interval)
}

// GetWheat returns wheat prices.
func (c *Client) GetWheat(interval string) ([]CommodityPrice, error) {
	return c.GetCommodity("WHEAT", interval)
}

// GetCorn returns corn prices.
func (c *Client) GetCorn(interval string) ([]CommodityPrice, error) {
	return c.GetCommodity("CORN", interval)
}

// GetCotton returns cotton prices.
func (c *Client) GetCotton(interval string) ([]CommodityPrice, error) {
	return c.GetCommodity("COTTON", interval)
}

// GetSugar returns sugar prices.
func (c *Client) GetSugar(interval string) ([]CommodityPrice, error) {
	return c.GetCommodity("SUGAR", interval)
}

// GetCoffee returns coffee prices.
func (c *Client) GetCoffee(interval string) ([]CommodityPrice, error) {
	return c.GetCommodity("COFFEE", interval)
}

// GetAllCommodities returns global commodities index data.
func (c *Client) GetAllCommodities(interval string) ([]CommodityPrice, error) {
	return c.GetCommodity("ALL_COMMODITIES", interval)
}

// =============================================================================
// Parsing Functions
// =============================================================================

func parseCommodityData(body []byte) ([]CommodityPrice, error) {
	var response struct {
		Name     string              `json:"name"`
		Interval string              `json:"interval"`
		Unit     string              `json:"unit"`
		Data     []map[string]string `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	prices := make([]CommodityPrice, 0, len(response.Data))
	for _, d := range response.Data {
		date := parseDate(d["date"])
		value := parseFloat64(d["value"])

		// Skip entries with no value (API returns "." for missing data)
		if value == 0 && d["value"] != "0" {
			continue
		}

		prices = append(prices, CommodityPrice{
			Date:  date,
			Value: value,
		})
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Date.After(prices[j].Date)
	})

	return prices, nil
}
