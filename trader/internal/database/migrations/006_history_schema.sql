-- History Database Schema
-- Migration 006: Create history.db schema for historical time-series data
--
-- This migration consolidates multiple databases into history.db:
-- - All history/{SYMBOL}.db files → single daily_prices table
-- - rates.db → exchange_rates table
-- - snapshots.db → portfolio_snapshots table (consolidated with portfolio.db)
--
-- Cleanup system:
-- - symbol_removals: Tracks 30-day grace period for removed securities
-- - cleanup_log: Audit trail of all cleanup operations
--
-- Data Migration Note:
-- During Phase 6, data will be migrated from 65+ per-symbol databases:
-- - history/AAPL_US.db, history/AMD_US.db, etc. → daily_prices table
-- - rates.db → exchange_rates table

-- Daily prices: OHLC data for all securities
-- Consolidates all history/{SYMBOL}.db files into single table
CREATE TABLE IF NOT EXISTS daily_prices (
    symbol TEXT NOT NULL,
    date TEXT NOT NULL,              -- YYYY-MM-DD format
    open REAL NOT NULL,
    high REAL NOT NULL,
    low REAL NOT NULL,
    close REAL NOT NULL,
    volume INTEGER,
    adjusted_close REAL,
    PRIMARY KEY (symbol, date)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_prices_symbol ON daily_prices(symbol);
CREATE INDEX IF NOT EXISTS idx_prices_date ON daily_prices(date DESC);
CREATE INDEX IF NOT EXISTS idx_prices_symbol_date ON daily_prices(symbol, date DESC);

-- Exchange rates: currency conversion history
-- Migrated from rates.db
CREATE TABLE IF NOT EXISTS exchange_rates (
    from_currency TEXT NOT NULL,
    to_currency TEXT NOT NULL,
    date TEXT NOT NULL,              -- YYYY-MM-DD format
    rate REAL NOT NULL,
    PRIMARY KEY (from_currency, to_currency, date)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_rates_pair ON exchange_rates(from_currency, to_currency);
CREATE INDEX IF NOT EXISTS idx_rates_date ON exchange_rates(date DESC);

-- Symbol removals: tracks grace period for security cleanup
-- When a security is deactivated, it's marked here with 30-day grace period
-- If reactivated within 30 days, row is deleted (data preserved)
-- After 30 days, daily cleanup job deletes historical data
CREATE TABLE IF NOT EXISTS symbol_removals (
    symbol TEXT PRIMARY KEY,
    removed_at INTEGER NOT NULL,     -- Unix timestamp
    grace_period_days INTEGER DEFAULT 30,
    row_count INTEGER,               -- Number of price rows for this symbol
    marked_by TEXT                   -- Who/what triggered the removal
) STRICT;

CREATE INDEX IF NOT EXISTS idx_removals_date ON symbol_removals(removed_at);

-- Cleanup log: audit trail of all data deletions
-- Tracks what was deleted, when, and why
CREATE TABLE IF NOT EXISTS cleanup_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    deleted_at INTEGER NOT NULL,     -- Unix timestamp
    row_count INTEGER,               -- How many rows were deleted
    cleanup_reason TEXT,             -- 'grace_period_expired', 'orphaned_data', etc.
    size_freed_bytes INTEGER         -- Approximate space freed
) STRICT;

CREATE INDEX IF NOT EXISTS idx_cleanup_symbol ON cleanup_log(symbol);
CREATE INDEX IF NOT EXISTS idx_cleanup_date ON cleanup_log(deleted_at DESC);

-- Database health tracking table
CREATE TABLE IF NOT EXISTS _database_health (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    checked_at INTEGER NOT NULL,  -- Unix timestamp
    integrity_check_passed INTEGER NOT NULL,  -- Boolean: 1 = passed, 0 = failed
    size_bytes INTEGER NOT NULL,
    wal_size_bytes INTEGER,
    page_count INTEGER,
    freelist_count INTEGER,
    vacuum_performed INTEGER DEFAULT 0,  -- Boolean: 1 = yes, 0 = no
    notes TEXT
) STRICT;

CREATE INDEX IF NOT EXISTS idx_health_checked_at ON _database_health(checked_at);
