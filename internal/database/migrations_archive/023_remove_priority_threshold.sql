-- Migration 023: Remove priority_threshold column from planner_settings
--
-- This migration removes the priority_threshold column from planner_settings table.
-- Priority threshold filtering has been removed from the planner logic.
--
-- Note: SQLite doesn't support DROP COLUMN IF EXISTS.
-- If the column doesn't exist, the migration handler will skip this migration.

-- Drop the priority_threshold column (SQLite 3.35.0+ required for DROP COLUMN)
-- Migration handler will skip if column doesn't exist (no such column error)
ALTER TABLE planner_settings DROP COLUMN priority_threshold;
