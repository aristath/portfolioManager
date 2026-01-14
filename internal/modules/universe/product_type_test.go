package universe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductTypeIndex_Value(t *testing.T) {
	// ProductTypeIndex constant should have value "INDEX"
	assert.Equal(t, ProductType("INDEX"), ProductTypeIndex)
}

func TestFromQuoteType_Index(t *testing.T) {
	// INDEX quote type should return ProductTypeIndex
	result := FromQuoteType("INDEX", "S&P 500")
	assert.Equal(t, ProductTypeIndex, result)

	// Case insensitive
	result = FromQuoteType("index", "NASDAQ Composite")
	assert.Equal(t, ProductTypeIndex, result)
}

func TestFromQuoteType_ExistingTypes(t *testing.T) {
	// Verify existing types still work correctly
	tests := []struct {
		quoteType   string
		productName string
		expected    ProductType
	}{
		{"EQUITY", "Apple Inc.", ProductTypeEquity},
		{"ETF", "Vanguard S&P 500 ETF", ProductTypeETF},
		{"MUTUALFUND", "Vanguard Total World ETF", ProductTypeETF},     // Has "ETF" in name
		{"MUTUALFUND", "WisdomTree Physical Gold ETC", ProductTypeETC}, // Has "ETC" in name
		{"MUTUALFUND", "iShares Gold Trust", ProductTypeETC},           // Has "Gold" in name
		{"MUTUALFUND", "Fidelity Contrafund", ProductTypeMutualFund},   // No indicators
		{"", "Some Security", ProductTypeUnknown},                      // Empty quote type
		{"CURRENCY", "US Dollar", ProductTypeUnknown},                  // Unknown type
	}

	for _, tt := range tests {
		t.Run(tt.quoteType+"_"+tt.productName, func(t *testing.T) {
			result := FromQuoteType(tt.quoteType, tt.productName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsTradable_TradableTypes(t *testing.T) {
	tradableTypes := []ProductType{
		ProductTypeEquity,
		ProductTypeETF,
		ProductTypeETC,
		ProductTypeMutualFund,
	}

	for _, pt := range tradableTypes {
		t.Run(string(pt), func(t *testing.T) {
			assert.True(t, pt.IsTradable(), "ProductType %s should be tradable", pt)
		})
	}
}

func TestIsTradable_NonTradableTypes(t *testing.T) {
	nonTradableTypes := []ProductType{
		ProductTypeIndex,
		ProductTypeUnknown,
	}

	for _, pt := range nonTradableTypes {
		t.Run(string(pt), func(t *testing.T) {
			assert.False(t, pt.IsTradable(), "ProductType %s should not be tradable", pt)
		})
	}
}

func TestIsIndex(t *testing.T) {
	// Only ProductTypeIndex should return true
	assert.True(t, ProductTypeIndex.IsIndex())

	// All other types should return false
	otherTypes := []ProductType{
		ProductTypeEquity,
		ProductTypeETF,
		ProductTypeETC,
		ProductTypeMutualFund,
		ProductTypeUnknown,
	}

	for _, pt := range otherTypes {
		t.Run(string(pt), func(t *testing.T) {
			assert.False(t, pt.IsIndex(), "ProductType %s should not be an index", pt)
		})
	}
}
