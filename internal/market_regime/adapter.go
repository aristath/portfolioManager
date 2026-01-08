package market_regime

// RegimeScoreProviderAdapter adapts RegimePersistence to consumers that expect a float64 regime score.
type RegimeScoreProviderAdapter struct {
	persistence *RegimePersistence
}

// NewRegimeScoreProviderAdapter creates a new adapter.
func NewRegimeScoreProviderAdapter(persistence *RegimePersistence) *RegimeScoreProviderAdapter {
	return &RegimeScoreProviderAdapter{persistence: persistence}
}

// GetCurrentRegimeScore returns the current regime score as a float64.
func (a *RegimeScoreProviderAdapter) GetCurrentRegimeScore() (float64, error) {
	score, err := a.persistence.GetCurrentRegimeScore()
	return float64(score), err
}
