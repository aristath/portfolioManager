package utils

// TemperamentMapping defines how a temperament slider affects a parameter.
// Each mapping specifies which temperament dimension controls the parameter,
// the range of possible values, and the progression curve for interpolation.
type TemperamentMapping struct {
	// Parameter is the unique identifier for this mapping (e.g., "evaluation_opportunity_weight")
	Parameter string

	// Temperament is which slider controls this parameter: "risk_tolerance", "aggression", "patience", or "fixed"
	Temperament string

	// Inverse indicates if higher temperament = lower value (true) or higher value (false)
	Inverse bool

	// Min is the minimum output value (when temperament favors lower)
	Min float64

	// Max is the maximum output value (when temperament favors higher)
	Max float64

	// Base is the neutral value returned when temperament = 0.5
	Base float64

	// Progression defines the transformation curve: "linear", "linear-reverse", "exponential",
	// "exponential-reverse", "logarithmic", "logarithmic-reverse", "sigmoid", "sigmoid-reverse"
	Progression string

	// AbsoluteMin is the hard minimum that can NEVER be violated regardless of temperament
	AbsoluteMin float64

	// AbsoluteMax is the hard maximum that can NEVER be violated regardless of temperament
	AbsoluteMax float64
}

// temperamentMappings contains all parameter mappings organized by category.
// This is the single source of truth for how temperament affects system behavior.
var temperamentMappings = map[string]TemperamentMapping{
	// ==================================================================================
	// Category 1: EVALUATION WEIGHTS (5 params)
	// Philosophy: Narrow bands only (±5% max) - what we value stays consistent
	// ==================================================================================
	"evaluation_opportunity_weight": {
		Parameter:   "evaluation_opportunity_weight",
		Temperament: "aggression",
		Inverse:     false,
		Min:         0.25,
		Max:         0.35,
		Base:        0.30,
		Progression: "linear",
		AbsoluteMin: 0.20,
		AbsoluteMax: 0.40,
	},
	"evaluation_quality_weight": {
		Parameter:   "evaluation_quality_weight",
		Temperament: "aggression",
		Inverse:     true, // Aggressive = slightly less quality focus
		Min:         0.22,
		Max:         0.28,
		Base:        0.25,
		Progression: "linear-reverse",
		AbsoluteMin: 0.18,
		AbsoluteMax: 0.32,
	},
	"evaluation_risk_adjusted_weight": {
		Parameter:   "evaluation_risk_adjusted_weight",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = more risk metrics focus
		Min:         0.12,
		Max:         0.20,
		Base:        0.15,
		Progression: "linear-reverse",
		AbsoluteMin: 0.08,
		AbsoluteMax: 0.25,
	},
	"evaluation_diversification_weight": {
		Parameter:   "evaluation_diversification_weight",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = more diversification focus
		Min:         0.17,
		Max:         0.23,
		Base:        0.20,
		Progression: "linear-reverse",
		AbsoluteMin: 0.12,
		AbsoluteMax: 0.28,
	},
	"evaluation_regime_weight": {
		Parameter:   "evaluation_regime_weight",
		Temperament: "fixed", // Always matters - narrowest range
		Inverse:     false,
		Min:         0.08,
		Max:         0.12,
		Base:        0.10,
		Progression: "linear",
		AbsoluteMin: 0.05,
		AbsoluteMax: 0.15,
	},

	// ==================================================================================
	// Category 2: PROFIT TAKING (3 params)
	// ==================================================================================
	"profit_taking_min_gain_threshold": {
		Parameter:   "profit_taking_min_gain_threshold",
		Temperament: "patience",
		Inverse:     true, // Impatient = take profits sooner (lower threshold)
		Min:         0.10,
		Max:         0.25,
		Base:        0.15,
		Progression: "exponential-reverse",
		AbsoluteMin: 0.05,
		AbsoluteMax: 0.35,
	},
	"profit_taking_windfall_threshold": {
		Parameter:   "profit_taking_windfall_threshold",
		Temperament: "patience",
		Inverse:     true, // Impatient = lower windfall bar
		Min:         0.20,
		Max:         0.50,
		Base:        0.30,
		Progression: "linear-reverse",
		AbsoluteMin: 0.15,
		AbsoluteMax: 0.60,
	},
	"profit_taking_sell_percentage": {
		Parameter:   "profit_taking_sell_percentage",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = sell full position
		Min:         0.50,
		Max:         1.0,
		Base:        1.0,
		Progression: "linear",
		AbsoluteMin: 0.25,
		AbsoluteMax: 1.0,
	},

	// ==================================================================================
	// Category 3: AVERAGING DOWN (3 params)
	// ==================================================================================
	"averaging_down_max_loss_threshold": {
		Parameter:   "averaging_down_max_loss_threshold",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = average down on deeper losses (more negative)
		Min:         -0.30,
		Max:         -0.10,
		Base:        -0.20,
		Progression: "linear",
		AbsoluteMin: -0.50,
		AbsoluteMax: -0.05,
	},
	"averaging_down_min_loss_threshold": {
		Parameter:   "averaging_down_min_loss_threshold",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = start at smaller dips (less negative)
		Min:         -0.10,
		Max:         -0.03,
		Base:        -0.05,
		Progression: "linear",
		AbsoluteMin: -0.15,
		AbsoluteMax: -0.01,
	},
	"averaging_down_percent": {
		Parameter:   "averaging_down_percent",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = add more
		Min:         0.05,
		Max:         0.20,
		Base:        0.10,
		Progression: "logarithmic",
		AbsoluteMin: 0.02,
		AbsoluteMax: 0.30,
	},

	// ==================================================================================
	// Category 4: OPPORTUNITY BUYS (4 params)
	// ==================================================================================
	"opportunity_buys_min_score": {
		Parameter:   "opportunity_buys_min_score",
		Temperament: "aggression",
		Inverse:     true, // Conservative (low aggression) = higher quality bar
		Min:         0.55,
		Max:         0.80,
		Base:        0.65,
		Progression: "sigmoid-reverse",
		AbsoluteMin: 0.50,
		AbsoluteMax: 0.90,
	},
	"opportunity_buys_max_value_per_position": {
		Parameter:   "opportunity_buys_max_value_per_position",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = larger positions
		Min:         200,
		Max:         1000,
		Base:        500,
		Progression: "logarithmic",
		AbsoluteMin: 100,
		AbsoluteMax: 2000,
	},
	"opportunity_buys_max_positions": {
		Parameter:   "opportunity_buys_max_positions",
		Temperament: "aggression",
		Inverse:     false, // More aggressive = more candidates
		Min:         2,
		Max:         10,
		Base:        5,
		Progression: "linear",
		AbsoluteMin: 1,
		AbsoluteMax: 15,
	},
	"opportunity_buys_target_return_threshold_pct": {
		Parameter:   "opportunity_buys_target_return_threshold_pct",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = stricter CAGR requirement
		Min:         0.60,
		Max:         0.90,
		Base:        0.70,
		Progression: "linear-reverse",
		AbsoluteMin: 0.50,
		AbsoluteMax: 1.0,
	},

	// ==================================================================================
	// Category 5: KELLY SIZING (12 params)
	// ==================================================================================
	"kelly_fixed_fractional": {
		Parameter:   "kelly_fixed_fractional",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = higher Kelly fraction
		Min:         0.25,
		Max:         0.75,
		Base:        0.50,
		Progression: "sigmoid",
		AbsoluteMin: 0.15,
		AbsoluteMax: 0.80,
	},
	"kelly_min_position_size": {
		Parameter:   "kelly_min_position_size",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = smaller minimum (can take smaller bets)
		Min:         0.005,
		Max:         0.02,
		Base:        0.01,
		Progression: "linear",
		AbsoluteMin: 0.001,
		AbsoluteMax: 0.05,
	},
	"kelly_max_position_size": {
		Parameter:   "kelly_max_position_size",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = larger max positions
		Min:         0.08,
		Max:         0.25,
		Base:        0.15,
		Progression: "logarithmic",
		AbsoluteMin: 0.05,
		AbsoluteMax: 0.35,
	},
	"kelly_bear_reduction": {
		Parameter:   "kelly_bear_reduction",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = less bear reduction (closer to 1.0)
		Min:         0.50,
		Max:         0.90,
		Base:        0.75,
		Progression: "sigmoid",
		AbsoluteMin: 0.30,
		AbsoluteMax: 1.0,
	},
	"kelly_base_multiplier": {
		Parameter:   "kelly_base_multiplier",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = higher Kelly fraction
		Min:         0.30,
		Max:         0.70,
		Base:        0.50,
		Progression: "linear",
		AbsoluteMin: 0.20,
		AbsoluteMax: 0.80,
	},
	"kelly_confidence_adjustment_range": {
		Parameter:   "kelly_confidence_adjustment_range",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = more confidence-sensitive
		Min:         0.10,
		Max:         0.25,
		Base:        0.15,
		Progression: "linear",
		AbsoluteMin: 0.05,
		AbsoluteMax: 0.35,
	},
	"kelly_regime_adjustment_range": {
		Parameter:   "kelly_regime_adjustment_range",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = more regime-sensitive
		Min:         0.05,
		Max:         0.20,
		Base:        0.10,
		Progression: "linear",
		AbsoluteMin: 0.02,
		AbsoluteMax: 0.30,
	},
	"kelly_min_multiplier": {
		Parameter:   "kelly_min_multiplier",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = higher floor
		Min:         0.15,
		Max:         0.35,
		Base:        0.25,
		Progression: "linear",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.45,
	},
	"kelly_max_multiplier": {
		Parameter:   "kelly_max_multiplier",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = higher ceiling
		Min:         0.60,
		Max:         0.85,
		Base:        0.75,
		Progression: "linear",
		AbsoluteMin: 0.50,
		AbsoluteMax: 0.95,
	},
	"kelly_bear_max_reduction": {
		Parameter:   "kelly_bear_max_reduction",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = more reduction
		Min:         0.15,
		Max:         0.35,
		Base:        0.25,
		Progression: "linear-reverse",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.50,
	},
	"kelly_bull_threshold": {
		Parameter:   "kelly_bull_threshold",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = trigger bull earlier (lower threshold)
		Min:         0.30,
		Max:         0.70,
		Base:        0.50,
		Progression: "linear",
		AbsoluteMin: 0.20,
		AbsoluteMax: 0.80,
	},
	"kelly_bear_threshold": {
		Parameter:   "kelly_bear_threshold",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = trigger bear earlier (less negative)
		Min:         -0.70,
		Max:         -0.30,
		Base:        -0.50,
		Progression: "linear-reverse",
		AbsoluteMin: -0.80,
		AbsoluteMax: -0.20,
	},

	// ==================================================================================
	// Category 6: RISK MANAGEMENT (7 params)
	// ==================================================================================
	"risk_min_hold_days": {
		Parameter:   "risk_min_hold_days",
		Temperament: "patience",
		Inverse:     false, // Patient = hold longer
		Min:         30,
		Max:         180,
		Base:        90,
		Progression: "linear",
		AbsoluteMin: 14,
		AbsoluteMax: 365,
	},
	"risk_sell_cooldown_days": {
		Parameter:   "risk_sell_cooldown_days",
		Temperament: "patience",
		Inverse:     false, // Patient = longer cooldown
		Min:         60,
		Max:         365,
		Base:        180,
		Progression: "linear",
		AbsoluteMin: 30,
		AbsoluteMax: 730,
	},
	"risk_max_loss_threshold": {
		Parameter:   "risk_max_loss_threshold",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = tolerate deeper losses (more negative)
		Min:         -0.35,
		Max:         -0.10,
		Base:        -0.20,
		Progression: "linear",
		AbsoluteMin: -0.50,
		AbsoluteMax: -0.05,
	},
	"risk_max_sell_percentage": {
		Parameter:   "risk_max_sell_percentage",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = sell larger portions
		Min:         0.10,
		Max:         0.50,
		Base:        0.20,
		Progression: "exponential",
		AbsoluteMin: 0.05,
		AbsoluteMax: 0.75,
	},
	"risk_min_time_between_trades": {
		Parameter:   "risk_min_time_between_trades",
		Temperament: "patience",
		Inverse:     false, // Patient = more time between trades
		Min:         15,
		Max:         240,
		Base:        60,
		Progression: "linear",
		AbsoluteMin: 5,
		AbsoluteMax: 480,
	},
	"risk_max_trades_per_day": {
		Parameter:   "risk_max_trades_per_day",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = more daily trades
		Min:         2,
		Max:         8,
		Base:        4,
		Progression: "linear",
		AbsoluteMin: 1,
		AbsoluteMax: 12,
	},
	"risk_max_trades_per_week": {
		Parameter:   "risk_max_trades_per_week",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = more weekly trades
		Min:         4,
		Max:         20,
		Base:        10,
		Progression: "linear",
		AbsoluteMin: 2,
		AbsoluteMax: 30,
	},

	// ==================================================================================
	// Category 7: QUALITY GATES (4 params)
	// ==================================================================================
	"quality_stability_threshold": {
		Parameter:   "quality_stability_threshold",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.45,
		Max:         0.65,
		Base:        0.55,
		Progression: "sigmoid-reverse",
		AbsoluteMin: 0.35,
		AbsoluteMax: 0.80,
	},
	"quality_long_term_threshold": {
		Parameter:   "quality_long_term_threshold",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.35,
		Max:         0.55,
		Base:        0.45,
		Progression: "sigmoid-reverse",
		AbsoluteMin: 0.25,
		AbsoluteMax: 0.70,
	},
	"quality_exceptional_threshold": {
		Parameter:   "quality_exceptional_threshold",
		Temperament: "aggression",
		Inverse:     true, // Conservative = need higher excellence
		Min:         0.70,
		Max:         0.85,
		Base:        0.75,
		Progression: "linear-reverse",
		AbsoluteMin: 0.60,
		AbsoluteMax: 0.95,
	},
	"quality_absolute_min_cagr": {
		Parameter:   "quality_absolute_min_cagr",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = higher floor
		Min:         0.05,
		Max:         0.08,
		Base:        0.06,
		Progression: "linear-reverse",
		AbsoluteMin: 0.04,
		AbsoluteMax: 0.10,
	},

	// ==================================================================================
	// Category 8: REBALANCING (3 params)
	// ==================================================================================
	"rebalancing_min_overweight_threshold": {
		Parameter:   "rebalancing_min_overweight_threshold",
		Temperament: "patience",
		Inverse:     true, // Impatient = rebalance sooner (lower threshold)
		Min:         0.03,
		Max:         0.10,
		Base:        0.05,
		Progression: "linear-reverse",
		AbsoluteMin: 0.01,
		AbsoluteMax: 0.15,
	},
	"rebalancing_position_drift_threshold": {
		Parameter:   "rebalancing_position_drift_threshold",
		Temperament: "patience",
		Inverse:     true, // Impatient = tighter tolerance
		Min:         0.03,
		Max:         0.10,
		Base:        0.05,
		Progression: "linear-reverse",
		AbsoluteMin: 0.01,
		AbsoluteMax: 0.15,
	},
	"rebalancing_cash_threshold_multiplier": {
		Parameter:   "rebalancing_cash_threshold_multiplier",
		Temperament: "patience",
		Inverse:     false, // Patient = need more excess
		Min:         1.5,
		Max:         3.0,
		Base:        2.0,
		Progression: "linear",
		AbsoluteMin: 1.0,
		AbsoluteMax: 5.0,
	},

	// ==================================================================================
	// Category 9: VOLATILITY ACCEPTANCE (4 params)
	// ==================================================================================
	"volatility_volatile_threshold": {
		Parameter:   "volatility_volatile_threshold",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = tolerate more volatility
		Min:         0.25,
		Max:         0.40,
		Base:        0.30,
		Progression: "linear",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.60,
	},
	"volatility_high_threshold": {
		Parameter:   "volatility_high_threshold",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = higher bar for "high"
		Min:         0.35,
		Max:         0.55,
		Base:        0.40,
		Progression: "logarithmic",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.60,
	},
	"volatility_max_acceptable": {
		Parameter:   "volatility_max_acceptable",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept volatile securities
		Min:         0.30,
		Max:         0.50,
		Base:        0.40,
		Progression: "sigmoid",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.60,
	},
	"volatility_max_acceptable_drawdown": {
		Parameter:   "volatility_max_acceptable_drawdown",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = tolerate larger drawdowns
		Min:         0.20,
		Max:         0.40,
		Base:        0.30,
		Progression: "linear",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.50,
	},

	// ==================================================================================
	// Category 10: TRANSACTION EFFICIENCY (2 params)
	// ==================================================================================
	"transaction_max_cost_ratio": {
		Parameter:   "transaction_max_cost_ratio",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = accept higher cost drag
		Min:         0.005,
		Max:         0.02,
		Base:        0.01,
		Progression: "linear",
		AbsoluteMin: 0.001,
		AbsoluteMax: 0.05,
	},
	"transaction_limit_order_buffer": {
		Parameter:   "transaction_limit_order_buffer",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = tighter buffer (more likely to fill)
		Min:         0.02,
		Max:         0.10,
		Base:        0.05,
		Progression: "linear",
		AbsoluteMin: 0.01,
		AbsoluteMax: 0.15,
	},

	// ==================================================================================
	// Category 11: PRIORITY BOOST - PROFIT TAKING (6 params)
	// ==================================================================================
	"boost_windfall_priority": {
		Parameter:   "boost_windfall_priority",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = stronger windfall priority
		Min:         1.2,
		Max:         1.8,
		Base:        1.5,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_bubble_risk": {
		Parameter:   "boost_bubble_risk",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = sell risky positions faster
		Min:         1.2,
		Max:         1.6,
		Base:        1.4,
		Progression: "linear-reverse",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_needs_rebalance": {
		Parameter:   "boost_needs_rebalance",
		Temperament: "patience",
		Inverse:     true, // Impatient = stronger rebalance priority
		Min:         1.1,
		Max:         1.5,
		Base:        1.3,
		Progression: "linear-reverse",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_overweight": {
		Parameter:   "boost_overweight",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = address concentration faster
		Min:         1.05,
		Max:         1.4,
		Base:        1.2,
		Progression: "linear-reverse",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_overvalued": {
		Parameter:   "boost_overvalued",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = act on overvaluation
		Min:         1.05,
		Max:         1.3,
		Base:        1.15,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_near_52w_high": {
		Parameter:   "boost_near_52w_high",
		Temperament: "patience",
		Inverse:     true, // Impatient = take profits at highs
		Min:         1.0,
		Max:         1.2,
		Base:        1.1,
		Progression: "linear-reverse",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},

	// ==================================================================================
	// Category 12: PRIORITY BOOST - AVERAGING DOWN (4 params)
	// ==================================================================================
	"boost_quality_value": {
		Parameter:   "boost_quality_value",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = stronger quality value priority
		Min:         1.2,
		Max:         1.8,
		Base:        1.5,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_recovery_candidate": {
		Parameter:   "boost_recovery_candidate",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = bet on recovery
		Min:         1.1,
		Max:         1.5,
		Base:        1.3,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_high_quality": {
		Parameter:   "boost_high_quality",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = favor quality
		Min:         1.05,
		Max:         1.3,
		Base:        1.15,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_value_opportunity": {
		Parameter:   "boost_value_opportunity",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = act on value
		Min:         1.0,
		Max:         1.2,
		Base:        1.1,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},

	// ==================================================================================
	// Category 13: PRIORITY BOOST - OPPORTUNITY BUYS (12 params)
	// ==================================================================================
	"boost_quantum_warning_penalty": {
		Parameter:   "boost_quantum_warning_penalty",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = less penalty for uncertainty
		Min:         0.5,
		Max:         0.9,
		Base:        0.7,
		Progression: "linear",
		AbsoluteMin: 0.3,
		AbsoluteMax: 1.0,
	},
	"boost_quality_value_buy": {
		Parameter:   "boost_quality_value_buy",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = stronger quality value priority
		Min:         1.2,
		Max:         1.6,
		Base:        1.4,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_high_quality_value": {
		Parameter:   "boost_high_quality_value",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = favor quality + value
		Min:         1.1,
		Max:         1.5,
		Base:        1.3,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_deep_value": {
		Parameter:   "boost_deep_value",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = buy deep discounts
		Min:         1.05,
		Max:         1.4,
		Base:        1.2,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_oversold_quality": {
		Parameter:   "boost_oversold_quality",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = buy oversold quality
		Min:         1.0,
		Max:         1.3,
		Base:        1.15,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_excellent_returns": {
		Parameter:   "boost_excellent_returns",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = chase performance
		Min:         1.1,
		Max:         1.4,
		Base:        1.25,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_high_returns": {
		Parameter:   "boost_high_returns",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = favor high returns
		Min:         1.0,
		Max:         1.3,
		Base:        1.15,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_quality_high_cagr": {
		Parameter:   "boost_quality_high_cagr",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = favor quality growth
		Min:         1.05,
		Max:         1.35,
		Base:        1.2,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_dividend_grower": {
		Parameter:   "boost_dividend_grower",
		Temperament: "patience",
		Inverse:     false, // Patient = favor dividend growers
		Min:         1.0,
		Max:         1.3,
		Base:        1.15,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_high_dividend": {
		Parameter:   "boost_high_dividend",
		Temperament: "patience",
		Inverse:     false, // Patient = favor income
		Min:         1.0,
		Max:         1.2,
		Base:        1.1,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_quality_penalty_reduction_exceptional": {
		Parameter:   "boost_quality_penalty_reduction_exceptional",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = more forgiveness for quality
		Min:         0.5,
		Max:         0.8,
		Base:        0.65,
		Progression: "linear",
		AbsoluteMin: 0.3,
		AbsoluteMax: 1.0,
	},
	"boost_quality_penalty_reduction_high": {
		Parameter:   "boost_quality_penalty_reduction_high",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = more forgiveness
		Min:         0.65,
		Max:         0.95,
		Base:        0.80,
		Progression: "linear",
		AbsoluteMin: 0.3,
		AbsoluteMax: 1.0,
	},

	// ==================================================================================
	// Category 14: PRIORITY BOOST - REGIME (7 params)
	// ==================================================================================
	"boost_low_risk": {
		Parameter:   "boost_low_risk",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = stronger safety preference
		Min:         1.0,
		Max:         1.3,
		Base:        1.15,
		Progression: "linear-reverse",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_medium_risk": {
		Parameter:   "boost_medium_risk",
		Temperament: "fixed", // Neutral for medium risk
		Inverse:     false,
		Min:         1.0,
		Max:         1.15,
		Base:        1.05,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_high_risk_penalty": {
		Parameter:   "boost_high_risk_penalty",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = less penalty
		Min:         0.75,
		Max:         1.0,
		Base:        0.90,
		Progression: "linear",
		AbsoluteMin: 0.3,
		AbsoluteMax: 1.0,
	},
	"boost_growth_bull": {
		Parameter:   "boost_growth_bull",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = stronger regime alignment
		Min:         1.0,
		Max:         1.3,
		Base:        1.15,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_value_bear": {
		Parameter:   "boost_value_bear",
		Temperament: "patience",
		Inverse:     false, // Patient = stronger bear market value
		Min:         1.0,
		Max:         1.3,
		Base:        1.15,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_dividend_sideways": {
		Parameter:   "boost_dividend_sideways",
		Temperament: "patience",
		Inverse:     false, // Patient = income focus in sideways
		Min:         1.0,
		Max:         1.25,
		Base:        1.12,
		Progression: "linear",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},
	"boost_strong_stability": {
		Parameter:   "boost_strong_stability",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = favor stability
		Min:         1.0,
		Max:         1.25,
		Base:        1.12,
		Progression: "linear-reverse",
		AbsoluteMin: 0.5,
		AbsoluteMax: 2.0,
	},

	// ==================================================================================
	// Category 15: TAG ASSIGNER - VALUE (5 params)
	// ==================================================================================
	"tag_value_opportunity_discount_pct": {
		Parameter:   "tag_value_opportunity_discount_pct",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = smaller discount needed
		Min:         0.10,
		Max:         0.25,
		Base:        0.15,
		Progression: "linear",
		AbsoluteMin: 0.05,
		AbsoluteMax: 0.35,
	},
	"tag_deep_value_discount_pct": {
		Parameter:   "tag_deep_value_discount_pct",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = smaller discount for deep
		Min:         0.20,
		Max:         0.35,
		Base:        0.25,
		Progression: "linear",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.50,
	},
	"tag_deep_value_extreme_pct": {
		Parameter:   "tag_deep_value_extreme_pct",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = smaller extreme threshold
		Min:         0.25,
		Max:         0.40,
		Base:        0.30,
		Progression: "linear",
		AbsoluteMin: 0.15,
		AbsoluteMax: 0.55,
	},
	"tag_undervalued_pe_threshold": {
		Parameter:   "tag_undervalued_pe_threshold",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = smaller PE discount needed (less negative)
		Min:         -0.30,
		Max:         -0.10,
		Base:        -0.20,
		Progression: "linear",
		AbsoluteMin: -0.50,
		AbsoluteMax: -0.05,
	},
	"tag_below_52w_high_threshold": {
		Parameter:   "tag_below_52w_high_threshold",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = smaller dip needed
		Min:         0.05,
		Max:         0.15,
		Base:        0.10,
		Progression: "linear",
		AbsoluteMin: 0.02,
		AbsoluteMax: 0.25,
	},

	// ==================================================================================
	// Category 16: TAG ASSIGNER - QUALITY (8 params)
	// ==================================================================================
	"tag_high_quality_stability": {
		Parameter:   "tag_high_quality_stability",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.60,
		Max:         0.80,
		Base:        0.70,
		Progression: "linear-reverse",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.90,
	},
	"tag_high_quality_long_term": {
		Parameter:   "tag_high_quality_long_term",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.60,
		Max:         0.80,
		Base:        0.70,
		Progression: "linear-reverse",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.90,
	},
	"tag_stable_stability": {
		Parameter:   "tag_stable_stability",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.65,
		Max:         0.85,
		Base:        0.75,
		Progression: "linear-reverse",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.90,
	},
	"tag_stable_volatility_max": {
		Parameter:   "tag_stable_volatility_max",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept more volatility
		Min:         0.20,
		Max:         0.35,
		Base:        0.25,
		Progression: "linear",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.60,
	},
	"tag_stable_consistency": {
		Parameter:   "tag_stable_consistency",
		Temperament: "patience",
		Inverse:     false, // Patient = require consistency
		Min:         0.65,
		Max:         0.85,
		Base:        0.75,
		Progression: "linear",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.90,
	},
	"tag_consistent_grower_consistency": {
		Parameter:   "tag_consistent_grower_consistency",
		Temperament: "patience",
		Inverse:     false, // Patient = require consistency
		Min:         0.65,
		Max:         0.85,
		Base:        0.75,
		Progression: "linear",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.90,
	},
	"tag_consistent_grower_cagr": {
		Parameter:   "tag_consistent_grower_cagr",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept lower growth
		Min:         0.06,
		Max:         0.12,
		Base:        0.09,
		Progression: "linear",
		AbsoluteMin: 0.03,
		AbsoluteMax: 0.20,
	},
	"tag_strong_stability_threshold": {
		Parameter:   "tag_strong_stability_threshold",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.65,
		Max:         0.85,
		Base:        0.75,
		Progression: "linear-reverse",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.90,
	},

	// ==================================================================================
	// Category 17: TAG ASSIGNER - TECHNICAL (5 params)
	// ==================================================================================
	"tag_rsi_oversold": {
		Parameter:   "tag_rsi_oversold",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = buy at higher RSI
		Min:         20,
		Max:         40,
		Base:        30,
		Progression: "linear",
		AbsoluteMin: 10,
		AbsoluteMax: 50,
	},
	"tag_rsi_overbought": {
		Parameter:   "tag_rsi_overbought",
		Temperament: "patience",
		Inverse:     true, // Impatient = sell at lower RSI
		Min:         60,
		Max:         80,
		Base:        70,
		Progression: "linear-reverse",
		AbsoluteMin: 50,
		AbsoluteMax: 90,
	},
	"tag_recovery_momentum_threshold": {
		Parameter:   "tag_recovery_momentum_threshold",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = act on smaller dips (less negative)
		Min:         -0.10,
		Max:         0.0,
		Base:        -0.05,
		Progression: "linear",
		AbsoluteMin: -0.20,
		AbsoluteMax: 0.05,
	},
	"tag_recovery_stability_min": {
		Parameter:   "tag_recovery_stability_min",
		Temperament: "aggression",
		Inverse:     true, // Conservative = need better stability
		Min:         0.55,
		Max:         0.75,
		Base:        0.65,
		Progression: "linear-reverse",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.90,
	},
	"tag_recovery_discount_min": {
		Parameter:   "tag_recovery_discount_min",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = smaller discount needed
		Min:         0.08,
		Max:         0.18,
		Base:        0.12,
		Progression: "linear",
		AbsoluteMin: 0.03,
		AbsoluteMax: 0.30,
	},

	// ==================================================================================
	// Category 18: TAG ASSIGNER - DIVIDEND (4 params)
	// ==================================================================================
	"tag_high_dividend_yield": {
		Parameter:   "tag_high_dividend_yield",
		Temperament: "patience",
		Inverse:     false, // Patient = higher yield requirement
		Min:         0.025,
		Max:         0.06,
		Base:        0.04,
		Progression: "linear",
		AbsoluteMin: 0.01,
		AbsoluteMax: 0.10,
	},
	"tag_dividend_opportunity_score": {
		Parameter:   "tag_dividend_opportunity_score",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.45,
		Max:         0.65,
		Base:        0.55,
		Progression: "linear-reverse",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.80,
	},
	"tag_dividend_opportunity_yield": {
		Parameter:   "tag_dividend_opportunity_yield",
		Temperament: "patience",
		Inverse:     false, // Patient = higher yield requirement
		Min:         0.015,
		Max:         0.04,
		Base:        0.025,
		Progression: "linear",
		AbsoluteMin: 0.005,
		AbsoluteMax: 0.08,
	},
	"tag_dividend_consistency_score": {
		Parameter:   "tag_dividend_consistency_score",
		Temperament: "patience",
		Inverse:     false, // Patient = require consistency
		Min:         0.60,
		Max:         0.80,
		Base:        0.70,
		Progression: "linear",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.90,
	},

	// ==================================================================================
	// Category 19: TAG ASSIGNER - DANGER (7 params)
	// ==================================================================================
	"tag_overvalued_pe_threshold": {
		Parameter:   "tag_overvalued_pe_threshold",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = flag earlier (smaller threshold)
		Min:         0.10,
		Max:         0.35,
		Base:        0.20,
		Progression: "linear-reverse",
		AbsoluteMin: 0.05,
		AbsoluteMax: 0.50,
	},
	"tag_overvalued_near_high_pct": {
		Parameter:   "tag_overvalued_near_high_pct",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = flag earlier
		Min:         0.02,
		Max:         0.10,
		Base:        0.05,
		Progression: "linear-reverse",
		AbsoluteMin: 0.01,
		AbsoluteMax: 0.20,
	},
	"tag_unsustainable_gains_return": {
		Parameter:   "tag_unsustainable_gains_return",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = flag earlier
		Min:         0.35,
		Max:         0.70,
		Base:        0.50,
		Progression: "linear-reverse",
		AbsoluteMin: 0.20,
		AbsoluteMax: 1.0,
	},
	"tag_valuation_stretch_ema": {
		Parameter:   "tag_valuation_stretch_ema",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = flag earlier
		Min:         0.20,
		Max:         0.45,
		Base:        0.30,
		Progression: "linear-reverse",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.60,
	},
	"tag_underperforming_days": {
		Parameter:   "tag_underperforming_days",
		Temperament: "patience",
		Inverse:     false, // Patient = wait longer
		Min:         90,
		Max:         270,
		Base:        180,
		Progression: "linear",
		AbsoluteMin: 30,
		AbsoluteMax: 365,
	},
	"tag_stagnant_return_threshold": {
		Parameter:   "tag_stagnant_return_threshold",
		Temperament: "patience",
		Inverse:     false, // Patient = accept lower returns longer
		Min:         0.03,
		Max:         0.08,
		Base:        0.05,
		Progression: "linear",
		AbsoluteMin: 0.01,
		AbsoluteMax: 0.15,
	},
	"tag_stagnant_days_threshold": {
		Parameter:   "tag_stagnant_days_threshold",
		Temperament: "patience",
		Inverse:     false, // Patient = wait longer
		Min:         180,
		Max:         540,
		Base:        365,
		Progression: "linear",
		AbsoluteMin: 90,
		AbsoluteMax: 730,
	},

	// ==================================================================================
	// Category 20: TAG ASSIGNER - PORTFOLIO RISK (4 params)
	// ==================================================================================
	"tag_overweight_deviation": {
		Parameter:   "tag_overweight_deviation",
		Temperament: "patience",
		Inverse:     true, // Impatient = tighter tolerance
		Min:         0.01,
		Max:         0.05,
		Base:        0.02,
		Progression: "linear-reverse",
		AbsoluteMin: 0.005,
		AbsoluteMax: 0.10,
	},
	"tag_overweight_absolute": {
		Parameter:   "tag_overweight_absolute",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = allow larger positions
		Min:         0.05,
		Max:         0.15,
		Base:        0.10,
		Progression: "linear",
		AbsoluteMin: 0.02,
		AbsoluteMax: 0.40,
	},
	"tag_concentration_risk_threshold": {
		Parameter:   "tag_concentration_risk_threshold",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = allow concentration
		Min:         0.08,
		Max:         0.25,
		Base:        0.15,
		Progression: "linear",
		AbsoluteMin: 0.05,
		AbsoluteMax: 0.40,
	},
	"tag_needs_rebalance_deviation": {
		Parameter:   "tag_needs_rebalance_deviation",
		Temperament: "patience",
		Inverse:     true, // Impatient = tighter tolerance
		Min:         0.02,
		Max:         0.05,
		Base:        0.03,
		Progression: "linear-reverse",
		AbsoluteMin: 0.01,
		AbsoluteMax: 0.10,
	},

	// ==================================================================================
	// Category 21: TAG ASSIGNER - RISK PROFILE (8 params)
	// ==================================================================================
	"tag_low_risk_volatility_max": {
		Parameter:   "tag_low_risk_volatility_max",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept more volatility
		Min:         0.10,
		Max:         0.20,
		Base:        0.15,
		Progression: "linear",
		AbsoluteMin: 0.05,
		AbsoluteMax: 0.30,
	},
	"tag_low_risk_stability_min": {
		Parameter:   "tag_low_risk_stability_min",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.60,
		Max:         0.80,
		Base:        0.70,
		Progression: "linear-reverse",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.90,
	},
	"tag_low_risk_drawdown_max": {
		Parameter:   "tag_low_risk_drawdown_max",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept larger drawdowns
		Min:         0.15,
		Max:         0.30,
		Base:        0.20,
		Progression: "linear",
		AbsoluteMin: 0.05,
		AbsoluteMax: 0.50,
	},
	"tag_medium_risk_volatility_min": {
		Parameter:   "tag_medium_risk_volatility_min",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = wider range
		Min:         0.10,
		Max:         0.20,
		Base:        0.15,
		Progression: "linear",
		AbsoluteMin: 0.05,
		AbsoluteMax: 0.30,
	},
	"tag_medium_risk_volatility_max": {
		Parameter:   "tag_medium_risk_volatility_max",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = wider range
		Min:         0.25,
		Max:         0.40,
		Base:        0.30,
		Progression: "linear",
		AbsoluteMin: 0.15,
		AbsoluteMax: 0.60,
	},
	"tag_medium_risk_stability_min": {
		Parameter:   "tag_medium_risk_stability_min",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.45,
		Max:         0.65,
		Base:        0.55,
		Progression: "linear-reverse",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.80,
	},
	"tag_high_risk_volatility_threshold": {
		Parameter:   "tag_high_risk_volatility_threshold",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = higher threshold
		Min:         0.25,
		Max:         0.40,
		Base:        0.30,
		Progression: "linear",
		AbsoluteMin: 0.15,
		AbsoluteMax: 0.60,
	},
	"tag_high_risk_stability_threshold": {
		Parameter:   "tag_high_risk_stability_threshold",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.40,
		Max:         0.60,
		Base:        0.50,
		Progression: "linear-reverse",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.80,
	},

	// ==================================================================================
	// Category 22: TAG ASSIGNER - BUBBLE & VALUE TRAP (12 params)
	// ==================================================================================
	"tag_bubble_cagr_threshold": {
		Parameter:   "tag_bubble_cagr_threshold",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = flag earlier
		Min:         0.12,
		Max:         0.20,
		Base:        0.15,
		Progression: "linear-reverse",
		AbsoluteMin: 0.08,
		AbsoluteMax: 0.30,
	},
	"tag_bubble_sharpe_threshold": {
		Parameter:   "tag_bubble_sharpe_threshold",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept lower Sharpe
		Min:         0.35,
		Max:         0.70,
		Base:        0.50,
		Progression: "linear",
		AbsoluteMin: 0.20,
		AbsoluteMax: 1.0,
	},
	"tag_bubble_volatility_threshold": {
		Parameter:   "tag_bubble_volatility_threshold",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept more volatility
		Min:         0.30,
		Max:         0.50,
		Base:        0.40,
		Progression: "linear",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.60,
	},
	"tag_bubble_stability_threshold": {
		Parameter:   "tag_bubble_stability_threshold",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.45,
		Max:         0.65,
		Base:        0.55,
		Progression: "linear-reverse",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.80,
	},
	"tag_value_trap_stability": {
		Parameter:   "tag_value_trap_stability",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.45,
		Max:         0.65,
		Base:        0.55,
		Progression: "linear-reverse",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.80,
	},
	"tag_value_trap_long_term": {
		Parameter:   "tag_value_trap_long_term",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.35,
		Max:         0.55,
		Base:        0.45,
		Progression: "linear-reverse",
		AbsoluteMin: 0.20,
		AbsoluteMax: 0.70,
	},
	"tag_value_trap_momentum": {
		Parameter:   "tag_value_trap_momentum",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = allow weaker momentum (less negative)
		Min:         -0.10,
		Max:         0.0,
		Base:        -0.05,
		Progression: "linear",
		AbsoluteMin: -0.20,
		AbsoluteMax: 0.05,
	},
	"tag_value_trap_volatility": {
		Parameter:   "tag_value_trap_volatility",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept more volatility
		Min:         0.25,
		Max:         0.45,
		Base:        0.35,
		Progression: "linear",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.60,
	},
	"tag_quantum_bubble_high_prob": {
		Parameter:   "tag_quantum_bubble_high_prob",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = act at lower probability
		Min:         0.60,
		Max:         0.85,
		Base:        0.70,
		Progression: "linear-reverse",
		AbsoluteMin: 0.50,
		AbsoluteMax: 0.95,
	},
	"tag_quantum_bubble_warning_prob": {
		Parameter:   "tag_quantum_bubble_warning_prob",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = warn earlier
		Min:         0.40,
		Max:         0.65,
		Base:        0.50,
		Progression: "linear-reverse",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.80,
	},
	"tag_quantum_trap_high_prob": {
		Parameter:   "tag_quantum_trap_high_prob",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = act at lower probability
		Min:         0.60,
		Max:         0.85,
		Base:        0.70,
		Progression: "linear-reverse",
		AbsoluteMin: 0.50,
		AbsoluteMax: 0.95,
	},
	"tag_quantum_trap_warning_prob": {
		Parameter:   "tag_quantum_trap_warning_prob",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = warn earlier
		Min:         0.40,
		Max:         0.65,
		Base:        0.50,
		Progression: "linear-reverse",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.80,
	},

	// ==================================================================================
	// Category 23: TAG ASSIGNER - TOTAL RETURN (5 params)
	// ==================================================================================
	"tag_excellent_total_return": {
		Parameter:   "tag_excellent_total_return",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.15,
		Max:         0.22,
		Base:        0.18,
		Progression: "linear-reverse",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.30,
	},
	"tag_high_total_return": {
		Parameter:   "tag_high_total_return",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.12,
		Max:         0.18,
		Base:        0.15,
		Progression: "linear-reverse",
		AbsoluteMin: 0.08,
		AbsoluteMax: 0.25,
	},
	"tag_moderate_total_return": {
		Parameter:   "tag_moderate_total_return",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.09,
		Max:         0.15,
		Base:        0.12,
		Progression: "linear-reverse",
		AbsoluteMin: 0.05,
		AbsoluteMax: 0.20,
	},
	"tag_dividend_total_return_yield": {
		Parameter:   "tag_dividend_total_return_yield",
		Temperament: "patience",
		Inverse:     false, // Patient = higher yield bar
		Min:         0.05,
		Max:         0.10,
		Base:        0.08,
		Progression: "linear",
		AbsoluteMin: 0.02,
		AbsoluteMax: 0.15,
	},
	"tag_dividend_total_return_cagr": {
		Parameter:   "tag_dividend_total_return_cagr",
		Temperament: "fixed", // Minimum growth always
		Inverse:     false,
		Min:         0.03,
		Max:         0.08,
		Base:        0.05,
		Progression: "linear",
		AbsoluteMin: 0.01,
		AbsoluteMax: 0.12,
	},

	// ==================================================================================
	// Category 24: TAG ASSIGNER - REGIME SPECIFIC (7 params)
	// ==================================================================================
	"tag_bear_safe_volatility": {
		Parameter:   "tag_bear_safe_volatility",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept more
		Min:         0.15,
		Max:         0.30,
		Base:        0.20,
		Progression: "linear",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.40,
	},
	"tag_bear_safe_stability": {
		Parameter:   "tag_bear_safe_stability",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.60,
		Max:         0.80,
		Base:        0.70,
		Progression: "linear-reverse",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.90,
	},
	"tag_bear_safe_drawdown": {
		Parameter:   "tag_bear_safe_drawdown",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept more
		Min:         0.15,
		Max:         0.30,
		Base:        0.20,
		Progression: "linear",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.50,
	},
	"tag_bull_growth_cagr": {
		Parameter:   "tag_bull_growth_cagr",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = lower bar
		Min:         0.09,
		Max:         0.16,
		Base:        0.12,
		Progression: "linear",
		AbsoluteMin: 0.05,
		AbsoluteMax: 0.25,
	},
	"tag_bull_growth_stability": {
		Parameter:   "tag_bull_growth_stability",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.60,
		Max:         0.80,
		Base:        0.70,
		Progression: "linear-reverse",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.90,
	},
	"tag_regime_volatile_volatility": {
		Parameter:   "tag_regime_volatile_volatility",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = higher threshold
		Min:         0.25,
		Max:         0.40,
		Base:        0.30,
		Progression: "linear",
		AbsoluteMin: 0.15,
		AbsoluteMax: 0.50,
	},
	"tag_sideways_value_stability": {
		Parameter:   "tag_sideways_value_stability",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.70,
		Max:         0.85,
		Base:        0.75,
		Progression: "linear-reverse",
		AbsoluteMin: 0.50,
		AbsoluteMax: 0.95,
	},

	// ==================================================================================
	// Category 25: TAG ASSIGNER - QUALITY GATE PATHS (19 params)
	// ==================================================================================
	// Path 2: Exceptional Excellence
	"tag_quality_exceptional_excellence_threshold": {
		Parameter:   "tag_quality_exceptional_excellence_threshold",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.70,
		Max:         0.85,
		Base:        0.75,
		Progression: "linear-reverse",
		AbsoluteMin: 0.60,
		AbsoluteMax: 0.95,
	},
	// Path 3: Quality Value Play
	"tag_quality_value_stability_min": {
		Parameter:   "tag_quality_value_stability_min",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.55,
		Max:         0.70,
		Base:        0.60,
		Progression: "linear-reverse",
		AbsoluteMin: 0.40,
		AbsoluteMax: 0.85,
	},
	"tag_quality_value_opportunity_min": {
		Parameter:   "tag_quality_value_opportunity_min",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.60,
		Max:         0.75,
		Base:        0.65,
		Progression: "linear-reverse",
		AbsoluteMin: 0.50,
		AbsoluteMax: 0.85,
	},
	"tag_quality_value_long_term_min": {
		Parameter:   "tag_quality_value_long_term_min",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.25,
		Max:         0.40,
		Base:        0.30,
		Progression: "linear-reverse",
		AbsoluteMin: 0.15,
		AbsoluteMax: 0.55,
	},
	// Path 4: Dividend Income Play
	"tag_dividend_income_stability_min": {
		Parameter:   "tag_dividend_income_stability_min",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.50,
		Max:         0.65,
		Base:        0.55,
		Progression: "linear-reverse",
		AbsoluteMin: 0.40,
		AbsoluteMax: 0.75,
	},
	"tag_dividend_income_score_min": {
		Parameter:   "tag_dividend_income_score_min",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.60,
		Max:         0.75,
		Base:        0.65,
		Progression: "linear-reverse",
		AbsoluteMin: 0.50,
		AbsoluteMax: 0.85,
	},
	"tag_dividend_income_yield_min": {
		Parameter:   "tag_dividend_income_yield_min",
		Temperament: "patience",
		Inverse:     false, // Patient = higher yield bar
		Min:         0.025,
		Max:         0.050,
		Base:        0.035,
		Progression: "linear",
		AbsoluteMin: 0.015,
		AbsoluteMax: 0.080,
	},
	// Path 5: Risk-Adjusted Excellence
	"tag_risk_adjusted_sharpe_threshold": {
		Parameter:   "tag_risk_adjusted_sharpe_threshold",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = higher bar
		Min:         0.60,
		Max:         0.85,
		Base:        0.70,
		Progression: "linear-reverse",
		AbsoluteMin: 0.50,
		AbsoluteMax: 1.0,
	},
	"tag_risk_adjusted_sortino_threshold": {
		Parameter:   "tag_risk_adjusted_sortino_threshold",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = higher bar
		Min:         0.60,
		Max:         0.85,
		Base:        0.70,
		Progression: "linear-reverse",
		AbsoluteMin: 0.50,
		AbsoluteMax: 1.0,
	},
	"tag_risk_adjusted_long_term_threshold": {
		Parameter:   "tag_risk_adjusted_long_term_threshold",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.50,
		Max:         0.65,
		Base:        0.55,
		Progression: "linear-reverse",
		AbsoluteMin: 0.40,
		AbsoluteMax: 0.75,
	},
	"tag_risk_adjusted_volatility_max": {
		Parameter:   "tag_risk_adjusted_volatility_max",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept more volatility
		Min:         0.30,
		Max:         0.45,
		Base:        0.35,
		Progression: "linear",
		AbsoluteMin: 0.20,
		AbsoluteMax: 0.60,
	},
	// Path 6: Composite Minimum
	"tag_composite_stability_weight": {
		Parameter:   "tag_composite_stability_weight",
		Temperament: "fixed", // Fixed weights - sum to 1.0
		Inverse:     false,
		Min:         0.55,
		Max:         0.65,
		Base:        0.60,
		Progression: "linear",
		AbsoluteMin: 0.50,
		AbsoluteMax: 0.70,
	},
	"tag_composite_long_term_weight": {
		Parameter:   "tag_composite_long_term_weight",
		Temperament: "fixed", // Fixed weights - sum to 1.0
		Inverse:     false,
		Min:         0.35,
		Max:         0.45,
		Base:        0.40,
		Progression: "linear",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.50,
	},
	"tag_composite_score_min": {
		Parameter:   "tag_composite_score_min",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.48,
		Max:         0.58,
		Base:        0.52,
		Progression: "linear-reverse",
		AbsoluteMin: 0.40,
		AbsoluteMax: 0.65,
	},
	"tag_composite_stability_floor": {
		Parameter:   "tag_composite_stability_floor",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.40,
		Max:         0.55,
		Base:        0.45,
		Progression: "linear-reverse",
		AbsoluteMin: 0.30,
		AbsoluteMax: 0.65,
	},
	// Path 7: Growth Opportunity
	"tag_growth_opportunity_cagr_min": {
		Parameter:   "tag_growth_opportunity_cagr_min",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.11,
		Max:         0.16,
		Base:        0.13,
		Progression: "linear-reverse",
		AbsoluteMin: 0.08,
		AbsoluteMax: 0.20,
	},
	"tag_growth_opportunity_stability_min": {
		Parameter:   "tag_growth_opportunity_stability_min",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.45,
		Max:         0.60,
		Base:        0.50,
		Progression: "linear-reverse",
		AbsoluteMin: 0.35,
		AbsoluteMax: 0.70,
	},
	"tag_growth_opportunity_volatility_max": {
		Parameter:   "tag_growth_opportunity_volatility_max",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept more volatility
		Min:         0.35,
		Max:         0.50,
		Base:        0.40,
		Progression: "linear",
		AbsoluteMin: 0.25,
		AbsoluteMax: 0.60,
	},
	// High Score Tag
	"tag_high_score_threshold": {
		Parameter:   "tag_high_score_threshold",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.65,
		Max:         0.80,
		Base:        0.70,
		Progression: "linear-reverse",
		AbsoluteMin: 0.55,
		AbsoluteMax: 0.90,
	},
	// Value Opportunity Score
	"tag_value_opportunity_score_threshold": {
		Parameter:   "tag_value_opportunity_score_threshold",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.60,
		Max:         0.75,
		Base:        0.65,
		Progression: "linear-reverse",
		AbsoluteMin: 0.50,
		AbsoluteMax: 0.85,
	},
	// Growth Tag
	"tag_growth_tag_cagr_threshold": {
		Parameter:   "tag_growth_tag_cagr_threshold",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.12,
		Max:         0.15,
		Base:        0.13,
		Progression: "linear-reverse",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.20,
	},

	// ==================================================================================
	// Category 26: EVALUATION SCORING (15 params)
	// Note: Category 25 merged with Category 5 Kelly sizing
	// ==================================================================================
	"scoring_windfall_excess_high": {
		Parameter:   "scoring_windfall_excess_high",
		Temperament: "patience",
		Inverse:     true, // Impatient = lower bar for windfall
		Min:         0.35,
		Max:         0.70,
		Base:        0.50,
		Progression: "linear-reverse",
		AbsoluteMin: 0.20,
		AbsoluteMax: 1.0,
	},
	"scoring_windfall_excess_medium": {
		Parameter:   "scoring_windfall_excess_medium",
		Temperament: "patience",
		Inverse:     true, // Impatient = lower bar for windfall
		Min:         0.15,
		Max:         0.40,
		Base:        0.25,
		Progression: "linear-reverse",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.60,
	},
	"scoring_deviation_scale": {
		Parameter:   "scoring_deviation_scale",
		Temperament: "patience",
		Inverse:     true, // Impatient = stricter deviation penalty
		Min:         0.20,
		Max:         0.40,
		Base:        0.30,
		Progression: "linear-reverse",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.60,
	},
	"scoring_regime_bull_threshold": {
		Parameter:   "scoring_regime_bull_threshold",
		Temperament: "aggression",
		Inverse:     false, // Aggressive = trigger bull earlier (lower threshold)
		Min:         0.20,
		Max:         0.50,
		Base:        0.30,
		Progression: "linear",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.70,
	},
	"scoring_regime_bear_threshold": {
		Parameter:   "scoring_regime_bear_threshold",
		Temperament: "risk_tolerance",
		Inverse:     true, // Lower risk = trigger bear earlier (less negative)
		Min:         -0.50,
		Max:         -0.20,
		Base:        -0.30,
		Progression: "linear-reverse",
		AbsoluteMin: -0.70,
		AbsoluteMax: -0.10,
	},
	"scoring_volatility_excellent": {
		Parameter:   "scoring_volatility_excellent",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept more
		Min:         0.10,
		Max:         0.20,
		Base:        0.15,
		Progression: "linear",
		AbsoluteMin: 0.05,
		AbsoluteMax: 0.30,
	},
	"scoring_volatility_good": {
		Parameter:   "scoring_volatility_good",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept more
		Min:         0.20,
		Max:         0.35,
		Base:        0.25,
		Progression: "linear",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.50,
	},
	"scoring_volatility_acceptable": {
		Parameter:   "scoring_volatility_acceptable",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept more
		Min:         0.30,
		Max:         0.50,
		Base:        0.40,
		Progression: "linear",
		AbsoluteMin: 0.15,
		AbsoluteMax: 0.60,
	},
	"scoring_drawdown_excellent": {
		Parameter:   "scoring_drawdown_excellent",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept more
		Min:         0.05,
		Max:         0.15,
		Base:        0.10,
		Progression: "linear",
		AbsoluteMin: 0.02,
		AbsoluteMax: 0.25,
	},
	"scoring_drawdown_good": {
		Parameter:   "scoring_drawdown_good",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept more
		Min:         0.15,
		Max:         0.30,
		Base:        0.20,
		Progression: "linear",
		AbsoluteMin: 0.05,
		AbsoluteMax: 0.40,
	},
	"scoring_drawdown_acceptable": {
		Parameter:   "scoring_drawdown_acceptable",
		Temperament: "risk_tolerance",
		Inverse:     false, // Higher risk = accept more
		Min:         0.20,
		Max:         0.40,
		Base:        0.30,
		Progression: "linear",
		AbsoluteMin: 0.10,
		AbsoluteMax: 0.50,
	},
	"scoring_sharpe_excellent": {
		Parameter:   "scoring_sharpe_excellent",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         1.5,
		Max:         2.5,
		Base:        2.0,
		Progression: "linear-reverse",
		AbsoluteMin: 1.0,
		AbsoluteMax: 3.5,
	},
	"scoring_sharpe_good": {
		Parameter:   "scoring_sharpe_good",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.7,
		Max:         1.3,
		Base:        1.0,
		Progression: "linear-reverse",
		AbsoluteMin: 0.4,
		AbsoluteMax: 2.0,
	},
	"scoring_sharpe_acceptable": {
		Parameter:   "scoring_sharpe_acceptable",
		Temperament: "aggression",
		Inverse:     true, // Conservative = higher bar
		Min:         0.3,
		Max:         0.7,
		Base:        0.5,
		Progression: "linear-reverse",
		AbsoluteMin: 0.1,
		AbsoluteMax: 1.0,
	},
}

