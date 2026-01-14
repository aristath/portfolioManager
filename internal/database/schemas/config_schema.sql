-- Config Database Schema
-- Single source of truth for config.db
-- This schema represents the final state after all migrations

-- Settings table: application configuration (key-value store)
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at INTEGER NOT NULL      -- Unix timestamp (seconds since epoch)
) STRICT;

-- Allocation targets table: direct allocation rules (no groups)
CREATE TABLE IF NOT EXISTS allocation_targets (
    id INTEGER PRIMARY KEY,
    type TEXT NOT NULL,      -- 'geography' or 'industry' only
    name TEXT NOT NULL,
    target_pct REAL NOT NULL,
    created_at INTEGER NOT NULL,     -- Unix timestamp (seconds since epoch)
    updated_at INTEGER NOT NULL,     -- Unix timestamp (seconds since epoch)
    UNIQUE(type, name)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_allocation_type ON allocation_targets(type);

-- Planner settings: Direct storage of planner configuration
-- Single row table (only one config exists)
-- These are now handled by the ExhaustiveGenerator with constraints.Enforcer during generation
CREATE TABLE IF NOT EXISTS planner_settings (
    -- Primary key (constant value - only one row exists)
    id TEXT PRIMARY KEY DEFAULT 'main',

    -- Basic identification
    name TEXT NOT NULL DEFAULT 'default',
    description TEXT DEFAULT '',

    -- Global planner settings
    enable_batch_generation INTEGER DEFAULT 1,  -- Boolean
    max_depth INTEGER DEFAULT 10,                -- Increased for exhaustive generation
    max_opportunities_per_category INTEGER DEFAULT 10,
    enable_diverse_selection INTEGER DEFAULT 1,  -- Boolean
    diversity_weight REAL DEFAULT 0.3,

    -- Transaction costs
    transaction_cost_fixed REAL DEFAULT 5.0,
    transaction_cost_percent REAL DEFAULT 0.001,

    -- Trade permissions
    allow_sell INTEGER DEFAULT 1,  -- Boolean
    allow_buy INTEGER DEFAULT 1,   -- Boolean

    -- Risk management settings
    min_hold_days INTEGER DEFAULT 90,
    sell_cooldown_days INTEGER DEFAULT 180,
    max_loss_threshold REAL DEFAULT -0.20,
    max_sell_percentage REAL DEFAULT 0.20,
    averaging_down_percent REAL DEFAULT 0.10,

    -- Opportunity Calculator enabled flags
    enable_profit_taking_calc INTEGER DEFAULT 1,
    enable_averaging_down_calc INTEGER DEFAULT 1,
    enable_opportunity_buys_calc INTEGER DEFAULT 1,
    enable_rebalance_sells_calc INTEGER DEFAULT 1,
    enable_rebalance_buys_calc INTEGER DEFAULT 1,
    enable_weight_based_calc INTEGER DEFAULT 1,

    -- Post-generation filter enabled flags
    -- Note: Eligibility and RecentlyTraded filtering is now done during generation
    enable_correlation_aware_filter INTEGER DEFAULT 1,
    enable_diversity_filter INTEGER DEFAULT 1,
    enable_tag_filtering INTEGER DEFAULT 1,

    -- Optimizer settings
    optimizer_blend REAL DEFAULT 0.5,
    optimizer_target_return REAL DEFAULT 0.11,
    min_cash_reserve REAL DEFAULT 500.0,

    -- Timestamps
    updated_at INTEGER NOT NULL      -- Unix timestamp (seconds since epoch)
) STRICT;

-- Insert default row (single row table - use INSERT OR REPLACE)
INSERT OR REPLACE INTO planner_settings (id, name, description, updated_at)
VALUES ('main', 'default', 'Default planner configuration', (strftime('%s', 'now')));

-- Market regime history: tracks continuous regime scores over time
-- Per-region regime detection: each region has independent scores
CREATE TABLE IF NOT EXISTS market_regime_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    recorded_at INTEGER NOT NULL,        -- Unix timestamp (seconds since epoch)
    region TEXT NOT NULL DEFAULT 'GLOBAL',  -- Region: US, EU, ASIA, or GLOBAL (for backwards compat)
    raw_score REAL NOT NULL,             -- Raw regime score before smoothing (-1.0 to +1.0)
    smoothed_score REAL NOT NULL,         -- Exponentially smoothed score (-1.0 to +1.0)
    discrete_regime TEXT NOT NULL,       -- Label (unused by code)
    created_at INTEGER DEFAULT (strftime('%s', 'now'))  -- Unix timestamp (seconds since epoch)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_regime_history_recorded ON market_regime_history(recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_regime_history_smoothed ON market_regime_history(smoothed_score);
CREATE INDEX IF NOT EXISTS idx_regime_history_region ON market_regime_history(region);

-- Market indices: configuration for regime detection indices
-- Index types:
--   PRICE: Normal price indices used in regime composite (SP500, DAX, HSI, etc.)
--   VOLATILITY: VIX-style indices excluded from composite (inverse correlation)
CREATE TABLE IF NOT EXISTS market_indices (
    symbol TEXT PRIMARY KEY,             -- Tradernet symbol (e.g., "SP500.IDX")
    name TEXT NOT NULL,                  -- Human-readable name
    market_code TEXT NOT NULL,           -- Tradernet market code (FIX, EU, HKEX)
    region TEXT NOT NULL,                -- Region: US, EU, ASIA
    index_type TEXT NOT NULL DEFAULT 'PRICE',  -- PRICE or VOLATILITY
    enabled INTEGER NOT NULL DEFAULT 1,  -- Boolean: 1 = use for regime detection
    created_at INTEGER NOT NULL,         -- Unix timestamp (seconds since epoch)
    updated_at INTEGER NOT NULL          -- Unix timestamp (seconds since epoch)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_market_indices_region ON market_indices(region);
CREATE INDEX IF NOT EXISTS idx_market_indices_enabled ON market_indices(enabled);
CREATE INDEX IF NOT EXISTS idx_market_indices_type ON market_indices(index_type);

-- Adaptive parameters: current active adaptive values
CREATE TABLE IF NOT EXISTS adaptive_parameters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parameter_type TEXT NOT NULL UNIQUE, -- 'scoring_weights', 'optimizer_blend', 'quality_gates'
    parameter_value TEXT NOT NULL,       -- JSON
    regime_score REAL NOT NULL,          -- Regime score when adapted (-1.0 to +1.0)
    adapted_at INTEGER NOT NULL,         -- Unix timestamp (seconds since epoch)
    created_at INTEGER DEFAULT (strftime('%s', 'now'))  -- Unix timestamp (seconds since epoch)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_adaptive_params_type ON adaptive_parameters(parameter_type);
CREATE INDEX IF NOT EXISTS idx_adaptive_params_adapted ON adaptive_parameters(adapted_at DESC);

-- Dismissed filters: user-dismissed pre-filter reasons for securities
-- Row exists = dismissed, delete row = re-enabled
-- Cleared automatically when a trade (BUY/SELL) is executed on the security
CREATE TABLE IF NOT EXISTS dismissed_filters (
    isin TEXT NOT NULL,
    calculator TEXT NOT NULL,
    reason TEXT NOT NULL,
    PRIMARY KEY (isin, calculator, reason)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_dismissed_filters_isin ON dismissed_filters(isin);
