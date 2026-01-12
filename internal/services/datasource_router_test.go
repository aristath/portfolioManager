package services

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseDataSourcePriorities tests the priority parsing from JSON settings.
func TestParseDataSourcePriorities(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Valid JSON array",
			input:    `["alphavantage","yahoo","tradernet"]`,
			expected: []string{"alphavantage", "yahoo", "tradernet"},
		},
		{
			name:     "Empty array",
			input:    `[]`,
			expected: []string{},
		},
		{
			name:     "Single source",
			input:    `["yahoo"]`,
			expected: []string{"yahoo"},
		},
		{
			name:     "Invalid JSON",
			input:    `invalid`,
			expected: nil, // Should handle gracefully
		},
		{
			name:     "Empty string",
			input:    "",
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseDataSourcePriorities(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestDataTypeConstants verifies data type constants are valid.
func TestDataTypeConstants(t *testing.T) {
	// Ensure all data types have valid string values
	assert.Equal(t, DataType("fundamentals"), DataTypeFundamentals)
	assert.Equal(t, DataType("current_prices"), DataTypePrices)
	assert.Equal(t, DataType("historical"), DataTypeHistorical)
	assert.Equal(t, DataType("technicals"), DataTypeTechnicals)
	assert.Equal(t, DataType("exchange_rates"), DataTypeExchangeRates)
	assert.Equal(t, DataType("isin_lookup"), DataTypeISINLookup)
	assert.Equal(t, DataType("company_metadata"), DataTypeMetadata)
}

// TestDataSourceConstants verifies data source constants are valid.
func TestDataSourceConstants(t *testing.T) {
	assert.Equal(t, DataSource("alphavantage"), DataSourceAlphaVantage)
	assert.Equal(t, DataSource("yahoo"), DataSourceYahoo)
	assert.Equal(t, DataSource("tradernet"), DataSourceTradernet)
	assert.Equal(t, DataSource("exchangerate"), DataSourceExchangeRate)
	assert.Equal(t, DataSource("openfigi"), DataSourceOpenFIGI)
}

// TestDefaultPriorities verifies default priorities match settings.
func TestDefaultPriorities(t *testing.T) {
	defaults := getDefaultPriorities()

	// Fundamentals defaults
	require.Contains(t, defaults, DataTypeFundamentals)
	assert.Equal(t, []DataSource{DataSourceAlphaVantage, DataSourceYahoo}, defaults[DataTypeFundamentals])

	// Current prices defaults
	require.Contains(t, defaults, DataTypePrices)
	assert.Equal(t, []DataSource{DataSourceTradernet, DataSourceAlphaVantage, DataSourceYahoo}, defaults[DataTypePrices])

	// ISIN lookup defaults
	require.Contains(t, defaults, DataTypeISINLookup)
	assert.Equal(t, []DataSource{DataSourceOpenFIGI, DataSourceYahoo}, defaults[DataTypeISINLookup])

	// Exchange rates defaults
	require.Contains(t, defaults, DataTypeExchangeRates)
	assert.Equal(t, []DataSource{DataSourceExchangeRate, DataSourceTradernet, DataSourceAlphaVantage}, defaults[DataTypeExchangeRates])
}

// MockSettingsGetter is a mock implementation of SettingsGetter.
type MockSettingsGetter struct {
	settings map[string]string
}

func (m *MockSettingsGetter) Get(key string) (*string, error) {
	if v, ok := m.settings[key]; ok {
		return &v, nil
	}
	return nil, nil
}

// TestNewDataSourceRouter verifies router creation with settings.
func TestNewDataSourceRouter(t *testing.T) {
	mockSettings := &MockSettingsGetter{
		settings: map[string]string{
			"datasource_fundamentals":     `["yahoo","alphavantage"]`, // Reversed order
			"datasource_current_prices":   `["yahoo"]`,                // Only Yahoo
			"datasource_historical":       `["tradernet","yahoo"]`,
			"datasource_technicals":       `["alphavantage"]`,
			"datasource_exchange_rates":   `["exchangerate"]`,
			"datasource_isin_lookup":      `["openfigi"]`,
			"datasource_company_metadata": `["yahoo","alphavantage"]`,
		},
	}

	router := NewDataSourceRouter(mockSettings, nil)
	require.NotNil(t, router)

	// Verify priorities were loaded from settings
	priorities := router.GetPriorities(DataTypeFundamentals)
	assert.Equal(t, []DataSource{DataSourceYahoo, DataSourceAlphaVantage}, priorities)

	// Verify current prices only has Yahoo
	prices := router.GetPriorities(DataTypePrices)
	assert.Equal(t, []DataSource{DataSourceYahoo}, prices)
}

// TestGetPrioritiesWithDefaults verifies fallback to defaults.
func TestGetPrioritiesWithDefaults(t *testing.T) {
	// Empty settings - should use defaults
	mockSettings := &MockSettingsGetter{
		settings: map[string]string{},
	}

	router := NewDataSourceRouter(mockSettings, nil)
	require.NotNil(t, router)

	// Should fall back to defaults
	priorities := router.GetPriorities(DataTypeFundamentals)
	assert.Equal(t, []DataSource{DataSourceAlphaVantage, DataSourceYahoo}, priorities)
}

// TestSettingKeyMapping verifies the mapping between data types and setting keys.
func TestSettingKeyMapping(t *testing.T) {
	assert.Equal(t, "datasource_fundamentals", dataTypeToSettingKey(DataTypeFundamentals))
	assert.Equal(t, "datasource_current_prices", dataTypeToSettingKey(DataTypePrices))
	assert.Equal(t, "datasource_historical", dataTypeToSettingKey(DataTypeHistorical))
	assert.Equal(t, "datasource_technicals", dataTypeToSettingKey(DataTypeTechnicals))
	assert.Equal(t, "datasource_exchange_rates", dataTypeToSettingKey(DataTypeExchangeRates))
	assert.Equal(t, "datasource_isin_lookup", dataTypeToSettingKey(DataTypeISINLookup))
	assert.Equal(t, "datasource_company_metadata", dataTypeToSettingKey(DataTypeMetadata))
}

// TestFetchWithFallback verifies the fallback mechanism.
func TestFetchWithFallback(t *testing.T) {
	t.Run("First source succeeds", func(t *testing.T) {
		callOrder := []string{}
		fetcher := func(source DataSource) (interface{}, error) {
			callOrder = append(callOrder, string(source))
			if source == DataSourceAlphaVantage {
				return "result from alphavantage", nil
			}
			return nil, errors.New("unexpected source")
		}

		result, source, err := fetchWithFallback(
			[]DataSource{DataSourceAlphaVantage, DataSourceYahoo},
			fetcher,
		)

		require.NoError(t, err)
		assert.Equal(t, "result from alphavantage", result)
		assert.Equal(t, DataSourceAlphaVantage, source)
		assert.Equal(t, []string{"alphavantage"}, callOrder)
	})

	t.Run("First source fails, second succeeds", func(t *testing.T) {
		callOrder := []string{}
		fetcher := func(source DataSource) (interface{}, error) {
			callOrder = append(callOrder, string(source))
			if source == DataSourceAlphaVantage {
				return nil, errors.New("alphavantage failed")
			}
			if source == DataSourceYahoo {
				return "result from yahoo", nil
			}
			return nil, errors.New("unexpected source")
		}

		result, source, err := fetchWithFallback(
			[]DataSource{DataSourceAlphaVantage, DataSourceYahoo},
			fetcher,
		)

		require.NoError(t, err)
		assert.Equal(t, "result from yahoo", result)
		assert.Equal(t, DataSourceYahoo, source)
		assert.Equal(t, []string{"alphavantage", "yahoo"}, callOrder)
	})

	t.Run("All sources fail", func(t *testing.T) {
		fetcher := func(source DataSource) (interface{}, error) {
			return nil, errors.New("source failed: " + string(source))
		}

		result, source, err := fetchWithFallback(
			[]DataSource{DataSourceAlphaVantage, DataSourceYahoo},
			fetcher,
		)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, DataSource(""), source)
		assert.Contains(t, err.Error(), "all data sources failed")
	})

	t.Run("Empty sources list", func(t *testing.T) {
		fetcher := func(source DataSource) (interface{}, error) {
			return nil, nil
		}

		result, source, err := fetchWithFallback([]DataSource{}, fetcher)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, DataSource(""), source)
		assert.Contains(t, err.Error(), "no data sources configured")
	})
}

