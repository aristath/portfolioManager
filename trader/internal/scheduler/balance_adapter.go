package scheduler

import (
	"github.com/aristath/portfolioManager/internal/modules/portfolio"
	"github.com/rs/zerolog"
)

// BalanceAdapter adapts portfolio.CashManager to implement BalanceServiceInterface
// This allows the sync cycle to check for negative balances using the cash manager
type BalanceAdapter struct {
	cashManager portfolio.CashManager
	log         zerolog.Logger
}

// NewBalanceAdapter creates a new balance adapter
func NewBalanceAdapter(cashManager portfolio.CashManager, log zerolog.Logger) *BalanceAdapter {
	return &BalanceAdapter{
		cashManager: cashManager,
		log:         log.With().Str("component", "balance_adapter").Logger(),
	}
}

// GetAllCurrencies returns all currencies that have cash balances
func (a *BalanceAdapter) GetAllCurrencies() ([]string, error) {
	balances, err := a.cashManager.GetAllCashBalances()
	if err != nil {
		return nil, err
	}

	currencies := make([]string, 0, len(balances))
	for currency := range balances {
		currencies = append(currencies, currency)
	}

	return currencies, nil
}

// GetTotalByCurrency returns the total cash balance for a specific currency
func (a *BalanceAdapter) GetTotalByCurrency(currency string) (float64, error) {
	balances, err := a.cashManager.GetAllCashBalances()
	if err != nil {
		return 0, err
	}

	// Return 0.0 if currency not found (not an error)
	return balances[currency], nil
}
