package cash_flows

import (
	"time"
)

// CashFlow represents a cash flow transaction (append-only ledger)
// Faithful translation from Python: app/modules/cash_flows/domain/models.py
type CashFlow struct {
	ID              int       `json:"id,omitempty"`
	TransactionID   string    `json:"transaction_id"`   // Unique from Tradernet
	TypeDocID       int       `json:"type_doc_id"`      // Transaction type code
	TransactionType *string   `json:"transaction_type"` // DEPOSIT, WITHDRAWAL, etc.
	Date            string    `json:"date"`             // YYYY-MM-DD
	Amount          float64   `json:"amount"`           // Original currency
	Currency        string    `json:"currency"`         // Currency code
	AmountEUR       float64   `json:"amount_eur"`       // Converted to EUR
	Status          *string   `json:"status"`           // Status string
	StatusC         *int      `json:"status_c"`         // Status code
	Description     *string   `json:"description"`      // Human description
	ParamsJSON      *string   `json:"-"`                // Raw params (not exposed in API)
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"` // Not used (append-only)
}

// APITransaction represents a transaction from the broker API (broker-agnostic).
// Used during sync_from_api to pass broker cash flows to the repository.
// This is an internal representation with clean field names.
type APITransaction struct {
	TransactionID   string                 `json:"transaction_id"`   // Transaction identifier
	TypeDocID       int                    `json:"type_doc_id"`      // Type document ID (broker-specific, may be 0)
	TransactionType string                 `json:"transaction_type"` // Transaction type (deposit, withdrawal, etc.)
	Date            string                 `json:"date"`             // Transaction date
	Amount          float64                `json:"amount"`           // Amount in original currency
	Currency        string                 `json:"currency"`         // Currency code
	AmountEUR       float64                `json:"amount_eur"`       // Amount in EUR
	Status          string                 `json:"status"`           // Status string
	StatusC         int                    `json:"status_c"`         // Status code
	Description     string                 `json:"description"`      // Human-readable description
	Params          map[string]interface{} `json:"params"`           // Additional parameters (broker-specific metadata)
}
