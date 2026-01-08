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

// BrokerCashFlow represents a cash flow transaction (broker-agnostic)
// Supports deposits, dividends, fees, and other cash operations
type BrokerCashFlow struct {
	ID              string                 // Transaction ID
	TransactionID   string                 // Alternative transaction ID field
	TypeDocID       int                    // Document type ID
	Type            string                 // Type field (flexible)
	TransactionType string                 // Transaction type field (flexible)
	DT              string                 // Alternate date field
	Date            string                 // Transaction date
	SM              float64                // Alternate amount field
	Amount          float64                // Transaction amount
	Curr            string                 // Alternate currency field
	Currency        string                 // Currency code
	SMEUR           float64                // Alternate EUR amount field
	AmountEUR       float64                // Amount in EUR
	Status          string                 // Transaction status
	StatusC         int                    // Status code
	Description     string                 // Transaction description
	Params          map[string]interface{} // Flexible parameters
}

// BrokerHealthResult represents broker connection health status (broker-agnostic)
type BrokerHealthResult struct {
	Connected bool   // Whether broker is connected
	Timestamp string // Timestamp of health check
}
