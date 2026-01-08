-- Migration: Add retry tracking to recommendations table
-- This prevents infinite retries of failed trades

-- Add retry_count column (default 0)
ALTER TABLE recommendations ADD COLUMN retry_count INTEGER NOT NULL DEFAULT 0;

-- Add last_attempt_at column (nullable, Unix timestamp)
ALTER TABLE recommendations ADD COLUMN last_attempt_at INTEGER;

-- Add failure_reason column (nullable, for debugging)
ALTER TABLE recommendations ADD COLUMN failure_reason TEXT;

-- Update status CHECK constraint to include 'failed'
-- SQLite doesn't support ALTER CHECK, so we need to recreate the table
CREATE TABLE recommendations_new (
    uuid TEXT PRIMARY KEY,
    symbol TEXT NOT NULL,
    name TEXT NOT NULL,
    side TEXT NOT NULL CHECK (side IN ('BUY', 'SELL')),
    quantity REAL NOT NULL CHECK (quantity > 0),
    estimated_price REAL NOT NULL CHECK (estimated_price > 0),
    estimated_value REAL NOT NULL,
    reason TEXT NOT NULL,
    currency TEXT NOT NULL,
    priority REAL NOT NULL,
    current_portfolio_score REAL NOT NULL,
    new_portfolio_score REAL NOT NULL,
    score_change REAL NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'executed', 'rejected', 'expired', 'failed')),
    portfolio_hash TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    executed_at INTEGER,
    retry_count INTEGER NOT NULL DEFAULT 0,
    last_attempt_at INTEGER,
    failure_reason TEXT
) STRICT;

-- Copy data from old table
INSERT INTO recommendations_new
SELECT
    uuid, symbol, name, side, quantity, estimated_price, estimated_value,
    reason, currency, priority, current_portfolio_score, new_portfolio_score,
    score_change, status, portfolio_hash, created_at, updated_at, executed_at,
    retry_count, last_attempt_at, failure_reason
FROM recommendations;

-- Drop old table
DROP TABLE recommendations;

-- Rename new table
ALTER TABLE recommendations_new RENAME TO recommendations;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_recommendations_status ON recommendations(status);
CREATE INDEX IF NOT EXISTS idx_recommendations_created_at ON recommendations(created_at);
CREATE INDEX IF NOT EXISTS idx_recommendations_portfolio_hash ON recommendations(portfolio_hash);
