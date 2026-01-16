package universe

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// fieldMapping maps Security struct field names to Tradernet JSON paths
// Empty string means the field is not provided by Tradernet
var fieldMapping = map[string]string{
	"name":                 "name",
	"geography":            "attributes.CntryOfRisk", // NOT issuer_country_code (that's "0")
	"industry":             "sector_code",
	"currency":             "face_curr_c",
	"market_code":          "mkt_name",
	"full_exchange_name":   "codesub_nm",
	"min_lot":              "quotes.x_lot", // Int field, prefer over lot_size_q (string)
	"product_type":         "type",         // "Regular stock", "ETF", etc.
	"min_portfolio_target": "",             // Not in Tradernet, preserve existing
	"max_portfolio_target": "",             // Not in Tradernet, preserve existing
}

// extractFromPath extracts value from nested JSON path (e.g. "attributes.CntryOfRisk")
func extractFromPath(data map[string]interface{}, path string) (interface{}, bool) {
	if path == "" {
		return nil, false
	}

	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current, ok = m[part]
		if !ok {
			return nil, false
		}
	}

	return current, true
}

// SecurityFromJSON creates Security struct from raw Tradernet JSON data
// This is the single entry point for parsing security data on read.
// Overrides are applied in security_repository.go after this function returns.
func SecurityFromJSON(isin, symbol, jsonData string, lastSynced *int64) (*Security, error) {
	if jsonData == "" {
		return nil, fmt.Errorf("empty JSON string")
	}

	// Parse raw Tradernet security object
	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Helper to get string value with field mapping
	getString := func(field string) string {
		path := fieldMapping[field]
		if val, ok := extractFromPath(rawData, path); ok {
			if s, ok := val.(string); ok {
				return s
			}
		}
		return ""
	}

	// Helper to get int value with type conversion
	getInt := func(field string) int {
		path := fieldMapping[field]
		if val, ok := extractFromPath(rawData, path); ok {
			// Handle both int and float64 from JSON
			switch v := val.(type) {
			case float64:
				return int(v)
			case int:
				return v
			case string:
				if i, err := strconv.Atoi(v); err == nil {
					return i
				}
			}
		}
		return 1 // Default min_lot
	}

	// Helper to get float64 value
	getFloat := func(field string) float64 {
		path := fieldMapping[field]
		if val, ok := extractFromPath(rawData, path); ok {
			if f, ok := val.(float64); ok {
				return f
			}
		}
		return 0.0
	}

	// Build Security struct from raw Tradernet data
	return &Security{
		ISIN:               isin,
		Symbol:             symbol,
		Name:               getString("name"),
		ProductType:        getString("product_type"),
		Industry:           getString("industry"),
		Geography:          getString("geography"),
		FullExchangeName:   getString("full_exchange_name"),
		MarketCode:         getString("market_code"),
		Currency:           getString("currency"),
		MinLot:             getInt("min_lot"),
		MinPortfolioTarget: getFloat("min_portfolio_target"),
		MaxPortfolioTarget: getFloat("max_portfolio_target"),
		LastSynced:         lastSynced,
	}, nil
}
