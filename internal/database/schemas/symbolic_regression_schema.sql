-- Symbolic Regression Database Schema
-- Stores discovered formulas for expected returns and security scoring

CREATE TABLE IF NOT EXISTS discovered_formulas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    formula_type TEXT NOT NULL,           -- 'expected_return' or 'scoring'
    security_type TEXT NOT NULL,           -- 'stock' or 'etf'
    regime_range_min REAL,                -- Optional: minimum regime score for this formula
    regime_range_max REAL,                -- Optional: maximum regime score for this formula
    formula_expression TEXT NOT NULL,     -- Formula as string (e.g., "0.65*cagr + 0.28*score")
    validation_metrics TEXT NOT NULL,     -- JSON: {"mae": 0.05, "rmse": 0.08, "spearman": 0.75, ...}
    fitness_score REAL NOT NULL,          -- Final fitness score (lower is better)
    complexity INTEGER NOT NULL,           -- Formula complexity (number of nodes)
    training_examples_count INTEGER,      -- Number of training examples used
    discovered_at INTEGER NOT NULL,        -- Unix timestamp (seconds since epoch)
    is_active INTEGER DEFAULT 1,          -- 1 = active (currently in use), 0 = inactive
    created_at INTEGER DEFAULT (strftime('%s', 'now'))  -- Unix timestamp (seconds since epoch)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_formulas_type ON discovered_formulas(formula_type);
CREATE INDEX IF NOT EXISTS idx_formulas_security_type ON discovered_formulas(security_type);
CREATE INDEX IF NOT EXISTS idx_formulas_active ON discovered_formulas(is_active);
CREATE INDEX IF NOT EXISTS idx_formulas_discovered ON discovered_formulas(discovered_at DESC);
