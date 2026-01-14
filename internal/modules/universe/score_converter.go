package universe

import (
	scoringdomain "github.com/aristath/sentinel/internal/modules/scoring/domain"
)

// ConvertToSecurityScore converts scoringdomain.CalculatedSecurityScore to SecurityScore
// After migration: accepts ISIN as primary identifier
// Exported for use in tests and other packages
func ConvertToSecurityScore(isin string, symbol string, calculated *scoringdomain.CalculatedSecurityScore) SecurityScore {
	// Extract group scores
	groupScores := calculated.GroupScores
	if groupScores == nil {
		groupScores = make(map[string]float64)
	}

	// Calculate quality score as average of long_term and stability
	qualityScore := 0.0
	if longTerm, ok := groupScores["long_term"]; ok {
		if stability, ok2 := groupScores["stability"]; ok2 {
			qualityScore = (longTerm + stability) / 2
		} else {
			qualityScore = longTerm
		}
	} else if stability, ok := groupScores["stability"]; ok {
		qualityScore = stability
	}

	// Extract sub-scores
	subScores := calculated.SubScores
	var cagrScore, consistencyScore float64
	var sharpeScore, drawdownScore, financialStrengthScore float64
	var rsi, ema200, below52wHighPct, dividendBonus float64

	if subScores != nil {
		if longTermSubs, ok := subScores["long_term"]; ok {
			if cagr, ok := longTermSubs["cagr"]; ok {
				cagrScore = cagr
			}
			// Extract raw Sharpe ratio
			if sharpeRaw, ok := longTermSubs["sharpe_raw"]; ok {
				sharpeScore = sharpeRaw
			}
		}
		if stabilitySubs, ok := subScores["stability"]; ok {
			if consistency, ok := stabilitySubs["consistency"]; ok {
				consistencyScore = consistency
			}
			// Extract financial strength score
			if financialStrength, ok := stabilitySubs["financial_strength"]; ok {
				financialStrengthScore = financialStrength
			}
		}
		if shortTermSubs, ok := subScores["short_term"]; ok {
			// Extract raw drawdown percentage
			if drawdownRaw, ok := shortTermSubs["drawdown_raw"]; ok {
				drawdownScore = drawdownRaw
			}
		}
		if technicalsSubs, ok := subScores["technicals"]; ok {
			// Extract raw RSI value
			if rsiRaw, ok := technicalsSubs["rsi_raw"]; ok {
				rsi = rsiRaw
			}
			// Extract raw EMA200 value
			if emaRaw, ok := technicalsSubs["ema_raw"]; ok {
				ema200 = emaRaw
			}
		}
		if opportunitySubs, ok := subScores["opportunity"]; ok {
			// Extract raw below_52w_high percentage
			if below52wRaw, ok := opportunitySubs["below_52w_high_raw"]; ok {
				below52wHighPct = below52wRaw
			}
		}
		if dividendsSubs, ok := subScores["dividends"]; ok {
			// Extract dividend bonus value
			if bonus, ok := dividendsSubs["dividend_bonus"]; ok {
				dividendBonus = bonus
			}
		}
	}

	// Approximate history years
	historyYears := 0.0
	if cagrScore > 0 {
		historyYears = 5.0
	}

	volatility := 0.0
	if calculated.Volatility != nil {
		volatility = *calculated.Volatility
	}

	return SecurityScore{
		ISIN:                   isin,   // Primary identifier after migration
		Symbol:                 symbol, // Keep for display/backward compatibility
		QualityScore:           qualityScore,
		OpportunityScore:       groupScores["opportunity"],
		AnalystScore:           groupScores["opinion"],
		AllocationFitScore:     groupScores["diversification"],
		CAGRScore:              cagrScore,
		ConsistencyScore:       consistencyScore,
		HistoryYears:           historyYears,
		TechnicalScore:         groupScores["technicals"],
		StabilityScore:       groupScores["stability"],
		TotalScore:             calculated.TotalScore,
		Volatility:             volatility,
		FinancialStrengthScore: financialStrengthScore,
		SharpeScore:            sharpeScore,
		DrawdownScore:          drawdownScore,
		DividendBonus:          dividendBonus,
		RSI:                    rsi,
		EMA200:                 ema200,
		Below52wHighPct:        below52wHighPct,
		SellScore:              0, // Position-specific, not stored in scores table
	}
}
