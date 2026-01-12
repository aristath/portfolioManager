-- Migration script for universe.db
-- Adds alphavantage_symbol column for Alpha Vantage API integration
-- Run this script on the Arduino device: sqlite3 ~/data/universe.db < migrate_universe_alphavantage.sql

-- ============================================================================
-- SECURITIES TABLE
-- ============================================================================
-- Add alphavantage_symbol column (nullable, for Alpha Vantage API symbol lookup)
ALTER TABLE securities ADD COLUMN alphavantage_symbol TEXT;

-- Create index for efficient symbol lookups
CREATE INDEX IF NOT EXISTS idx_securities_alphavantage_symbol ON securities(alphavantage_symbol);
