-- Migration 018: Remove bucket_id from cash_flows table
--
-- This migration removes the bucket_id column from the cash_flows table.
-- The system is being simplified to use a single, unified portfolio without
-- per-bucket cash flow filtering.
--
-- Data Migration Note:
-- Any existing data in the bucket_id column will be lost.
-- All cash flows are now part of the single global portfolio.

-- Drop the index first
DROP INDEX IF EXISTS idx_cashflows_bucket;

-- Remove the bucket_id column
ALTER TABLE cash_flows DROP COLUMN bucket_id;
