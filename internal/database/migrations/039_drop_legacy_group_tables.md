# Migration 039: Drop Legacy Group Tables

## Purpose
Remove unused `country_groups` and `industry_groups` tables from universe.db. These tables were deprecated in commit 3ec93a69 when the groups abstraction was removed in favor of direct geography/industry allocation.

## Background
- Originally, the system used intermediate mapping tables to group countries into geographies and industries into industry groups
- Migration 037 converted deprecated `country_group` and `industry_group` allocation types to `geography` and `industry`
- The current system uses direct country codes (`securities.data.attributes.CntryOfRisk`) and sector codes (`securities.data.sector_code`) for allocation
- No active code references these tables

## Changes

### 1. Drop country_groups table (universe.db)
```sql
DROP TABLE IF EXISTS country_groups;
```

### 2. Drop industry_groups table (universe.db)
```sql
DROP TABLE IF EXISTS industry_groups;
```

## Verification
```sql
-- Verify tables are dropped (should not appear in list)
.tables
```

## Impact
- **None** - These tables were not used by any active code
- The allocation system works directly with country/sector codes from securities data
- All allocation targets are stored in `allocation_targets` table in config.db with type='geography' or type='industry'

## Related Changes
- Commit 3ec93a69: Removed groups abstraction
- Migration 037: Converted deprecated allocation types
