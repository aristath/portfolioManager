package scorers

import (
	"math"
	"time"

	"github.com/aristath/sentinel/internal/modules/quantum"
	"github.com/aristath/sentinel/internal/modules/scoring"
	"github.com/aristath/sentinel/internal/modules/scoring/domain"
	"github.com/aristath/sentinel/internal/modules/symbolic_regression"
	"github.com/aristath/sentinel/pkg/formulas"
)

// AdaptiveWeightsProvider interface for getting adaptive weights
type AdaptiveWeightsProvider interface {
	CalculateAdaptiveWeights(regimeScore float64) map[string]float64
}

// SecurityScorer orchestrates all scoring groups for a security
// Faithful translation from Python: app/modules/scoring/domain/security_scorer.py
type SecurityScorer struct {
	technicals          *TechnicalsScorer
	longTerm            *LongTermScorer
	opportunity         *OpportunityScorer
	dividend            *DividendScorer
	stability           *StabilityScorer // Replaces StabilityScorer (internal data only)
	shortTerm           *ShortTermScorer
	diversification     *DiversificationScorer
	adaptiveService     AdaptiveWeightsProvider               // Optional: adaptive market service
	regimeScoreProvider RegimeScoreProvider                   // Optional: regime score provider
	formulaStorage      *symbolic_regression.FormulaStorage   // Optional: discovered formula storage
	quantumCalculator   *quantum.QuantumProbabilityCalculator // Quantum probability calculator
}

// ScoreWeights defines the weight for each scoring group
// Quality-focused weights for 15-20 year retirement fund strategy
// Uses internal data only (no external stability or opinions)
var ScoreWeights = map[string]float64{
	"long_term":   0.30, // CAGR, Sortino, Sharpe - core value metric (↑ from 25%)
	"stability":   0.20, // CAGR consistency, volatility, recovery (replaces stability)
	"dividends":   0.18, // Yield, Consistency, Growth - unchanged
	"opportunity": 0.15, // 52W high distance (no P/E, ↑ from 12%)
	"short_term":  0.10, // Recent momentum, Drawdown (↑ from 8%)
	"technicals":  0.07, // RSI, Bollinger, EMA - unchanged
	// Total: 100%
	//
	// Changes from previous version:
	// - "stability" → "stability" (internal calculation from price history)
	// - "opinion" REMOVED (required external analyst data)
	// - "diversification" REMOVED (moved to planning layer)
	// - Redistributed weights: +5% long_term, +3% opportunity, +2% short_term
}

// NewSecurityScorer creates a new security scorer
func NewSecurityScorer() *SecurityScorer {
	return &SecurityScorer{
		technicals:        NewTechnicalsScorer(),
		longTerm:          NewLongTermScorer(),
		opportunity:       NewOpportunityScorer(),
		dividend:          NewDividendScorer(),
		stability:         NewStabilityScorer(),
		shortTerm:         NewShortTermScorer(),
		diversification:   NewDiversificationScorer(), // Kept for optional portfolio-aware scoring
		quantumCalculator: quantum.NewQuantumProbabilityCalculator(),
	}
}

// SetAdaptiveService sets the adaptive market service for dynamic weights
func (ss *SecurityScorer) SetAdaptiveService(service AdaptiveWeightsProvider) {
	ss.adaptiveService = service
}

// SetRegimeScoreProvider sets the regime score provider for getting current regime
func (ss *SecurityScorer) SetRegimeScoreProvider(provider RegimeScoreProvider) {
	ss.regimeScoreProvider = provider
}

// SetFormulaStorage sets the formula storage for discovered scoring formulas
func (ss *SecurityScorer) SetFormulaStorage(storage *symbolic_regression.FormulaStorage) {
	ss.formulaStorage = storage
}

// ScoreSecurityInput contains all data needed to score a security
// Uses only internal data (prices, dividends) - no external stability or analyst data
type ScoreSecurityInput struct {
	PortfolioContext    *domain.PortfolioContext
	Industry            *string
	Geography           *string // Comma-separated for multiple geographies
	MarketCode          string  // Tradernet market code for per-region regime scoring (e.g., "FIX", "EU", "HKEX")
	ProductType         string  // Product type: EQUITY, ETF, MUTUALFUND, ETC, CASH, UNKNOWN
	SortinoRatio        *float64
	MaxDrawdown         *float64
	DividendYield       *float64 // Internally calculated from ledger.db
	FiveYearAvgDivYield *float64 // Internally calculated from ledger.db
	PayoutRatio         *float64 // Optional - estimated from dividend/price if needed
	Symbol              string
	DailyPrices         []float64
	MonthlyPrices       []formulas.MonthlyPrice
	TargetAnnualReturn  float64
}

