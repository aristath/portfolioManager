# Migration Guide: Stock → Security Refactoring

This guide covers the database migration from `stocks` table to `securities` table with product type support.

## Overview

**Migration versions:**
- v9: Add `product_type` column to `stocks` table
- v10: Rename `stocks` table to `securities`

**Important:** These migrations are **forward-only** and cannot be rolled back automatically. Follow backup procedures below.

## Prerequisites

Before running migrations:

1. **Stop the application** on the Arduino device
2. **Create database backups** (see Backup Procedure below)
3. **Test migrations locally** before deploying to production
4. **Ensure sufficient disk space** for database copies

## Backup Procedure

### Critical Files to Backup

```bash
# On Arduino device, backup all databases
ssh arduino@192.168.1.11
cd /home/arduino/arduino-trader/data

# Create timestamped backups
timestamp=$(date +%Y%m%d_%H%M%S)
mkdir -p backups/$timestamp

cp config.db backups/$timestamp/
cp trades.db backups/$timestamp/
cp history.db backups/$timestamp/

# Verify backups
ls -lah backups/$timestamp/
```

### Backup from Development Machine

```bash
# Using sshpass (credentials stored in .env.aristath)
SSHPASS="aristath" sshpass -e ssh arduino@192.168.1.11 \
  "cd /home/arduino/arduino-trader/data && tar czf backup.tar.gz *.db"

SSHPASS="aristath" sshpass -e scp arduino@192.168.1.11:/home/arduino/arduino-trader/data/backup.tar.gz \
  ./backups/arduino_db_backup_$(date +%Y%m%d_%H%M%S).tar.gz
```

## Migration Process

### Step 1: Deploy Code

```bash
# From development machine
git push origin support-multiple-security-types

# On Arduino device
ssh arduino@192.168.1.11
cd /home/arduino/arduino-trader
git pull origin support-multiple-security-types

# Install dependencies if needed
source venv/bin/activate
pip install -r requirements.txt
```

### Step 2: Run Database Migrations

Migrations run automatically on application startup via `DatabaseManager.init()`.

```bash
# On Arduino device
cd /home/arduino/arduino-trader

# Check current migration version
sqlite3 data/config.db "SELECT version FROM schema_version"

# Start application (migrations run automatically)
sudo systemctl restart arduino-trader

# Monitor logs for migration success
sudo journalctl -u arduino-trader -n 50 --no-pager

# Verify final version
sqlite3 data/config.db "SELECT version FROM schema_version"
# Expected: 10
```

**Migration v9 → v10 performs:**
1. Creates new `securities` table with product_type column
2. Copies all data: `INSERT INTO securities SELECT * FROM stocks`
3. Recreates indexes on securities table
4. Drops old stocks table
5. Updates schema_version to 10

### Step 3: Backfill Product Types

After migration v9 adds the `product_type` column, run the backfill script:

```bash
# On Arduino device
cd /home/arduino/arduino-trader
source venv/bin/activate

# Dry run to see what would be updated
python scripts/backfill_product_types.py --dry-run

# Review output, then run actual backfill
python scripts/backfill_product_types.py

# Review manual review list
# For any securities needing manual classification:
curl -X PUT "http://localhost:8000/api/securities/ALUM.EU" \
     -H "Content-Type: application/json" \
     -d '{"product_type": "ETC"}'
```

### Step 4: Verify Migration

```bash
# Check table structure
sqlite3 data/config.db ".schema securities"

# Count records
sqlite3 data/config.db "SELECT COUNT(*) FROM securities"

# Check product_type distribution
sqlite3 data/config.db "SELECT product_type, COUNT(*) FROM securities GROUP BY product_type"

# Verify no NULL product_types for active securities
sqlite3 data/config.db "SELECT symbol, product_type FROM securities WHERE active=1 AND product_type IS NULL"
# Should return 0 rows

# Verify indexes exist
sqlite3 data/config.db ".indexes securities"
# Expected: idx_securities_active, idx_securities_country, idx_securities_isin
```

### Step 5: Test Application

```bash
# Test API endpoints
curl http://localhost:8000/health
curl http://localhost:8000/api/securities
curl http://localhost:8000/api/portfolio/summary

# Check LED display shows status
# Monitor logs for errors
sudo journalctl -u arduino-trader -f
```

## Rollback Procedure

**IMPORTANT:** Migrations v9 and v10 are **forward-only**. There is no automatic rollback.

If migration fails or data corruption occurs:

