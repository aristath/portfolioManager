package settings

import (
	"github.com/aristath/sentinel/internal/utils"
)

// getTemperamentValues retrieves the current temperament slider values
func (s *Service) getTemperamentValues() (riskTolerance, aggression, patience float64) {
	// Get risk_tolerance
	rtValue, err := s.Get("risk_tolerance")
	if err != nil || rtValue == nil {
		riskTolerance = 0.5
	} else if rt, ok := rtValue.(float64); ok {
		riskTolerance = rt
	} else {
		riskTolerance = 0.5
	}

	// Get temperament_aggression
	aggValue, err := s.Get("temperament_aggression")
	if err != nil || aggValue == nil {
		aggression = 0.5
	} else if agg, ok := aggValue.(float64); ok {
		aggression = agg
	} else {
		aggression = 0.5
	}

	// Get temperament_patience
	patValue, err := s.Get("temperament_patience")
	if err != nil || patValue == nil {
		patience = 0.5
	} else if pat, ok := patValue.(float64); ok {
		patience = pat
	} else {
		patience = 0.5
	}

	return
}

// getAdjustedParam retrieves a single adjusted parameter value
func (s *Service) getAdjustedParam(paramName string) float64 {
	mapping, exists := utils.GetTemperamentMapping(paramName)
	if !exists {
		s.log.Warn().Str("param", paramName).Msg("Temperament mapping not found, using base value")
		return 0
	}

	riskTolerance, aggression, patience := s.getTemperamentValues()
	return utils.GetAdjustedValue(mapping, riskTolerance, aggression, patience)
}

// ============================================================================
// EVALUATION WEIGHTS
// ============================================================================

// GetAdjustedEvaluationWeights returns normalized evaluation weights adjusted by temperament.
// Uses pure end-state scoring with 4 components.
func (s *Service) GetAdjustedEvaluationWeights() EvaluationWeights {
	weights := EvaluationWeights{
		PortfolioQuality:         s.getAdjustedParam("evaluation_quality_weight"),
		DiversificationAlignment: s.getAdjustedParam("evaluation_diversification_weight"),
		RiskAdjustedMetrics:      s.getAdjustedParam("evaluation_risk_adjusted_weight"),
		EndStateImprovement:      s.getAdjustedParam("evaluation_improvement_weight"),
	}

	return weights.Normalize()
}

// ============================================================================
// CORE TRADING PARAMS
// ============================================================================

// GetAdjustedProfitTakingParams returns profit-taking parameters adjusted by temperament
func (s *Service) GetAdjustedProfitTakingParams() ProfitTakingParams {
	return ProfitTakingParams{
		MinGainThreshold:  s.getAdjustedParam("profit_taking_min_gain_threshold"),
		WindfallThreshold: s.getAdjustedParam("profit_taking_windfall_threshold"),
		SellPercentage:    s.getAdjustedParam("profit_taking_sell_percentage"),
	}
}

// GetAdjustedAveragingDownParams returns averaging-down parameters adjusted by temperament
func (s *Service) GetAdjustedAveragingDownParams() AveragingDownParams {
	return AveragingDownParams{
		MaxLossThreshold: s.getAdjustedParam("averaging_down_max_loss_threshold"),
		MinLossThreshold: s.getAdjustedParam("averaging_down_min_loss_threshold"),
		Percent:          s.getAdjustedParam("averaging_down_percent"),
	}
}

// GetAdjustedOpportunityBuysParams returns opportunity buy parameters adjusted by temperament
func (s *Service) GetAdjustedOpportunityBuysParams() OpportunityBuysParams {
	return OpportunityBuysParams{
		MinScore:                 s.getAdjustedParam("opportunity_buys_min_score"),
		MaxValuePerPosition:      s.getAdjustedParam("opportunity_buys_max_value_per_position"),
		MaxPositions:             int(s.getAdjustedParam("opportunity_buys_max_positions")),
		TargetReturnThresholdPct: s.getAdjustedParam("opportunity_buys_target_return_threshold_pct"),
	}
}

// ============================================================================
// KELLY SIZING
// ============================================================================

