-- Cache Database Schema
-- Single source of truth for cache.db
-- This schema represents the final state after all migrations

-- REMOVED: recommendations table
-- Recommendations are now stored in-memory via InMemoryRecommendationRepository
-- This provides better performance and automatic cleanup on restart

-- REMOVED: cache_data table
-- Generic cache table was unused in the codebase

-- REMOVED: job_history table
-- Job execution history was unused in the codebase

-- Generic cache table for key-value storage with expiration
CREATE TABLE IF NOT EXISTS cache (
    key TEXT PRIMARY KEY,
    value TEXT,
    expires_at INTEGER  -- Unix timestamp (seconds since epoch)
) STRICT;
