package calculators

import (
	"fmt"
	"sort"

	"github.com/aristath/sentinel/internal/modules/planning/domain"
	"github.com/aristath/sentinel/internal/utils"
	"github.com/rs/zerolog"
)

// RebalanceBuysCalculator identifies underweight positions to buy for rebalancing.
// Uses mandatory tag-based pre-filtering for performance.
type RebalanceBuysCalculator struct {
	*BaseCalculator
	tagFilter    TagFilter
	securityRepo SecurityRepository
}

// NewRebalanceBuysCalculator creates a new rebalance buys calculator.
func NewRebalanceBuysCalculator(
	tagFilter TagFilter,
	securityRepo SecurityRepository,
	log zerolog.Logger,
) *RebalanceBuysCalculator {
	return &RebalanceBuysCalculator{
		BaseCalculator: NewBaseCalculator(log, "rebalance_buys"),
		tagFilter:      tagFilter,
		securityRepo:   securityRepo,
	}
}

// Name returns the calculator name.
func (c *RebalanceBuysCalculator) Name() string {
	return "rebalance_buys"
}

// Category returns the opportunity category.
func (c *RebalanceBuysCalculator) Category() domain.OpportunityCategory {
	return domain.OpportunityCategoryRebalanceBuys
}

// Calculate identifies rebalance buy opportunities.
func (c *RebalanceBuysCalculator) Calculate(
	ctx *domain.OpportunityContext,
	params map[string]interface{},
) (domain.CalculatorResult, error) {
	// Parameters with defaults
	minUnderweightThreshold := GetFloatParam(params, "min_underweight_threshold", 0.05) // 5% underweight
	
	// Calculate minimum trade amount based on transaction costs (default: 1% max cost ratio)
	// Formula: minTrade = fixedCost / (maxCostRatio - transactionCostPercent)
	// With €2 + 0.2% and 1% max ratio: 2 / (0.01 - 0.002) = €250
	maxCostRatio := GetFloatParam(params, "max_cost_ratio", 0.01) // Default 1% max cost
	minTradeAmount := ctx.CalculateMinTradeAmount(maxCostRatio)
	
	// Maximum position size constraints (allocation-based sizing with caps)
	// Percentage-based cap: maximum percentage of portfolio per position (default: 5%)
	maxPerPositionPct := GetFloatParam(params, "max_per_position_pct", 0.05)
	// Fixed cap: absolute maximum per position (default: €5000)
	maxPerPositionFixed := GetFloatParam(params, "max_per_position_fixed", 5000.0)
	maxPositions := GetIntParam(params, "max_positions", 0) // 0 = unlimited

	// Initialize exclusion collector
	exclusions := NewExclusionCollector(c.Name())

	// Extract config for tag filtering
	var config *domain.PlannerConfiguration
	if cfg, ok := params["config"].(*domain.PlannerConfiguration); ok && cfg != nil {
		config = cfg
	} else {
		config = domain.NewDefaultConfiguration()
	}

	// Tag-based pre-filtering (mandatory)
	var candidateMap map[string]bool
	if c.tagFilter != nil {
		candidateSymbols, err := c.tagFilter.GetOpportunityCandidates(ctx, config)
		if err != nil {
			return domain.CalculatorResult{PreFiltered: exclusions.Result()}, fmt.Errorf("failed to get tag-based candidates: %w", err)
		}

		if len(candidateSymbols) == 0 {
			c.log.Debug().Msg("No tag-based candidates found")
			return domain.CalculatorResult{PreFiltered: exclusions.Result()}, nil
		}

		// Build lookup map
		candidateMap = make(map[string]bool)
		for _, symbol := range candidateSymbols {
			candidateMap[symbol] = true
		}

		c.log.Debug().
			Int("tag_candidates", len(candidateSymbols)).
			Msg("Tag-based pre-filtering complete")
	}

	if !ctx.AllowBuy {
		c.log.Debug().Msg("Buying not allowed, skipping rebalance buys")
		return domain.CalculatorResult{PreFiltered: exclusions.Result()}, nil
	}

	// NOTE: Cash checks removed - sequence generator handles cash feasibility
	// This allows SELL→BUY sequences where sells generate cash for buys

	// Check if we have geography allocations and weights
	if ctx.GeographyAllocations == nil || ctx.GeographyWeights == nil {
		c.log.Debug().Msg("No geography allocation data available")
		return domain.CalculatorResult{PreFiltered: exclusions.Result()}, nil
	}

	c.log.Debug().
		Float64("min_underweight_threshold", minUnderweightThreshold).
		Msg("Calculating rebalance buys")

	// Identify underweight geographies
	underweightGeographies := make(map[string]float64)
	for geo, targetAllocation := range ctx.GeographyWeights {
		currentAllocation := ctx.GeographyAllocations[geo]
		underweight := targetAllocation - currentAllocation
		if underweight > minUnderweightThreshold {
			underweightGeographies[geo] = underweight
			c.log.Debug().
				Str("geography", geo).
				Float64("current", currentAllocation).
				Float64("target", targetAllocation).
				Float64("underweight", underweight).
				Msg("Underweight geography identified")
		}
	}

	if len(underweightGeographies) == 0 {
		c.log.Info().Msg("No underweight geographies - all geographies are at or above target")
		return domain.CalculatorResult{PreFiltered: exclusions.Result()}, nil
	}

	// Build candidates for securities in underweight geographies
	type scoredCandidate struct {
		isin        string
		symbol      string
		geography   string
		underweight float64
		score       float64
	}
	var scoredCandidates []scoredCandidate

	for isin, security := range ctx.StocksByISIN {
		symbol := security.Symbol
		securityName := security.Name

		// Skip if recently bought (ISIN lookup)
		if ctx.RecentlyBoughtISINs[isin] { // ISIN key ✅
			c.log.Debug().Str("symbol", symbol).Msg("FILTER: recently bought")
			exclusions.Add(isin, symbol, securityName, "recently bought (cooling off period)")
			continue
		}

		// Skip if tag-based pre-filtering excluded symbol (mandatory)
		// (excluded due to bad tags like quality-gate-fail, bubble-risk, etc.)
		if candidateMap != nil {
			if !candidateMap[symbol] {
				c.log.Debug().Str("symbol", symbol).Str("geography", security.Geography).Msg("FILTER: excluded by tag filter")
				exclusions.Add(isin, symbol, securityName, "excluded by tag filter (bad tags)")
				continue
			}
		}

		// Get security and extract geography
		geography := security.Geography
		if geography == "" {
			c.log.Debug().Str("symbol", symbol).Msg("FILTER: no geography")
			exclusions.Add(isin, symbol, securityName, "no geography assigned")
			continue
		}

		// Check per-security constraint: AllowBuy must be true
		if !security.AllowBuy {
			c.log.Debug().Str("symbol", symbol).Str("geography", geography).Msg("FILTER: allow_buy=false")
			exclusions.Add(isin, symbol, securityName, "allow_buy=false")
			continue
		}

		// Check if any of the security's geographies are underweight
		// Parse comma-separated geographies
		geos := utils.ParseCSV(geography)
		var underweight float64
		var matchedGeo string
		for _, geo := range geos {
			if uw, ok := underweightGeographies[geo]; ok {
				if uw > underweight {
					underweight = uw
					matchedGeo = geo
				}
			}
		}
		if matchedGeo == "" {
			c.log.Debug().Str("symbol", symbol).Str("geography", geography).Msg("FILTER: geography not underweight")
			exclusions.Add(isin, symbol, securityName, "geography not underweight")
			continue
		}

		c.log.Debug().Str("symbol", symbol).Str("matched_geo", matchedGeo).Float64("underweight", underweight).Msg("PASSED geography filter")

		// Get security score (ISIN lookup) - used for prioritization, not filtering
		score := 0.5 // Default neutral score
		if ctx.SecurityScores != nil {
			if s, ok := ctx.SecurityScores[isin]; ok { // ISIN key ✅
				score = s
			}
		}

		// Quality gate checks - CRITICAL protection against bad trades
		// Tags are mandatory - always rely on tags (quality-gate-fail, below-minimum-return, bubble-risk)
		// Tags encode explicit quality judgments from 7-path quality gates, minimum return requirements, and bubble detection.
		if c.securityRepo != nil {
			// Tag-based quality checks (when enabled)
			securityTags, err := c.securityRepo.GetTagsForSecurity(symbol)
			if err == nil {
				// Check for exclusion tags (inverted logic - skip if present)
				if contains(securityTags, "value-trap") || contains(securityTags, "ensemble-value-trap") {
					c.log.Debug().
						Str("symbol", symbol).
						Msg("Skipping - value trap detected (tag-based check)")
					exclusions.Add(isin, symbol, securityName, "value trap detected (tag-based)")
					continue
				}
				if contains(securityTags, "bubble-risk") || contains(securityTags, "ensemble-bubble-risk") {
					c.log.Debug().
						Str("symbol", symbol).
						Msg("Skipping - bubble risk detected (tag-based check)")
					exclusions.Add(isin, symbol, securityName, "bubble risk detected (tag-based)")
					continue
				}
				if contains(securityTags, "below-minimum-return") {
					c.log.Debug().
						Str("symbol", symbol).
						Msg("Skipping - below minimum return (tag-based check)")
					exclusions.Add(isin, symbol, securityName, "below minimum return (tag-based)")
					continue
				}
				// Skip if quality gate failed (inverted logic - cleaner)
				if contains(securityTags, "quality-gate-fail") {
					c.log.Debug().
						Str("symbol", symbol).
						Msg("Skipping - quality gate failed (tag-based check)")
					exclusions.Add(isin, symbol, securityName, "quality gate failed (tag-based)")
					continue
				}
			}
		}

		scoredCandidates = append(scoredCandidates, scoredCandidate{
			isin:        isin,
			symbol:      symbol,
			geography:   matchedGeo,
			underweight: underweight,
			score:       score,
		})
	}

	// Sort by combined priority (underweight * score)
	sort.Slice(scoredCandidates, func(i, j int) bool {
		priorityI := scoredCandidates[i].underweight * scoredCandidates[i].score
		priorityJ := scoredCandidates[j].underweight * scoredCandidates[j].score
		return priorityI > priorityJ
	})

	// Limit if needed
	if maxPositions > 0 && len(scoredCandidates) > maxPositions {
		scoredCandidates = scoredCandidates[:maxPositions]
	}

	c.log.Info().
		Int("scored_candidates", len(scoredCandidates)).
		Int("underweight_geographies", len(underweightGeographies)).
		Msg("Starting candidate creation from scored candidates")

	// Create action candidates
	var candidates []domain.ActionCandidate
	filteredCount := 0
	for _, scored := range scoredCandidates {
		isin := scored.isin
		symbol := scored.symbol

		// Get security info (direct ISIN lookup)
		security, ok := ctx.StocksByISIN[isin] // ISIN key ✅
		if !ok {
			filteredCount++
			c.log.Debug().
				Str("symbol", symbol).
				Str("isin", isin).
				Msg("FILTER: security not found in StocksByISIN")
			exclusions.Add(isin, symbol, "", "security not found in StocksByISIN")
			continue
		}

		securityName := security.Name

		// Get current price (direct ISIN lookup)
		currentPrice, ok := ctx.CurrentPrices[isin] // ISIN key ✅
		if !ok || currentPrice <= 0 {
			filteredCount++
			c.log.Info().
				Str("symbol", symbol).
				Str("isin", isin).
				Bool("price_exists", ok).
				Float64("price", currentPrice).
				Msg("FILTER: no current price available")
			exclusions.Add(isin, symbol, securityName, "no current price available")
			continue
		}

		// Calculate target value based on allocation gap (allocation-based sizing)
		// This is mathematically correct for rebalancing: buy enough to move toward target allocation
		underweightValue := scored.underweight * ctx.TotalPortfolioValueEUR
		
		// Apply maximum position size constraints
		maxPerPositionValue := maxPerPositionPct * ctx.TotalPortfolioValueEUR
		if maxPerPositionValue > maxPerPositionFixed {
			maxPerPositionValue = maxPerPositionFixed
		}
		
		// Start with allocation-based target
		targetValue := underweightValue
		
		// Cap at maximum per position
		if targetValue > maxPerPositionValue {
			targetValue = maxPerPositionValue
		}
		
		// Ensure minimum trade amount (transaction cost constraint)
		if targetValue < minTradeAmount {
			targetValue = minTradeAmount
		}
		
		c.log.Debug().
			Str("symbol", symbol).
			Str("geography", scored.geography).
			Float64("underweight_pct", scored.underweight*100).
			Float64("underweight_value", underweightValue).
			Float64("max_per_position", maxPerPositionValue).
			Float64("target_value", targetValue).
			Float64("min_trade_amount", minTradeAmount).
			Msg("Allocation-based position sizing")
		
		// NOTE: Cash cap removed - sequence generator handles cash feasibility

		quantity := int(targetValue / currentPrice)
		if quantity == 0 {
			quantity = 1
		}

		// Round quantity to lot size and validate
		quantityBeforeLot := quantity
		quantity = RoundToLotSize(quantity, security.MinLot)
		if quantity <= 0 {
			filteredCount++
			c.log.Info().
				Str("symbol", symbol).
				Int("quantity_before_lot", quantityBeforeLot).
				Int("min_lot", security.MinLot).
				Int("quantity_after_lot", quantity).
				Msg("FILTER: quantity below minimum lot size after rounding")
			exclusions.Add(isin, symbol, securityName, fmt.Sprintf("quantity below minimum lot size %d", security.MinLot))
			continue
		}

		// Recalculate value based on rounded quantity
		valueEUR := float64(quantity) * currentPrice

		// Check if rounded quantity still meets minimum trade amount
		if valueEUR < minTradeAmount {
			filteredCount++
			c.log.Info().
				Str("symbol", symbol).
				Float64("trade_value", valueEUR).
				Float64("min_trade_amount", minTradeAmount).
				Int("quantity", quantity).
				Float64("price", currentPrice).
				Msg("FILTER: trade below minimum trade amount after lot size rounding")
			exclusions.Add(isin, symbol, securityName, fmt.Sprintf("trade value €%.2f below minimum €%.2f", valueEUR, minTradeAmount))
			continue
		}

		// Concentration guardrail - block if would exceed limits
		// Note: For rebalancing, we're specifically trying to buy underweight geographies,
		// but we still respect position-level limits
		passes, concentrationReason := CheckConcentrationGuardrail(isin, security.Geography, valueEUR, ctx)
		if !passes {
			filteredCount++
			c.log.Info().
				Str("symbol", symbol).
				Str("isin", isin).
				Str("reason", concentrationReason).
				Float64("trade_value", valueEUR).
				Msg("FILTER: concentration limit exceeded")
			exclusions.Add(isin, symbol, securityName, concentrationReason)
			continue
		}

		// Apply transaction costs
		transactionCost := ctx.TransactionCostFixed + (valueEUR * ctx.TransactionCostPercent)
		totalCostEUR := valueEUR + transactionCost

		// NOTE: Cash check removed - sequence generator handles cash feasibility

		// Priority based on underweight and score
		priority := scored.underweight * scored.score * 0.6

		// Apply quantum warning penalty and priority boosts (30% for rebalance buys - new positions)
		if c.securityRepo != nil {
			securityTags, err := c.securityRepo.GetTagsForSecurity(symbol)
			if err == nil && len(securityTags) > 0 {
				priority = ApplyQuantumWarningPenalty(priority, securityTags, "rebalance_buys")
				priority = ApplyTagBasedPriorityBoosts(priority, securityTags, "rebalance_buys", c.securityRepo)
			}
		}

		// Build reason
		reason := fmt.Sprintf("Rebalance: %s underweight by %.1f%% (score: %.2f)",
			scored.geography, scored.underweight*100, scored.score)

		// Build tags
		tags := []string{"rebalance", "buy", "underweight"}

		candidate := domain.ActionCandidate{
			Side:     "BUY",
			ISIN:     isin,   // PRIMARY identifier ✅
			Symbol:   symbol, // BOUNDARY identifier
			Name:     security.Name,
			Quantity: quantity,
			Price:    currentPrice,
			ValueEUR: totalCostEUR,
			Currency: string(security.Currency),
			Priority: priority,
			Reason:   reason,
			Tags:     tags,
		}

		candidates = append(candidates, candidate)
	}

	c.log.Info().
		Int("candidates", len(candidates)).
		Int("underweight_countries", len(underweightGeographies)).
		Int("pre_filtered", len(exclusions.Result())).
		Int("scored_candidates_before_filtering", len(scoredCandidates)).
		Int("filtered_during_candidate_creation", filteredCount).
		Int("tag_candidates", len(candidateMap)).
		Msg("Rebalance buy opportunities identified")

	return domain.CalculatorResult{
		Candidates:  candidates,
		PreFiltered: exclusions.Result(),
	}, nil
}