// GetAdjustedKellyParams returns Kelly criterion parameters adjusted by temperament
func (s *Service) GetAdjustedKellyParams() KellyParams {
	return KellyParams{
		FixedFractional:           s.getAdjustedParam("kelly_fixed_fractional"),
		MinPositionSize:           s.getAdjustedParam("kelly_min_position_size"),
		MaxPositionSize:           s.getAdjustedParam("kelly_max_position_size"),
		BearReduction:             s.getAdjustedParam("kelly_bear_reduction"),
		BaseMultiplier:            s.getAdjustedParam("kelly_base_multiplier"),
		ConfidenceAdjustmentRange: s.getAdjustedParam("kelly_confidence_adjustment_range"),
		RegimeAdjustmentRange:     s.getAdjustedParam("kelly_regime_adjustment_range"),
		MinMultiplier:             s.getAdjustedParam("kelly_min_multiplier"),
		MaxMultiplier:             s.getAdjustedParam("kelly_max_multiplier"),
		BearMaxReduction:          s.getAdjustedParam("kelly_bear_max_reduction"),
		BullThreshold:             s.getAdjustedParam("kelly_bull_threshold"),
		BearThreshold:             s.getAdjustedParam("kelly_bear_threshold"),
	}
}

// ============================================================================
// RISK MANAGEMENT
// ============================================================================

// GetAdjustedRiskManagementParams returns risk management parameters.
// MinHoldDays and SellCooldownDays are read directly from settings (independent of temperament).
// Other parameters are still adjusted by temperament.
func (s *Service) GetAdjustedRiskManagementParams() RiskManagementParams {
	// Read cooloff settings directly from settings (independent of temperament)
	minHoldDays := s.getSettingFloat("min_hold_days", 90.0)
	sellCooldownDays := s.getSettingFloat("sell_cooldown_days", 180.0)

	return RiskManagementParams{
		MinHoldDays:          int(minHoldDays),
		SellCooldownDays:     int(sellCooldownDays),
		MaxLossThreshold:     s.getAdjustedParam("risk_max_loss_threshold"),
		MaxSellPercentage:    s.getAdjustedParam("risk_max_sell_percentage"),
		MinTimeBetweenTrades: int(s.getAdjustedParam("risk_min_time_between_trades")),
		MaxTradesPerDay:      int(s.getAdjustedParam("risk_max_trades_per_day")),
		MaxTradesPerWeek:     int(s.getAdjustedParam("risk_max_trades_per_week")),
	}
}

// getSettingFloat retrieves a float setting value directly from settings (not temperament-adjusted).
// Returns the default value if the setting doesn't exist or can't be parsed.
func (s *Service) getSettingFloat(key string, defaultValue float64) float64 {
	val, err := s.Get(key)
	if err != nil {
		return defaultValue
	}
	if floatVal, ok := val.(float64); ok {
		return floatVal
	}
	return defaultValue
}

// ============================================================================
// QUALITY GATES
// ============================================================================

// GetAdjustedQualityGateParams returns quality gate parameters adjusted by temperament
func (s *Service) GetAdjustedQualityGateParams() QualityGateParams {
	return QualityGateParams{
		StabilityThreshold:   s.getAdjustedParam("quality_stability_threshold"),
		LongTermThreshold:    s.getAdjustedParam("quality_long_term_threshold"),
		ExceptionalThreshold: s.getAdjustedParam("quality_exceptional_threshold"),
		AbsoluteMinCAGR:      s.getAdjustedParam("quality_absolute_min_cagr"),
		// Path 2: Exceptional Excellence
		ExceptionalExcellenceThreshold: s.getAdjustedParam("tag_quality_exceptional_excellence_threshold"),
		// Path 3: Quality Value Play
		QualityValueStabilityMin:   s.getAdjustedParam("tag_quality_value_stability_min"),
		QualityValueOpportunityMin: s.getAdjustedParam("tag_quality_value_opportunity_min"),
		QualityValueLongTermMin:    s.getAdjustedParam("tag_quality_value_long_term_min"),
		// Path 4: Dividend Income Play
		DividendIncomeStabilityMin: s.getAdjustedParam("tag_dividend_income_stability_min"),
		DividendIncomeScoreMin:     s.getAdjustedParam("tag_dividend_income_score_min"),
		DividendIncomeYieldMin:     s.getAdjustedParam("tag_dividend_income_yield_min"),
		// Path 5: Risk-Adjusted Excellence
		RiskAdjustedLongTermThreshold: s.getAdjustedParam("tag_risk_adjusted_long_term_threshold"),
		RiskAdjustedSharpeThreshold:   s.getAdjustedParam("tag_risk_adjusted_sharpe_threshold"),
		RiskAdjustedSortinoThreshold:  s.getAdjustedParam("tag_risk_adjusted_sortino_threshold"),
		RiskAdjustedVolatilityMax:     s.getAdjustedParam("tag_risk_adjusted_volatility_max"),
		// Path 6: Composite Minimum
		CompositeStabilityWeight: s.getAdjustedParam("tag_composite_stability_weight"),
		CompositeLongTermWeight:  s.getAdjustedParam("tag_composite_long_term_weight"),
		CompositeScoreMin:        s.getAdjustedParam("tag_composite_score_min"),
		CompositeStabilityFloor:  s.getAdjustedParam("tag_composite_stability_floor"),
		// Path 7: Growth Opportunity
		GrowthOpportunityCAGRMin:       s.getAdjustedParam("tag_growth_opportunity_cagr_min"),
		GrowthOpportunityStabilityMin:  s.getAdjustedParam("tag_growth_opportunity_stability_min"),
		GrowthOpportunityVolatilityMax: s.getAdjustedParam("tag_growth_opportunity_volatility_max"),
		// High Score Tag
		HighScoreThreshold: s.getAdjustedParam("tag_high_score_threshold"),
	}
}

