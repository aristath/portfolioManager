# Schema Verification Report

This document verifies that the consolidated schema files accurately represent the final state after applying all migrations.

## Verification Method

For each database, I traced through all migrations that affect it and verified the consolidated schema matches the final state.

## Universe.db

### Migrations Applied:
1. **003** - Creates securities (symbol PK), country_groups, industry_groups
2. **014** - Removes bucket_id from securities
3. **028** - Creates tags and security_tags (with symbol FK)
4. **030** - Changes securities PK to isin, updates security_tags to use isin
5. **032** - Adds enhanced tags

### Verification:
✅ **securities table**:
- PK changed from symbol → isin (migration 030)
- No bucket_id column (removed in 014)
- Index: idx_securities_symbol exists (030 creates it, replaces idx_securities_isin)
- Index: idx_securities_isin does NOT exist (removed in 030 since isin is now PK)

✅ **country_groups table**: Matches migration 003 exactly

✅ **industry_groups table**: Matches migration 003 exactly

✅ **tags table**: Created in 028, matches

✅ **security_tags table**:
- PK changed from (symbol, tag_id) → (isin, tag_id) (migration 030)
- Foreign keys to securities(isin) and tags(id) (migration 030)
- Indexes: idx_security_tags_isin, idx_security_tags_tag_id (migration 030)
- Note: Migration 030 comment says "idx_security_tags_symbol" but actually creates index on isin

✅ **Enhanced tags**: All 20 tags from migration 032 are inserted

## Portfolio.db

### Migrations Applied:
1. **004** - Creates positions (symbol PK), scores (symbol PK), calculated_metrics, portfolio_snapshots
2. **016** - Removes bucket_id from positions
3. **022** - Removes portfolio_snapshots table
4. **029** - Adds score columns (sharpe_score, drawdown_score, etc.)
5. **033** - Changes scores PK from symbol → isin (NO foreign key)
6. **034** - Changes positions PK from symbol → isin (NO foreign key)

### Verification:
✅ **positions table**:
- PK changed from symbol → isin (migration 034)
- No bucket_id column (removed in 016)
- No foreign key (migration 034 doesn't include it, and SQLite doesn't support cross-db FKs)
- Symbol column kept for display/API conversion

✅ **scores table**:
- PK changed from symbol → isin (migration 033)
- All columns from 029 added (sharpe_score, drawdown_score, etc.)
- No foreign key (migration 033 doesn't include it)
- All indexes match

✅ **calculated_metrics table**: Matches migration 004 exactly

✅ **portfolio_snapshots table**: Removed (migration 022)

## Ledger.db

### Migrations Applied:
1. **010** - Creates trades, cash_flows (old schema), dividend_history, drip_tracking
2. **013** - Adds date column to cash_flows
3. **017** - Removes bucket_id from trades
4. **018** - Removes bucket_id from cash_flows
5. **019** - Removes bucket_id from dividend_history
6. **030** - Adds isin column to trades and dividend_history, creates indexes
7. **036** - Migrates cash_flows to new schema

### Verification:
✅ **trades table**:
- No bucket_id column (removed in 017)
- Has isin column (added in 030)
- Indexes: idx_trades_symbol, idx_trades_isin, idx_trades_executed (030 adds isin index)

✅ **cash_flows table**:
- Schema matches migration 036 exactly
- No bucket_id column (removed in 018)
- Has date column (added in 013, kept in 036)
- New schema: transaction_id, type_doc_id, transaction_type, etc.

✅ **dividend_history table**:
- No bucket_id column (removed in 019)
- Has isin column (added in 030)
- Indexes: idx_dividends_symbol, idx_dividends_isin, idx_dividends_payment_date

✅ **drip_tracking table**: Matches migration 010 exactly

## Config.db

### Migrations Applied:
1. **009** - Creates settings, allocation_targets
2. **015** - Creates planner_settings
3. **026** - Adds risk management columns to planner_settings
4. **027** - Adds optimizer_blend to planner_settings

### Verification:
✅ **settings table**: Matches migration 009 exactly

✅ **allocation_targets table**: Matches migration 009 exactly

✅ **planner_settings table**:
- All columns from 015 present
- Risk management columns from 026 present (min_hold_days, sell_cooldown_days, etc.)
- optimizer_blend from 027 present
- Default row inserted

## Agents.db

### Migrations Applied:
1. **001** - Creates sequences, evaluations, best_result, planner_configs, planner_config_history
2. **005** - Recreates sequences, evaluations, best_result (updates best_result schema)
3. **021** - Removes planner_configs and planner_config_history

### Verification:
✅ **sequences table**: Matches migration 005 exactly

✅ **evaluations table**: Matches migration 005 exactly

✅ **best_result table**:
- Schema updated from 001 (best_sequence_hash, best_score) → 005 (sequence_hash, plan_data, score)
- Matches migration 005 exactly

✅ **planner_configs table**: Removed (migration 021)

✅ **planner_config_history table**: Removed (migration 021)

## History.db

### Migrations Applied:
1. **006** - Creates daily_prices, exchange_rates, monthly_prices

### Verification:
✅ **daily_prices table**: Matches migration 006 exactly

✅ **exchange_rates table**: Matches migration 006 exactly

✅ **monthly_prices table**: Matches migration 006 exactly

## Cache.db

### Migrations Applied:
1. **002** - Creates recommendations (old schema)
2. **008** - Creates recommendations (new schema) and cache_data
3. **035** - Migrates recommendations to final schema

### Verification:
✅ **recommendations table**:
- Schema matches migration 035 exactly
- All columns present: uuid, symbol, name, side, quantity, etc.
- All indexes match

✅ **cache_data table**: Matches migration 008 exactly

## Issues Found and Fixed

1. **Portfolio schema** - Had incorrect FOREIGN KEY constraints that:
   - Don't exist in migrations 033 and 034
   - Can't work in SQLite (cross-database foreign keys not supported)
   - **FIXED**: Removed foreign keys from positions and scores tables

## Summary

✅ All schema files accurately represent the final state after all migrations
✅ All tables, columns, indexes, and constraints match the migrations
✅ One issue found and fixed (foreign keys in portfolio schema)

The consolidated schemas are accurate and ready for use.
