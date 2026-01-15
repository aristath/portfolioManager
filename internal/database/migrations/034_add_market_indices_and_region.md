# Add Market Indices Table and Region to Regime History

## Overview

This migration adds:
1. `region` column to `market_regime_history` table for per-region regime scores
2. `market_indices` table for storing index configuration

## Purpose

Enables per-region market regime detection:
- **US indices** (SP500, NASDAQ, DJI30, RUT, NQX) calculate US regime score
- **EU indices** (DAX, FTSE, CAC40, etc.) calculate EU regime score
- **Asia indices** (HSI) calculate Asia regime score
- VIX.IDX is classified as VOLATILITY and excluded from price composite

## Migration SQL (For Existing Databases)

### Step 1: Add region column to market_regime_history

```sql
-- Add region column with GLOBAL default for backwards compatibility
ALTER TABLE market_regime_history ADD COLUMN region TEXT NOT NULL DEFAULT 'GLOBAL';

-- Create index for efficient region queries
CREATE INDEX IF NOT EXISTS idx_regime_history_region ON market_regime_history(region);
```

### Step 2: Create market_indices table

```sql
CREATE TABLE IF NOT EXISTS market_indices (
    symbol TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    market_code TEXT NOT NULL,
    region TEXT NOT NULL,
    index_type TEXT NOT NULL DEFAULT 'PRICE',
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS idx_market_indices_region ON market_indices(region);
CREATE INDEX IF NOT EXISTS idx_market_indices_enabled ON market_indices(enabled);
CREATE INDEX IF NOT EXISTS idx_market_indices_type ON market_indices(index_type);
```

### Step 3: Seed known indices

```sql
-- US Indices (PRICE type)
INSERT OR IGNORE INTO market_indices (symbol, name, market_code, region, index_type, enabled, created_at, updated_at)
VALUES
    ('SP500.IDX', 'S&P 500', 'FIX', 'US', 'PRICE', 1, strftime('%s','now'), strftime('%s','now')),
    ('NASDAQ.IDX', 'NASDAQ Composite', 'FIX', 'US', 'PRICE', 1, strftime('%s','now'), strftime('%s','now')),
    ('DJI30.IDX', 'Dow Jones Industrial Average', 'FIX', 'US', 'PRICE', 1, strftime('%s','now'), strftime('%s','now')),
    ('RUT.IDX', 'Russell 2000', 'FIX', 'US', 'PRICE', 1, strftime('%s','now'), strftime('%s','now')),
    ('NQX.IDX', 'NASDAQ 100', 'FIX', 'US', 'PRICE', 1, strftime('%s','now'), strftime('%s','now'));

-- US Volatility Index (VOLATILITY type - excluded from composite)
INSERT OR IGNORE INTO market_indices (symbol, name, market_code, region, index_type, enabled, created_at, updated_at)
VALUES
    ('VIX.IDX', 'CBOE Volatility Index', 'FIX', 'US', 'VOLATILITY', 1, strftime('%s','now'), strftime('%s','now'));

-- EU Indices (PRICE type)
INSERT OR IGNORE INTO market_indices (symbol, name, market_code, region, index_type, enabled, created_at, updated_at)
VALUES
    ('DAX.IDX', 'DAX (Germany)', 'EU', 'EU', 'PRICE', 1, strftime('%s','now'), strftime('%s','now')),
    ('FTSE.IDX', 'FTSE 100 (UK)', 'EU', 'EU', 'PRICE', 1, strftime('%s','now'), strftime('%s','now')),
    ('FCHI.IDX', 'CAC 40 (France)', 'EU', 'EU', 'PRICE', 1, strftime('%s','now'), strftime('%s','now')),
    ('FTMIB.IDX', 'FTSE MIB (Italy)', 'EU', 'EU', 'PRICE', 1, strftime('%s','now'), strftime('%s','now')),
    ('IBEX.IDX', 'IBEX 35 (Spain)', 'EU', 'EU', 'PRICE', 1, strftime('%s','now'), strftime('%s','now')),
    ('OMXS30.IDX', 'OMX Stockholm 30 (Sweden)', 'EU', 'EU', 'PRICE', 1, strftime('%s','now'), strftime('%s','now'));

-- Asia Indices (PRICE type)
INSERT OR IGNORE INTO market_indices (symbol, name, market_code, region, index_type, enabled, created_at, updated_at)
VALUES
    ('HSI.IDX', 'Hang Seng Index', 'HKEX', 'ASIA', 'PRICE', 1, strftime('%s','now'), strftime('%s','now'));
```

## Impact

- **New installations**: Tables created automatically with correct schema
- **Existing installations**: Run migration SQL to add new column and table
- **Backwards compatibility**: Existing regime history records have `region='GLOBAL'`

## Verification

```sql
-- Verify region column exists
SELECT region, COUNT(*) FROM market_regime_history GROUP BY region;

-- Verify indices are seeded
SELECT region, index_type, COUNT(*) FROM market_indices GROUP BY region, index_type;
-- Expected: US PRICE=5, US VOLATILITY=1, EU PRICE=6, ASIA PRICE=1
```