// ============================================================================
// REBALANCING
// ============================================================================

// GetAdjustedRebalancingParams returns rebalancing parameters adjusted by temperament
func (s *Service) GetAdjustedRebalancingParams() RebalancingParams {
	return RebalancingParams{
		MinOverweightThreshold:  s.getAdjustedParam("rebalancing_min_overweight_threshold"),
		PositionDriftThreshold:  s.getAdjustedParam("rebalancing_position_drift_threshold"),
		CashThresholdMultiplier: s.getAdjustedParam("rebalancing_cash_threshold_multiplier"),
	}
}

// ============================================================================
// VOLATILITY
// ============================================================================

// GetAdjustedVolatilityParams returns volatility acceptance parameters adjusted by temperament
func (s *Service) GetAdjustedVolatilityParams() VolatilityParams {
	return VolatilityParams{
		VolatileThreshold:     s.getAdjustedParam("volatility_volatile_threshold"),
		HighThreshold:         s.getAdjustedParam("volatility_high_threshold"),
		MaxAcceptable:         s.getAdjustedParam("volatility_max_acceptable"),
		MaxAcceptableDrawdown: s.getAdjustedParam("volatility_max_acceptable_drawdown"),
	}
}

// ============================================================================
// TRANSACTION EFFICIENCY
// ============================================================================

// GetAdjustedTransactionParams returns transaction efficiency parameters adjusted by temperament
func (s *Service) GetAdjustedTransactionParams() TransactionParams {
	return TransactionParams{
		MaxCostRatio:     s.getAdjustedParam("transaction_max_cost_ratio"),
		LimitOrderBuffer: s.getAdjustedParam("transaction_limit_order_buffer"),
	}
}

// ============================================================================
// PRIORITY BOOSTS - PROFIT TAKING
// ============================================================================

// GetAdjustedProfitTakingBoosts returns profit-taking priority boosts adjusted by temperament
func (s *Service) GetAdjustedProfitTakingBoosts() ProfitTakingBoosts {
	return ProfitTakingBoosts{
		WindfallPriority: s.getAdjustedParam("boost_windfall_priority"),
		BubbleRisk:       s.getAdjustedParam("boost_bubble_risk"),
		NeedsRebalance:   s.getAdjustedParam("boost_needs_rebalance"),
		Overweight:       s.getAdjustedParam("boost_overweight"),
		Overvalued:       s.getAdjustedParam("boost_overvalued"),
		Near52wHigh:      s.getAdjustedParam("boost_near_52w_high"),
	}
}

// ============================================================================
// PRIORITY BOOSTS - AVERAGING DOWN
// ============================================================================

// GetAdjustedAveragingDownBoosts returns averaging-down priority boosts adjusted by temperament
func (s *Service) GetAdjustedAveragingDownBoosts() AveragingDownBoosts {
	return AveragingDownBoosts{
		QualityValue:      s.getAdjustedParam("boost_quality_value"),
		RecoveryCandidate: s.getAdjustedParam("boost_recovery_candidate"),
		HighQuality:       s.getAdjustedParam("boost_high_quality"),
		ValueOpportunity:  s.getAdjustedParam("boost_value_opportunity"),
	}
}

// ============================================================================
// PRIORITY BOOSTS - OPPORTUNITY BUYS
// ============================================================================

