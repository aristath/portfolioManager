package dividends

// DividendRepositoryInterface defines the contract for dividend repository operations
type DividendRepositoryInterface interface {
	// Create creates a new dividend record
	Create(dividend *DividendRecord) error

	// GetByID retrieves a dividend record by ID
	GetByID(id int) (*DividendRecord, error)

	// GetByCashFlowID retrieves a dividend record linked to a cash flow
	GetByCashFlowID(cashFlowID int) (*DividendRecord, error)

	// ExistsForCashFlow checks if a dividend record already exists for a cash flow
	ExistsForCashFlow(cashFlowID int) (bool, error)

	// GetBySymbol retrieves all dividend records for a symbol (helper method - looks up ISIN first)
	GetBySymbol(symbol string) ([]DividendRecord, error)

	// GetByISIN retrieves all dividend records for an ISIN
	GetByISIN(isin string) ([]DividendRecord, error)

	// GetByIdentifier retrieves dividend records by symbol or ISIN
	GetByIdentifier(identifier string) ([]DividendRecord, error)

	// GetAll retrieves all dividend records, optionally limited
	GetAll(limit int) ([]DividendRecord, error)

	// GetPendingBonuses retrieves all pending dividend bonuses by symbol
	GetPendingBonuses() (map[string]float64, error)

	// GetPendingBonus retrieves pending dividend bonus for a specific symbol
	GetPendingBonus(symbol string) (float64, error)

	// MarkReinvested marks a dividend as reinvested (DRIP executed)
	MarkReinvested(dividendID int, quantity int) error

	// SetPendingBonus sets pending bonus for a dividend that couldn't be reinvested
	SetPendingBonus(dividendID int, bonus float64) error

	// ClearBonus clears pending bonuses for a symbol (after security is bought)
	ClearBonus(symbol string) (int, error)

	// GetUnreinvestedDividends retrieves dividends that haven't been reinvested yet
	GetUnreinvestedDividends(minAmountEUR float64) ([]DividendRecord, error)

	// GetTotalDividendsBySymbol retrieves total dividends received per symbol (in EUR)
	GetTotalDividendsBySymbol() (map[string]float64, error)

	// GetTotalReinvested retrieves total amount of dividends that were reinvested (in EUR)
	GetTotalReinvested() (float64, error)

	// GetReinvestmentRate retrieves dividend reinvestment rate (0.0 to 1.0)
	GetReinvestmentRate() (float64, error)
}

// Compile-time check that DividendRepository implements DividendRepositoryInterface
var _ DividendRepositoryInterface = (*DividendRepository)(nil)
