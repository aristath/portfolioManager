package dividends

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DividendRecord represents a dividend payment with DRIP tracking
// After Unix timestamp migration: PaymentDate uses Unix timestamp (int64) at midnight UTC
// Converted to YYYY-MM-DD string only at JSON boundary for API compatibility
type DividendRecord struct {
	ReinvestedAt       *time.Time `json:"reinvested_at,omitempty"`
	ReinvestedQuantity *int       `json:"reinvested_quantity,omitempty"`
	CreatedAt          *time.Time `json:"created_at,omitempty"`
	CashFlowID         *int       `json:"cash_flow_id,omitempty"`
	ClearedAt          *time.Time `json:"cleared_at,omitempty"`
	PaymentDate        *int64     `json:"-"` // Unix timestamp at midnight UTC (date only), converted to string in MarshalJSON
	Symbol             string     `json:"symbol"`
	Currency           string     `json:"currency"`
	ISIN               string     `json:"isin,omitempty"`
	AmountEUR          float64    `json:"amount_eur"`
	ID                 int        `json:"id"`
	PendingBonus       float64    `json:"pending_bonus"`
	Amount             float64    `json:"amount"`
	Reinvested         bool       `json:"reinvested"`
	BonusCleared       bool       `json:"bonus_cleared"`
}

// MarshalJSON customizes JSON serialization to convert Unix timestamp to YYYY-MM-DD string
func (d DividendRecord) MarshalJSON() ([]byte, error) {
	type Alias DividendRecord
	aux := &struct {
		PaymentDate string `json:"payment_date"`
		*Alias
	}{
		Alias: (*Alias)(&d),
	}

	// Convert Unix timestamp to YYYY-MM-DD string for API
	if d.PaymentDate != nil {
		t := time.Unix(*d.PaymentDate, 0).UTC()
		aux.PaymentDate = t.Format("2006-01-02")
	}

	return json.Marshal(aux)
}

// Validate validates dividend record data
func (d *DividendRecord) Validate() error {
	if d.Symbol == "" || strings.TrimSpace(d.Symbol) == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	if d.Amount <= 0 {
		return fmt.Errorf("dividend amount must be positive")
	}

	if d.AmountEUR <= 0 {
		return fmt.Errorf("amount_eur must be positive")
	}

	if d.Currency == "" {
		return fmt.Errorf("currency cannot be empty")
	}

	if d.PaymentDate == nil {
		return fmt.Errorf("payment_date cannot be empty")
	}

	// Normalize symbol
	d.Symbol = strings.ToUpper(strings.TrimSpace(d.Symbol))

	return nil
}
