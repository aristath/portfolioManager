// Package domain provides core domain interfaces.
//
// This package defines interfaces that break circular dependencies and provide
// clean contracts between packages. These interfaces follow the Dependency
// Inversion Principle - high-level modules depend on abstractions, not concretions.
//
// Key Interfaces:
// - CashManager: Cash balance management (breaks circular dependency)
// - BrokerClient: Broker-agnostic trading operations
// - CurrencyExchangeServiceInterface: Currency exchange operations
// - AllocationTargetProvider: Allocation target access
// - PortfolioSummaryProvider: Portfolio summary data
// - ConcentrationAlertProvider: Concentration alert detection
// - ScoresRepository: Security scores access (breaks circular dependency)
package domain

/**
 * CashManager defines operations for managing cash balances.
 *
 * This interface breaks circular dependencies between portfolio, cash_flows,
 * and services packages. It provides a clean contract for cash balance management
 * without creating import cycles.
 *
 * Merged from:
 *   - portfolio.CashManager (UpdateCashPosition, GetAllCashBalances)
 *   - services.CashManagerInterface (GetCashBalance)
 */
type CashManager interface {
	/**
	 * UpdateCashPosition updates or creates a cash balance for the given currency.
	 *
	 * @param currency - Currency code (EUR, USD, etc.)
	 * @param balance - New balance amount
	 * @returns error - Error if update fails
	 */
	UpdateCashPosition(currency string, balance float64) error

	/**
	 * GetAllCashBalances returns all cash balances as a map of currency -> balance.
	 *
	 * @returns map[string]float64 - Map of currency codes to balances
	 * @returns error - Error if retrieval fails
	 */
	GetAllCashBalances() (map[string]float64, error)

	/**
	 * GetCashBalance returns the cash balance for the given currency.
	 *
	 * Returns 0 if no balance exists (not an error).
	 *
	 * @param currency - Currency code (EUR, USD, etc.)
	 * @returns float64 - Cash balance (0 if not found)
	 * @returns error - Error if retrieval fails
	 */
	GetCashBalance(currency string) (float64, error)
}

/**
 * BrokerClient defines broker-agnostic trading and portfolio operations.
 *
 * This interface abstracts away broker-specific implementations (Tradernet, IBKR, etc.).
 * All broker operations should go through this interface for maximum flexibility.
 *
 * The interface provides:
 * - Portfolio operations (positions, cash balances)
 * - Trading operations (place orders, get trades, pending orders)
 * - Market data operations (quotes, historical prices, security lookup)
 * - Cash operations (cash flows, movements)
 * - Connection & health (connection status, health checks, credential updates)
 *
 * Implementations:
 * - Tradernet: internal/clients/tradernet/adapter.go
 * - Future brokers can implement this interface without changing business logic
 */