### Option 1: Restore from Backup (Recommended)

```bash
# Stop application
sudo systemctl stop arduino-trader

# Restore backup
cd /home/arduino/arduino-trader/data
timestamp="20250131_120000"  # Use your backup timestamp

cp backups/$timestamp/config.db config.db
cp backups/$timestamp/trades.db trades.db
cp backups/$timestamp/history.db history.db

# Revert code
git reset --hard origin/main

# Restart application
sudo systemctl start arduino-trader
```

### Option 2: Manual Rollback (Advanced)

If you need to manually rollback after v10 migration:

```bash
# This requires recreating the stocks table from securities
sqlite3 data/config.db <<EOF
-- Create stocks table (original schema without product_type)
CREATE TABLE IF NOT EXISTS stocks (
    symbol TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    country TEXT,
    fullExchangeName TEXT,
    yahoo_symbol TEXT,
    isin TEXT,
    industry TEXT,
    priority_multiplier REAL DEFAULT 1.0,
    min_lot INTEGER DEFAULT 1,
    active INTEGER DEFAULT 1,
    allow_buy INTEGER DEFAULT 1,
    allow_sell INTEGER DEFAULT 0,
    currency TEXT,
    last_synced TEXT,
    min_portfolio_target REAL,
    max_portfolio_target REAL
);

-- Copy data (excluding product_type column)
INSERT INTO stocks (
    symbol, name, country, fullExchangeName, yahoo_symbol, isin, industry,
    priority_multiplier, min_lot, active, allow_buy, allow_sell, currency,
    last_synced, min_portfolio_target, max_portfolio_target
)
SELECT
    symbol, name, country, fullExchangeName, yahoo_symbol, isin, industry,
    priority_multiplier, min_lot, active, allow_buy, allow_sell, currency,
    last_synced, min_portfolio_target, max_portfolio_target
FROM securities;

-- Recreate indexes
CREATE INDEX idx_stocks_active ON stocks(active);
CREATE INDEX idx_stocks_country ON stocks(country);
CREATE INDEX idx_stocks_isin ON stocks(isin);

-- Drop securities table
DROP TABLE securities;

-- Revert schema version
UPDATE schema_version SET version = 8;
EOF

# Revert code to before refactoring
git reset --hard <commit_before_refactoring>

# Restart application
sudo systemctl restart arduino-trader
```

## Troubleshooting

### Migration Hangs or Times Out

```bash
# Check database locks
sqlite3 data/config.db "PRAGMA database_list"

# Check application processes
ps aux | grep python | grep arduino-trader

# If needed, stop all processes and retry
sudo systemctl stop arduino-trader
pkill -f "arduino-trader"

# Verify no lock files
ls -la data/*.db-*
```

### Data Integrity Errors

```bash
# Run integrity check
sqlite3 data/config.db "PRAGMA integrity_check"

# Verify foreign key constraints
sqlite3 data/config.db "PRAGMA foreign_key_check"

# If corruption detected, restore from backup
```

### Product Type Validation Errors

If application fails to start due to active securities without product_type:

```bash
# Temporarily mark securities as inactive
sqlite3 data/config.db "UPDATE securities SET active=0 WHERE product_type IS NULL"

# Or set a default product_type
sqlite3 data/config.db "UPDATE securities SET product_type='UNKNOWN' WHERE product_type IS NULL"

# Then run backfill script and manually classify
```

## Post-Migration Checklist

- [ ] All database backups created
- [ ] Migration v9 completed (product_type column added)
- [ ] Migration v10 completed (table renamed to securities)
- [ ] Backfill script executed successfully
- [ ] All active securities have known product_type (not NULL or UNKNOWN)
- [ ] API endpoints responding correctly
- [ ] LED display showing status
- [ ] Application logs show no errors
- [ ] Portfolio calculations working
- [ ] Scoring service functioning
- [ ] Trading functionality operational

## Support

If you encounter issues:

1. Check application logs: `sudo journalctl -u arduino-trader -n 100`
2. Verify database integrity: `sqlite3 data/config.db "PRAGMA integrity_check"`
3. Review this migration guide
4. Restore from backup if necessary
5. Report issues at https://github.com/aristath/portfolioManager/issues

## Migration History

- **v8 → v9**: Add product_type column to stocks table
- **v9 → v10**: Rename stocks table to securities

Previous migrations:
- v7 → v8: Add isin column and index to stocks table
- v5 → v6: Create allocation_targets table
- v4 → v5: Add country index to stocks table
