package universe

// ScoreRepositoryInterface defines the contract for score repository operations
type ScoreRepositoryInterface interface {
	// GetByISIN returns a score by ISIN (primary method)
	GetByISIN(isin string) (*SecurityScore, error)

	// GetBySymbol returns a score by symbol (helper method - looks up ISIN first)
	// This requires universeDB to lookup ISIN from securities table
	GetBySymbol(symbol string) (*SecurityScore, error)

	// GetByIdentifier returns a score by symbol or ISIN
	GetByIdentifier(identifier string) (*SecurityScore, error)

	// GetAll returns all scores
	GetAll() ([]SecurityScore, error)

	// GetTop returns top scored securities
	GetTop(limit int) ([]SecurityScore, error)

	// Upsert inserts or updates a score
	Upsert(score SecurityScore) error

	// Delete deletes score by ISIN
	Delete(isin string) error

	// DeleteAll deletes all scores
	DeleteAll() error
}

// Compile-time check that ScoreRepository implements ScoreRepositoryInterface
var _ ScoreRepositoryInterface = (*ScoreRepository)(nil)
