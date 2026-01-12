// Package symbols provides unified symbol mapping across data providers.
package symbols

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider_String(t *testing.T) {
	tests := []struct {
		provider Provider
		expected string
	}{
		{ProviderTradernet, "tradernet"},
		{ProviderYahoo, "yahoo"},
		{ProviderAlphaVantage, "alphavantage"},
		{ProviderOpenFIGI, "openfigi"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.provider))
		})
	}
}

func TestExchangeMap_ContainsExpectedExchanges(t *testing.T) {
	expectedExchanges := []string{
		"ATH",    // Athens
		"LSE",    // London
		"FRA",    // Frankfurt
		"XETRA",  // Deutsche Börse Xetra
		"PAR",    // Paris
		"AMS",    // Amsterdam
		"MIL",    // Milan
		"HKG",    // Hong Kong
		"TYO",    // Tokyo
		"NYSE",   // New York Stock Exchange
		"NASDAQ", // NASDAQ
	}

	for _, code := range expectedExchanges {
		t.Run(code, func(t *testing.T) {
			exchange, exists := ExchangeMap[code]
			require.True(t, exists, "Exchange %s should exist in ExchangeMap", code)
			assert.Equal(t, code, exchange.Code)
			assert.NotEmpty(t, exchange.Name)
			assert.NotEmpty(t, exchange.Country)
			assert.NotEmpty(t, exchange.Currency)
			assert.NotEmpty(t, exchange.Timezone)
		})
	}
}

func TestGetExchangeByCode(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected *Exchange
		wantNil  bool
	}{
		{
			name:    "Athens Exchange",
			code:    "ATH",
			wantNil: false,
		},
		{
			name:    "Unknown Exchange",
			code:    "UNKNOWN",
			wantNil: true,
		},
		{
			name:    "Empty Code",
			code:    "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetExchangeByCode(tt.code)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.code, result.Code)
			}
		})
	}
}

func TestGetExchangeFromTradernetSuffix(t *testing.T) {
	tests := []struct {
		name         string
		suffix       string
		expectedCode string
		wantNil      bool
	}{
		{
			name:         "Greek suffix",
			suffix:       ".GR",
			expectedCode: "ATH",
			wantNil:      false,
		},
		{
			name:         "Japanese suffix",
			suffix:       ".JP",
			expectedCode: "TYO",
			wantNil:      false,
		},
		{
			name:         "US suffix",
			suffix:       ".US",
			expectedCode: "NYSE", // Default US exchange
			wantNil:      false,
		},
		{
			name:    "Unknown suffix",
			suffix:  ".XX",
			wantNil: true,
		},
		{
			name:    "Empty suffix",
			suffix:  "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetExchangeFromTradernetSuffix(tt.suffix)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedCode, result.Code)
			}
		})
	}
}

func TestGetExchangeFromYahooSuffix(t *testing.T) {
	tests := []struct {
		name         string
		suffix       string
		expectedCode string
		wantNil      bool
	}{
		{
			name:         "Athens suffix",
			suffix:       ".AT",
			expectedCode: "ATH",
			wantNil:      false,
		},
		{
			name:         "London suffix",
			suffix:       ".L",
			expectedCode: "LSE",
			wantNil:      false,
		},
		{
			name:         "Tokyo suffix",
			suffix:       ".T",
			expectedCode: "TYO",
			wantNil:      false,
		},
		{
			name:         "Hong Kong suffix",
			suffix:       ".HK",
			expectedCode: "HKG",
			wantNil:      false,
		},
		{
			name:         "No suffix (US)",
			suffix:       "",
			expectedCode: "NYSE",
			wantNil:      false,
		},
		{
			name:    "Unknown suffix",
			suffix:  ".XX",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetExchangeFromYahooSuffix(tt.suffix)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedCode, result.Code)
			}
		})
	}
}

