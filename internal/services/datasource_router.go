// Package services provides core business services.
package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/rs/zerolog"
)

// DataType represents the type of data being requested.
type DataType string

// Data type constants.
const (
	DataTypeFundamentals  DataType = "fundamentals"
	DataTypePrices        DataType = "current_prices"
	DataTypeHistorical    DataType = "historical"
	DataTypeTechnicals    DataType = "technicals"
	DataTypeExchangeRates DataType = "exchange_rates"
	DataTypeISINLookup    DataType = "isin_lookup"
	DataTypeMetadata      DataType = "company_metadata"
)

// DataSource represents a data provider.
type DataSource string

// Data source constants.
const (
	DataSourceAlphaVantage DataSource = "alphavantage"
	DataSourceYahoo        DataSource = "yahoo"
	DataSourceTradernet    DataSource = "tradernet"
	DataSourceExchangeRate DataSource = "exchangerate"
	DataSourceOpenFIGI     DataSource = "openfigi"
)

// Source aliases for cleaner code in data fetcher.
const (
	SourceAlphaVantage = DataSourceAlphaVantage
	SourceYahoo        = DataSourceYahoo
	SourceTradernet    = DataSourceTradernet
	SourceExchangeRate = DataSourceExchangeRate
	SourceOpenFIGI     = DataSourceOpenFIGI
)

// SettingsGetter is an interface for getting settings values.
type SettingsGetter interface {
	Get(key string) (*string, error)
}

// DataSourceClients holds references to all available data source clients.
type DataSourceClients struct {
	// Clients are set externally - this allows testing without real clients
	AlphaVantageAPIKey string
	OpenFIGIAPIKey     string
}

// DataSourceRouter manages data source priorities and fallback logic.
type DataSourceRouter struct {
	settings   SettingsGetter
	clients    *DataSourceClients
	priorities map[DataType][]DataSource
	mu         sync.RWMutex
	log        zerolog.Logger
}

// NewDataSourceRouter creates a new data source router.
func NewDataSourceRouter(settings SettingsGetter, clients *DataSourceClients) *DataSourceRouter {
	r := &DataSourceRouter{
		settings:   settings,
		clients:    clients,
		priorities: make(map[DataType][]DataSource),
		log:        zerolog.Nop(),
	}
	r.RefreshPriorities()
	return r
}

// SetLogger sets the logger for the router.
func (r *DataSourceRouter) SetLogger(log zerolog.Logger) {
	r.log = log.With().Str("component", "datasource_router").Logger()
}

// GetPriorities returns the priority order for a data type.
func (r *DataSourceRouter) GetPriorities(dataType DataType) []DataSource {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if priorities, ok := r.priorities[dataType]; ok {
		return priorities
	}

	// Fall back to defaults
	defaults := getDefaultPriorities()
	return defaults[dataType]
}

// RefreshPriorities reloads priorities from settings.
func (r *DataSourceRouter) RefreshPriorities() {
	r.mu.Lock()
	defer r.mu.Unlock()

	dataTypes := []DataType{
		DataTypeFundamentals,
		DataTypePrices,
		DataTypeHistorical,
		DataTypeTechnicals,
		DataTypeExchangeRates,
		DataTypeISINLookup,
		DataTypeMetadata,
	}

	defaults := getDefaultPriorities()

	for _, dt := range dataTypes {
		key := dataTypeToSettingKey(dt)
		if r.settings != nil {
			if val, err := r.settings.Get(key); err == nil && val != nil && *val != "" {
				sources := parseDataSourcePrioritiesFromJSON(*val)
				if len(sources) > 0 {
					r.priorities[dt] = sources
					continue
				}
			}
		}
		// Use defaults
		r.priorities[dt] = defaults[dt]
	}
}

// IsSourceAvailable checks if a data source is available (has required configuration).
func (r *DataSourceRouter) IsSourceAvailable(source DataSource) bool {
	switch source {
	case DataSourceAlphaVantage:
		// Alpha Vantage can work without API key (limited)
		// Return true if we have a key, but it's not strictly required
		if r.clients != nil && r.clients.AlphaVantageAPIKey != "" {
			return true
		}
		// Can still use with limited requests
		return true
	case DataSourceOpenFIGI:
		// OpenFIGI works without API key (lower limits)
		return true
	case DataSourceYahoo:
		// Yahoo Finance is always available (no API key required)
		return true
	case DataSourceTradernet:
		// Tradernet is always available (credentials handled elsewhere)
		return true
	case DataSourceExchangeRate:
		// ExchangeRateAPI is always available (free tier)
		return true
	default:
		return false
	}
}

// GetAvailableSources returns only the available sources for a data type.
func (r *DataSourceRouter) GetAvailableSources(dataType DataType) []DataSource {
	priorities := r.GetPriorities(dataType)
	available := make([]DataSource, 0, len(priorities))
	for _, source := range priorities {
		if r.IsSourceAvailable(source) {
			available = append(available, source)
		}
	}
	return available
}

// FetcherFunc is a function that fetches data from a specific source.
type FetcherFunc func(source DataSource) (interface{}, error)

// FetchWithFallback attempts to fetch data from sources in priority order.
// It returns the result, the source that succeeded, and any error.
func (r *DataSourceRouter) FetchWithFallback(dataType DataType, fetcher FetcherFunc) (interface{}, DataSource, error) {
	sources := r.GetAvailableSources(dataType)
	return fetchWithFallback(sources, fetcher)
}

// dataTypeToSettingKey converts a data type to its settings key.
func dataTypeToSettingKey(dt DataType) string {
	return "datasource_" + string(dt)
}

// getDefaultPriorities returns the default priority configuration.
// These match the defaults in settings/models.go.
func getDefaultPriorities() map[DataType][]DataSource {
	return map[DataType][]DataSource{
		DataTypeFundamentals:  {DataSourceAlphaVantage, DataSourceYahoo},
		DataTypePrices:        {DataSourceTradernet, DataSourceAlphaVantage, DataSourceYahoo},
		DataTypeHistorical:    {DataSourceTradernet, DataSourceAlphaVantage, DataSourceYahoo},
		DataTypeTechnicals:    {DataSourceAlphaVantage, DataSourceYahoo},
		DataTypeExchangeRates: {DataSourceExchangeRate, DataSourceTradernet, DataSourceAlphaVantage},
		DataTypeISINLookup:    {DataSourceOpenFIGI, DataSourceYahoo},
		DataTypeMetadata:      {DataSourceYahoo, DataSourceAlphaVantage, DataSourceOpenFIGI},
	}
}

// parseDataSourcePrioritiesFromJSON parses a JSON array of data source strings.
func parseDataSourcePrioritiesFromJSON(jsonStr string) []DataSource {
	if jsonStr == "" {
		return nil
	}
	var strings []string
	if err := json.Unmarshal([]byte(jsonStr), &strings); err != nil {
		return nil
	}
	sources := make([]DataSource, 0, len(strings))
	for _, s := range strings {
		sources = append(sources, DataSource(s))
	}
	return sources
}

// fetchWithFallback implements the core fallback logic.
func fetchWithFallback(sources []DataSource, fetcher FetcherFunc) (interface{}, DataSource, error) {
	if len(sources) == 0 {
		return nil, "", errors.New("no data sources configured")
	}

	var lastErr error
	var allErrors []string

	for _, source := range sources {
		result, err := fetcher(source)
		if err == nil {
			return result, source, nil
		}
		lastErr = err
		allErrors = append(allErrors, fmt.Sprintf("%s: %v", source, err))
	}

	return nil, "", fmt.Errorf("all data sources failed: %v (last error: %w)", allErrors, lastErr)
}
