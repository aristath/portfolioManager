package universe

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityFromJSON_TradernetFormat(t *testing.T) {
	// Raw Tradernet response for CAT.3750.AS
	rawJSON := `{
		"id": 23844374,
		"ticker": "CAT.3750.AS",
		"name": "Contemporary Amperex Technology",
		"issue_nb": "CNE100006WS8",
		"face_curr_c": "HKD",
		"mkt_name": "HKEX",
		"codesub_nm": "HKEX",
		"lot_size_q": "100.00000000",
		"issuer_country_code": "0",
		"sector_code": "Machinery, Tools, Heavy Vehicles, Trains & Ships",
		"type": "Regular stock",
		"quotes": {
			"x_lot": 100,
			"min_step": 0.01
		},
		"attributes": {
			"CntryOfRisk": "CN",
			"base_mkt_id": "HKEX"
		}
	}`

	lastSynced := int64(1704067200)
	security, err := SecurityFromJSON("CNE100006WS8", "CAT.3750.AS", rawJSON, &lastSynced)

	require.NoError(t, err)
	assert.Equal(t, "CNE100006WS8", security.ISIN)
	assert.Equal(t, "CAT.3750.AS", security.Symbol)
	assert.Equal(t, "Contemporary Amperex Technology", security.Name)
	assert.Equal(t, "CN", security.Geography) // From attributes.CntryOfRisk, NOT issuer_country_code
	assert.Equal(t, "Machinery, Tools, Heavy Vehicles, Trains & Ships", security.Industry)
	assert.Equal(t, "HKD", security.Currency)
	assert.Equal(t, "HKEX", security.MarketCode)
	assert.Equal(t, "HKEX", security.FullExchangeName)
	assert.Equal(t, 100, security.MinLot) // From quotes.x_lot
	assert.Equal(t, "Regular stock", security.ProductType)
	assert.Equal(t, int64(1704067200), *security.LastSynced)
}

func TestSecurityFromJSON_ETFFormat(t *testing.T) {
	// Test ETF with different structure
	rawJSON := `{
		"id": 12345,
		"ticker": "SPY.US",
		"name": "SPDR S&P 500 ETF Trust",
		"issue_nb": "US78462F1030",
		"face_curr_c": "USD",
		"mkt_name": "NYSE",
		"codesub_nm": "NYSE Arca",
		"lot_size_q": "1.00000000",
		"sector_code": "ETFs",
		"type": "ETF",
		"quotes": {
			"x_lot": 1
		},
		"attributes": {
			"CntryOfRisk": "US"
		}
	}`

	lastSynced := int64(1704153600)
	security, err := SecurityFromJSON("US78462F1030", "SPY.US", rawJSON, &lastSynced)

	require.NoError(t, err)
	assert.Equal(t, "US78462F1030", security.ISIN)
	assert.Equal(t, "SPY.US", security.Symbol)
	assert.Equal(t, "SPDR S&P 500 ETF Trust", security.Name)
	assert.Equal(t, "US", security.Geography)
	assert.Equal(t, "ETFs", security.Industry)
	assert.Equal(t, "USD", security.Currency)
	assert.Equal(t, "NYSE", security.MarketCode)
	assert.Equal(t, "NYSE Arca", security.FullExchangeName)
	assert.Equal(t, 1, security.MinLot)
	assert.Equal(t, "ETF", security.ProductType)
}

func TestSecurityFromJSON_MissingFields(t *testing.T) {
	// Test with minimal required fields
	rawJSON := `{
		"ticker": "TEST.US",
		"name": "Test Security",
		"face_curr_c": "USD"
	}`

	security, err := SecurityFromJSON("US1234567890", "TEST.US", rawJSON, nil)

	require.NoError(t, err)
	assert.Equal(t, "US1234567890", security.ISIN)
	assert.Equal(t, "TEST.US", security.Symbol)
	assert.Equal(t, "Test Security", security.Name)
	assert.Equal(t, "USD", security.Currency)
	assert.Equal(t, "", security.Geography) // Missing attributes.CntryOfRisk
	assert.Equal(t, "", security.Industry)  // Missing sector_code
	assert.Equal(t, 1, security.MinLot)     // Default value
	assert.Nil(t, security.LastSynced)
}

func TestSecurityFromJSON_EmptyString(t *testing.T) {
	security, err := SecurityFromJSON("US1234567890", "TEST.US", "", nil)

	assert.Error(t, err)
	assert.Nil(t, security)
	assert.Contains(t, err.Error(), "empty JSON string")
}

func TestSecurityFromJSON_InvalidJSON(t *testing.T) {
	security, err := SecurityFromJSON("US1234567890", "TEST.US", "not valid json", nil)

	assert.Error(t, err)
	assert.Nil(t, security)
	assert.Contains(t, err.Error(), "failed to parse JSON")
}

