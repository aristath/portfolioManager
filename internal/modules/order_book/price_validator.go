package order_book

// PriceValidator provides price validation from secondary sources
// This abstraction allows the order book module to validate prices without
// depending on specific price providers.
//
// The interface follows Dependency Inversion Principle:
// - High-level module (order_book) defines the interface it needs
// - Low-level modules (price adapters) implement the interface
// - Dependencies point inward (infrastructure → business logic)
type PriceValidator interface {
	// GetValidationPrice fetches a validation price for the given symbol
	// This price is used to cross-validate the broker's quote data
	//
	// Parameters:
	//   - symbol: Broker symbol (e.g., "AAPL.US")
	//
	// Returns:
	//   - *float64: Validation price, or nil if unavailable
	//   - error: Error if fetch failed (service unavailable, invalid data, etc.)
	//
	// Implementation Notes:
	// - Implementations should handle symbol transformation (e.g., "AAPL.US" → "AAPL")
	// - Implementations should retry on transient failures
	// - Returning nil price with nil error means "price not available" (graceful degradation)
	GetValidationPrice(symbol string) (*float64, error)
}