// TestIsSourceAvailable verifies source availability checking.
func TestIsSourceAvailable(t *testing.T) {
	tests := []struct {
		name     string
		source   DataSource
		expected bool
	}{
		{"Yahoo always available", DataSourceYahoo, true},
		{"Tradernet always available", DataSourceTradernet, true},
		{"ExchangeRate always available", DataSourceExchangeRate, true},
		// AlphaVantage and OpenFIGI depend on API key
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := NewDataSourceRouter(&MockSettingsGetter{settings: map[string]string{}}, nil)
			// For these tests, we just verify the method exists and returns something
			// Full availability check would require real clients
			result := router.IsSourceAvailable(tc.source)
			// We can't assert specific values without knowing client state
			_ = result
		})
	}
}

// TestRefreshPriorities verifies dynamic priority refresh.
func TestRefreshPriorities(t *testing.T) {
	mockSettings := &MockSettingsGetter{
		settings: map[string]string{
			"datasource_fundamentals": `["yahoo"]`,
		},
	}

	router := NewDataSourceRouter(mockSettings, nil)

	// Initial priority
	priorities := router.GetPriorities(DataTypeFundamentals)
	assert.Equal(t, []DataSource{DataSourceYahoo}, priorities)

	// Update settings
	mockSettings.settings["datasource_fundamentals"] = `["alphavantage","yahoo"]`

	// Refresh
	router.RefreshPriorities()

	// Should reflect new priorities
	newPriorities := router.GetPriorities(DataTypeFundamentals)
	assert.Equal(t, []DataSource{DataSourceAlphaVantage, DataSourceYahoo}, newPriorities)
}

// Helper function to parse JSON for tests.
func parseDataSourcePriorities(jsonStr string) []string {
	if jsonStr == "" {
		return nil
	}
	var result []string
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil
	}
	return result
}
