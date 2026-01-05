-- Migration 029: Add Missing Score Columns
--
-- Adds raw value columns to scores table for performance optimization.
-- These values are currently recalculated on-demand during planning and
-- optimization, causing performance issues. Storing them enables faster
-- access without recalculation.
--
-- Columns added:
-- - sharpe_score: Raw Sharpe ratio (used in long-term scoring, optimization)
-- - drawdown_score: Raw max drawdown percentage (used in risk assessment, tag assignment)
-- - dividend_bonus: Dividend bonus value (used in optimization returns calculation)
-- - financial_strength_score: Financial strength component score (used in fundamentals scoring)
-- - rsi: Raw RSI value 0-100 (used in technicals scoring, tag assignment)
-- - ema_200: Raw 200-day EMA price value (used in technicals scoring, tag assignment)
-- - below_52w_high_pct: Raw percentage below 52-week high (used in opportunity scoring, tag assignment)

-- Portfolio Database (portfolio.db)
-- Add new columns to scores table
ALTER TABLE scores ADD COLUMN sharpe_score REAL;
ALTER TABLE scores ADD COLUMN drawdown_score REAL;
ALTER TABLE scores ADD COLUMN dividend_bonus REAL;
ALTER TABLE scores ADD COLUMN financial_strength_score REAL;
ALTER TABLE scores ADD COLUMN rsi REAL;
ALTER TABLE scores ADD COLUMN ema_200 REAL;
ALTER TABLE scores ADD COLUMN below_52w_high_pct REAL;
