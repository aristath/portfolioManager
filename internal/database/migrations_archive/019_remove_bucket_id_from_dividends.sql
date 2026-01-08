-- Migration 019: Remove bucket_id from dividend_history table
--
-- This migration removes the bucket_id column from the dividend_history table.
-- The system is being simplified to use a single, unified portfolio without
-- per-bucket dividend filtering.
--
-- Data Migration Note:
-- Any existing data in the bucket_id column will be lost.
-- All dividends are now part of the single global portfolio.

-- Drop the index first
DROP INDEX IF EXISTS idx_dividends_bucket;

-- Remove the bucket_id column
ALTER TABLE dividend_history DROP COLUMN bucket_id;
