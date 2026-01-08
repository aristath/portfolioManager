package cash_flows

// RepositoryInterface defines the contract for cash flow repository operations
type RepositoryInterface interface {
	// Create inserts a new cash flow
	Create(cashFlow *CashFlow) (*CashFlow, error)

	// GetByTransactionID retrieves a cash flow by transaction ID
	GetByTransactionID(transactionID string) (*CashFlow, error)

	// Exists checks if a transaction ID exists
	Exists(transactionID string) (bool, error)

	// GetAll retrieves all cash flows with optional limit
	GetAll(limit *int) ([]CashFlow, error)

	// GetByDateRange retrieves cash flows within a date range
	// startDate and endDate are in YYYY-MM-DD format
	GetByDateRange(startDate, endDate string) ([]CashFlow, error)

	// GetByType retrieves cash flows by transaction type
	GetByType(txType string) ([]CashFlow, error)

	// SyncFromAPI syncs transactions from API, returns count of newly inserted
	SyncFromAPI(transactions []APITransaction) (int, error)

	// GetTotalDeposits calculates total deposits
	GetTotalDeposits() (float64, error)

	// GetTotalWithdrawals calculates total withdrawals
	GetTotalWithdrawals() (float64, error)

	// GetCashBalanceHistory calculates running cash balance
	// startDate and endDate are in YYYY-MM-DD format
	GetCashBalanceHistory(startDate, endDate string, initialCash float64) ([]BalancePoint, error)
}

// Compile-time check that Repository implements RepositoryInterface
var _ RepositoryInterface = (*Repository)(nil)
