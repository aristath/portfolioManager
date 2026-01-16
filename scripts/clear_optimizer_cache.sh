#!/bin/bash
# Clear optimizer cache to remove old target weights with deprecated allocation types

set -e

DB_PATH="/opt/sentinel/cache.db"

echo "=== Clearing Optimizer Cache ==="
echo

echo "Clearing optimizer weights cache..."
sqlite3 "$DB_PATH" << SQL
-- Clear all optimizer-related cache entries
DELETE FROM calculation_cache
WHERE key LIKE 'optimizer:%'
   OR key LIKE 'hrp_weights:%'
   OR key LIKE 'mv_weights:%';

-- Show remaining cache entries count
SELECT COUNT(*) as remaining_entries FROM calculation_cache;
SQL

echo
echo "Optimizer cache cleared successfully!"
echo "Note: Cache will be rebuilt automatically on next optimization run."
