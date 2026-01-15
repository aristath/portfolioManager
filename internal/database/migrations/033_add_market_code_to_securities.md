# Add Market Code to Securities Migration

## Overview

This migration adds the `market_code` column to the `securities` table. The market code is the Tradernet exchange/market identifier (e.g., "FIX" for US markets, "EU" for European markets, "HKEX" for Hong Kong) used for per-region market regime detection.

## Purpose

Market codes enable:
1. **Per-region regime detection**: Securities are evaluated against their own region's market conditions
2. **Region mapping**: Map securities to regions (US, EU, ASIA, RUSSIA, MIDDLE_EAST, CENTRAL_ASIA)
3. **Global average fallback**: Regions without indices use weighted average of regions with indices

## Market Code to Region Mapping

| Market Code | Region | Markets |
|-------------|--------|---------|
| FIX | US | NYSE, NASDAQ |
| EU | EU | European exchanges |
| ATHEX | EU | Athens Stock Exchange |
| HKEX | ASIA | Hong Kong |
| HKG | ASIA | Hong Kong (alternative) |
| FORTS | RUSSIA | Moscow Exchange derivatives |
| MCX | RUSSIA | Moscow Exchange |
| TABADUL | MIDDLE_EAST | Saudi Arabia |
| KASE | CENTRAL_ASIA | Kazakhstan |

## Code Changes (Already Applied)

The code has been updated to:
1. Add `MarketCode` field to `Security` and `SecurityWithScore` models
2. Handle `market_code` column in all repository queries (Create, GetByISIN, Update, GetAllActive)
3. Add `GetByMarketCode()` repository method
4. Enrich `MarketCode` from broker API via `MetadataEnricher`
5. Add index for efficient market code queries
6. Create region mapper (`internal/market_regime/region_mapper.go`)

## Migration SQL (For Existing Databases)

```sql
-- Add market_code column
ALTER TABLE securities ADD COLUMN market_code TEXT;

-- Create index for efficient queries
CREATE INDEX IF NOT EXISTS idx_securities_market_code ON securities(market_code);
```

## Backfilling Existing Data

Existing securities will have `market_code = NULL` after migration. The value will be populated automatically when:
1. **On next sync**: `MetadataEnricher.Enrich()` will fetch and store the market code
2. **On universe sync**: `SyncService` processes securities through enricher

Alternatively, for immediate backfill based on symbol suffix (less reliable than API):

```sql
-- Optional: Derive market code from symbol suffix (fallback only)
UPDATE securities SET market_code = 'FIX' WHERE symbol LIKE '%.US' AND market_code IS NULL;
UPDATE securities SET market_code = 'EU' WHERE symbol LIKE '%.EU' AND market_code IS NULL;
UPDATE securities SET market_code = 'ATHEX' WHERE symbol LIKE '%.GR' AND market_code IS NULL;
UPDATE securities SET market_code = 'HKEX' WHERE symbol LIKE '%.AS' AND market_code IS NULL;
```

**Note**: Symbol-based derivation is a fallback only. The authoritative source is the Tradernet API's `mkt` field, which is populated automatically during sync.

## Impact

- **New installations**: Will have clean schema with `market_code` column
- **Existing installations**: Column is added, values populated on next sync
- **Immediate operation**: System works correctly with NULL market_code values (uses global average)

## Verification

After migration and sync:

```sql
-- Check market code distribution
SELECT market_code, COUNT(*) FROM securities WHERE active = 1 GROUP BY market_code;

-- Verify no active securities without market code (after sync completes)
SELECT COUNT(*) FROM securities WHERE active = 1 AND market_code IS NULL;
```
