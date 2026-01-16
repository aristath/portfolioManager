/**
 * Package domain provides core domain models and types.
 *
 * This package defines the core business domain models and value types used throughout
 * the application. These types are pure domain models with no infrastructure dependencies,
 * following clean architecture principles.
 *
 * Domain Models:
 * - Currency: Currency codes (EUR, USD, GBP, TEST)
 * - ProductType: Financial product types (EQUITY, ETF, ETC, etc.)
 * - Position: Portfolio position (holding in a security)
 * - Trade: Executed trade transaction
 * - Money: Monetary value with currency
 *
 * Note: Security type was removed after migration 038. Use universe.Security directly
 * as the single source of truth for security data.
 */
package domain

import "time"

/**
 * Currency represents a currency code.
 *
 * Uses ISO 4217 currency codes. TEST currency is used for research mode.
 */
type Currency string

const (
	CurrencyEUR  Currency = "EUR"  // Euro
	CurrencyUSD  Currency = "USD"  // US Dollar
	CurrencyGBP  Currency = "GBP"  // British Pound
	CurrencyTEST Currency = "TEST" // For research mode (virtual currency)
)

/**
 * ProductType represents the type of financial product/instrument.
 *
 * Used to categorize securities by their product type.
 */
type ProductType string

const (
	// ProductTypeEquity represents individual stocks/shares
	ProductTypeEquity ProductType = "EQUITY"
	// ProductTypeETF represents Exchange Traded Funds
	ProductTypeETF ProductType = "ETF"
	// ProductTypeETC represents Exchange Traded Commodities
	ProductTypeETC ProductType = "ETC"
	// ProductTypeMutualFund represents mutual funds (some UCITS products)
	ProductTypeMutualFund ProductType = "MUTUALFUND"
	// ProductTypeIndex represents market indices (non-tradeable)
	ProductTypeIndex ProductType = "INDEX"
	// ProductTypeUnknown represents unknown type
	ProductTypeUnknown ProductType = "UNKNOWN"
)

// Security type removed - use universe.Security directly (single source of truth).
// After migration 038: All security data is in universe.Security with JSON storage.

/**
 * Position represents a portfolio position (holding in a security).
 *
 * A position represents the current state of a holding in the portfolio,
 * including quantity, average cost, current price, and unrealized P&L.
 */
type Position struct {
	LastUpdated  time.Time `json:"last_updated"`  // Last update timestamp
	Symbol       string    `json:"symbol"`        // Security symbol
	ISIN         string    `json:"isin"`          // Primary identifier (after migration)
	Currency     Currency  `json:"currency"`      // Position currency
	ID           int64     `json:"id"`            // Position ID
	SecurityID   int64     `json:"security_id"`   // Security ID (foreign key)
	Quantity     float64   `json:"quantity"`      // Number of shares held
	AverageCost  float64   `json:"average_cost"`  // Average purchase price
	CurrentPrice float64   `json:"current_price"` // Current market price
	MarketValue  float64   `json:"market_value"`  // Position value (quantity * current_price)
	UnrealizedPL float64   `json:"unrealized_pl"` // Unrealized profit/loss
}

/**
 * Trade represents an executed trade transaction.
 *
 * A trade is an immutable record of a completed buy or sell transaction.
 * All trades are stored in ledger.db as part of the financial audit trail.
 */
type Trade struct {
	ExecutedAt time.Time `json:"executed_at"` // Trade execution timestamp
	CreatedAt  time.Time `json:"created_at"`  // Trade record creation timestamp
	Symbol     string    `json:"symbol"`      // Security symbol
	Side       string    `json:"side"`        // "BUY" or "SELL"
	Currency   Currency  `json:"currency"`    // Trade currency
	ID         int64     `json:"id"`          // Trade ID
	SecurityID int64     `json:"security_id"` // Security ID (foreign key)
	Quantity   float64   `json:"quantity"`    // Traded quantity
	Price      float64   `json:"price"`       // Execution price
	Fees       float64   `json:"fees"`        // Transaction fees
	Total      float64   `json:"total"`       // Total trade value (quantity * price + fees)
}

/**
 * Money represents a monetary value with currency.
 *
 * This is a value type that combines an amount with its currency,
 * ensuring type safety in monetary calculations.
 */
type Money struct {
	Currency Currency `json:"currency"` // Currency code
	Amount   float64  `json:"amount"`   // Monetary amount
}

/**
 * NewMoney creates a new Money value.
 *
 * @param amount - Monetary amount
 * @param currency - Currency code
 * @returns Money - New Money value
 */
func NewMoney(amount float64, currency Currency) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}
