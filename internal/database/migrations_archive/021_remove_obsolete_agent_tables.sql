-- Migration 021: Remove obsolete agent_configs and config_history tables
--
-- This migration removes tables that were used for multi-agent functionality:
-- - agent_configs: TOML strategy configurations (replaced by planner_settings in config.db)
-- - config_history: Version tracking for agent configs (no longer needed)
--
-- These tables are no longer used since the system was simplified to use
-- a single planner configuration stored in config.db (planner_settings table).
--
-- Data Migration Note:
-- Any existing data in these tables will be lost.
-- If agent_configs table exists, planner configuration should be migrated
-- to planner_settings table in config.db before running this migration.

-- Drop config_history table first (has foreign key to agent_configs)
DROP TABLE IF EXISTS config_history;

-- Drop agent_configs table
DROP TABLE IF EXISTS agent_configs;

-- Drop associated indexes
DROP INDEX IF EXISTS idx_agent_configs_bucket;
DROP INDEX IF EXISTS idx_agent_configs_name;
DROP INDEX IF EXISTS idx_config_history_agent;
DROP INDEX IF EXISTS idx_config_history_saved;
DROP INDEX IF EXISTS idx_config_history_performance;
