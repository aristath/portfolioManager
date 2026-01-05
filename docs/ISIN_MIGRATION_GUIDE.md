# ISIN Migration Guide

## Overview

This guide documents the migration from using `symbol` (Tradernet format) as the primary identifier to using `ISIN` (International Securities Identification Number) across all database tables and application code.

## Migration Status

✅ **COMPLETE** - All code has been updated, all tests are passing, and the system is ready for production use.

## Pre-Migration Checklist

Before running the migration on a production database, ensure:

1. **All securities have ISIN values**
   ```go
   validator := database.NewISINValidator(db)
   missing, err := validator.ValidateAllSecuritiesHaveISIN()
   if err != nil {
       // Handle error
   }
   if len(missing) > 0 {
       // Migration cannot proceed - fix missing ISINs first
   }
   ```

2. **No duplicate ISINs exist**
   ```go
   duplicates, err := validator.ValidateNoDuplicateISINs()
   if err != nil {
       // Handle error
   }
   if len(duplicates) > 0 {
       // Migration cannot proceed - fix duplicates first
   }
   ```

3. **All foreign key references are valid**
   ```go
   result, err := validator.ValidateAll()
   if err != nil {
       // Handle error
   }
   if !result.IsValid {
       // Fix issues before migration
       fmt.Println(result.FormatErrors())
   }
   ```

## Running the Migration

The migration script `030_migrate_to_isin_primary_key.sql` will be automatically executed when the database is initialized via the `DB.Migrate()` method.

### Manual Execution

If you need to run the migration manually:

```bash
# Connect to your database
sqlite3 universe.db < trader/internal/database/migrations/030_migrate_to_isin_primary_key.sql
```

**WARNING**: Always backup your database before running migrations!

## What Changed

### Database Schema

**Tables with PRIMARY KEY changed to ISIN:**
- `securities` - `isin TEXT PRIMARY KEY` (was `symbol TEXT PRIMARY KEY`)
- `scores` - `isin TEXT PRIMARY KEY` (was `symbol TEXT PRIMARY KEY`)
- `positions` - `isin TEXT PRIMARY KEY` (was `symbol TEXT PRIMARY KEY`)
- `security_tags` - `(isin, tag_id) PRIMARY KEY` (was `(symbol, tag_id) PRIMARY KEY`)

**Tables with ISIN column added (id remains PRIMARY KEY):**
- `trades` - `isin TEXT` column added and populated
- `dividend_history` - `isin TEXT` column added and populated
- `recommendations` - `isin TEXT` column (to be added in future migration)

**Symbol columns:**
- All `symbol` columns remain as indexed fields for display and API conversion
- Symbol is no longer used as PRIMARY KEY but is still indexed for lookups

### Application Code

**Repositories:**
- All repositories now use `GetByISIN()` as the primary lookup method
- `GetBySymbol()` is a helper that resolves symbol → ISIN internally
- `Update()` and `Delete()` methods now accept ISIN as the primary identifier

**Services:**
- All service layer code uses ISIN internally
- Symbol conversion happens at API boundaries only

**Handlers:**
- API endpoints accept ISIN in URL parameters (e.g., `/api/securities/{isin}`)
- `AddSecurityByIdentifier` still accepts symbols but resolves to ISIN internally
- All internal lookups use ISIN

**Domain Models:**
- `Security.ISIN` is now required (PRIMARY KEY)
- `SecurityScore.ISIN` is now required (PRIMARY KEY)
- `Position.ISIN` is now required (PRIMARY KEY)

## Post-Migration Verification

After migration, verify:

1. **All securities are accessible by ISIN**
   ```sql
   SELECT COUNT(*) FROM securities WHERE isin IS NOT NULL AND isin != '';
   ```

2. **All foreign keys are valid**
   ```sql
   -- Check scores
   SELECT COUNT(*) FROM scores s
   LEFT JOIN securities sec ON s.isin = sec.isin
   WHERE sec.isin IS NULL;
   -- Should return 0
   ```

3. **All data preserved**
   ```sql
   -- Compare counts before/after migration
   SELECT COUNT(*) FROM securities;
   SELECT COUNT(*) FROM scores;
   SELECT COUNT(*) FROM positions;
   ```

## Rollback Plan

**⚠️ IMPORTANT**: This migration is **NOT easily reversible**. The migration recreates tables with new PRIMARY KEYs, which means:

1. Always backup your database before migration
2. Test the migration on a copy of production data first
3. Ensure all application code is updated before running migration

If rollback is needed:
1. Restore from backup
2. Revert application code to previous version

## Testing

All migration tests are in `trader/internal/database/migrations/migration_isin_test.go`:

```bash
go test ./internal/database/migrations -v
```

## Future Improvements

- **History Database**: The `history.db` database still uses `symbol` column name but stores ISIN values. A future migration (`031_rename_history_symbol_to_isin.sql`) will rename the column for consistency. This is a naming improvement only - functionality is already correct.

## Support

If you encounter issues during migration:
1. Check validation results before migration
2. Review migration logs
3. Verify all securities have ISIN values from Tradernet API
4. Check for orphaned foreign key references