func TestExtractFromPath(t *testing.T) {
	data := map[string]interface{}{
		"name": "Test",
		"attributes": map[string]interface{}{
			"CntryOfRisk": "CN",
			"nested": map[string]interface{}{
				"value": float64(123),
			},
		},
		"quotes": map[string]interface{}{
			"x_lot": float64(100),
		},
		"simple_string": "hello",
		"simple_number": float64(42),
	}

	tests := []struct {
		name     string
		path     string
		expected interface{}
		found    bool
	}{
		{
			name:     "simple string field",
			path:     "name",
			expected: "Test",
			found:    true,
		},
		{
			name:     "nested field level 2",
			path:     "attributes.CntryOfRisk",
			expected: "CN",
			found:    true,
		},
		{
			name:     "nested field level 3",
			path:     "attributes.nested.value",
			expected: float64(123),
			found:    true,
		},
		{
			name:     "nested in quotes",
			path:     "quotes.x_lot",
			expected: float64(100),
			found:    true,
		},
		{
			name:     "nonexistent top level",
			path:     "nonexistent",
			expected: nil,
			found:    false,
		},
		{
			name:     "nonexistent nested",
			path:     "attributes.nonexistent",
			expected: nil,
			found:    false,
		},
		{
			name:     "empty path",
			path:     "",
			expected: nil,
			found:    false,
		},
		{
			name:     "path through non-map",
			path:     "simple_string.nested",
			expected: nil,
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, found := extractFromPath(data, tt.path)
			assert.Equal(t, tt.found, found, "found mismatch for path: %s", tt.path)
			if found {
				assert.Equal(t, tt.expected, val, "value mismatch for path: %s", tt.path)
			}
		})
	}
}

func TestFieldMapping_AllFields(t *testing.T) {
	// Verify all Security struct fields have mappings (even if empty for non-Tradernet fields)
	requiredFields := []string{
		"name",
		"geography",
		"industry",
		"currency",
		"market_code",
		"full_exchange_name",
		"min_lot",
		"product_type",
		"min_portfolio_target",
		"max_portfolio_target",
	}

	for _, field := range requiredFields {
		_, exists := fieldMapping[field]
		assert.True(t, exists, "Missing mapping for field: %s", field)
	}
}

func TestFieldMapping_ValidPaths(t *testing.T) {
	// Verify all non-empty paths are valid format
	for field, path := range fieldMapping {
		if path == "" {
			// Empty paths are OK for fields not in Tradernet
			continue
		}

		// Path should not start or end with dots
		assert.NotEqual(t, ".", path[0:1], "Path for %s starts with dot", field)
		assert.NotEqual(t, ".", path[len(path)-1:], "Path for %s ends with dot", field)

		// Path should not have double dots
		assert.NotContains(t, path, "..", "Path for %s contains double dots", field)
	}
}

func TestSecurityFromJSON_MinLotTypeConversions(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		expectedLot int
	}{
		{
			name: "x_lot as integer",
			json: `{
				"ticker": "TEST.US",
				"name": "Test",
				"quotes": {"x_lot": 100}
			}`,
			expectedLot: 100,
		},
		{
			name: "x_lot as float64",
			json: `{
				"ticker": "TEST.US",
				"name": "Test",
				"quotes": {"x_lot": 50.0}
			}`,
			expectedLot: 50,
		},
		{
			name: "x_lot as string",
			json: `{
				"ticker": "TEST.US",
				"name": "Test",
				"quotes": {"x_lot": "25"}
			}`,
			expectedLot: 25,
		},
		{
			name: "x_lot missing - defaults to 1",
			json: `{
				"ticker": "TEST.US",
				"name": "Test"
			}`,
			expectedLot: 1,
		},
		{
			name: "x_lot invalid - defaults to 1",
			json: `{
				"ticker": "TEST.US",
				"name": "Test",
				"quotes": {"x_lot": "invalid"}
			}`,
			expectedLot: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			security, err := SecurityFromJSON("US1234567890", "TEST.US", tt.json, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedLot, security.MinLot)
		})
	}
}

func TestSecurityFromJSON_PreservesPortfolioTargets(t *testing.T) {
	// Portfolio targets are NOT in Tradernet data, should default to 0
	rawJSON := `{
		"ticker": "TEST.US",
		"name": "Test Security",
		"face_curr_c": "USD"
	}`

	security, err := SecurityFromJSON("US1234567890", "TEST.US", rawJSON, nil)

	require.NoError(t, err)
	assert.Equal(t, 0.0, security.MinPortfolioTarget)
	assert.Equal(t, 0.0, security.MaxPortfolioTarget)
}
