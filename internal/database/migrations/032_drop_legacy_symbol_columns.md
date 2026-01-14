# Drop Legacy Symbol Columns Migration

## Overview

This migration removes the deprecated `yahoo_symbol` and `alphavantage_symbol` columns from the `securities` table. These columns are no longer used - client-specific symbols are now stored in the `client_symbols` table.

## Code Changes (Already Applied)

The code has been updated to:
1. Not read or write these columns in `security_repository.go`
2. Remove these fields from `Security` model struct
3. Remove these fields from API responses
4. Use `client_symbols` table for all client-specific symbol mappings

## Migration SQL (For Existing Databases)

SQLite 3.35.0+ supports `ALTER TABLE DROP COLUMN`. Run these commands to clean up existing databases:

```sql
-- Drop legacy columns from securities table
ALTER TABLE securities DROP COLUMN yahoo_symbol;
ALTER TABLE securities DROP COLUMN alphavantage_symbol;
```

## Alternative (SQLite < 3.35.0)

For older SQLite versions, recreate the table:

```sql
-- Start transaction
BEGIN TRANSACTION;

-- Create new table without legacy columns
CREATE TABLE securities_new (
    isin TEXT PRIMARY KEY,
    symbol TEXT NOT NULL,
    name TEXT NOT NULL,
    product_type TEXT,
    industry TEXT,
    country TEXT,
    fullExchangeName TEXT,
    priority_multiplier REAL DEFAULT 1.0,
    min_lot INTEGER DEFAULT 1,
    active INTEGER DEFAULT 1,
    allow_buy INTEGER DEFAULT 1,
    allow_sell INTEGER DEFAULT 1,
    currency TEXT,
    last_synced INTEGER,
    min_portfolio_target REAL,
    max_portfolio_target REAL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
) STRICT;

-- Copy data (excluding legacy columns)
INSERT INTO securities_new
SELECT isin, symbol, name, product_type, industry, country, fullExchangeName,
       priority_multiplier, min_lot, active, allow_buy, allow_sell, currency,
       last_synced, min_portfolio_target, max_portfolio_target, created_at, updated_at
FROM securities;

-- Drop old table
DROP TABLE securities;

-- Rename new table
ALTER TABLE securities_new RENAME TO securities;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_securities_active ON securities(active);
CREATE INDEX IF NOT EXISTS idx_securities_country ON securities(country);
CREATE INDEX IF NOT EXISTS idx_securities_industry ON securities(industry);
CREATE INDEX IF NOT EXISTS idx_securities_symbol ON securities(symbol);

-- Commit
COMMIT;
```

## Impact

- **New installations**: Will have clean schema without legacy columns
- **Existing installations**: Legacy columns remain but are unused (ignored by code)
- **Optional cleanup**: Run migration SQL above to remove unused columns and reclaim space

## Notes

- This is a cleanup migration, not a functional requirement
- The system works correctly with or without these columns present
- Columns are simply ignored if present in existing databases
