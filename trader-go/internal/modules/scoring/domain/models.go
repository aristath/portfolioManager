package domain

import "time"

// CalculatedSecurityScore represents a complete security score with all components
// Faithful translation from Python: app/modules/scoring/domain/models.py
type CalculatedSecurityScore struct {
	Symbol       string                        `json:"symbol"`
	TotalScore   float64                       `json:"total_score"` // Final weighted score
	Volatility   *float64                      `json:"volatility"`  // Raw annualized volatility
	CalculatedAt time.Time                     `json:"calculated_at"`
	GroupScores  map[string]float64            `json:"group_scores"` // long_term, fundamentals, opportunity, etc.
	SubScores    map[string]map[string]float64 `json:"sub_scores"`   // Group -> Component -> Score
}

// PrefetchedSecurityData represents pre-fetched data to avoid duplicate API calls
type PrefetchedSecurityData struct {
	DailyPrices   []DailyPrice   `json:"daily_prices"`   // Daily OHLCV data
	MonthlyPrices []MonthlyPrice `json:"monthly_prices"` // Monthly averages
	Fundamentals  interface{}    `json:"fundamentals"`   // Yahoo fundamentals (can be nil)
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
	Symbol                string  `json:"symbol"`
	Eligible              bool    `json:"eligible"`     // Whether sell is allowed at all
	BlockReason           *string `json:"block_reason"` // If not eligible, why
	UnderperformanceScore float64 `json:"underperformance_score"`
	TimeHeldScore         float64 `json:"time_held_score"`
	PortfolioBalanceScore float64 `json:"portfolio_balance_score"`
	InstabilityScore      float64 `json:"instability_score"`
	TotalScore            float64 `json:"total_score"`
	SuggestedSellPct      float64 `json:"suggested_sell_pct"` // 0.10 to 0.50
	SuggestedSellQuantity int     `json:"suggested_sell_quantity"`
	SuggestedSellValue    float64 `json:"suggested_sell_value"`
	ProfitPct             float64 `json:"profit_pct"`
	DaysHeld              int     `json:"days_held"`
}

// PortfolioContext represents portfolio context for allocation fit calculations
type PortfolioContext struct {
	CountryWeights  map[string]float64 `json:"country_weights"`  // group_name -> weight (-1 to +1)
	IndustryWeights map[string]float64 `json:"industry_weights"` // group_name -> weight (-1 to +1)
	Positions       map[string]float64 `json:"positions"`        // symbol -> position_value
	TotalValue      float64            `json:"total_value"`

	// Additional data for portfolio scoring
	SecurityCountries  map[string]string  `json:"security_countries"`  // symbol -> country
	SecurityIndustries map[string]string  `json:"security_industries"` // symbol -> industry
	SecurityScores     map[string]float64 `json:"security_scores"`     // symbol -> quality_score
	SecurityDividends  map[string]float64 `json:"security_dividends"`  // symbol -> dividend_yield

	// Group mappings
	CountryToGroup  map[string]string `json:"country_to_group"`  // country -> group_name
	IndustryToGroup map[string]string `json:"industry_to_group"` // industry -> group_name

	// Cost basis data for averaging down
	PositionAvgPrices map[string]float64 `json:"position_avg_prices"` // symbol -> avg_purchase_price
	CurrentPrices     map[string]float64 `json:"current_prices"`      // symbol -> current_market_price
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
	Name       string             `json:"name"`       // long_term, fundamentals, etc.
	Score      float64            `json:"score"`      // Overall group score (0-1)
	Weight     float64            `json:"weight"`     // Weight in total score
	Components map[string]float64 `json:"components"` // Individual component scores
}

// ScoreRequest represents a request to calculate scores for a security
type ScoreRequest struct {
	Symbol         string                  `json:"symbol"`
	ISIN           string                  `json:"isin"`
	FetchPrices    bool                    `json:"fetch_prices"`    // Whether to fetch price data
	PrefetchedData *PrefetchedSecurityData `json:"prefetched_data"` // Optional pre-fetched data
}

// ScoreResponse represents the response from score calculation
type ScoreResponse struct {
	Symbol string                   `json:"symbol"`
	Score  *CalculatedSecurityScore `json:"score"`
	Error  *string                  `json:"error,omitempty"`
}

// BulkScoreRequest represents a request to score multiple securities
type BulkScoreRequest struct {
	Symbols []string `json:"symbols"`
}

// BulkScoreResponse represents the response from bulk scoring
type BulkScoreResponse struct {
	Scores []ScoreResponse `json:"scores"`
}
