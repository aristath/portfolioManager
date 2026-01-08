package portfolio

// PositionRepositoryInterface defines the contract for position repository operations
// Used by PortfolioService to enable testing with mocks
type PositionRepositoryInterface interface {
	// GetAll returns all positions
	GetAll() ([]Position, error)

	// GetWithSecurityInfo returns all positions with security information
	GetWithSecurityInfo() ([]PositionWithSecurity, error)

	// GetBySymbol returns a position by symbol (helper method - looks up ISIN first)
	GetBySymbol(symbol string) (*Position, error)

	// GetByISIN returns a position by ISIN (primary method)
	GetByISIN(isin string) (*Position, error)

	// GetByIdentifier returns a position by symbol or ISIN
	GetByIdentifier(identifier string) (*Position, error)

	// GetCount returns the total number of positions
	GetCount() (int, error)

	// GetTotalValue returns the total value of all positions
	GetTotalValue() (float64, error)

	// Upsert inserts or updates a position
	Upsert(position Position) error

	// Delete deletes a position by ISIN
	Delete(isin string) error

	// DeleteAll deletes all positions
	DeleteAll() error

	// UpdatePrice updates the price for a position by ISIN
	UpdatePrice(isin string, price float64, currencyRate float64) error

	// UpdateLastSoldAt updates the last sold timestamp for a position by ISIN
	UpdateLastSoldAt(isin string) error
}

// Compile-time check that PositionRepository implements PositionRepositoryInterface
var _ PositionRepositoryInterface = (*PositionRepository)(nil)

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

// Note: TradernetClientInterface and CurrencyExchangeServiceInterface have been moved to domain/interfaces.go
// They are now available as domain.TradernetClientInterface and domain.CurrencyExchangeServiceInterface
