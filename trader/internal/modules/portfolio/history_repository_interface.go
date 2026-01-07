package portfolio

// HistoryRepositoryInterface defines the contract for history repository operations
type HistoryRepositoryInterface interface {
	// GetDailyRange retrieves daily price history within a date range
	// startDate and endDate are in YYYY-MM-DD format
	GetDailyRange(startDate, endDate string) ([]DailyPrice, error)

	// GetLatestPrice retrieves the most recent price
	GetLatestPrice() (*DailyPrice, error)
}

// Compile-time check that HistoryRepository implements HistoryRepositoryInterface
var _ HistoryRepositoryInterface = (*HistoryRepository)(nil)
