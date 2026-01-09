-- Cache Database Schema
-- Single source of truth for cache.db
-- This schema represents the final state after all migrations

-- REMOVED: recommendations table
-- Recommendations are now stored in-memory via InMemoryRecommendationRepository
-- This provides better performance and automatic cleanup on restart

-- REMOVED: cache_data table
-- Generic cache table was unused in the codebase

-- Job execution history (tracks last run time per job type)
-- Used by time-based scheduler to determine if jobs should run
CREATE TABLE IF NOT EXISTS job_history (
    job_type TEXT PRIMARY KEY,
    last_run_at INTEGER NOT NULL,    -- Unix timestamp (seconds since epoch)
    last_status TEXT NOT NULL DEFAULT 'success'
) STRICT;

CREATE INDEX IF NOT EXISTS idx_job_history_last_run ON job_history(last_run_at);