// ScoreSecurityWithDefaults scores a security with default values for missing data
func (ss *SecurityScorer) ScoreSecurityWithDefaults(input ScoreSecurityInput) *domain.CalculatedSecurityScore {
	// Set defaults
	if input.TargetAnnualReturn == 0 {
		input.TargetAnnualReturn = scoring.OptimalCAGR
	}

	return ss.ScoreSecurity(input)
}

// ScoreSecurity calculates complete security score with all groups
// Uses only internal data (prices, dividends) - no external stability or analyst data
func (ss *SecurityScorer) ScoreSecurity(input ScoreSecurityInput) *domain.CalculatedSecurityScore {
	groupScores := make(map[string]float64)
	subScores := make(map[string]map[string]float64)

	// 1. Long-term Performance (30%) - CAGR, Sortino, Sharpe
	longTermScore := ss.longTerm.Calculate(
		input.MonthlyPrices,
		input.DailyPrices,
		input.SortinoRatio,
		input.TargetAnnualReturn,
	)
	groupScores["long_term"] = longTermScore.Score
	subScores["long_term"] = longTermScore.Components

	// 2. Stability (20%) - CAGR consistency, volatility, recovery (replaces Stability)
	stabilityScore := ss.stability.Calculate(input.MonthlyPrices, input.DailyPrices)
	groupScores["stability"] = stabilityScore.Score
	subScores["stability"] = stabilityScore.Components

	// 3. Opportunity (15%) - 52W high distance, quality gates (P/E ratio removed)
	opportunityScore := ss.opportunity.CalculateWithQualityGate(
		input.DailyPrices,
		&stabilityScore.Score, // Use stability score for quality gate
		&longTermScore.Score,  // Pass long-term score for quality gate
		input.ProductType,     // Pass product type
	)
	groupScores["opportunity"] = opportunityScore.Score
	subScores["opportunity"] = opportunityScore.Components

	// 4. Dividends (18%) - Yield, Consistency, Growth (with total return boost)
	// Extract CAGR from long-term components for total return calculation
	var expectedCAGR *float64
	if cagrRaw, hasCAGR := longTermScore.Components["cagr_raw"]; hasCAGR && cagrRaw > 0 {
		expectedCAGR = &cagrRaw
	}
	dividendScore := ss.dividend.CalculateEnhanced(
		input.DividendYield,
		input.PayoutRatio,
		input.FiveYearAvgDivYield,
		expectedCAGR,
	)
	groupScores["dividends"] = dividendScore.Score
	subScores["dividends"] = dividendScore.Components

	// 5. Short-term Performance (10%) - Recent momentum, Drawdown
	shortTermScore := ss.shortTerm.Calculate(
		input.DailyPrices,
		input.MaxDrawdown,
	)
	groupScores["short_term"] = shortTermScore.Score
	subScores["short_term"] = shortTermScore.Components

	// 6. Technicals (7%) - RSI, Bollinger, EMA
	technicalsScore := ss.technicals.Calculate(input.DailyPrices)
	groupScores["technicals"] = technicalsScore.Score
	subScores["technicals"] = technicalsScore.Components

	// Note: Opinion scorer removed (required external analyst data)
	// Note: Diversification scoring moved to planning layer (not included in base score)

	// Try to use discovered formula first
	var totalScore float64
	useDiscoveredFormula := false

	if ss.formulaStorage != nil {
		// Determine security type
		securityType := symbolic_regression.SecurityTypeStock
		if input.ProductType == "ETF" || input.ProductType == "MUTUALFUND" {
			securityType = symbolic_regression.SecurityTypeETF
		}

		// Get regime score if available (use per-region score when MarketCode is available)
		var regimePtr *float64
		if ss.regimeScoreProvider != nil {
			var regimeScore float64
			var err error
			if input.MarketCode != "" {
				// Use per-region regime score based on security's market
				regimeScore, err = ss.regimeScoreProvider.GetRegimeScoreForMarketCode(input.MarketCode)
			} else {
				// Fallback to global score
				regimeScore, err = ss.regimeScoreProvider.GetCurrentRegimeScore()
			}
			if err == nil {
				regimePtr = &regimeScore
			}
		}

		// Try to get discovered formula
		discoveredFormula, err := ss.formulaStorage.GetActiveFormula(
			symbolic_regression.FormulaTypeScoring,
			securityType,
			regimePtr,
		)

		if err == nil && discoveredFormula != nil {
			// Parse and evaluate discovered formula
			parsedFormula, parseErr := symbolic_regression.ParseFormula(discoveredFormula.FormulaExpression)
			if parseErr == nil {
				// Build training inputs with group scores
				inputs := symbolic_regression.TrainingInputs{
					LongTermScore:        groupScores["long_term"],
					StabilityScore:       groupScores["stability"],
					DividendsScore:       groupScores["dividends"],
					OpportunityScore:     groupScores["opportunity"],
					ShortTermScore:       groupScores["short_term"],
					TechnicalsScore:      groupScores["technicals"],
					DiversificationScore: groupScores["diversification"],
				}

				// Add regime score if available
				if regimePtr != nil {
					inputs.RegimeScore = *regimePtr
				}

				// Evaluate formula
				formulaFn := symbolic_regression.FormulaToFunction(parsedFormula)
				totalScore = formulaFn(inputs)
				useDiscoveredFormula = true
			}
		}
	}

	// Fall back to static weighted sum if no discovered formula
	if !useDiscoveredFormula {
		// Get product-type-aware weights
		weights := ss.getScoreWeights(input.ProductType, input.MarketCode)

		// Normalize weights
		normalizedWeights := normalizeWeights(weights)

		// Calculate weighted total using static formula
		totalScore = 0.0
		for group, score := range groupScores {
			weight := normalizedWeights[group]
			totalScore += score * weight
		}
	}

	// Calculate volatility
	var volatility *float64
	var returns []float64
	if len(input.DailyPrices) >= 30 {
		returns = formulas.CalculateReturns(input.DailyPrices)
		vol := formulas.AnnualizedVolatility(returns)
		volatility = &vol
	}

	// Calculate quantum metrics if we have sufficient data
	if len(input.DailyPrices) >= 30 && volatility != nil {
		// Get Sharpe and Sortino ratios from long-term components
		var sharpe, sortino float64
		if sharpeRaw, hasSharpe := subScores["long_term"]["sharpe_raw"]; hasSharpe {
			sharpe = sharpeRaw
		}
		if sortinoRaw, hasSortino := subScores["long_term"]["sortino_raw"]; hasSortino {
			sortino = sortinoRaw
		}

		// Calculate quantum metrics
		quantumMetrics := ss.quantumCalculator.CalculateQuantumScore(
			returns,
			*volatility,
			sharpe,
			sortino,
			nil, // kurtosis not available
		)

		// Add quantum metrics to subScores
		if subScores["quantum"] == nil {
			subScores["quantum"] = make(map[string]float64)
		}
		subScores["quantum"]["risk_adjusted"] = round3(quantumMetrics.RiskAdjusted)
		subScores["quantum"]["interference"] = round3(quantumMetrics.Interference)
		subScores["quantum"]["multimodal"] = round3(quantumMetrics.Multimodal)
	}

	return &domain.CalculatedSecurityScore{
		Symbol:       input.Symbol,
		TotalScore:   round4(totalScore),
		Volatility:   volatility,
		CalculatedAt: time.Now(),
		GroupScores:  roundScores(groupScores),
		SubScores:    roundSubScores(subScores),
	}
}

