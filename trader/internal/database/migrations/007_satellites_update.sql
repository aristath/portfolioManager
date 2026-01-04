-- Satellites Database Schema Update
-- Migration 007: Update satellites.db schema for agent integration
--
-- ⚠️ OBSOLETE: This migration is no longer used.
-- Satellites/buckets functionality has been removed from the system.
-- The system now uses a single, unified portfolio without buckets.
--
-- This migration file is kept for historical reference only.
-- It will be skipped automatically on databases where the buckets table doesn't exist.
--
-- Original purpose (now obsolete):
-- - Add agent_id column to buckets table (references agent_configs in agents.db)
-- - Remove strategy_type column (replaced by agent_id)
--
-- Each satellite bucket would reference a TOML strategy configuration (agent)
-- from the agents database.

-- Attempt to add the column (will fail silently if it already exists via migration system error handling)
ALTER TABLE buckets ADD COLUMN agent_id TEXT;

CREATE INDEX IF NOT EXISTS idx_buckets_agent ON buckets(agent_id);

-- Note: This migration will be skipped on databases without buckets table
-- (all databases now since satellites.db is removed)
