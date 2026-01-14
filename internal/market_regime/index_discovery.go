// Package market_regime provides market regime detection and analysis functionality.
package market_regime

// Index type constants
const (
	IndexTypePrice      = "PRICE"      // Normal price indices used in regime composite
	IndexTypeVolatility = "VOLATILITY" // VIX-style volatility indices, excluded from composite
)

// KnownIndex represents a market index with its metadata
type KnownIndex struct {
	Symbol     string // Tradernet symbol (e.g., "SP500.IDX")
	Name       string // Human-readable name
	MarketCode string // Tradernet market code (FIX, EU, HKEX)
	Region     string // Region constant (US, EU, ASIA)
	IndexType  string // PRICE or VOLATILITY
}

// knownIndices is the master list of market indices available on Tradernet.
// These are used for per-region market regime detection.
//
// Index Types:
// - PRICE: Normal price indices used to calculate regime composite score
// - VOLATILITY: VIX-style indices that move inversely to prices (excluded from composite)
//
// Regions:
// - US (FIX): NYSE/NASDAQ indices
// - EU (EU): European indices
// - ASIA (HKEX): Asian indices
var knownIndices = []KnownIndex{
	// US Indices (market_code = FIX)
	{Symbol: "SP500.IDX", Name: "S&P 500", MarketCode: "FIX", Region: RegionUS, IndexType: IndexTypePrice},
	{Symbol: "NASDAQ.IDX", Name: "NASDAQ Composite", MarketCode: "FIX", Region: RegionUS, IndexType: IndexTypePrice},
	{Symbol: "DJI30.IDX", Name: "Dow Jones Industrial Average", MarketCode: "FIX", Region: RegionUS, IndexType: IndexTypePrice},
	{Symbol: "RUT.IDX", Name: "Russell 2000", MarketCode: "FIX", Region: RegionUS, IndexType: IndexTypePrice},
	{Symbol: "NQX.IDX", Name: "NASDAQ 100", MarketCode: "FIX", Region: RegionUS, IndexType: IndexTypePrice},
	{Symbol: "VIX.IDX", Name: "CBOE Volatility Index", MarketCode: "FIX", Region: RegionUS, IndexType: IndexTypeVolatility},

	// EU Indices (market_code = EU)
	{Symbol: "DAX.IDX", Name: "DAX (Germany)", MarketCode: "EU", Region: RegionEU, IndexType: IndexTypePrice},
	{Symbol: "FTSE.IDX", Name: "FTSE 100 (UK)", MarketCode: "EU", Region: RegionEU, IndexType: IndexTypePrice},
	{Symbol: "FCHI.IDX", Name: "CAC 40 (France)", MarketCode: "EU", Region: RegionEU, IndexType: IndexTypePrice},
	{Symbol: "FTMIB.IDX", Name: "FTSE MIB (Italy)", MarketCode: "EU", Region: RegionEU, IndexType: IndexTypePrice},
	{Symbol: "IBEX.IDX", Name: "IBEX 35 (Spain)", MarketCode: "EU", Region: RegionEU, IndexType: IndexTypePrice},
	{Symbol: "OMXS30.IDX", Name: "OMX Stockholm 30 (Sweden)", MarketCode: "EU", Region: RegionEU, IndexType: IndexTypePrice},

	// Asia Indices (market_code = HKEX)
	{Symbol: "HSI.IDX", Name: "Hang Seng Index", MarketCode: "HKEX", Region: RegionAsia, IndexType: IndexTypePrice},
}

// GetKnownIndices returns all known market indices
func GetKnownIndices() []KnownIndex {
	// Return a copy to prevent modification
	result := make([]KnownIndex, len(knownIndices))
	copy(result, knownIndices)
	return result
}

// GetPriceIndicesForRegion returns only PRICE-type indices for a given region.
// Excludes VOLATILITY indices (VIX) from the result.
func GetPriceIndicesForRegion(region string) []KnownIndex {
	var result []KnownIndex
	for _, idx := range knownIndices {
		if idx.Region == region && idx.IndexType == IndexTypePrice {
			result = append(result, idx)
		}
	}
	return result
}

// GetAllPriceIndices returns all PRICE-type indices across all regions.
// Excludes VOLATILITY indices (VIX) from the result.
func GetAllPriceIndices() []KnownIndex {
	var result []KnownIndex
	for _, idx := range knownIndices {
		if idx.IndexType == IndexTypePrice {
			result = append(result, idx)
		}
	}
	return result
}

// GetIndexSymbolsForRegion returns just the symbol strings for a region's PRICE indices.
func GetIndexSymbolsForRegion(region string) []string {
	indices := GetPriceIndicesForRegion(region)
	symbols := make([]string, len(indices))
	for i, idx := range indices {
		symbols[i] = idx.Symbol
	}
	return symbols
}
