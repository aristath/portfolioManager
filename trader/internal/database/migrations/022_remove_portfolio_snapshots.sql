-- Migration 022: Remove portfolio_snapshots table
--
-- This migration removes the portfolio_snapshots table and its associated index.
-- Portfolio snapshots were used for historical tracking and analytics, which
-- are no longer needed as the system focuses on decision-making only.
--
-- Data Migration Note:
-- Any existing snapshot data will be lost.
-- Historical portfolio values can be reconstructed from position history
-- and price history if needed in the future.

-- Drop the index first
DROP INDEX IF EXISTS idx_snapshots_date;

-- Drop the portfolio_snapshots table
DROP TABLE IF EXISTS portfolio_snapshots;