func TestGetExchangeFromAlphaVantageSuffix(t *testing.T) {
	tests := []struct {
		name         string
		suffix       string
		expectedCode string
		wantNil      bool
	}{
		{
			name:         "Athens suffix",
			suffix:       ".AT",
			expectedCode: "ATH",
			wantNil:      false,
		},
		{
			name:         "London suffix",
			suffix:       ".LON",
			expectedCode: "LSE",
			wantNil:      false,
		},
		{
			name:         "Xetra suffix",
			suffix:       ".DEX",
			expectedCode: "XETRA",
			wantNil:      false,
		},
		{
			name:         "No suffix (US)",
			suffix:       "",
			expectedCode: "NYSE",
			wantNil:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetExchangeFromAlphaVantageSuffix(tt.suffix)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedCode, result.Code)
			}
		})
	}
}

func TestExtractSymbolAndSuffix(t *testing.T) {
	tests := []struct {
		name           string
		fullSymbol     string
		expectedSymbol string
		expectedSuffix string
	}{
		{
			name:           "Greek stock",
			fullSymbol:     "OPAP.GR",
			expectedSymbol: "OPAP",
			expectedSuffix: ".GR",
		},
		{
			name:           "US stock with suffix",
			fullSymbol:     "AAPL.US",
			expectedSymbol: "AAPL",
			expectedSuffix: ".US",
		},
		{
			name:           "US stock without suffix",
			fullSymbol:     "AAPL",
			expectedSymbol: "AAPL",
			expectedSuffix: "",
		},
		{
			name:           "Hong Kong stock",
			fullSymbol:     "1810.HK",
			expectedSymbol: "1810",
			expectedSuffix: ".HK",
		},
		{
			name:           "Empty symbol",
			fullSymbol:     "",
			expectedSymbol: "",
			expectedSuffix: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbol, suffix := ExtractSymbolAndSuffix(tt.fullSymbol)
			assert.Equal(t, tt.expectedSymbol, symbol)
			assert.Equal(t, tt.expectedSuffix, suffix)
		})
	}
}