// GetAdjustedOpportunityBuyBoosts returns opportunity buy priority boosts adjusted by temperament
func (s *Service) GetAdjustedOpportunityBuyBoosts() OpportunityBuyBoosts {
	return OpportunityBuyBoosts{
		QuantumWarningPenalty:              s.getAdjustedParam("boost_quantum_warning_penalty"),
		QualityValueBuy:                    s.getAdjustedParam("boost_quality_value_buy"),
		HighQualityValue:                   s.getAdjustedParam("boost_high_quality_value"),
		DeepValue:                          s.getAdjustedParam("boost_deep_value"),
		OversoldQuality:                    s.getAdjustedParam("boost_oversold_quality"),
		ExcellentReturns:                   s.getAdjustedParam("boost_excellent_returns"),
		HighReturns:                        s.getAdjustedParam("boost_high_returns"),
		QualityHighCAGR:                    s.getAdjustedParam("boost_quality_high_cagr"),
		DividendGrower:                     s.getAdjustedParam("boost_dividend_grower"),
		HighDividend:                       s.getAdjustedParam("boost_high_dividend"),
		QualityPenaltyReductionExceptional: s.getAdjustedParam("boost_quality_penalty_reduction_exceptional"),
		QualityPenaltyReductionHigh:        s.getAdjustedParam("boost_quality_penalty_reduction_high"),
	}
}

// ============================================================================
// PRIORITY BOOSTS - REGIME
// ============================================================================

// GetAdjustedRegimeBoosts returns regime-based priority boosts adjusted by temperament
func (s *Service) GetAdjustedRegimeBoosts() RegimeBoosts {
	return RegimeBoosts{
		LowRisk:          s.getAdjustedParam("boost_low_risk"),
		MediumRisk:       s.getAdjustedParam("boost_medium_risk"),
		HighRiskPenalty:  s.getAdjustedParam("boost_high_risk_penalty"),
		GrowthBull:       s.getAdjustedParam("boost_growth_bull"),
		ValueBear:        s.getAdjustedParam("boost_value_bear"),
		DividendSideways: s.getAdjustedParam("boost_dividend_sideways"),
		HighStability:    s.getAdjustedParam("boost_high_stability"),
	}
}

// ============================================================================
// TAG THRESHOLDS - VALUE
// ============================================================================

// GetAdjustedValueThresholds returns value-related tag thresholds adjusted by temperament
func (s *Service) GetAdjustedValueThresholds() ValueThresholds {
	return ValueThresholds{
		ValueOpportunityDiscountPct: s.getAdjustedParam("tag_value_opportunity_discount_pct"),
		DeepValueDiscountPct:        s.getAdjustedParam("tag_deep_value_discount_pct"),
		DeepValueExtremePct:         s.getAdjustedParam("tag_deep_value_extreme_pct"),
		Below52wHighThreshold:       s.getAdjustedParam("tag_below_52w_high_threshold"),
	}
}

// ============================================================================
// TAG THRESHOLDS - QUALITY
// ============================================================================

// GetAdjustedQualityThresholds returns quality-related tag thresholds adjusted by temperament
func (s *Service) GetAdjustedQualityThresholds() QualityThresholds {
	return QualityThresholds{
		HighQualityStability:           s.getAdjustedParam("tag_high_quality_stability"),
		HighQualityLongTerm:            s.getAdjustedParam("tag_high_quality_long_term"),
		StableStability:                s.getAdjustedParam("tag_stable_stability"),
		StableVolatilityMax:            s.getAdjustedParam("tag_stable_volatility_max"),
		StableConsistency:              s.getAdjustedParam("tag_stable_consistency"),
		ConsistentGrowerConsistency:    s.getAdjustedParam("tag_consistent_grower_consistency"),
		ConsistentGrowerCAGR:           s.getAdjustedParam("tag_consistent_grower_cagr"),
		HighStabilityThreshold:         s.getAdjustedParam("tag_high_stability_threshold"),
		ValueOpportunityScoreThreshold: s.getAdjustedParam("tag_value_opportunity_score_threshold"),
	}
}

// ============================================================================
// TAG THRESHOLDS - TECHNICAL
// ============================================================================

