package domain

// Broker-agnostic types for portfolio management
// These types abstract away broker-specific implementations (Tradernet, IBKR, etc.)

// BrokerPosition represents a portfolio position (broker-agnostic)
type BrokerPosition struct {
	Symbol         string  // Security symbol
	Quantity       float64 // Number of shares held
	AvgPrice       float64 // Average purchase price
	CurrentPrice   float64 // Current market price
	MarketValue    float64 // Position value in position currency
	MarketValueEUR float64 // Position value in EUR
	UnrealizedPnL  float64 // Unrealized profit/loss
	Currency       string  // Position currency
	CurrencyRate   float64 // Exchange rate to EUR
}

// BrokerCashBalance represents cash balance in a currency (broker-agnostic)
type BrokerCashBalance struct {
	Currency string  // Currency code (EUR, USD, etc.)
	Amount   float64 // Cash amount
}

// BrokerOrderResult represents the result of placing an order (broker-agnostic)
type BrokerOrderResult struct {
	OrderID  string  // Order confirmation ID
	Symbol   string  // Security symbol
	Side     string  // "BUY" or "SELL"
	Quantity float64 // Executed quantity
	Price    float64 // Execution price
}

// BrokerTrade represents an executed trade (broker-agnostic)
type BrokerTrade struct {
	OrderID    string  // Order ID
	Symbol     string  // Security symbol
	Side       string  // "BUY" or "SELL"
	Quantity   float64 // Traded quantity
	Price      float64 // Execution price
	ExecutedAt string  // Execution timestamp
}

// BrokerQuote represents a security quote (broker-agnostic)
type BrokerQuote struct {
	Symbol    string  // Security symbol
	Price     float64 // Current price
	Change    float64 // Absolute change
	ChangePct float64 // Percentage change
	Volume    int64   // Trading volume
	Timestamp string  // Quote timestamp
}

// BrokerPendingOrder represents a pending order (broker-agnostic)
type BrokerPendingOrder struct {
	OrderID  string  // Pending order ID
	Symbol   string  // Security symbol
	Side     string  // "BUY" or "SELL"
	Quantity float64 // Order quantity
	Price    float64 // Order price
	Currency string  // Currency
}

// BrokerSecurityInfo represents security lookup result (broker-agnostic)
type BrokerSecurityInfo struct {
	Symbol       string  // Security symbol
	Name         *string // Company name (nullable)
	ISIN         *string // ISIN code (nullable)
	Currency     *string // Trading currency (nullable)
	Market       *string // Market name (nullable)
	ExchangeCode *string // Exchange code (nullable)
}

// BrokerCashMovement represents cash withdrawal history (broker-agnostic)
type BrokerCashMovement struct {
	TotalWithdrawals float64                  // Total withdrawals amount
	Withdrawals      []map[string]interface{} // List of withdrawal records (flexible schema)
	Note             string                   // Additional notes
}

// BrokerCashFlow represents a single cash flow transaction in the account.
// Cash flows include deposits, withdrawals, dividends, interest, fees, taxes, etc.
// This is distinct from trades, which exchange cash for securities.
type BrokerCashFlow struct {
	ID            string                 // Unique transaction identifier
	TransactionID string                 // Alternative/external transaction reference ID
	Type          string                 // Transaction type: "deposit", "withdrawal", "dividend", "interest", "fee", "tax"
	Date          string                 // Transaction date in YYYY-MM-DD format
	Amount        float64                // Transaction amount in the original currency (positive for inflows, negative for outflows)
	Currency      string                 // Currency code (ISO 4217: EUR, USD, GBP, etc.)
	AmountEUR     float64                // Transaction amount converted to EUR for reporting consistency
	Status        string                 // Transaction status: "completed", "pending", "cancelled", etc.
	StatusC       int                    // Numeric status code (broker-specific, for internal use)
	Description   string                 // Human-readable transaction description
	Params        map[string]interface{} // Additional transaction parameters (flexible schema for broker-specific data)
}

// BrokerHealthResult represents broker connection health status (broker-agnostic)
type BrokerHealthResult struct {
	Connected bool   // Whether broker is connected
	Timestamp string // Timestamp of health check
}
