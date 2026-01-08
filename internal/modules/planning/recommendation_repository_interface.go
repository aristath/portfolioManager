package planning

// RecommendationRepositoryInterface defines the contract for recommendation repository operations
type RecommendationRepositoryInterface interface {
	// CreateOrUpdate creates or updates a recommendation
	CreateOrUpdate(rec Recommendation) (string, error)

	// FindMatchingForExecution finds recommendations matching execution criteria
	FindMatchingForExecution(symbol, side, portfolioHash string) ([]Recommendation, error)

	// MarkExecuted marks a recommendation as executed
	MarkExecuted(recUUID string) error

	// CountPendingBySide returns count of pending recommendations by side (buy/sell)
	CountPendingBySide() (buyCount int, sellCount int, err error)

	// DismissAllByPortfolioHash dismisses all recommendations for a portfolio hash
	DismissAllByPortfolioHash(portfolioHash string) (int, error)

	// DismissAllPending dismisses all pending recommendations
	DismissAllPending() (int, error)

	// GetPendingRecommendations retrieves pending recommendations with optional limit
	GetPendingRecommendations(limit int) ([]Recommendation, error)

	// GetRecommendationsAsPlan returns recommendations formatted as a plan
	// startingCashEUR is the starting cash balance in EUR (optional, defaults to 0 if not provided)
	GetRecommendationsAsPlan(getEvaluatedCount func(portfolioHash string) (int, error), startingCashEUR float64) (map[string]interface{}, error)
}

// Compile-time check that RecommendationRepository implements RecommendationRepositoryInterface
var _ RecommendationRepositoryInterface = (*RecommendationRepository)(nil)