// normalizeWeights ensures weights sum to 1.0
func normalizeWeights(weights map[string]float64) map[string]float64 {
	sum := 0.0
	for _, weight := range weights {
		sum += weight
	}

	if sum == 0 {
		return weights
	}

	normalized := make(map[string]float64, len(weights))
	for group, weight := range weights {
		normalized[group] = weight / sum
	}

	return normalized
}

// RegimeScoreProvider interface for getting current regime score
// Supports both global scores (backward compatible) and per-region scores.
type RegimeScoreProvider interface {
	// GetCurrentRegimeScore returns the global average regime score.
	// For backward compatibility with existing code.
	GetCurrentRegimeScore() (float64, error)

	// GetRegimeScoreForMarketCode returns the regime score for a security
	// based on its market code (e.g., "FIX" -> US, "EU" -> EU, "HKEX" -> ASIA).
	// For regions without dedicated indices, returns the global average.
	GetRegimeScoreForMarketCode(marketCode string) (float64, error)
}

// getBaseWeights returns the base score weights for a product type without any adaptive adjustments.
// This is an internal method used by getScoreWeights and GetScoreWeightsWithRegime.
func (ss *SecurityScorer) getBaseWeights(productType string) map[string]float64 {
	// Treat ETFs and Mutual Funds identically (both are diversified products)
	if productType == "ETF" || productType == "MUTUALFUND" {
		// Diversified product weights (ETFs & Mutual Funds)
		// Uses internal data only - no external stability or analyst data
		return map[string]float64{
			"long_term":   0.35, // ↑ from 30% (tracking quality matters most)
			"stability":   0.15, // ↓ from 20% (less relevant for diversified products)
			"dividends":   0.18, // Unchanged
			"opportunity": 0.12, // Unchanged
			"short_term":  0.10, // Unchanged
			"technicals":  0.10, // ↑ from 7% (compensate for removed opinion/diversification)
		}
	}

	// Default weights for stocks (EQUITY) and other types
	return ScoreWeights
}

