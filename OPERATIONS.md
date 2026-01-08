# Sentinel Operations Guide

This guide provides operational procedures for running, monitoring, and maintaining the Sentinel portfolio management system.

## Table of Contents

1. [Daily Operations](#daily-operations)
2. [Database Backup](#database-backup)
3. [Data Integrity Verification](#data-integrity-verification)
4. [Performance Monitoring](#performance-monitoring)
5. [Troubleshooting](#troubleshooting)
6. [Emergency Procedures](#emergency-procedures)

---

## Daily Operations

### Starting the System

```bash
# On Arduino Uno Q
cd ~/sentinel
./sentinel
```

The system will:
- Initialize all 7 databases
- Start HTTP server on port 8080
- Begin background job scheduler
- Connect to Tradernet API (if configured)
- Start LED display updates

### Monitoring System Health

#### Via Web UI
1. Open `http://localhost:8080` in browser
2. Check "System Status" tab
3. Verify:
   - Tradernet connection: ✅ Connected
   - Market status: Shows current market hours
   - Recent jobs: All successful
   - Cash balances: No negative balances

#### Via Logs
```bash
# View real-time logs
tail -f ~/sentinel-data/logs/sentinel.log

# Check for errors
grep ERROR ~/sentinel-data/logs/sentinel.log | tail -20

# Check for warnings
grep WARN ~/sentinel-data/logs/sentinel.log | tail -20
```

### Normal Shutdown

```bash
# Send SIGTERM (Ctrl+C if running in foreground)
kill -TERM $(pgrep sentinel)

# System will:
# - Close all database connections gracefully
# - Complete running jobs
# - Flush WAL files
# - Exit cleanly
```

---

## Database Backup

### Automated Backup

Run the backup script daily (recommended via cron):

```bash
# Manual backup
~/sentinel/scripts/backup_databases.sh

# Setup daily backup (cron)
crontab -e
# Add line:
0 2 * * * ~/sentinel/scripts/backup_databases.sh
```

**Backup Storage:**
- Location: `~/sentinel-backups/backup_YYYYMMDD_HHMMSS/`
- Compressed archive: `backup_YYYYMMDD_HHMMSS.tar.gz`
- Retention: 7 days (automatically cleaned)

### Manual Database Backup

```bash
# Backup specific database
sqlite3 ~/sentinel-data/ledger.db ".backup '/path/to/backup/ledger_backup.db'"

# Verify backup integrity
sqlite3 /path/to/backup/ledger_backup.db "PRAGMA integrity_check;"
```

### Restore from Backup

```bash
# 1. Stop Sentinel
kill -TERM $(pgrep sentinel)

# 2. Backup current database (just in case)
cp ~/sentinel-data/ledger.db ~/sentinel-data/ledger.db.before_restore

# 3. Restore from backup
cp /path/to/backup/ledger_backup.db ~/sentinel-data/ledger.db

# 4. Verify integrity
sqlite3 ~/sentinel-data/ledger.db "PRAGMA integrity_check;"

# 5. Restart Sentinel
~/sentinel/sentinel
```

---

## Data Integrity Verification

Run integrity checks weekly:

```bash
# Check ledger database (immutable audit trail - most critical)
sqlite3 ~/sentinel-data/ledger.db < ~/sentinel/scripts/verify_data_integrity.sql

# Check portfolio database
sqlite3 ~/sentinel-data/portfolio.db < ~/sentinel/scripts/verify_data_integrity.sql

# Check universe database
sqlite3 ~/sentinel-data/universe.db < ~/sentinel/scripts/verify_data_integrity.sql
```

**Expected Output:**
- All checks should show "PASS"
- "WARNING" status requires investigation
- "FAIL" status requires immediate action

### Common Integrity Issues

#### Orphaned Records

**Symptom:** Script reports orphaned dividend_history or kelly_sizes

**Cause:** Records exist that reference deleted parent records

**Fix:**
```sql
-- Remove orphaned dividends (after backing up!)
DELETE FROM dividend_history
WHERE cash_flow_id NOT IN (SELECT id FROM cash_flows);

-- Remove orphaned kelly_sizes
DELETE FROM kelly_sizes
WHERE isin NOT IN (SELECT isin FROM scores);
```

#### Negative Cash Balances

**Symptom:** cash_balances table shows negative values

**Cause:** Trades executed without sufficient cash, sync issues

**Investigation:**
```sql
-- Find negative balances
SELECT * FROM cash_balances WHERE balance < 0;

-- Check recent cash flows
SELECT * FROM cash_flows
ORDER BY created_at DESC
LIMIT 20;
```

**Fix:** Depends on root cause. May require manual cash adjustment or trade reversal.

---

## Performance Monitoring

### Using the Performance Utility

The system includes built-in performance monitoring in `internal/utils/performance.go`.

#### Example Usage in Code:

```go
import "github.com/aristath/sentinel/internal/utils"

// Method 1: Manual timer
func MySlowOperation() {
    timer := utils.NewTimer("my_operation", log)
    defer timer.Stop()

    // ... do work ...
}

// Method 2: Defer-friendly
func AnotherOperation() {
    defer utils.OperationTimer("another_operation", log)()

    // ... do work ...
}

// Method 3: Database query timing
func QueryDatabase() {
    defer := utils.MeasureDBQuery("fetch_positions", log)
    rows, err := db.Query("SELECT * FROM positions")
    // ... process rows ...
    defer(rowsAffected)
}
```

### Performance Thresholds

**Automatic warnings are logged for:**
- Operations > 10 seconds: INFO log
- Operations > 30 seconds: WARN log
- Database queries > 5 seconds: WARN log

### Finding Slow Operations

```bash
# Find slow operations in logs
grep "Slow operation detected" ~/sentinel-data/logs/sentinel.log

# Find slow database queries
grep "Slow database query" ~/sentinel-data/logs/sentinel.log

# Get performance metrics summary
grep "Performance measurement" ~/sentinel-data/logs/sentinel.log | \
  jq -r '[.operation, .duration_seconds] | @tsv' | \
  sort -k2 -nr | \
  head -20
```

---

## Troubleshooting

### System Won't Start

#### Check 1: Database Files
```bash
# Verify databases exist
ls -lh ~/sentinel-data/*.db

# Check file permissions
chmod 644 ~/sentinel-data/*.db
```

#### Check 2: Port Already in Use
```bash
# Check if port 8080 is in use
lsof -i :8080

# Kill process if needed
kill $(lsof -t -i:8080)
```

#### Check 3: Database Corruption
```bash
# Check each database
for db in ~/sentinel-data/*.db; do
    echo "Checking $db..."
    sqlite3 "$db" "PRAGMA integrity_check;"
done
```

### Tradernet Connection Issues

#### Symptom: "Tradernet not connected" in logs

**Check 1: API Credentials**
```bash
# Via Settings UI
# Navigate to Settings → Credentials tab
# Verify API Key and Secret are configured
```

**Check 2: Network Connectivity**
```bash
# Test Tradernet API endpoint
curl -v https://tradernet.com/api/v1/ping
```

**Check 3: API Rate Limits**
```bash
# Check logs for rate limit errors
grep "rate limit" ~/sentinel-data/logs/sentinel.log -i
```

### Frontend Not Loading

#### Check 1: Frontend Build
```bash
cd ~/sentinel/frontend
npm run build

# Verify dist/ directory exists
ls -lh dist/
```

#### Check 2: Backend Serving Static Files
```bash
# Check if embedded assets are being served
curl -I http://localhost:8080/

# Should return 200 OK
```

### Database Lock Errors

#### Symptom: "database is locked" errors

**Cause:** Multiple processes accessing database simultaneously, or unclean shutdown

**Fix:**
```bash
# 1. Stop all Sentinel instances
pkill sentinel

# 2. Check for WAL files
ls ~/sentinel-data/*.db-wal

# 3. Checkpoint WAL files
for db in ~/sentinel-data/*.db; do
    sqlite3 "$db" "PRAGMA wal_checkpoint(TRUNCATE);"
done

# 4. Restart
~/sentinel/sentinel
```

---

## Emergency Procedures

### Rollback a Bad Trade

**IMPORTANT:** Only possible if trade hasn't been sent to broker yet.

```sql
-- 1. Check recent trades
SELECT * FROM trades
ORDER BY executed_at DESC
LIMIT 10;

-- 2. Find the trade ID
-- Note: Order ID from Tradernet

-- 3. If trade was logged but not executed, delete it
-- WARNING: This should only be done if trade failed at broker
DELETE FROM trades WHERE order_id = 'ORDER_ID_HERE';

-- 4. Verify deletion
SELECT COUNT(*) FROM trades WHERE order_id = 'ORDER_ID_HERE';
-- Should return 0
```

**Note:** This does NOT cancel orders at the broker. Use Tradernet UI for that.

### Recover from Corrupted Database

**Scenario:** integrity_check reports corruption

```bash
# 1. Stop Sentinel immediately
kill -9 $(pgrep sentinel)

# 2. Backup corrupted database
cp ~/sentinel-data/corrupted.db ~/sentinel-data/corrupted.db.CORRUPT

# 3. Attempt recovery
sqlite3 ~/sentinel-data/corrupted.db ".recover" | \
  sqlite3 ~/sentinel-data/corrupted.db.recovered

# 4. Verify recovered database
sqlite3 ~/sentinel-data/corrupted.db.recovered "PRAGMA integrity_check;"

# 5. If successful, replace
mv ~/sentinel-data/corrupted.db ~/sentinel-data/corrupted.db.CORRUPT.backup
mv ~/sentinel-data/corrupted.db.recovered ~/sentinel-data/corrupted.db

# 6. Restart
~/sentinel/sentinel
```

### System Not Responding

```bash
# 1. Check if process is running
ps aux | grep sentinel

# 2. Check CPU/memory usage
top -p $(pgrep sentinel)

# 3. Get process trace (to see what it's doing)
strace -p $(pgrep sentinel) -e trace=all

# 4. If hung, force kill and restart
kill -9 $(pgrep sentinel)
~/sentinel/sentinel

# 5. Check logs for panic/crash
tail -100 ~/sentinel-data/logs/sentinel.log
```

---

## Maintenance Schedule

### Daily
- ✅ Check system status via UI
- ✅ Verify no ERROR logs
- ✅ Check for negative cash balances

### Weekly
- ✅ Run data integrity verification
- ✅ Review performance logs
- ✅ Check disk space usage

### Monthly
- ✅ Review all trades for accuracy
- ✅ Verify portfolio vs broker positions
- ✅ Audit cash flow records
- ✅ Update dependencies (Go modules, npm packages)

### Quarterly
- ✅ Full database backup to external storage
- ✅ Review and optimize database indexes
- ✅ Performance testing
- ✅ Security audit (API keys rotation)

---

## Important File Locations

### Data
- **Databases:** `~/sentinel-data/*.db`
- **Logs:** `~/sentinel-data/logs/`
- **Backups:** `~/sentinel-backups/`

### Configuration
- **Settings:** Stored in `config.db` (access via UI)
- **Environment:** `~/.env` (deprecated, use UI)

### Documentation
- **Database Constraints:** `internal/database/schemas/CONSTRAINTS.md`
- **Verification Script:** `scripts/verify_data_integrity.sql`
- **Backup Script:** `scripts/backup_databases.sh`
- **Project Instructions:** `CLAUDE.md`

---

## Support

For issues not covered in this guide:

1. Check logs: `~/sentinel-data/logs/sentinel.log`
2. Review `CLAUDE.md` for architecture details
3. Check database constraints: `internal/database/schemas/CONSTRAINTS.md`
4. Review bug fix plan: `/Users/aristath/.claude/plans/frolicking-nibbling-graham.md`

**Remember:** This system manages real money. When in doubt, stop the system and investigate before taking action.

---

**Last Updated:** 2026-01-08
**Version:** 1.0 (After comprehensive bug fix session)
