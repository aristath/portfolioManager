# Migration 035: Drop dismissed_filters Table

## Description
Removes the dismissed_filters table and its index as the dismiss filter functionality is being removed from the application.

## Changes
- Drops the dismissed_filters table
- Drops the idx_dismissed_filters_isin index

## SQL
```sql
DROP INDEX IF EXISTS idx_dismissed_filters_isin;
DROP TABLE IF EXISTS dismissed_filters;
```

## Rollback
If needed, restore from config.db backup before applying this migration.

## Notes
- This is a destructive operation - all dismissed filter data will be lost
- No data migration needed as this feature is being completely removed
- Safe to apply as feature is being removed from codebase