type BrokerClient interface {
	// Portfolio operations
	/**
	 * GetPortfolio retrieves all portfolio positions from the broker.
	 *
	 * @returns []BrokerPosition - List of portfolio positions
	 * @returns error - Error if retrieval fails
	 */
	GetPortfolio() ([]BrokerPosition, error)

	/**
	 * GetCashBalances retrieves all cash balances from the broker.
	 *
	 * @returns []BrokerCashBalance - List of cash balances by currency
	 * @returns error - Error if retrieval fails
	 */
	GetCashBalances() ([]BrokerCashBalance, error)

	// Trading operations
	/**
	 * PlaceOrder places a limit order with the broker.
	 *
	 * @param symbol - Security symbol
	 * @param side - "BUY" or "SELL"
	 * @param quantity - Order quantity
	 * @param limitPrice - Limit price for the order
	 * @returns *BrokerOrderResult - Order result with execution details
	 * @returns error - Error if order placement fails
	 */
	PlaceOrder(symbol, side string, quantity, limitPrice float64) (*BrokerOrderResult, error)

	/**
	 * GetExecutedTrades retrieves executed trades from the broker.
	 *
	 * @param limit - Maximum number of trades to retrieve
	 * @returns []BrokerTrade - List of executed trades
	 * @returns error - Error if retrieval fails
	 */
	GetExecutedTrades(limit int) ([]BrokerTrade, error)

	/**
	 * GetPendingOrders retrieves pending orders from the broker.
	 *
	 * @returns []BrokerPendingOrder - List of pending orders
	 * @returns error - Error if retrieval fails
	 */
	GetPendingOrders() ([]BrokerPendingOrder, error)

	// Market data operations
	/**
	 * GetQuote retrieves a single security quote from the broker.
	 *
	 * @param symbol - Security symbol
	 * @returns *BrokerQuote - Security quote
	 * @returns error - Error if retrieval fails
	 */
	GetQuote(symbol string) (*BrokerQuote, error)

	/**
	 * GetQuotes fetches quotes for multiple symbols in a single batch call.
	 *
	 * Returns a map of symbol -> quote for all successfully retrieved quotes.
	 * Symbols not found are simply omitted from the result (not an error).
	 *
	 * @param symbols - List of security symbols
	 * @returns map[string]*BrokerQuote - Map of symbol to quote
	 * @returns error - Error if batch retrieval fails
	 */
	GetQuotes(symbols []string) (map[string]*BrokerQuote, error)

	/**
	 * GetLevel1Quote fetches Level 1 market data (best bid and best ask only).
	 *
	 * Returns BrokerOrderBook with Bids[0] and Asks[0] populated.
	 *
	 * @param symbol - Security symbol
	 * @returns *BrokerOrderBook - Order book with best bid/ask
	 * @returns error - Error if retrieval fails
	 */
	GetLevel1Quote(symbol string) (*BrokerOrderBook, error)

	/**
	 * GetHistoricalPrices fetches OHLCV candlestick data for a symbol.
	 *
	 * @param symbol - Security symbol
	 * @param start - Start timestamp (Unix seconds)
	 * @param end - End timestamp (Unix seconds)
	 * @param timeframeSeconds - Timeframe in seconds (86400 for daily, 3600 for hourly, etc.)
	 * @returns []BrokerOHLCV - List of OHLCV candlesticks
	 * @returns error - Error if retrieval fails
	 */
	GetHistoricalPrices(symbol string, start, end int64, timeframeSeconds int) ([]BrokerOHLCV, error)

	/**
	 * FindSymbol searches for securities by symbol.
	 *
	 * @param symbol - Security symbol to search for
	 * @param exchange - Optional exchange code to filter results
	 * @returns []BrokerSecurityInfo - List of matching securities
	 * @returns error - Error if search fails
	 */
	FindSymbol(symbol string, exchange *string) ([]BrokerSecurityInfo, error)

	/**
	 * GetSecurityMetadata gets full security metadata including country and sector.
	 *
	 * This uses getAllSecurities API which returns issuer_country_code and sector_code,
	 * unlike FindSymbol (tickerFinder) which doesn't return these fields.
	 *
	 * @param symbol - Security symbol
	 * @returns *BrokerSecurityInfo - Security metadata
	 * @returns error - Error if retrieval fails
	 */
	GetSecurityMetadata(symbol string) (*BrokerSecurityInfo, error)

	/**
	 * GetSecurityMetadataRaw gets raw security metadata from broker API without transformation.
	 *
	 * Returns the raw response from getAllSecurities API as-is, preserving all original field names
	 * and structure. This is used by the metadata sync system to store complete raw data.
	 *
	 * @param symbol - Security symbol
	 * @returns interface{} - Raw API response (map[string]interface{} with "securities" array)
	 * @returns error - Error if retrieval fails
	 */
	GetSecurityMetadataRaw(symbol string) (interface{}, error)

	/**
	 * GetSecurityMetadataBatch gets raw security metadata for multiple symbols in a single batch API call.
	 *
	 * This method avoids rate limit errors (429) by fetching metadata for all symbols at once
	 * instead of making individual requests. Used for efficient bulk metadata sync.
	 *
	 * @param symbols - Array of security symbols to fetch
	 * @returns interface{} - Raw API response with structure: {"securities": [...], "total": N}
	 * @returns error - Error if batch retrieval fails
	 */
	GetSecurityMetadataBatch(symbols []string) (interface{}, error)

	/**
	 * GetFXRates retrieves currency exchange rates for today's date.
	 *
	 * Returns a map of currency codes to exchange rates relative to baseCurrency.
	 *
	 * @param baseCurrency - Base currency code (e.g., "EUR")
	 * @param currencies - List of currency codes to get rates for
	 * @returns map[string]float64 - Map of currency code to exchange rate
	 * @returns error - Error if retrieval fails
	 */
	GetFXRates(baseCurrency string, currencies []string) (map[string]float64, error)

	// Cash operations
	/**
	 * GetAllCashFlows retrieves cash flow transactions from the broker.
	 *
	 * Cash flows include deposits, withdrawals, dividends, interest, fees, taxes, etc.
	 *
	 * @param limit - Maximum number of cash flows to retrieve
	 * @returns []BrokerCashFlow - List of cash flow transactions
	 * @returns error - Error if retrieval fails
	 */
	GetAllCashFlows(limit int) ([]BrokerCashFlow, error)

	/**
	 * GetCashMovements retrieves cash withdrawal history from the broker.
	 *
	 * @returns *BrokerCashMovement - Cash movement summary
	 * @returns error - Error if retrieval fails
	 */
	GetCashMovements() (*BrokerCashMovement, error)

	// Connection & health
	/**
	 * IsConnected checks if the broker client is connected.
	 *
	 * @returns bool - True if connected, false otherwise
	 */
	IsConnected() bool

	/**
	 * HealthCheck performs a health check on the broker connection.
	 *
	 * @returns *BrokerHealthResult - Health check result
	 * @returns error - Error if health check fails
	 */
	HealthCheck() (*BrokerHealthResult, error)

	/**
	 * SetCredentials updates the broker API credentials.
	 *
	 * This allows credentials to be updated at runtime (e.g., from Settings UI).
	 *
	 * @param apiKey - API key
	 * @param apiSecret - API secret
	 */
	SetCredentials(apiKey, apiSecret string)
}

