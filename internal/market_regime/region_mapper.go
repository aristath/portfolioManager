// Package market_regime provides market regime detection and analysis functionality.
package market_regime

// Region constants for market classification
const (
	RegionUS          = "US"
	RegionEU          = "EU"
	RegionAsia        = "ASIA"
	RegionRussia      = "RUSSIA"
	RegionMiddleEast  = "MIDDLE_EAST"
	RegionCentralAsia = "CENTRAL_ASIA"
	RegionUnknown     = "UNKNOWN"
)

// regionsWithIndices tracks which regions have dedicated indices for regime detection.
// Regions not in this map use global average fallback.
var regionsWithIndices = map[string]bool{
	RegionUS:   true,
	RegionEU:   true,
	RegionAsia: true,
}

// marketCodeToRegion maps Tradernet market codes to regions.
// Market codes come from the 'mkt' field in Tradernet API responses.
var marketCodeToRegion = map[string]string{
	// US markets (NYSE, NASDAQ)
	"FIX": RegionUS,

	// European markets
	"EU":    RegionEU,
	"ATHEX": RegionEU, // Athens Stock Exchange

	// Asian markets
	"HKEX": RegionAsia, // Hong Kong
	"HKG":  RegionAsia, // Hong Kong alternative code

	// Russian markets (no indices available)
	"FORTS": RegionRussia, // Moscow Exchange derivatives
	"MCX":   RegionRussia, // Moscow Exchange

	// Middle East (no indices available)
	"TABADUL": RegionMiddleEast, // Saudi Arabia

	// Central Asia (no indices available)
	"KASE": RegionCentralAsia, // Kazakhstan
}

// GetRegionForMarketCode returns the region for a given market code.
// Returns RegionUnknown if the market code is not recognized.
func GetRegionForMarketCode(marketCode string) string {
	if region, ok := marketCodeToRegion[marketCode]; ok {
		return region
	}
	return RegionUnknown
}

// RegionHasIndices returns true if the region has dedicated indices for regime detection.
// Regions without indices (RUSSIA, MIDDLE_EAST, CENTRAL_ASIA, UNKNOWN) use global average.
func RegionHasIndices(region string) bool {
	return regionsWithIndices[region]
}

// GetAllRegionsWithIndices returns a list of all regions that have dedicated indices.
func GetAllRegionsWithIndices() []string {
	regions := make([]string, 0, len(regionsWithIndices))
	for region := range regionsWithIndices {
		regions = append(regions, region)
	}
	return regions
}

// CalculateGlobalAverage calculates the weighted average of regime scores
// from regions that have dedicated indices (US, EU, ASIA).
// Returns 0.0 (neutral) if no regions with indices have data.
func CalculateGlobalAverage(regionScores map[string]float64) float64 {
	sum := 0.0
	count := 0.0

	for region, score := range regionScores {
		if RegionHasIndices(region) {
			sum += score
			count++
		}
	}

	if count == 0 {
		return 0.0 // True neutral only if ALL regions have no data
	}

	return sum / count
}
