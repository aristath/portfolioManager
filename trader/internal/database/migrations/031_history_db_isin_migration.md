# History Database ISIN Migration (Future)

## Current State

The history database (`history.db`) currently uses a `symbol` column in `daily_prices` and `monthly_prices` tables, but the application code stores ISIN values in this column.

## Migration Plan

### Step 1: Create Migration Script
Create migration `031_rename_history_symbol_to_isin.sql` to:
- Rename `symbol` column to `isin` in `daily_prices` table
- Rename `symbol` column to `isin` in `monthly_prices` table
- Update indexes to use `isin` instead of `symbol`

### Step 2: Update Code
After migration:
- Update `history_db.go` queries to use `isin` instead of `symbol`
- Remove TODO comments

## Notes

- This is a separate database (history.db) from the main universe.db and portfolio.db
- The functionality is correct - ISINs are already being stored and retrieved
- This is purely a naming consistency improvement
- Can be done as a separate migration after the main ISIN migration is complete

