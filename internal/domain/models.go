// Package domain provides core domain models and types.
package domain

import "time"

// Currency represents a currency code
type Currency string

const (
	CurrencyEUR  Currency = "EUR"
	CurrencyUSD  Currency = "USD"
	CurrencyGBP  Currency = "GBP"
	CurrencyTEST Currency = "TEST" // For research mode
)

// Security represents a tradable security
type Security struct {
	LastUpdated time.Time `json:"last_updated"`
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	Exchange    string    `json:"exchange"`
	Country     string    `json:"country"`
	Currency    Currency  `json:"currency"`
	ISIN        string    `json:"isin"`
	ID          int64     `json:"id"`
	Active      bool      `json:"active"`
	// Constraint fields for opportunity filtering
	AllowSell bool `json:"allow_sell"` // Whether selling is allowed for this security
	AllowBuy  bool `json:"allow_buy"`  // Whether buying is allowed for this security
	MinLot    int  `json:"min_lot"`    // Minimum lot size for trading
}

// Position represents a portfolio position
type Position struct {
	LastUpdated  time.Time `json:"last_updated"`
	Symbol       string    `json:"symbol"`
	ISIN         string    `json:"isin"` // Primary identifier (after migration)
	Currency     Currency  `json:"currency"`
	ID           int64     `json:"id"`
	SecurityID   int64     `json:"security_id"`
	Quantity     float64   `json:"quantity"`
	AverageCost  float64   `json:"average_cost"`
	CurrentPrice float64   `json:"current_price"`
	MarketValue  float64   `json:"market_value"`
	UnrealizedPL float64   `json:"unrealized_pl"`
}

// Trade represents an executed trade
type Trade struct {
	ExecutedAt time.Time `json:"executed_at"`
	CreatedAt  time.Time `json:"created_at"`
	Symbol     string    `json:"symbol"`
	Side       string    `json:"side"`
	Currency   Currency  `json:"currency"`
	ID         int64     `json:"id"`
	SecurityID int64     `json:"security_id"`
	Quantity   float64   `json:"quantity"`
	Price      float64   `json:"price"`
	Fees       float64   `json:"fees"`
	Total      float64   `json:"total"`
}

// Money represents a monetary value with currency
type Money struct {
	Currency Currency `json:"currency"`
	Amount   float64  `json:"amount"`
}

// NewMoney creates a new Money value
func NewMoney(amount float64, currency Currency) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}