/**
 * CurrencyExchangeServiceInterface defines the contract for currency exchange operations.
 *
 * Used by portfolio, services, optimization, and universe packages.
 * This interface avoids import cycles with the services package.
 */
type CurrencyExchangeServiceInterface interface {
	/**
	 * GetRate returns the exchange rate from one currency to another.
	 *
	 * @param fromCurrency - Source currency code
	 * @param toCurrency - Target currency code
	 * @returns float64 - Exchange rate (how many units of toCurrency per unit of fromCurrency)
	 * @returns error - Error if rate retrieval fails
	 */
	GetRate(fromCurrency, toCurrency string) (float64, error)

	/**
	 * EnsureBalance ensures there is sufficient balance in the target currency.
	 *
	 * If insufficient, it will convert from source currency automatically.
	 *
	 * @param currency - Target currency code
	 * @param minAmount - Minimum required balance
	 * @param sourceCurrency - Source currency to convert from if needed
	 * @returns bool - True if successful, false if unable to ensure balance
	 * @returns error - Error if operation fails
	 */
	EnsureBalance(currency string, minAmount float64, sourceCurrency string) (bool, error)
}

/**
 * AllocationTargetProvider provides access to allocation targets.
 *
 * This interface breaks the circular dependency between portfolio and allocation packages.
 */
type AllocationTargetProvider interface {
	/**
	 * GetAll returns all allocation targets as a map of name -> target percentage.
	 *
	 * @returns map[string]float64 - Map of allocation name to target percentage (0-1)
	 * @returns error - Error if retrieval fails
	 */
	GetAll() (map[string]float64, error)
}

/**
 * PortfolioSummaryProvider provides portfolio summary data without creating
 * a dependency on the portfolio package.
 *
 * This interface breaks the circular dependency:
 * allocation → portfolio → cash_flows → trading → allocation
 */
type PortfolioSummaryProvider interface {
	/**
	 * GetPortfolioSummary returns the current portfolio summary.
	 *
	 * @returns PortfolioSummary - Portfolio allocation summary
	 * @returns error - Error if retrieval fails
	 */
	GetPortfolioSummary() (PortfolioSummary, error)
}

/**
 * ConcentrationAlertProvider provides concentration alert detection without
 * requiring direct dependency on ConcentrationAlertService.
 *
 * This interface breaks the circular dependency: trading → allocation
 */
type ConcentrationAlertProvider interface {
	/**
	 * DetectAlerts detects concentration alerts from a portfolio summary.
	 *
	 * @param summary - Portfolio summary to analyze
	 * @returns []ConcentrationAlert - List of concentration alerts
	 * @returns error - Error if detection fails
	 */
	DetectAlerts(summary PortfolioSummary) ([]ConcentrationAlert, error)
}

/**
 * PositionChecker provides minimal position checking without requiring
 * a dependency on the portfolio package.
 *
 * This interface breaks the circular dependency:
 * allocation → universe → portfolio → universe
 *
 * Used by SecurityDeletionService to check if a security has open positions.
 */
type PositionChecker interface {
	/**
	 * GetPositionQuantity returns the quantity of a position by ISIN.
	 *
	 * Returns 0 if the position doesn't exist (not an error).
	 *
	 * @param isin - Security ISIN
	 * @returns float64 - Position quantity (0 if not found)
	 * @returns error - Error if lookup fails
	 */
	GetPositionQuantity(isin string) (float64, error)
}

