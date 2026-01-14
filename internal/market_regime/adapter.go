package market_regime

// RegimeScoreProviderAdapter adapts RegimePersistence to consumers
// that expect float64 regime scores. Supports both global and per-region queries.
type RegimeScoreProviderAdapter struct {
	persistence *RegimePersistence
}

// NewRegimeScoreProviderAdapter creates a new adapter.
func NewRegimeScoreProviderAdapter(persistence *RegimePersistence) *RegimeScoreProviderAdapter {
	return &RegimeScoreProviderAdapter{persistence: persistence}
}

// GetCurrentRegimeScore returns the current global regime score as a float64.
// Returns the GLOBAL_AVERAGE score (average of US, EU, ASIA regions).
// Falls back to calculating from stored regional scores if GLOBAL_AVERAGE not stored.
func (a *RegimeScoreProviderAdapter) GetCurrentRegimeScore() (float64, error) {
	// Try to get the stored GLOBAL_AVERAGE first
	score, err := a.persistence.GetCurrentRegimeScoreForRegion("GLOBAL_AVERAGE")
	if err == nil && score != NeutralScore {
		return float64(score), nil
	}

	// Fallback: calculate from stored regional scores
	allScores, err := a.persistence.GetAllCurrentScores()
	if err != nil {
		return 0.0, err
	}

	// If no regional scores exist, return neutral
	if len(allScores) == 0 {
		return 0.0, nil
	}

	return CalculateGlobalAverage(allScores), nil
}

// GetRegimeScoreForRegion returns the regime score for a specific region.
// Regions with dedicated indices (US, EU, ASIA) return their calculated score.
// Other regions (RUSSIA, MIDDLE_EAST, CENTRAL_ASIA, UNKNOWN) return the global average.
func (a *RegimeScoreProviderAdapter) GetRegimeScoreForRegion(region string) (float64, error) {
	// Check if this region has dedicated indices
	if RegionHasIndices(region) {
		score, err := a.persistence.GetCurrentRegimeScoreForRegion(region)
		if err != nil {
			return 0.0, err
		}
		return float64(score), nil
	}

	// For regions without indices, return global average
	// First try to get it from stored GLOBAL_AVERAGE
	score, err := a.persistence.GetCurrentRegimeScoreForRegion("GLOBAL_AVERAGE")
	if err == nil && score != NeutralScore {
		return float64(score), nil
	}

	// Fallback: calculate global average from stored regional scores
	allScores, err := a.persistence.GetAllCurrentScores()
	if err != nil {
		return 0.0, err
	}

	return CalculateGlobalAverage(allScores), nil
}

// GetRegimeScoreForMarketCode returns the regime score for a security based on its market code.
// This maps market code -> region -> regime score.
func (a *RegimeScoreProviderAdapter) GetRegimeScoreForMarketCode(marketCode string) (float64, error) {
	region := GetRegionForMarketCode(marketCode)
	return a.GetRegimeScoreForRegion(region)
}
