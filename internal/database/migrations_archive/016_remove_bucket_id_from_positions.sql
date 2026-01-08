-- Migration 016: Remove bucket_id from positions table
--
-- This migration removes the bucket_id column from the positions table.
-- The system is being simplified to use a single, unified portfolio without
-- per-bucket position filtering.
--
-- Data Migration Note:
-- Any existing data in the bucket_id column will be lost.
-- All positions are now part of the single global portfolio.

-- Drop the index first
DROP INDEX IF EXISTS idx_positions_bucket;

-- Remove the bucket_id column
ALTER TABLE positions DROP COLUMN bucket_id;
