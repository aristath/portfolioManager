-- Migration 023: Remove priority_threshold column from planner_settings
--
-- This migration removes the priority_threshold column from planner_settings table.
-- Priority threshold filtering has been removed from the planner logic.
--
-- Note: SQLite doesn't support DROP COLUMN IF EXISTS, so we check if column exists first.
-- If the column doesn't exist, this migration is a no-op.

-- Check if column exists and drop it (SQLite 3.35.0+ required for DROP COLUMN)
-- If column doesn't exist, this will fail but migration handler will skip it
ALTER TABLE planner_settings DROP COLUMN priority_threshold;
