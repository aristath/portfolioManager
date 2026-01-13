package domain

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

// BrokerClient defines broker-agnostic trading and portfolio operations
// This interface abstracts away broker-specific implementations (Tradernet, IBKR, etc.)
// All broker operations should go through this interface for maximum flexibility
type BrokerClient interface {
	// Portfolio operations
	GetPortfolio() ([]BrokerPosition, error)
	GetCashBalances() ([]BrokerCashBalance, error)

	// Trading operations
	PlaceOrder(symbol, side string, quantity, limitPrice float64) (*BrokerOrderResult, error)
	GetExecutedTrades(limit int) ([]BrokerTrade, error)
	GetPendingOrders() ([]BrokerPendingOrder, error)

	// Market data operations
	GetQuote(symbol string) (*BrokerQuote, error)
	// GetQuotes fetches quotes for multiple symbols in a single batch call
	// Returns a map of symbol -> quote for all successfully retrieved quotes
	// Symbols not found are simply omitted from the result (not an error)
	GetQuotes(symbols []string) (map[string]*BrokerQuote, error)
	// GetLevel1Quote fetches Level 1 market data (best bid and best ask only)
	// Returns BrokerOrderBook with Bids[0] and Asks[0] populated
	GetLevel1Quote(symbol string) (*BrokerOrderBook, error)
	// GetHistoricalPrices fetches OHLCV candlestick data for a symbol
	// timeframeSeconds: 86400 for daily, 3600 for hourly, etc.
	GetHistoricalPrices(symbol string, start, end int64, timeframeSeconds int) ([]BrokerOHLCV, error)
	FindSymbol(symbol string, exchange *string) ([]BrokerSecurityInfo, error)
	// GetFXRates retrieves currency exchange rates for today's date
	// Returns a map of currency codes to exchange rates relative to baseCurrency
	GetFXRates(baseCurrency string, currencies []string) (map[string]float64, error)

	// Cash operations
	GetAllCashFlows(limit int) ([]BrokerCashFlow, error)
	GetCashMovements() (*BrokerCashMovement, error)

	// Connection & health
	IsConnected() bool
	HealthCheck() (*BrokerHealthResult, error)
	SetCredentials(apiKey, apiSecret string)
}

// CurrencyExchangeServiceInterface defines the contract for currency exchange operations
// Used by portfolio, services, optimization, and universe packages
// This interface avoids import cycles with the services package
type CurrencyExchangeServiceInterface interface {
	// GetRate returns the exchange rate from one currency to another
	GetRate(fromCurrency, toCurrency string) (float64, error)

	// EnsureBalance ensures there is sufficient balance in the target currency
	// If insufficient, it will convert from source currency automatically
	// Returns true if successful, false if unable to ensure balance
	EnsureBalance(currency string, minAmount float64, sourceCurrency string) (bool, error)
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

// BrokerSymbolRepositoryInterface defines the contract for broker symbol mapping operations.
// This interface is defined in the domain layer to avoid circular dependencies between
// services and universe packages (clean architecture - dependency inversion).
type BrokerSymbolRepositoryInterface interface {
	// GetBrokerSymbol returns the broker-specific symbol for a given ISIN and broker name.
	// Returns error if mapping doesn't exist (fail-fast approach).
	GetBrokerSymbol(isin, brokerName string) (string, error)

	// SetBrokerSymbol creates or updates a broker symbol mapping.
	// Replaces existing mapping if present (upsert operation).
	SetBrokerSymbol(isin, brokerName, symbol string) error

	// GetAllBrokerSymbols returns all broker symbols for a given ISIN.
	// Returns a map of broker_name -> broker_symbol.
	GetAllBrokerSymbols(isin string) (map[string]string, error)

	// GetISINByBrokerSymbol performs reverse lookup: finds ISIN by broker symbol.
	// Returns error if mapping doesn't exist.
	GetISINByBrokerSymbol(brokerName, brokerSymbol string) (string, error)

	// DeleteBrokerSymbol removes a broker symbol mapping for an ISIN/broker pair.
	DeleteBrokerSymbol(isin, brokerName string) error
}
