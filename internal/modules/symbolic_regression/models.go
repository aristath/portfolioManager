package symbolic_regression

import "time"

// TrainingExample represents a single training data point
type TrainingExample struct {
	SecurityISIN   string
	SecuritySymbol string
	ProductType    string // EQUITY, ETF, MUTUALFUND
	Date           string // YYYY-MM-DD format
	TargetDate     string // YYYY-MM-DD format (date for target return)

	// Inputs (features) at Date
	Inputs TrainingInputs

	// Target (what we're trying to predict)
	TargetReturn float64 // Actual return from Date to TargetDate
}

// TrainingInputs contains all input features for a training example
type TrainingInputs struct {
	// Scoring components
	LongTermScore        float64
	FundamentalsScore    float64
	DividendsScore       float64
	OpportunityScore     float64
	ShortTermScore       float64
	TechnicalsScore      float64
	OpinionScore         float64
	DiversificationScore float64
	TotalScore           float64

	// Metrics
	CAGR          float64
	DividendYield float64
	Volatility    float64
	SharpeRatio   *float64
	SortinoRatio  *float64
	RSI           *float64
	MaxDrawdown   *float64

	// Regime
	RegimeScore float64 // Continuous -1.0 to +1.0

	// Additional metrics from calculated_metrics table
	AdditionalMetrics map[string]float64
}

// FormulaType represents the type of formula being discovered
type FormulaType string

const (
	FormulaTypeExpectedReturn FormulaType = "expected_return"
	FormulaTypeScoring        FormulaType = "scoring"
)

// SecurityType represents the security type for formula discovery
type SecurityType string

const (
	SecurityTypeStock SecurityType = "stock" // EQUITY
	SecurityTypeETF   SecurityType = "etf"   // ETF, MUTUALFUND
)

// DiscoveredFormula represents a discovered formula
type DiscoveredFormula struct {
	ID                int64
	FormulaType       FormulaType
	SecurityType      SecurityType
	RegimeRangeMin    *float64           // Optional: regime range minimum
	RegimeRangeMax    *float64           // Optional: regime range maximum
	FormulaExpression string             // Formula as string (e.g., "0.65*cagr + 0.28*score")
	ValidationMetrics map[string]float64 // MAE, RMSE, Spearman correlation, etc.
	DiscoveredAt      time.Time
}
