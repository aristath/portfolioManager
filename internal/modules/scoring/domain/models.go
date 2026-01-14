// Package domain provides scoring domain models.
package domain

import "time"

// CalculatedSecurityScore represents a complete security score with all components
// Faithful translation from Python: app/modules/scoring/domain/models.py
type CalculatedSecurityScore struct {
	CalculatedAt time.Time                     `json:"calculated_at"`
	Volatility   *float64                      `json:"volatility"`
	GroupScores  map[string]float64            `json:"group_scores"`
	SubScores    map[string]map[string]float64 `json:"sub_scores"`
	Symbol       string                        `json:"symbol"`
	TotalScore   float64                       `json:"total_score"`
}

// PrefetchedSecurityData represents pre-fetched data to avoid duplicate API calls
type PrefetchedSecurityData struct {
	Stability     interface{}    `json:"stability"` // Price-derived stability metrics (replaces external fundamentals)
	DailyPrices   []DailyPrice   `json:"daily_prices"`
	MonthlyPrices []MonthlyPrice `json:"monthly_prices"`
}

// DailyPrice represents a daily price data point
type DailyPrice struct {
	Date   string  `json:"date"`
	Close  float64 `json:"close"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Open   float64 `json:"open"`
	Volume int64   `json:"volume"`
}

// MonthlyPrice represents a monthly price data point
type MonthlyPrice struct {
	Month       string  `json:"month"` // YYYY-MM format
	AvgAdjClose float64 `json:"avg_adj_close"`
}

// TechnicalData represents technical indicators for instability detection
type TechnicalData struct {
	CurrentVolatility    float64 `json:"current_volatility"`    // Last 60 days
	HistoricalVolatility float64 `json:"historical_volatility"` // Last 365 days
	DistanceFromMA200    float64 `json:"distance_from_ma_200"`  // Positive = above MA, negative = below
}

// SellScore represents the result of sell score calculation
type SellScore struct {
	BlockReason           *string `json:"block_reason"`
	Symbol                string  `json:"symbol"`
	InstabilityScore      float64 `json:"instability_score"`
	UnderperformanceScore float64 `json:"underperformance_score"`
	TimeHeldScore         float64 `json:"time_held_score"`
	PortfolioBalanceScore float64 `json:"portfolio_balance_score"`
	TotalScore            float64 `json:"total_score"`
	SuggestedSellPct      float64 `json:"suggested_sell_pct"`
	SuggestedSellQuantity int     `json:"suggested_sell_quantity"`
	SuggestedSellValue    float64 `json:"suggested_sell_value"`
	ProfitPct             float64 `json:"profit_pct"`
	DaysHeld              int     `json:"days_held"`
	Eligible              bool    `json:"eligible"`
}

// PortfolioContext represents portfolio context for allocation fit calculations
type PortfolioContext struct {
	// Core allocation data
	GeographyWeights    map[string]float64 `json:"geography_weights"`
	IndustryWeights     map[string]float64 `json:"industry_weights"`
	Positions           map[string]float64 `json:"positions"`
	SecurityGeographies map[string]string  `json:"security_geographies"` // ISIN -> geography (comma-separated for multiple)
	SecurityIndustries  map[string]string  `json:"security_industries"`  // ISIN -> industry (comma-separated for multiple)
	SecurityScores      map[string]float64 `json:"security_scores"`
	SecurityDividends   map[string]float64 `json:"security_dividends"`
	PositionAvgPrices   map[string]float64 `json:"position_avg_prices"`
	CurrentPrices       map[string]float64 `json:"current_prices"`
	TotalValue          float64            `json:"total_value"`

	// Extended metrics for comprehensive evaluation
	SecurityCAGRs       map[string]float64 `json:"security_cagrs,omitempty"`        // symbol -> historical CAGR
	SecurityVolatility  map[string]float64 `json:"security_volatility,omitempty"`   // symbol -> annual volatility
	SecuritySharpe      map[string]float64 `json:"security_sharpe,omitempty"`       // symbol -> Sharpe ratio
	SecuritySortino     map[string]float64 `json:"security_sortino,omitempty"`      // symbol -> Sortino ratio
	SecurityMaxDrawdown map[string]float64 `json:"security_max_drawdown,omitempty"` // symbol -> max drawdown (negative)

	// Market regime and optimizer targets
	MarketRegimeScore      float64            `json:"market_regime_score,omitempty"`      // -1 (bear) to +1 (bull)
	OptimizerTargetWeights map[string]float64 `json:"optimizer_target_weights,omitempty"` // symbol -> target weight
}

// PortfolioScore represents overall portfolio health score
type PortfolioScore struct {
	DiversificationScore float64 `json:"diversification_score"` // Country + industry balance (0-100)
	DividendScore        float64 `json:"dividend_score"`        // Weighted average dividend yield score (0-100)
	QualityScore         float64 `json:"quality_score"`         // Weighted average security quality (0-100)
	Total                float64 `json:"total"`                 // Combined score (0-100)
}

// ScoreGroup represents a scoring group result
type ScoreGroup struct {
	Components map[string]float64 `json:"components"`
	Name       string             `json:"name"`
	Score      float64            `json:"score"`
	Weight     float64            `json:"weight"`
}

// ScoreRequest represents a request to calculate scores for a security
type ScoreRequest struct {
	PrefetchedData *PrefetchedSecurityData `json:"prefetched_data"`
	Symbol         string                  `json:"symbol"`
	ISIN           string                  `json:"isin"`
	FetchPrices    bool                    `json:"fetch_prices"`
}

// ScoreResponse represents the response from score calculation
type ScoreResponse struct {
	Score  *CalculatedSecurityScore `json:"score"`
	Error  *string                  `json:"error,omitempty"`
	Symbol string                   `json:"symbol"`
}

// BulkScoreRequest represents a request to score multiple securities
type BulkScoreRequest struct {
	Symbols []string `json:"symbols"`
}

// BulkScoreResponse represents the response from bulk scoring
type BulkScoreResponse struct {
	Scores []ScoreResponse `json:"scores"`
}
