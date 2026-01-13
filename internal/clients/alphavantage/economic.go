package alphavantage

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/aristath/sentinel/internal/clientdata"
)

// =============================================================================
// Economic Indicator Endpoints
// =============================================================================

// GetEconomicIndicator returns data for a specific economic indicator.
func (c *Client) GetEconomicIndicator(indicator, interval string) (*EconomicIndicator, error) {
	// Economic indicators use indicator name as cache key (not ISIN)
	cacheKey := indicator
	if interval != "" {
		cacheKey = indicator + ":" + interval
	}

	// Check cache
	table := "alphavantage_economic"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var indicatorData EconomicIndicator
			if err := json.Unmarshal(data, &indicatorData); err == nil {
				return &indicatorData, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{}
	if interval != "" {
		params["interval"] = interval
	}

	body, err := c.doRequest(indicator, params)
	if err != nil {
		return nil, err
	}

	data, err := parseEconomicData(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", indicator, err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, data, clientdata.TTLEconomic); err != nil {
			c.log.Warn().Err(err).Str("indicator", cacheKey).Msg("Failed to cache economic indicator")
		}
	}

	return data, nil
}

// GetRealGDP returns real GDP data.
func (c *Client) GetRealGDP(interval string) (*EconomicIndicator, error) {
	return c.GetEconomicIndicator("REAL_GDP", interval)
}

// GetRealGDPPerCapita returns real GDP per capita data.
func (c *Client) GetRealGDPPerCapita() (*EconomicIndicator, error) {
	return c.GetEconomicIndicator("REAL_GDP_PER_CAPITA", "")
}

// GetTreasuryYield returns treasury yield data.
func (c *Client) GetTreasuryYield(interval, maturity string) (*EconomicIndicator, error) {
	// Treasury yield uses indicator name + maturity as cache key
	cacheKey := "TREASURY_YIELD"
	if maturity != "" {
		cacheKey = "TREASURY_YIELD:" + maturity
	}
	if interval != "" {
		cacheKey = cacheKey + ":" + interval
	}

	// Check cache
	table := "alphavantage_economic"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var indicatorData EconomicIndicator
			if err := json.Unmarshal(data, &indicatorData); err == nil {
				return &indicatorData, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{}
	if interval != "" {
		params["interval"] = interval
	}
	if maturity != "" {
		params["maturity"] = maturity
	}

	body, err := c.doRequest("TREASURY_YIELD", params)
	if err != nil {
		return nil, err
	}

	data, err := parseEconomicData(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse treasury yield: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, data, clientdata.TTLEconomic); err != nil {
			c.log.Warn().Err(err).Str("indicator", cacheKey).Msg("Failed to cache treasury yield")
		}
	}

	return data, nil
}

// GetFederalFundsRate returns federal funds rate data.
func (c *Client) GetFederalFundsRate(interval string) (*EconomicIndicator, error) {
	return c.GetEconomicIndicator("FEDERAL_FUNDS_RATE", interval)
}

// GetCPI returns Consumer Price Index data.
func (c *Client) GetCPI(interval string) (*EconomicIndicator, error) {
	return c.GetEconomicIndicator("CPI", interval)
}

// GetInflation returns inflation rate data.
func (c *Client) GetInflation() (*EconomicIndicator, error) {
	return c.GetEconomicIndicator("INFLATION", "")
}

// GetRetailSales returns retail sales data.
func (c *Client) GetRetailSales() (*EconomicIndicator, error) {
	return c.GetEconomicIndicator("RETAIL_SALES", "")
}

// GetDurableGoodsOrders returns durable goods orders data.
func (c *Client) GetDurableGoodsOrders() (*EconomicIndicator, error) {
	return c.GetEconomicIndicator("DURABLES", "")
}

// GetUnemployment returns unemployment rate data.
func (c *Client) GetUnemployment() (*EconomicIndicator, error) {
	return c.GetEconomicIndicator("UNEMPLOYMENT", "")
}

// GetNonfarmPayroll returns nonfarm payroll data.
func (c *Client) GetNonfarmPayroll() (*EconomicIndicator, error) {
	return c.GetEconomicIndicator("NONFARM_PAYROLL", "")
}

// =============================================================================
// Parsing Functions
// =============================================================================

func parseEconomicData(body []byte) (*EconomicIndicator, error) {
	var response struct {
		Name     string              `json:"name"`
		Interval string              `json:"interval"`
		Unit     string              `json:"unit"`
		Data     []map[string]string `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	dataPoints := make([]EconomicDataPoint, 0, len(response.Data))
	for _, d := range response.Data {
		date := parseDate(d["date"])
		value := parseFloat64(d["value"])

		// Skip entries with no value (API returns "." for missing data)
		if value == 0 && d["value"] != "0" && d["value"] != "" {
			continue
		}

		dataPoints = append(dataPoints, EconomicDataPoint{
			Date:  date,
			Value: value,
		})
	}

	sort.Slice(dataPoints, func(i, j int) bool {
		return dataPoints[i].Date.After(dataPoints[j].Date)
	})

	return &EconomicIndicator{
		Name:     response.Name,
		Interval: response.Interval,
		Unit:     response.Unit,
		Data:     dataPoints,
	}, nil
}
