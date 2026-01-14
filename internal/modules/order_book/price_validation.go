// Package order_book provides order book analysis for optimal trade execution
package order_book

// PriceValidationService provides price validation from secondary sources
// This abstraction allows the order book module to validate prices without
// depending on specific price providers
type PriceValidationService interface {
	// GetValidationPrice fetches a validation price for the given symbol
	// This price is used to cross-validate the broker's quote data
	// Returns nil if price is unavailable (graceful degradation)
	GetValidationPrice(symbol string) (*float64, error)
}