/**
 * ScoresRepository defines operations for accessing security scores.
 *
 * This interface breaks circular dependencies between scheduler, services,
 * and other packages that need access to security scoring data.
 *
 * All scores in the database are stored in normalized 0-1 format.
 *
 * Merged from:
 *   - scheduler.ScoresRepositoryInterface (GetCAGRs, GetQualityScores, GetValueTrapData, GetTotalScores, GetRiskMetrics)
 *   - services.ScoresRepository (same methods)
 */
type ScoresRepository interface {
	/**
	 * GetCAGRs retrieves Compound Annual Growth Rate values for securities.
	 *
	 * Converts normalized cagr_score (0-1) to CAGR percentage values using
	 * piecewise linear conversion.
	 *
	 * @param isinList - List of security ISINs to fetch
	 * @returns map[string]float64 - Map of ISIN to CAGR value (e.g., 0.11 for 11%)
	 * @returns error - Error if retrieval fails
	 */
	GetCAGRs(isinList []string) (map[string]float64, error)

	/**
	 * GetQualityScores retrieves long-term and stability scores for securities.
	 *
	 * @param isinList - List of security ISINs to fetch
	 * @returns map[string]float64 - Long-term scores (normalized 0-1)
	 * @returns map[string]float64 - Stability scores (normalized 0-1)
	 * @returns error - Error if retrieval fails
	 */
	GetQualityScores(isinList []string) (map[string]float64, map[string]float64, error)

	/**
	 * GetValueTrapData retrieves opportunity scores, momentum scores, and volatility.
	 *
	 * @param isinList - List of security ISINs to fetch
	 * @returns map[string]float64 - Opportunity scores (normalized 0-1)
	 * @returns map[string]float64 - Momentum scores (normalized -1 to 1)
	 * @returns map[string]float64 - Volatility values
	 * @returns error - Error if retrieval fails
	 */
	GetValueTrapData(isinList []string) (map[string]float64, map[string]float64, map[string]float64, error)

	/**
	 * GetTotalScores retrieves total composite scores for securities.
	 *
	 * @param isinList - List of security ISINs to fetch
	 * @returns map[string]float64 - Map of ISIN to total score
	 * @returns error - Error if retrieval fails
	 */
	GetTotalScores(isinList []string) (map[string]float64, error)

	/**
	 * GetRiskMetrics retrieves Sharpe ratio and max drawdown for securities.
	 *
	 * @param isinList - List of security ISINs to fetch
	 * @returns map[string]float64 - Sharpe ratios (raw values from database)
	 * @returns map[string]float64 - Max drawdown values (negative percentages, e.g., -0.25 for 25%)
	 * @returns error - Error if retrieval fails
	 */
	GetRiskMetrics(isinList []string) (map[string]float64, map[string]float64, error)
}

/**
 * PortfolioSummary represents complete portfolio allocation summary.
 *
 * This is the domain model used by the allocation package.
 * Note: This is different from portfolio.PortfolioSummary (which uses AllocationStatus).
 * This version uses PortfolioAllocation for compatibility with allocation package.
 */
type PortfolioSummary struct {
	GeographyAllocations []PortfolioAllocation // Geography allocation breakdown
	IndustryAllocations  []PortfolioAllocation // Industry allocation breakdown
	TotalValue           float64               // Total portfolio value in EUR
	CashBalance          float64               // Total cash balance in EUR
}

/**
 * PortfolioAllocation represents allocation info for display.
 *
 * Used to show current vs target allocation for geographies and industries.
 */
type PortfolioAllocation struct {
	Name         string  // Allocation name (e.g., "US", "Technology")
	TargetPct    float64 // Target percentage (0-1)
	CurrentPct   float64 // Current percentage (0-1)
	CurrentValue float64 // Current value in EUR
	Deviation    float64 // Deviation from target (current - target)
}

/**
 * ConcentrationAlert represents alert for approaching concentration limit.
 *
 * Alerts are generated when portfolio concentration exceeds thresholds.
 */
type ConcentrationAlert struct {
	Type              string  // Alert type (e.g., "security", "sector", "geography")
	Name              string  // Alert name (security symbol, sector name, etc.)
	Severity          string  // Alert severity ("critical", "warning")
	CurrentPct        float64 // Current percentage (0-1)
	LimitPct          float64 // Concentration limit (0-1)
	AlertThresholdPct float64 // Alert threshold (0-1)
}
