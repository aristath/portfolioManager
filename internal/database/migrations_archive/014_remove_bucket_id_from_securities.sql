-- Migration 014: Remove bucket_id from securities table
--
-- This migration removes bucket_id column from securities table as part of
-- simplifying to a single universe (no multi-bucket system).
--
-- All securities are now in a unified universe without bucket filtering.

-- Drop the index first (required before dropping column in some SQLite versions)
DROP INDEX IF EXISTS idx_securities_bucket;

-- Drop the bucket_id column
ALTER TABLE securities DROP COLUMN bucket_id;
