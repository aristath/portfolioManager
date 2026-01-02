package scorers

import (
	"math"
	"strings"

	"github.com/aristath/arduino-trader/internal/modules/scoring"
	"github.com/aristath/arduino-trader/internal/modules/scoring/domain"
)

// DiversificationScorer calculates portfolio fit and balance score
// Faithful translation from Python: app/modules/scoring/domain/diversification.py
type DiversificationScorer struct{}

// DiversificationScore represents the result of diversification scoring
type DiversificationScore struct {
	Components map[string]float64 `json:"components"`
	Score      float64            `json:"score"`
}

// NewDiversificationScorer creates a new diversification scorer
func NewDiversificationScorer() *DiversificationScorer {
	return &DiversificationScorer{}
}

// Calculate calculates the diversification score based on portfolio awareness
// Components:
// - Geography Gap (40%): Boost underweight regions
// - Industry Gap (30%): Boost underweight sectors
// - Averaging Down (30%): Bonus for quality dips we own
func (ds *DiversificationScorer) Calculate(
	symbol string,
	country string,
	industry *string,
	qualityScore float64,
	opportunityScore float64,
	portfolioContext *domain.PortfolioContext,
) DiversificationScore {
	// Default neutral scores if no portfolio context
	if portfolioContext == nil {
		return DiversificationScore{
			Score: 0.5,
			Components: map[string]float64{
				"country":   0.5,
				"industry":  0.5,
				"averaging": 0.5,
			},
		}
	}

	geoGapScore := calculateGeoGapScore(country, portfolioContext)
	industryGapScore := calculateIndustryGapScore(industry, portfolioContext)
	averagingDownScore := calculateAveragingDownScore(symbol, qualityScore, opportunityScore, portfolioContext)

	// Weights: 40% geography, 30% industry, 30% averaging
	totalScore := geoGapScore*0.40 + industryGapScore*0.30 + averagingDownScore*0.30
	totalScore = math.Min(1.0, totalScore)

	return DiversificationScore{
		Score: round3(totalScore),
		Components: map[string]float64{
			"country":   round3(geoGapScore),
			"industry":  round3(industryGapScore),
			"averaging": round3(averagingDownScore),
		},
	}
}

// calculateGeoGapScore calculates country gap score (40% weight)
// Higher weight = underweight region = higher score (buy to rebalance)
func calculateGeoGapScore(country string, portfolioContext *domain.PortfolioContext) float64 {
	// Map individual country to group
	group := "OTHER"
	if portfolioContext.CountryToGroup != nil {
		if g, ok := portfolioContext.CountryToGroup[country]; ok {
			group = g
		}
	}

	// Look up weight for the group (-1 to +1, where positive = underweight)
	geoWeight := 0.0
	if portfolioContext.CountryWeights != nil {
		geoWeight = portfolioContext.CountryWeights[group]
	}

	// Convert weight to score: 0.5 + (weight * 0.4)
	// weight=+1 (very underweight) → score=0.9
	// weight=0 (balanced) → score=0.5
	// weight=-1 (very overweight) → score=0.1
	geoGapScore := 0.5 + (geoWeight * 0.4)

	return math.Max(0.1, math.Min(0.9, geoGapScore))
}

