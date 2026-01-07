package portfolio

// PositionRepositoryInterface defines the contract for position repository operations
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
