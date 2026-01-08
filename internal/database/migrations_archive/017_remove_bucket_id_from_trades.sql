-- Migration 017: Remove bucket_id from trades table
--
-- This migration removes the bucket_id column from the trades table.
-- The system is being simplified to use a single, unified portfolio without
-- per-bucket trade filtering.
--
-- Data Migration Note:
-- Any existing data in the bucket_id column will be lost.
-- All trades are now part of the single global portfolio.

-- Drop the index first
DROP INDEX IF EXISTS idx_trades_bucket;

-- Remove the bucket_id column
ALTER TABLE trades DROP COLUMN bucket_id;
