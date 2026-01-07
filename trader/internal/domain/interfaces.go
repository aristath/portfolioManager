package domain

import (
	"github.com/aristath/portfolioManager/internal/clients/tradernet"
)

// CashManager defines operations for managing cash balances
// This interface breaks circular dependencies between portfolio, cash_flows, and services packages
// Merged from:
//   - portfolio.CashManager (UpdateCashPosition, GetAllCashBalances)
//   - services.CashManagerInterface (GetCashBalance)
type CashManager interface {
	// UpdateCashPosition updates or creates a cash balance for the given currency
	UpdateCashPosition(currency string, balance float64) error

	// GetAllCashBalances returns all cash balances as a map of currency -> balance
	GetAllCashBalances() (map[string]float64, error)

	// GetCashBalance returns the cash balance for the given currency
	// Returns 0 if no balance exists (not an error)
	GetCashBalance(currency string) (float64, error)
}

// TradernetClientInterface defines the contract for Tradernet client operations
// Merged from multiple packages to provide a unified interface
// Used by portfolio, trading, services, and optimization packages
type TradernetClientInterface interface {
	// GetPortfolio retrieves all positions from Tradernet
	GetPortfolio() ([]tradernet.Position, error)

	// GetCashBalances retrieves all cash balances from Tradernet
	GetCashBalances() ([]tradernet.CashBalance, error)

	// GetExecutedTrades retrieves executed trades from Tradernet
	GetExecutedTrades(limit int) ([]tradernet.Trade, error)

	// PlaceOrder places an order via Tradernet
	PlaceOrder(symbol, side string, quantity float64) (*tradernet.OrderResult, error)

	// IsConnected checks if the Tradernet client is connected
	IsConnected() bool
}

// CurrencyExchangeServiceInterface defines the contract for currency exchange operations
// Used by portfolio, services, optimization, and universe packages
// This interface avoids import cycles with the services package
type CurrencyExchangeServiceInterface interface {
	// GetRate returns the exchange rate from one currency to another
	GetRate(fromCurrency, toCurrency string) (float64, error)
}

// AllocationTargetProvider provides access to allocation targets
// This interface breaks the circular dependency between portfolio and allocation packages
type AllocationTargetProvider interface {
	// GetAll returns all allocation targets as a map of name -> target percentage
	GetAll() (map[string]float64, error)
}

// PortfolioSummaryProvider provides portfolio summary data without creating
// a dependency on the portfolio package. This interface breaks the circular
// dependency: allocation → portfolio → cash_flows → trading → allocation
type PortfolioSummaryProvider interface {
	// GetPortfolioSummary returns the current portfolio summary
	GetPortfolioSummary() (PortfolioSummary, error)
}

// ConcentrationAlertProvider provides concentration alert detection without
// requiring direct dependency on ConcentrationAlertService. This interface
// breaks the circular dependency: trading → allocation
type ConcentrationAlertProvider interface {
	// DetectAlerts detects concentration alerts from a portfolio summary
	DetectAlerts(summary PortfolioSummary) ([]ConcentrationAlert, error)
}

// PortfolioSummary represents complete portfolio allocation summary
// This is the domain model used by the allocation package
// Note: This is different from portfolio.PortfolioSummary (which uses AllocationStatus)
// This version uses PortfolioAllocation for compatibility with allocation package
type PortfolioSummary struct {
	CountryAllocations  []PortfolioAllocation
	IndustryAllocations []PortfolioAllocation
	TotalValue          float64
	CashBalance         float64
}

// PortfolioAllocation represents allocation info for display
type PortfolioAllocation struct {
	Name         string
	TargetPct    float64
	CurrentPct   float64
	CurrentValue float64
	Deviation    float64
}

// ConcentrationAlert represents alert for approaching concentration limit
type ConcentrationAlert struct {
	Type              string
	Name              string
	Severity          string
	CurrentPct        float64
	LimitPct          float64
	AlertThresholdPct float64
}
