-- Calculations Database Schema
-- Single source of truth for calculations.db
-- This schema represents the final state after all migrations
-- Used for caching expensive per-security and portfolio-wide calculations

-- Per-security technical indicators (EMA, RSI, Sharpe, etc.)
-- Cached with TTL-based expiration for idle-time processing
CREATE TABLE IF NOT EXISTS technical_cache (
    isin TEXT NOT NULL,
    metric TEXT NOT NULL,
    period INTEGER NOT NULL DEFAULT 0,
    value REAL NOT NULL,
    expires_at INTEGER NOT NULL,
    PRIMARY KEY (isin, metric, period)
) STRICT;

-- Portfolio-wide optimizer calculations (covariance matrix, HRP, etc.)
-- Keyed by hash of ISINs to detect when portfolio composition changes
CREATE TABLE IF NOT EXISTS optimizer_cache (
    cache_type TEXT NOT NULL,
    isin_hash TEXT NOT NULL,
    value BLOB NOT NULL,
    expires_at INTEGER NOT NULL,
    PRIMARY KEY (cache_type, isin_hash)
) STRICT;

-- Indexes for efficient cleanup of expired entries
CREATE INDEX IF NOT EXISTS idx_tech_expires ON technical_cache(expires_at);
CREATE INDEX IF NOT EXISTS idx_opt_expires ON optimizer_cache(expires_at);
