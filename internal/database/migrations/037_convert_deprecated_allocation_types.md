# Migration 037: Convert Deprecated Allocation Types

## Purpose
Convert deprecated `country_group` and `industry_group` allocation types to the current `geography` and `industry` format introduced in commit 3ec93a69.

## Changes

### 1. Update allocation_targets (config.db)
```sql
-- Convert country_group -> geography
UPDATE allocation_targets
SET type = 'geography'
WHERE type = 'country_group';

-- Convert industry_group -> industry
UPDATE allocation_targets
SET type = 'industry'
WHERE type = 'industry_group';
```

## Verification
```sql
-- Check for any remaining deprecated types (should return 0 rows)
SELECT * FROM allocation_targets
WHERE type IN ('country_group', 'industry_group');

-- Verify conversion
SELECT type, COUNT(*) as count
FROM allocation_targets
GROUP BY type;
```

## Rollback
```sql
-- If needed to rollback (NOT RECOMMENDED)
UPDATE allocation_targets
SET type = 'country_group'
WHERE type = 'geography';

UPDATE allocation_targets
SET type = 'industry_group'
WHERE type = 'industry';
```

## Related Changes
- Commit 3ec93a69: Removed groups abstraction
- Test file fixes: get_optimizer_weights_test.go mock data updated
- Cache cleared: Optimizer cache cleared to remove old target weights
