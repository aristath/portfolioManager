package market_regime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Known Indices Tests
// ============================================================================

func TestGetKnownIndices_ReturnsAllRegions(t *testing.T) {
	indices := GetKnownIndices()

	// Count indices by region
	regionCounts := make(map[string]int)
	for _, idx := range indices {
		regionCounts[idx.Region]++
	}

	// US should have indices (SP500, NASDAQ, RUT, DJI30, NQX - VIX is VOLATILITY type)
	assert.GreaterOrEqual(t, regionCounts[RegionUS], 5, "US should have at least 5 PRICE indices")

	// EU should have indices (DAX, FTSE, FCHI, FTMIB, IBEX, OMXS30)
	assert.GreaterOrEqual(t, regionCounts[RegionEU], 6, "EU should have at least 6 indices")

	// Asia should have at least 1 index (HSI)
	assert.GreaterOrEqual(t, regionCounts[RegionAsia], 1, "Asia should have at least 1 index")
}

func TestGetKnownIndices_ContainsExpectedSymbols(t *testing.T) {
	indices := GetKnownIndices()

	// Build a set of symbols
	symbolSet := make(map[string]bool)
	for _, idx := range indices {
		symbolSet[idx.Symbol] = true
	}

	// Check expected US indices exist
	assert.True(t, symbolSet["SP500.IDX"], "SP500.IDX should exist")
	assert.True(t, symbolSet["NASDAQ.IDX"], "NASDAQ.IDX should exist")
	assert.True(t, symbolSet["DJI30.IDX"], "DJI30.IDX should exist")

	// Check expected EU indices exist
	assert.True(t, symbolSet["DAX.IDX"], "DAX.IDX should exist")
	assert.True(t, symbolSet["FTSE.IDX"], "FTSE.IDX should exist")

	// Check expected Asia indices exist
	assert.True(t, symbolSet["HSI.IDX"], "HSI.IDX should exist")
}

func TestGetKnownIndices_VIXIsVolatilityType(t *testing.T) {
	indices := GetKnownIndices()

	var vixIndex *KnownIndex
	for i := range indices {
		if indices[i].Symbol == "VIX.IDX" {
			vixIndex = &indices[i]
			break
		}
	}

	assert.NotNil(t, vixIndex, "VIX.IDX should exist")
	assert.Equal(t, IndexTypeVolatility, vixIndex.IndexType, "VIX.IDX should be VOLATILITY type")
	assert.Equal(t, RegionUS, vixIndex.Region, "VIX.IDX should be US region")
}

func TestGetKnownIndices_AllNonVIXArePriceType(t *testing.T) {
	indices := GetKnownIndices()

	for _, idx := range indices {
		if idx.Symbol == "VIX.IDX" {
			continue
		}
		assert.Equal(t, IndexTypePrice, idx.IndexType, "%s should be PRICE type", idx.Symbol)
	}
}

func TestGetKnownIndices_HasCorrectMarketCodes(t *testing.T) {
	indices := GetKnownIndices()

	for _, idx := range indices {
		switch idx.Region {
		case RegionUS:
			assert.Equal(t, "FIX", idx.MarketCode, "%s should have FIX market code", idx.Symbol)
		case RegionEU:
			assert.Equal(t, "EU", idx.MarketCode, "%s should have EU market code", idx.Symbol)
		case RegionAsia:
			assert.Equal(t, "HKEX", idx.MarketCode, "%s should have HKEX market code", idx.Symbol)
		}
	}
}

func TestGetPriceIndicesForRegion_ExcludesVIX(t *testing.T) {
	usIndices := GetPriceIndicesForRegion(RegionUS)

	for _, idx := range usIndices {
		assert.NotEqual(t, "VIX.IDX", idx.Symbol, "VIX.IDX should not be in PRICE indices")
		assert.Equal(t, IndexTypePrice, idx.IndexType, "All returned indices should be PRICE type")
	}
}

func TestGetPriceIndicesForRegion_EU(t *testing.T) {
	euIndices := GetPriceIndicesForRegion(RegionEU)

	assert.GreaterOrEqual(t, len(euIndices), 6, "EU should have at least 6 PRICE indices")

	for _, idx := range euIndices {
		assert.Equal(t, RegionEU, idx.Region, "All indices should be EU region")
		assert.Equal(t, IndexTypePrice, idx.IndexType, "All indices should be PRICE type")
	}
}

func TestGetPriceIndicesForRegion_Asia(t *testing.T) {
	asiaIndices := GetPriceIndicesForRegion(RegionAsia)

	assert.GreaterOrEqual(t, len(asiaIndices), 1, "Asia should have at least 1 PRICE index")

	for _, idx := range asiaIndices {
		assert.Equal(t, RegionAsia, idx.Region, "All indices should be ASIA region")
		assert.Equal(t, IndexTypePrice, idx.IndexType, "All indices should be PRICE type")
	}
}

func TestGetPriceIndicesForRegion_UnknownRegion(t *testing.T) {
	indices := GetPriceIndicesForRegion(RegionRussia)

	assert.Empty(t, indices, "RUSSIA should have no indices")
}