// GetAdjustedTechnicalThresholds returns technical indicator thresholds adjusted by temperament
func (s *Service) GetAdjustedTechnicalThresholds() TechnicalThresholds {
	return TechnicalThresholds{
		RSIOversold:               s.getAdjustedParam("tag_rsi_oversold"),
		RSIOverbought:             s.getAdjustedParam("tag_rsi_overbought"),
		RecoveryMomentumThreshold: s.getAdjustedParam("tag_recovery_momentum_threshold"),
		RecoveryStabilityMin:      s.getAdjustedParam("tag_recovery_stability_min"),
		RecoveryDiscountMin:       s.getAdjustedParam("tag_recovery_discount_min"),
	}
}

// ============================================================================
// TAG THRESHOLDS - DIVIDEND
// ============================================================================

// GetAdjustedDividendThresholds returns dividend-related tag thresholds adjusted by temperament
func (s *Service) GetAdjustedDividendThresholds() DividendThresholds {
	return DividendThresholds{
		HighDividendYield:        s.getAdjustedParam("tag_high_dividend_yield"),
		DividendOpportunityScore: s.getAdjustedParam("tag_dividend_opportunity_score"),
		DividendOpportunityYield: s.getAdjustedParam("tag_dividend_opportunity_yield"),
		DividendConsistencyScore: s.getAdjustedParam("tag_dividend_consistency_score"),
	}
}

// ============================================================================
// TAG THRESHOLDS - DANGER
// ============================================================================

// GetAdjustedDangerThresholds returns danger/warning tag thresholds adjusted by temperament
func (s *Service) GetAdjustedDangerThresholds() DangerThresholds {
	return DangerThresholds{
		UnsustainableGainsReturn: s.getAdjustedParam("tag_unsustainable_gains_return"),
		ValuationStretchEMA:      s.getAdjustedParam("tag_valuation_stretch_ema"),
		UnderperformingDays:      int(s.getAdjustedParam("tag_underperforming_days")),
		StagnantReturnThreshold:  s.getAdjustedParam("tag_stagnant_return_threshold"),
		StagnantDaysThreshold:    int(s.getAdjustedParam("tag_stagnant_days_threshold")),
	}
}

// ============================================================================
// TAG THRESHOLDS - PORTFOLIO RISK
// ============================================================================

// GetAdjustedPortfolioRiskThresholds returns portfolio risk tag thresholds adjusted by temperament
func (s *Service) GetAdjustedPortfolioRiskThresholds() PortfolioRiskThresholds {
	return PortfolioRiskThresholds{
		OverweightDeviation:        s.getAdjustedParam("tag_overweight_deviation"),
		OverweightAbsolute:         s.getAdjustedParam("tag_overweight_absolute"),
		ConcentrationRiskThreshold: s.getAdjustedParam("tag_concentration_risk_threshold"),
		NeedsRebalanceDeviation:    s.getAdjustedParam("tag_needs_rebalance_deviation"),
	}
}

// ============================================================================
// TAG THRESHOLDS - RISK PROFILE
// ============================================================================

// GetAdjustedRiskProfileThresholds returns risk profile classification thresholds adjusted by temperament
func (s *Service) GetAdjustedRiskProfileThresholds() RiskProfileThresholds {
	return RiskProfileThresholds{
		LowRiskVolatilityMax:        s.getAdjustedParam("tag_low_risk_volatility_max"),
		LowRiskStabilityMin:         s.getAdjustedParam("tag_low_risk_stability_min"),
		LowRiskDrawdownMax:          s.getAdjustedParam("tag_low_risk_drawdown_max"),
		MediumRiskVolatilityMin:     s.getAdjustedParam("tag_medium_risk_volatility_min"),
		MediumRiskVolatilityMax:     s.getAdjustedParam("tag_medium_risk_volatility_max"),
		MediumRiskStabilityMin:      s.getAdjustedParam("tag_medium_risk_stability_min"),
		HighRiskVolatilityThreshold: s.getAdjustedParam("tag_high_risk_volatility_threshold"),
		HighRiskStabilityThreshold:  s.getAdjustedParam("tag_high_risk_stability_threshold"),
	}
}

// ============================================================================
// TAG THRESHOLDS - BUBBLE & VALUE TRAP
// ============================================================================