// calculateIndustryGapScore calculates industry gap score (30% weight)
// Higher weight = underweight sector = higher score
func calculateIndustryGapScore(industry *string, portfolioContext *domain.PortfolioContext) float64 {
	if industry == nil || *industry == "" {
		return 0.5
	}

	// Split comma-separated industries
	industries := strings.Split(*industry, ",")
	if len(industries) == 0 {
		return 0.5
	}

	indScores := make([]float64, 0, len(industries))
	for _, ind := range industries {
		ind = strings.TrimSpace(ind)
		if ind == "" {
			continue
		}

		// Map individual industry to group
		group := "OTHER"
		if portfolioContext.IndustryToGroup != nil {
			if g, ok := portfolioContext.IndustryToGroup[ind]; ok {
				group = g
			}
		}

		// Look up weight for the group
		indWeight := 0.0
		if portfolioContext.IndustryWeights != nil {
			indWeight = portfolioContext.IndustryWeights[group]
		}

		// Convert weight to score
		indScore := 0.5 + (indWeight * 0.4)
		indScores = append(indScores, math.Max(0.1, math.Min(0.9, indScore)))
	}

	if len(indScores) == 0 {
		return 0.5
	}

	// Average across all industries
	sum := 0.0
	for _, score := range indScores {
		sum += score
	}
	return sum / float64(len(indScores))
}

// calculateAveragingDownScore calculates averaging down score (30% weight)
// Rewards buying more of quality positions that have dipped
func calculateAveragingDownScore(
	symbol string,
	qualityScore float64,
	opportunityScore float64,
	portfolioContext *domain.PortfolioContext,
) float64 {
	// Check if we own this position
	positionValue := 0.0
	if portfolioContext.Positions != nil {
		positionValue = portfolioContext.Positions[symbol]
	}

	// If we don't own it, return neutral
	if positionValue <= 0 {
		return 0.5
	}

	// Calculate averaging down potential
	avgDownPotential := qualityScore * opportunityScore

	// Base score based on potential
	averagingDownScore := 0.3
	if avgDownPotential >= 0.5 {
		averagingDownScore = 0.7 + (avgDownPotential-0.5)*0.6
	} else if avgDownPotential >= 0.3 {
		averagingDownScore = 0.5 + (avgDownPotential-0.3)*1.0
	}

	// Apply cost basis bonus
	averagingDownScore = applyCostBasisBonus(symbol, averagingDownScore, portfolioContext)

	// Apply concentration penalty
	averagingDownScore = applyConcentrationPenalty(positionValue, averagingDownScore, portfolioContext)

	return averagingDownScore
}

// applyCostBasisBonus applies bonus if current price is below average purchase price
// Rewards buying the dip on positions we're already in
func applyCostBasisBonus(symbol string, score float64, portfolioContext *domain.PortfolioContext) float64 {
	if portfolioContext.PositionAvgPrices == nil || portfolioContext.CurrentPrices == nil {
		return score
	}

	avgPrice, hasAvg := portfolioContext.PositionAvgPrices[symbol]
	currentPrice, hasCurrent := portfolioContext.CurrentPrices[symbol]

	if !hasAvg || !hasCurrent || avgPrice <= 0 {
		return score
	}

	// Calculate price vs average
	priceVsAvg := (currentPrice - avgPrice) / avgPrice

	// Only apply bonus if we're below average (loss)
	if priceVsAvg >= 0 {
		return score
	}

	// loss_pct is absolute value
	lossPct := math.Abs(priceVsAvg)

	// Only apply bonus up to COST_BASIS_BOOST_THRESHOLD (default 0.15 = 15%)
	if lossPct <= scoring.CostBasisBoostThreshold {
		costBasisBoost := math.Min(scoring.MaxCostBasisBoost, lossPct*2)
		return math.Min(1.0, score+costBasisBoost)
	}

	return score
}

// applyConcentrationPenalty penalizes over-concentration in single positions
// Prevents position from becoming too large relative to portfolio
func applyConcentrationPenalty(positionValue float64, score float64, portfolioContext *domain.PortfolioContext) float64 {
	totalValue := portfolioContext.TotalValue
	if totalValue <= 0 {
		return score
	}

	positionPct := positionValue / totalValue

	// Apply penalties for concentration
	if positionPct > scoring.ConcentrationHigh {
		// High concentration (>25%): 70% of original score
		return score * 0.7
	} else if positionPct > scoring.ConcentrationMed {
		// Medium concentration (>15%): 90% of original score
		return score * 0.9
	}

	return score
}