// GetTemperamentMapping returns the mapping for a specific parameter.
// Returns the mapping and true if found, or an empty mapping and false if not found.
func GetTemperamentMapping(parameter string) (TemperamentMapping, bool) {
	mapping, exists := temperamentMappings[parameter]
	return mapping, exists
}

// GetAllTemperamentMappings returns a copy of all temperament mappings.
func GetAllTemperamentMappings() map[string]TemperamentMapping {
	result := make(map[string]TemperamentMapping, len(temperamentMappings))
	for k, v := range temperamentMappings {
		result[k] = v
	}
	return result
}

// GetAdjustedValue calculates the adjusted parameter value based on the mapping
// and the current temperament slider values.
//
// Parameters:
//   - mapping: The temperament mapping for the parameter
//   - riskTolerance: Risk tolerance slider value (0.0-1.0)
//   - aggression: Aggression slider value (0.0-1.0)
//   - patience: Patience slider value (0.0-1.0)
//
// Returns the adjusted value, clamped to absolute bounds.
func GetAdjustedValue(mapping TemperamentMapping, riskTolerance, aggression, patience float64) float64 {
	// For fixed parameters, always return Base
	if mapping.Temperament == "fixed" {
		return mapping.Base
	}

	// Select the appropriate temperament value
	var temperamentValue float64
	switch mapping.Temperament {
	case "risk_tolerance":
		temperamentValue = riskTolerance
	case "aggression":
		temperamentValue = aggression
	case "patience":
		temperamentValue = patience
	default:
		return mapping.Base
	}

	// Determine progression based on Inverse flag
	progression := mapping.Progression
	// Note: For inverse relationships, we use the reverse progression
	// but we also want value=0 to give Max and value=1 to give Min
	// The existing GetTemperamentValue handles reverse progressions
	// by swapping the output endpoints

	// Use the existing GetTemperamentValue function
	params := TemperamentParams{
		Type:        mapping.Temperament,
		Value:       temperamentValue,
		Min:         mapping.Min,
		Max:         mapping.Max,
		Progression: progression,
		Base:        mapping.Base,
	}

	result := GetTemperamentValue(params)

	// Enforce absolute bounds
	if result < mapping.AbsoluteMin {
		result = mapping.AbsoluteMin
	}
	if result > mapping.AbsoluteMax {
		result = mapping.AbsoluteMax
	}

	return result
}
