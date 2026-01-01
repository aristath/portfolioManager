package models

// TradeSide represents the direction of a trade (BUY or SELL)
type TradeSide string

const (
	TradeSideBuy  TradeSide = "BUY"
	TradeSideSell TradeSide = "SELL"
)

// IsBuy checks if this trade side is BUY
func (t TradeSide) IsBuy() bool {
	return t == TradeSideBuy
}

// IsSell checks if this trade side is SELL
func (t TradeSide) IsSell() bool {
	return t == TradeSideSell
}

// ActionCandidate represents a potential trade action with associated metadata
// for priority-based selection and sequencing.
type ActionCandidate struct {
	Side     TradeSide `json:"side"`      // Trade direction ("BUY" or "SELL")
	Symbol   string    `json:"symbol"`    // Security symbol
	Name     string    `json:"name"`      // Security name for display
	Quantity int       `json:"quantity"`  // Number of units to trade
	Price    float64   `json:"price"`     // Price per unit
	ValueEUR float64   `json:"value_eur"` // Total value in EUR
	Currency string    `json:"currency"`  // Trading currency
	Priority float64   `json:"priority"`  // Higher values indicate higher priority
	Reason   string    `json:"reason"`    // Human-readable explanation for this action
	Tags     []string  `json:"tags"`      // Classification tags (e.g., ["windfall", "underweight_asia"])
}

// Security represents a security in the investment universe
// (simplified version with only fields needed for evaluation)
type Security struct {
	Symbol   string  `json:"symbol"`   // Security symbol
	Name     string  `json:"name"`     // Security name
	Country  *string `json:"country"`  // Country (optional)
	Industry *string `json:"industry"` // Industry (optional)
	Currency string  `json:"currency"` // Trading currency
}

// Position represents a current position in a security
type Position struct {
	Symbol         string  `json:"symbol"`           // Security symbol
	Quantity       float64 `json:"quantity"`         // Number of shares
	AvgPrice       float64 `json:"avg_price"`        // Average purchase price
	Currency       string  `json:"currency"`         // Position currency
	CurrencyRate   float64 `json:"currency_rate"`    // Currency conversion rate to EUR
	CurrentPrice   float64 `json:"current_price"`    // Current market price
	MarketValueEUR float64 `json:"market_value_eur"` // Market value in EUR
}

// PortfolioContext contains portfolio state for allocation fit calculations
type PortfolioContext struct {
	// Core portfolio weights and positions
	CountryWeights  map[string]float64 `json:"country_weights"`  // group_name -> weight (-1 to +1)
	IndustryWeights map[string]float64 `json:"industry_weights"` // group_name -> weight (-1 to +1)
	Positions       map[string]float64 `json:"positions"`        // symbol -> position_value in EUR
	TotalValue      float64            `json:"total_value"`      // Total portfolio value in EUR

	// Additional data for portfolio scoring
	SecurityCountries  map[string]string  `json:"security_countries,omitempty"`  // symbol -> country (individual)
	SecurityIndustries map[string]string  `json:"security_industries,omitempty"` // symbol -> industry (individual)
	SecurityScores     map[string]float64 `json:"security_scores,omitempty"`     // symbol -> quality_score
	SecurityDividends  map[string]float64 `json:"security_dividends,omitempty"`  // symbol -> dividend_yield

	// Group mappings (for mapping individual countries/industries to groups)
	CountryToGroup  map[string]string `json:"country_to_group,omitempty"`  // country -> group_name
	IndustryToGroup map[string]string `json:"industry_to_group,omitempty"` // industry -> group_name

	// Cost basis data for averaging down
	PositionAvgPrices map[string]float64 `json:"position_avg_prices,omitempty"` // symbol -> avg_purchase_price
	CurrentPrices     map[string]float64 `json:"current_prices,omitempty"`      // symbol -> current_market_price
}

// EvaluationContext contains all data needed to simulate and score action sequences
type EvaluationContext struct {
	// Portfolio state
	PortfolioContext      PortfolioContext `json:"portfolio_context"`
	Positions             []Position       `json:"positions"`
	Securities            []Security       `json:"securities"`
	AvailableCashEUR      float64          `json:"available_cash_eur"`
	TotalPortfolioValueEUR float64         `json:"total_portfolio_value_eur"`

	// Market data
	CurrentPrices    map[string]float64 `json:"current_prices"`     // symbol -> current price
	StocksBySymbol   map[string]Security `json:"stocks_by_symbol"`  // symbol -> Security (computed)

	// Configuration
	TransactionCostFixed   float64 `json:"transaction_cost_fixed"`   // Fixed transaction cost (EUR)
	TransactionCostPercent float64 `json:"transaction_cost_percent"` // Percentage transaction cost (0.002 = 0.2%)

	// Optional: Price adjustment scenarios for stochastic evaluation
	PriceAdjustments map[string]float64 `json:"price_adjustments,omitempty"` // symbol -> multiplier (e.g., 1.05 for +5%)
}

// SequenceEvaluationResult represents the result of evaluating a single sequence
type SequenceEvaluationResult struct {
	Sequence         []ActionCandidate `json:"sequence"`          // The sequence that was evaluated
	Score            float64           `json:"score"`             // Total score
	EndCashEUR       float64           `json:"end_cash_eur"`      // Final cash after sequence
	EndPortfolio     PortfolioContext  `json:"end_portfolio"`     // Final portfolio state
	TransactionCosts float64           `json:"transaction_costs"` // Total transaction costs incurred
	Feasible         bool              `json:"feasible"`          // Whether the sequence was feasible
}

// BatchEvaluationRequest represents a request to evaluate multiple sequences
type BatchEvaluationRequest struct {
	Sequences        [][]ActionCandidate `json:"sequences"`         // List of sequences to evaluate
	EvaluationContext EvaluationContext  `json:"evaluation_context"` // Context for evaluation
}

// BatchEvaluationResponse represents the response from batch evaluation
type BatchEvaluationResponse struct {
	Results []SequenceEvaluationResult `json:"results"` // Results for each sequence
	Errors  []string                   `json:"errors"`  // Any errors encountered (per-sequence)
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string `json:"status"`  // "healthy" or "unhealthy"
	Version string `json:"version"` // Service version
}