func TestConvertSymbol(t *testing.T) {
	tests := []struct {
		name        string
		symbol      string
		from        Provider
		to          Provider
		expected    string
		expectError bool
	}{
		// Tradernet to Yahoo
		{
			name:     "Tradernet GR to Yahoo",
			symbol:   "OPAP.GR",
			from:     ProviderTradernet,
			to:       ProviderYahoo,
			expected: "OPAP.AT",
		},
		{
			name:     "Tradernet JP to Yahoo",
			symbol:   "7203.JP",
			from:     ProviderTradernet,
			to:       ProviderYahoo,
			expected: "7203.T",
		},
		{
			name:     "Tradernet US to Yahoo",
			symbol:   "AAPL.US",
			from:     ProviderTradernet,
			to:       ProviderYahoo,
			expected: "AAPL",
		},
		// Tradernet to Alpha Vantage
		{
			name:     "Tradernet GR to AlphaVantage",
			symbol:   "OPAP.GR",
			from:     ProviderTradernet,
			to:       ProviderAlphaVantage,
			expected: "OPAP.AT",
		},
		{
			name:     "Tradernet EU (London) to AlphaVantage",
			symbol:   "BA.EU",
			from:     ProviderTradernet,
			to:       ProviderAlphaVantage,
			expected: "BA.LON",
		},
		// Yahoo to Alpha Vantage
		{
			name:     "Yahoo AT to AlphaVantage",
			symbol:   "OPAP.AT",
			from:     ProviderYahoo,
			to:       ProviderAlphaVantage,
			expected: "OPAP.AT",
		},
		{
			name:     "Yahoo L to AlphaVantage",
			symbol:   "BA.L",
			from:     ProviderYahoo,
			to:       ProviderAlphaVantage,
			expected: "BA.LON",
		},
		{
			name:     "Yahoo DE to AlphaVantage (Xetra)",
			symbol:   "BMW.DE",
			from:     ProviderYahoo,
			to:       ProviderAlphaVantage,
			expected: "BMW.DEX",
		},
		{
			name:     "Yahoo HK to AlphaVantage",
			symbol:   "1810.HK",
			from:     ProviderYahoo,
			to:       ProviderAlphaVantage,
			expected: "1810.HKG",
		},
		{
			name:     "Yahoo T to AlphaVantage",
			symbol:   "7203.T",
			from:     ProviderYahoo,
			to:       ProviderAlphaVantage,
			expected: "7203.TYO",
		},
		{
			name:     "Yahoo US (no suffix) to AlphaVantage",
			symbol:   "AAPL",
			from:     ProviderYahoo,
			to:       ProviderAlphaVantage,
			expected: "AAPL",
		},
		// Alpha Vantage to Yahoo
		{
			name:     "AlphaVantage LON to Yahoo",
			symbol:   "BA.LON",
			from:     ProviderAlphaVantage,
			to:       ProviderYahoo,
			expected: "BA.L",
		},
		{
			name:     "AlphaVantage DEX to Yahoo",
			symbol:   "BMW.DEX",
			from:     ProviderAlphaVantage,
			to:       ProviderYahoo,
			expected: "BMW.DE",
		},
		// Same provider (no conversion)
		{
			name:     "Same provider Yahoo",
			symbol:   "AAPL",
			from:     ProviderYahoo,
			to:       ProviderYahoo,
			expected: "AAPL",
		},
		// Error cases
		{
			name:        "Unknown suffix",
			symbol:      "TEST.XX",
			from:        ProviderTradernet,
			to:          ProviderYahoo,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertSymbol(tt.symbol, tt.from, tt.to)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestConvertSymbol_EuropeanExchanges(t *testing.T) {
	// Test that .EU suffix maps correctly to different European exchanges
	// based on the stored yahoo_symbol
	tests := []struct {
		name        string
		yahooSymbol string
		from        Provider
		to          Provider
		expected    string
	}{
		{
			name:        "Paris via Yahoo",
			yahooSymbol: "HO.PA",
			from:        ProviderYahoo,
			to:          ProviderAlphaVantage,
			expected:    "HO.PAR",
		},
		{
			name:        "Amsterdam via Yahoo",
			yahooSymbol: "ASML.AS",
			from:        ProviderYahoo,
			to:          ProviderAlphaVantage,
			expected:    "ASML.AMS",
		},
		{
			name:        "Milan via Yahoo",
			yahooSymbol: "LDO.MI",
			from:        ProviderYahoo,
			to:          ProviderAlphaVantage,
			expected:    "LDO.MIL",
		},
		{
			name:        "Frankfurt via Yahoo",
			yahooSymbol: "BMW.F",
			from:        ProviderYahoo,
			to:          ProviderAlphaVantage,
			expected:    "BMW.FRA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertSymbol(tt.yahooSymbol, tt.from, tt.to)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAllExchanges(t *testing.T) {
	exchanges := GetAllExchanges()
	assert.NotEmpty(t, exchanges)
	assert.GreaterOrEqual(t, len(exchanges), 11) // At least the documented exchanges
}

func TestGetExchangesByCountry(t *testing.T) {
	tests := []struct {
		name           string
		country        string
		expectedCount  int
		shouldHaveCode string
	}{
		{
			name:           "Germany",
			country:        "DE",
			expectedCount:  2, // FRA and XETRA
			shouldHaveCode: "XETRA",
		},
		{
			name:           "United States",
			country:        "US",
			expectedCount:  2, // NYSE and NASDAQ
			shouldHaveCode: "NYSE",
		},
		{
			name:          "Unknown country",
			country:       "XX",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exchanges := GetExchangesByCountry(tt.country)
			assert.Len(t, exchanges, tt.expectedCount)
			if tt.shouldHaveCode != "" {
				found := false
				for _, ex := range exchanges {
					if ex.Code == tt.shouldHaveCode {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find exchange %s", tt.shouldHaveCode)
			}
		})
	}
}
