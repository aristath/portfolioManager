-- Migration 037: Convert Deprecated Allocation Types
-- Run this on config.db to convert old allocation types to current format

-- Step 1: Backup (run this in shell before running SQL)
-- cp /opt/sentinel/config.db /opt/sentinel/config.db.backup-$(date +%Y%m%d-%H%M%S)

-- Step 2: Check current state
SELECT 'Before migration:' as status;
SELECT type, name, target_pct
FROM allocation_targets
WHERE type IN ('country_group', 'industry_group');

-- Step 3: Convert country_group -> geography
UPDATE allocation_targets
SET type = 'geography',
    updated_at = unixepoch()
WHERE type = 'country_group';

-- Step 4: Convert industry_group -> industry
UPDATE allocation_targets
SET type = 'industry',
    updated_at = unixepoch()
WHERE type = 'industry_group';

-- Step 5: Verify migration
SELECT 'After migration:' as status;
SELECT type, COUNT(*) as count
FROM allocation_targets
GROUP BY type;

-- Step 6: Check for any remaining deprecated types (should return 0 rows)
SELECT 'Deprecated types remaining (should be empty):' as status;
SELECT * FROM allocation_targets
WHERE type IN ('country_group', 'industry_group');