// getScoreWeights returns score weights based on product type and market regime
// Implements product-type-aware scoring weights as per PRODUCT_TYPE_DIFFERENTIATION.md
// If adaptive service is available, uses adaptive weights based on regime score
// marketCode is used for per-region regime scoring (can be empty for fallback to global)
func (ss *SecurityScorer) getScoreWeights(productType, marketCode string) map[string]float64 {
	// Get base weights first
	baseWeights := ss.getBaseWeights(productType)

	// If no adaptive service, return base weights
	if ss.adaptiveService == nil {
		return baseWeights
	}

	// Get regime score if provider is available (use per-region when marketCode is available)
	regimeScore := 0.0
	if ss.regimeScoreProvider != nil {
		var currentScore float64
		var err error
		if marketCode != "" {
			currentScore, err = ss.regimeScoreProvider.GetRegimeScoreForMarketCode(marketCode)
		} else {
			currentScore, err = ss.regimeScoreProvider.GetCurrentRegimeScore()
		}
		if err == nil {
			regimeScore = currentScore
		}
	}

	// Apply adaptive weights
	return ss.applyAdaptiveWeights(baseWeights, regimeScore)
}

// GetScoreWeightsWithRegime returns score weights with explicit regime score
// This allows callers to provide regime score when available
func (ss *SecurityScorer) GetScoreWeightsWithRegime(productType string, regimeScore float64) map[string]float64 {
	// Get base weights first
	baseWeights := ss.getBaseWeights(productType)

	// Apply adaptive weights if service is available
	return ss.applyAdaptiveWeights(baseWeights, regimeScore)
}

// applyAdaptiveWeights applies adaptive weights to base weights based on regime score
func (ss *SecurityScorer) applyAdaptiveWeights(baseWeights map[string]float64, regimeScore float64) map[string]float64 {
	if ss.adaptiveService == nil {
		return baseWeights
	}

	adaptiveWeights := ss.adaptiveService.CalculateAdaptiveWeights(regimeScore)
	if len(adaptiveWeights) == 0 {
		return baseWeights
	}

	// Merge: use adaptive weights for common keys, keep base weights for others
	result := make(map[string]float64)
	for key, baseWeight := range baseWeights {
		if adaptiveWeight, ok := adaptiveWeights[key]; ok {
			result[key] = adaptiveWeight
		} else {
			result[key] = baseWeight
		}
	}
	// Add any adaptive weights not in base
	for key, adaptiveWeight := range adaptiveWeights {
		if _, ok := result[key]; !ok {
			result[key] = adaptiveWeight
		}
	}
	return result
}

// round4 rounds to 4 decimal places
func round4(f float64) float64 {
	return math.Round(f*10000) / 10000
}

// roundScores rounds all scores in map to 3 decimal places
func roundScores(scores map[string]float64) map[string]float64 {
	rounded := make(map[string]float64, len(scores))
	for k, v := range scores {
		rounded[k] = round3(v)
	}
	return rounded
}

// roundSubScores rounds all sub-scores to 3 decimal places
func roundSubScores(subScores map[string]map[string]float64) map[string]map[string]float64 {
	rounded := make(map[string]map[string]float64, len(subScores))
	for group, components := range subScores {
		rounded[group] = make(map[string]float64, len(components))
		for component, score := range components {
			rounded[group][component] = round3(score)
		}
	}
	return rounded
}
