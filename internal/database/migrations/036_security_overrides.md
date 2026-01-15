# Migration 036: Security Overrides System

## Purpose
Implement a clean separation between Tradernet-provided data (defaults) and user customizations (overrides). This enables:
1. Storing Tradernet metadata (geography, industry) as defaults
2. Allowing users to override any field without losing the original data
3. Easy "reset to default" functionality
4. Cleaner data architecture with no mixed concerns

## Changes

### New Table: security_overrides
```sql
CREATE TABLE IF NOT EXISTS security_overrides (
    isin TEXT NOT NULL,
    field TEXT NOT NULL,               -- Field name (e.g., 'allow_buy', 'geography', 'min_lot')
    value TEXT NOT NULL,               -- Value as string (converted to appropriate type at read time)
    created_at INTEGER NOT NULL,       -- Unix timestamp
    updated_at INTEGER NOT NULL,       -- Unix timestamp
    PRIMARY KEY (isin, field),
    FOREIGN KEY (isin) REFERENCES securities(isin) ON DELETE CASCADE
) STRICT;

CREATE INDEX IF NOT EXISTS idx_security_overrides_isin ON security_overrides(isin);
```

### Columns Removed from securities
The following columns are removed from the securities table. Their values are now:
- Stored in `security_overrides` when user changes them
- Default values applied at read time when no override exists

| Column | Default Value | Notes |
|--------|--------------|-------|
| allow_buy | true | User-configurable |
| allow_sell | true | User-configurable |
| min_lot | 1 | User-configurable |
| priority_multiplier | 1.0 | User-configurable |

### Overridable Fields
Any field in the Security struct can be overridden via the EAV pattern:
- `geography` - Override Tradernet's issuer_country_code
- `industry` - Override Tradernet's sector_code mapping
- `product_type` - Override auto-detected type
- `name` - Override Tradernet's company name
- `min_lot` - Set minimum lot size
- `allow_buy` - Enable/disable buying
- `allow_sell` - Enable/disable selling
- `priority_multiplier` - Set priority weight

## Data Migration

### Step 1: Create security_overrides table
```sql
CREATE TABLE IF NOT EXISTS security_overrides (
    isin TEXT NOT NULL,
    field TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY (isin, field),
    FOREIGN KEY (isin) REFERENCES securities(isin) ON DELETE CASCADE
) STRICT;

CREATE INDEX IF NOT EXISTS idx_security_overrides_isin ON security_overrides(isin);
```

### Step 2: Migrate non-default values to overrides
```sql
-- Migrate allow_buy = false (default is true)
INSERT INTO security_overrides (isin, field, value, created_at, updated_at)
SELECT isin, 'allow_buy', 'false', strftime('%s', 'now'), strftime('%s', 'now')
FROM securities WHERE allow_buy = 0;

-- Migrate allow_sell = false (default is true)
INSERT INTO security_overrides (isin, field, value, created_at, updated_at)
SELECT isin, 'allow_sell', 'false', strftime('%s', 'now'), strftime('%s', 'now')
FROM securities WHERE allow_sell = 0;

-- Migrate min_lot != 1 (default is 1)
INSERT INTO security_overrides (isin, field, value, created_at, updated_at)
SELECT isin, 'min_lot', CAST(min_lot AS TEXT), strftime('%s', 'now'), strftime('%s', 'now')
FROM securities WHERE min_lot != 1;

-- Migrate priority_multiplier != 1.0 (default is 1.0)
INSERT INTO security_overrides (isin, field, value, created_at, updated_at)
SELECT isin, 'priority_multiplier', CAST(priority_multiplier AS TEXT), strftime('%s', 'now'), strftime('%s', 'now')
FROM securities WHERE priority_multiplier != 1.0;
```

### Step 3: Drop columns from securities
SQLite doesn't support DROP COLUMN directly. Options:
1. Create new table without columns, copy data, drop old, rename new
2. Leave columns (they'll be ignored by code)

For simplicity, we use option 2 initially - the columns remain but are ignored.
The schema file reflects the target state without these columns.

## Code Changes
1. `override_repository.go` - New repository for CRUD on overrides
2. `security_repository.go` - Merge overrides when reading securities
3. `security_setup_service.go` - Remove allow_buy/allow_sell/min_lot/priority_multiplier from Create
4. `handlers.go` - Add override endpoints, update PUT to write to overrides
5. DI wiring - Wire OverrideRepository
6. Frontend - Show override indicators, reset functionality

## Verification
```bash
# Check schema
sqlite3 data/universe.db ".schema security_overrides"

# Check existing overrides (should be empty initially)
sqlite3 data/universe.db "SELECT COUNT(*) FROM security_overrides;"

# After editing a security via UI, verify override created
sqlite3 data/universe.db "SELECT * FROM security_overrides;"
```

## Rollback
To rollback:
1. Re-add columns to securities table
2. Copy values from security_overrides back to columns
3. Drop security_overrides table
