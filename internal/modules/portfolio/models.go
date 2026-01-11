package portfolio

import (
	"encoding/json"
	"time"
)

// Position represents current position in a security
// After Unix timestamp migration: date/timestamp fields use Unix timestamps (int64)
// Converted to strings only at JSON boundary for API compatibility
type Position struct {
	LastUpdated      *int64  `json:"-"` // Unix timestamp (seconds since epoch), converted to string in MarshalJSON
	LastSoldAt       *int64  `json:"-"` // Unix timestamp at midnight UTC (date only), converted to string in MarshalJSON
	ISIN             string  `json:"isin,omitempty"`
	Currency         string  `json:"currency"`
	FirstBoughtAt    *int64  `json:"-"` // Unix timestamp at midnight UTC (date only), converted to string in MarshalJSON
	Symbol           string  `json:"symbol"`
	CurrentPrice     float64 `json:"current_price,omitempty"`
	CostBasisEUR     float64 `json:"cost_basis_eur,omitempty"`
	UnrealizedPnL    float64 `json:"unrealized_pnl,omitempty"`
	UnrealizedPnLPct float64 `json:"unrealized_pnl_pct,omitempty"`
	MarketValueEUR   float64 `json:"market_value_eur,omitempty"`
	CurrencyRate     float64 `json:"currency_rate"`
	AvgPrice         float64 `json:"avg_price"`
	Quantity         float64 `json:"quantity"`
}

// MarshalJSON customizes JSON serialization to convert Unix timestamps to strings
func (p Position) MarshalJSON() ([]byte, error) {
	type Alias Position
	aux := &struct {
		LastUpdated   string `json:"last_updated,omitempty"`
		FirstBoughtAt string `json:"first_bought_at,omitempty"`
		LastSoldAt    string `json:"last_sold_at,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(&p),
	}

	// Convert Unix timestamps to strings for API
	if p.LastUpdated != nil {
		t := time.Unix(*p.LastUpdated, 0).UTC()
		aux.LastUpdated = t.Format(time.RFC3339)
	}
	if p.FirstBoughtAt != nil {
		t := time.Unix(*p.FirstBoughtAt, 0).UTC()
		aux.FirstBoughtAt = t.Format("2006-01-02")
	}
	if p.LastSoldAt != nil {
		t := time.Unix(*p.LastSoldAt, 0).UTC()
		aux.LastSoldAt = t.Format("2006-01-02")
	}

	return json.Marshal(aux)
}

// AllocationStatus represents current allocation vs target
// Faithful translation from Python: app/domain/models.py
type AllocationStatus struct {
	Category     string  `json:"category"`      // "country" or "industry"
	Name         string  `json:"name"`          // Country name or Industry name
	TargetPct    float64 `json:"target_pct"`    // Target allocation percentage
	CurrentPct   float64 `json:"current_pct"`   // Current allocation percentage
	CurrentValue float64 `json:"current_value"` // Current value in EUR
	Deviation    float64 `json:"deviation"`     // current - target (negative = underweight)
}

// PortfolioSummary represents complete portfolio allocation summary
// Faithful translation from Python: app/domain/models.py
type PortfolioSummary struct {
	CountryAllocations  []AllocationStatus `json:"country_allocations"`
	IndustryAllocations []AllocationStatus `json:"industry_allocations"`
	TotalValue          float64            `json:"total_value"`
	CashBalance         float64            `json:"cash_balance"`
}

// PositionWithSecurity represents position with security information
// Used by get_with_security_info() - combines Position + Security data
type PositionWithSecurity struct {
	ISIN             string  `db:"isin"` // Primary identifier for internal operations
	Country          string  `db:"country"`
	StockName        string  `db:"name"`
	Symbol           string  `db:"symbol"` // For broker API and UI display
	Currency         string  `db:"currency"`
	FullExchangeName string  `db:"fullExchangeName"`
	Industry         string  `db:"industry"`
	LastUpdated      *int64  `db:"last_updated"` // Unix timestamp, converted to string in handler
	CurrentPrice     float64 `db:"current_price"`
	MarketValueEUR   float64 `db:"market_value_eur"`
	CurrencyRate     float64 `db:"currency_rate"`
	AvgPrice         float64 `db:"avg_price"`
	Quantity         float64 `db:"quantity"`
	AllowSell        bool    `db:"allow_sell"`
}
