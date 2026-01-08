package tradernet

// Tradernet Order Type Codes
// Tradernet API uses numeric string codes for order types in some endpoints
const (
	TradernetOrderTypeBuy  = "1" // BUY order
	TradernetOrderTypeSell = "2" // SELL order
)

// Normalized order sides for domain (broker-agnostic)
// All order types should be normalized to these standard values
const (
	OrderSideBuy  = "BUY"  // Buy/long position
	OrderSideSell = "SELL" // Sell/short position
)
