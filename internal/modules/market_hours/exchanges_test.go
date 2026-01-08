package market_hours

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExchangeCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Valid exchange codes
		{"XNYS code", "XNYS", "XNYS"},
		{"XNAS code", "XNAS", "XNAS"},
		{"XETR code", "XETR", "XETR"},
		{"XHKG code", "XHKG", "XHKG"},
		{"XSHG code", "XSHG", "XSHG"},
		{"XTSE code", "XTSE", "XTSE"},
		{"XASX code", "XASX", "XASX"},

		// Database names (case-sensitive)
		{"NYSE database name", "NYSE", "XNYS"},
		{"NASDAQ database name", "NASDAQ", "XNAS"},
		{"XETRA database name", "XETRA", "XETR"},
		{"HKSE database name", "HKSE", "XHKG"},
		{"London database name", "London", "XLON"},
		{"LSE database name", "LSE", "XLON"},
		{"Paris database name", "Paris", "XPAR"},
		{"Milan database name", "Milan", "XMIL"},
		{"Amsterdam database name", "Amsterdam", "XAMS"},
		{"Copenhagen database name", "Copenhagen", "XCSE"},
		{"Athens database name", "Athens", "ASEX"},
		{"Hong Kong database name", "Hong Kong", "XHKG"},
		{"Shanghai database name", "Shanghai", "XSHG"},
		{"Shenzhen database name", "Shenzhen", "XSHG"},
		{"Frankfurt database name", "Frankfurt", "XETR"},
		{"Tokyo database name", "Tokyo", "XTSE"},
		{"Sydney database name", "Sydney", "XASX"},

		// Alternative names
		{"NasdaqCM", "NasdaqCM", "XNAS"},
		{"NasdaqGS", "NasdaqGS", "XNAS"},
		{"New York", "New York", "XNYS"},

		// Legacy mappings
		{"TSE legacy", "TSE", "XTSE"},
		{"ASX legacy", "ASX", "XASX"},

		// Case-insensitive lookups
		{"nyse lowercase", "nyse", "XNYS"},
		{"nasdaq lowercase", "nasdaq", "XNAS"},
		{"london lowercase", "london", "XLON"},
		{"HONG KONG mixed case", "HONG KONG", "XHKG"},

		// Whitespace handling
		{"NYSE with whitespace", "  NYSE  ", "XNYS"},
		{"  London  with spaces", "  London  ", "XLON"},

		// Unknown exchange (should default to XNYS)
		{"unknown exchange", "UnknownExchange", "XNYS"},
		{"empty string", "", "XNYS"},
		{"random string", "RandomExchange123", "XNYS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetExchangeCode(tt.input)
			assert.Equal(t, tt.expected, result, "GetExchangeCode(%q) = %q, want %q", tt.input, result, tt.expected)
		})
	}
}

func TestGetExchangeCode_StrictMarketHours(t *testing.T) {
	// Test that strict market hours exchanges are properly identified
	strictExchanges := []string{"XHKG", "XSHG", "XTSE", "XASX", "HKSE", "Shenzhen", "Tokyo", "Sydney", "ASX", "TSE"}

	for _, exchange := range strictExchanges {
		t.Run(exchange, func(t *testing.T) {
			code := GetExchangeCode(exchange)
			// These should map to codes that have StrictHours: true
			assert.Contains(t, []string{"XHKG", "XSHG", "XTSE", "XASX"}, code, "Exchange %s should map to a strict hours exchange", exchange)
		})
	}
}

func TestGetExchangeConfig(t *testing.T) {
	// Test that getExchangeConfig returns correct configs
	tests := []struct {
		name           string
		exchangeCode   string
		shouldExist    bool
		expectedName   string
		expectedStrict bool
	}{
		{"XNYS exists", "XNYS", true, "New York Stock Exchange", false},
		{"XNAS exists", "XNAS", true, "NASDAQ", false},
		{"XETR exists", "XETR", true, "XETRA (Frankfurt)", false},
		{"XLON exists", "XLON", true, "London Stock Exchange", false},
		{"XPAR exists", "XPAR", true, "Euronext Paris", false},
		{"XMIL exists", "XMIL", true, "Borsa Italiana (Milan)", false},
		{"XAMS exists", "XAMS", true, "Euronext Amsterdam", false},
		{"XCSE exists", "XCSE", true, "Copenhagen Stock Exchange", false},
		{"ASEX exists", "ASEX", true, "Athens Stock Exchange", false},
		{"XHKG exists", "XHKG", true, "Hong Kong Stock Exchange", true},
		{"XSHG exists", "XSHG", true, "Shanghai Stock Exchange", true},
		{"XTSE exists", "XTSE", true, "Tokyo Stock Exchange", true},
		{"XASX exists", "XASX", true, "Australian Securities Exchange", true},
		{"unknown defaults to XNYS", "UNKNOWN", true, "New York Stock Exchange", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := getExchangeConfig(tt.exchangeCode)
			if tt.shouldExist {
				assert.NotNil(t, config, "Config for %s should exist", tt.exchangeCode)
				if config != nil {
					assert.Equal(t, tt.expectedName, config.Name, "Config name mismatch")
					assert.Equal(t, tt.expectedStrict, config.StrictHours, "StrictHours mismatch for %s", tt.exchangeCode)
					assert.Equal(t, tt.exchangeCode == "UNKNOWN" && config.Code == "XNYS" || config.Code == tt.exchangeCode, true, "Code should match or default to XNYS")
				}
			} else {
				assert.Nil(t, config, "Config for %s should not exist", tt.exchangeCode)
			}
		})
	}
}