// GetAdjustedBubbleTrapThresholds returns bubble and value trap detection thresholds adjusted by temperament
func (s *Service) GetAdjustedBubbleTrapThresholds() BubbleTrapThresholds {
	return BubbleTrapThresholds{
		BubbleCAGRThreshold:       s.getAdjustedParam("tag_bubble_cagr_threshold"),
		BubbleSharpeThreshold:     s.getAdjustedParam("tag_bubble_sharpe_threshold"),
		BubbleVolatilityThreshold: s.getAdjustedParam("tag_bubble_volatility_threshold"),
		BubbleStabilityThreshold:  s.getAdjustedParam("tag_bubble_stability_threshold"),
		ValueTrapStability:        s.getAdjustedParam("tag_value_trap_stability"),
		ValueTrapLongTerm:         s.getAdjustedParam("tag_value_trap_long_term"),
		ValueTrapMomentum:         s.getAdjustedParam("tag_value_trap_momentum"),
		ValueTrapVolatility:       s.getAdjustedParam("tag_value_trap_volatility"),
		QuantumBubbleHighProb:     s.getAdjustedParam("tag_quantum_bubble_high_prob"),
		QuantumBubbleWarningProb:  s.getAdjustedParam("tag_quantum_bubble_warning_prob"),
		QuantumTrapHighProb:       s.getAdjustedParam("tag_quantum_trap_high_prob"),
		QuantumTrapWarningProb:    s.getAdjustedParam("tag_quantum_trap_warning_prob"),
		GrowthTagCAGRThreshold:    s.getAdjustedParam("tag_growth_tag_cagr_threshold"),
	}
}

// ============================================================================
// TAG THRESHOLDS - TOTAL RETURN
// ============================================================================

// GetAdjustedTotalReturnThresholds returns total return classification thresholds adjusted by temperament
func (s *Service) GetAdjustedTotalReturnThresholds() TotalReturnThresholds {
	return TotalReturnThresholds{
		ExcellentTotalReturn:     s.getAdjustedParam("tag_excellent_total_return"),
		HighTotalReturn:          s.getAdjustedParam("tag_high_total_return"),
		ModerateTotalReturn:      s.getAdjustedParam("tag_moderate_total_return"),
		DividendTotalReturnYield: s.getAdjustedParam("tag_dividend_total_return_yield"),
		DividendTotalReturnCAGR:  s.getAdjustedParam("tag_dividend_total_return_cagr"),
	}
}

// ============================================================================
// TAG THRESHOLDS - REGIME SPECIFIC
// ============================================================================

// GetAdjustedRegimeThresholds returns regime-specific tag thresholds adjusted by temperament
func (s *Service) GetAdjustedRegimeThresholds() RegimeThresholds {
	return RegimeThresholds{
		BearSafeVolatility:       s.getAdjustedParam("tag_bear_safe_volatility"),
		BearSafeStability:        s.getAdjustedParam("tag_bear_safe_stability"),
		BearSafeDrawdown:         s.getAdjustedParam("tag_bear_safe_drawdown"),
		BullGrowthCAGR:           s.getAdjustedParam("tag_bull_growth_cagr"),
		BullGrowthStability:      s.getAdjustedParam("tag_bull_growth_stability"),
		RegimeVolatileVolatility: s.getAdjustedParam("tag_regime_volatile_volatility"),
		SidewaysValueStability:   s.getAdjustedParam("tag_sideways_value_stability"),
	}
}

// ============================================================================
// EVALUATION SCORING
// ============================================================================

// GetAdjustedScoringParams returns evaluation scoring parameters adjusted by temperament.
// Uses pure end-state scoring (no windfall-related params).
func (s *Service) GetAdjustedScoringParams() ScoringParams {
	return ScoringParams{
		DeviationScale:       s.getAdjustedParam("scoring_deviation_scale"),
		RegimeBullThreshold:  s.getAdjustedParam("scoring_regime_bull_threshold"),
		RegimeBearThreshold:  s.getAdjustedParam("scoring_regime_bear_threshold"),
		VolatilityExcellent:  s.getAdjustedParam("scoring_volatility_excellent"),
		VolatilityGood:       s.getAdjustedParam("scoring_volatility_good"),
		VolatilityAcceptable: s.getAdjustedParam("scoring_volatility_acceptable"),
		DrawdownExcellent:    s.getAdjustedParam("scoring_drawdown_excellent"),
		DrawdownGood:         s.getAdjustedParam("scoring_drawdown_good"),
		DrawdownAcceptable:   s.getAdjustedParam("scoring_drawdown_acceptable"),
		SharpeExcellent:      s.getAdjustedParam("scoring_sharpe_excellent"),
		SharpeGood:           s.getAdjustedParam("scoring_sharpe_good"),
		SharpeAcceptable:     s.getAdjustedParam("scoring_sharpe_acceptable"),
	}
}
