-- Migration 020: Remove bucket_id from portfolio_snapshots table
--
-- This migration removes the bucket_id column from the portfolio_snapshots table.
-- The system is being simplified to use a single, unified portfolio without
-- per-bucket snapshot filtering.
--
-- Data Migration Note:
-- Any existing data in the bucket_id column will be lost.
-- All snapshots are now part of the single global portfolio.

-- Drop the index first
DROP INDEX IF EXISTS idx_snapshots_bucket;

-- Remove the bucket_id column
ALTER TABLE portfolio_snapshots DROP COLUMN bucket_id;
