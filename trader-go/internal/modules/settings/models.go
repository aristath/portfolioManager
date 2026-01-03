package settings

// SettingDefaults holds all default values for configurable settings
// Faithful translation from Python: app/api/settings.py -> SETTING_DEFAULTS
var SettingDefaults = map[string]interface{}{
	// Security scoring
	"min_security_score":   0.5,  // Minimum score for security to be recommended (0-1)
	"target_annual_return": 0.11, // Optimal CAGR for scoring (11%)
	"market_avg_pe":        22.0, // Reference P/E for valuation

	// Trading mode
	"trading_mode": "research", // "live" or "research" - blocks trades in research mode

	// Portfolio Optimizer settings
	"optimizer_blend":         0.5,  // 0.0 = pure Mean-Variance, 1.0 = pure HRP
	"optimizer_target_return": 0.11, // Target annual return for MV component

	// Cash management
	"min_cash_reserve": 500.0, // Minimum cash to keep (never fully deploy)

	// LED Matrix settings
	"ticker_speed":         50.0,  // Ticker scroll speed in ms per frame (lower = faster)
	"led_brightness":       150.0, // LED brightness (0-255)
	"ticker_show_value":    1.0,   // Show portfolio value
	"ticker_show_cash":     1.0,   // Show cash balance
	"ticker_show_actions":  1.0,   // Show next actions (BUY/SELL)
	"ticker_show_amounts":  1.0,   // Show amounts for actions
	"ticker_max_actions":   3.0,   // Max recommendations to show (buy + sell)

	// Job scheduling intervals
	"job_sync_cycle_minutes": 15.0, // Unified sync cycle interval
	"job_maintenance_hour":   3.0,  // Daily maintenance hour (0-23)
	"job_auto_deploy_minutes": 5.0, // Auto-deploy check interval (minutes)

	// Universe Pruning settings
	"universe_pruning_enabled":         1.0,  // 1.0 = enabled, 0.0 = disabled
	"universe_pruning_score_threshold": 0.50, // Minimum average score to keep security
	"universe_pruning_months":          3.0,  // Number of months to look back for scores
	"universe_pruning_min_samples":     2.0,  // Minimum number of score samples required
	"universe_pruning_check_delisted":  1.0,  // 1.0 = check for delisted securities

	// Event-Driven Rebalancing settings
	"event_driven_rebalancing_enabled":     1.0,  // 1.0 = enabled, 0.0 = disabled
	"rebalance_position_drift_threshold":   0.05, // Position drift threshold (0.05 = 5%)
	"rebalance_cash_threshold_multiplier":  2.0,  // Cash threshold = multiplier × min_trade_size

	// Trade Frequency Limits settings
	"trade_frequency_limits_enabled":     1.0,   // 1.0 = enabled, 0.0 = disabled
	"min_time_between_trades_minutes":    60.0,  // Minimum minutes between any trades
	"max_trades_per_day":                 4.0,   // Maximum trades per calendar day
	"max_trades_per_week":                10.0,  // Maximum trades per rolling 7-day window

	// Security Discovery settings
	"security_discovery_enabled":                1.0,               // 1.0 = enabled, 0.0 = disabled
	"security_discovery_score_threshold":        0.75,              // Minimum score to add security
	"security_discovery_max_per_month":          2.0,               // Maximum securities to add per month
	"security_discovery_require_manual_review":  0.0,               // 1.0 = require review, 0.0 = auto-add
	"security_discovery_geographies":            "EU,US,ASIA",      // Comma-separated geography list
	"security_discovery_exchanges":              "usa,europe",      // Comma-separated exchange list
	"security_discovery_min_volume":             1000000.0,         // Minimum daily volume for liquidity
	"security_discovery_fetch_limit":            50.0,              // Maximum candidates to fetch from API

	// Market Regime Detection settings
	"market_regime_detection_enabled":      1.0,   // 1.0 = enabled, 0.0 = disabled
	"market_regime_bull_cash_reserve":      0.02,  // Cash reserve percentage in bull market (2%)
	"market_regime_bear_cash_reserve":      0.05,  // Cash reserve percentage in bear market (5%)
	"market_regime_sideways_cash_reserve":  0.03,  // Cash reserve percentage in sideways market (3%)
	"market_regime_bull_threshold":         0.05,  // Threshold for bull market (5% above MA)
	"market_regime_bear_threshold":         -0.05, // Threshold for bear market (-5% below MA)

	// Virtual test currency (for testing planner in research mode)
	"virtual_test_cash": 0.0, // TEST currency amount (only visible in research mode)
}

// StringSettings defines which settings should be treated as strings rather than floats
var StringSettings = map[string]bool{
	"trading_mode":                     true,
	"security_discovery_geographies":   true,
	"security_discovery_exchanges":     true,
	"risk_profile":                     true,
}

// SettingUpdate represents a setting value update request
type SettingUpdate struct {
	Value interface{} `json:"value"`
}

// TradingModeResponse represents the trading mode response
type TradingModeResponse struct{
	TradingMode string `json:"trading_mode"`
}

// TradingModeToggleResponse represents the trading mode toggle response
type TradingModeToggleResponse struct {
	TradingMode  string `json:"trading_mode"`
	PreviousMode string `json:"previous_mode"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	SimpleCache      SimpleCacheStats      `json:"simple_cache"`
	CalculationsDB   CalculationsDBStats   `json:"calculations_db"`
}

// SimpleCacheStats represents simple cache statistics
type SimpleCacheStats struct {
	Entries int `json:"entries"`
}

// CalculationsDBStats represents calculations database statistics
type CalculationsDBStats struct {
	Entries        int `json:"entries"`
	ExpiredCleaned int `json:"expired_cleaned"`
}
