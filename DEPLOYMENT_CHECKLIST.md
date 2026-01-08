# Deployment Checklist - January 8, 2026 Updates

This checklist guides you through deploying the comprehensive bug fixes and improvements from the 2026-01-08 session.

## Pre-Deployment

### 1. Review Changes
- [x] Read `IMPROVEMENTS_2026-01-08.md` for complete list of changes
- [x] All builds passing (backend + frontend)
- [x] Documentation reviewed

### 2. Backup Current System
```bash
# Run backup script
~/sentinel/scripts/backup_databases.sh

# Verify backup created
ls -lh ~/sentinel-backups/backup_*/ | tail -1
```

### 3. Verify Current State
```bash
# Check if Sentinel is running
pgrep sentinel

# Check current logs for any errors
grep ERROR ~/sentinel-data/logs/sentinel.log | tail -20
```

## Deployment Steps

### 1. Stop Sentinel
```bash
# Graceful shutdown
kill -TERM $(pgrep sentinel)

# Wait for shutdown
sleep 5

# Verify stopped
pgrep sentinel || echo "Sentinel stopped successfully"
```

### 2. Deploy New Code
```bash
cd ~/sentinel

# If deploying from git
git pull origin main

# If deploying pre-built binary
# (Copy sentinel-arm64 to ~/sentinel/sentinel)

# Rebuild frontend (if needed)
cd frontend
npm install
npm run build
cd ..
```

### 3. Apply Schema Updates
```bash
# Run schema update script
~/sentinel/scripts/apply_schema_updates.sh

# This will:
# - Check Sentinel is stopped
# - Verify backup exists
# - Check schema constraints
# - Run integrity checks
```

### 4. Start Sentinel
```bash
# Start in background
nohup ~/sentinel/sentinel > ~/sentinel-data/logs/startup.log 2>&1 &

# Or start in foreground for monitoring
~/sentinel/sentinel
```

### 5. Verify System Health
```bash
# Check process is running
pgrep sentinel

# Monitor startup logs
tail -f ~/sentinel-data/logs/sentinel.log

# Look for:
# - "HTTP server listening on :8080"
# - No ERROR or FATAL messages
# - All 7 databases initialized successfully
```

## Post-Deployment Verification

### 1. Web UI Check
```bash
# Open in browser
# http://localhost:8080

# Verify:
# - UI loads without errors
# - Settings tab accessible
# - Portfolio data displays
# - System Status shows green
```

### 2. EventSource Connection
```bash
# In browser DevTools (F12):
# - Network tab → Filter: EventStream
# - Should see ONE connection to /api/events
# - Connection should stay open (not reconnecting)
# - No errors in console
```

### 3. Database Integrity
```bash
# Run full integrity verification
sqlite3 ~/sentinel-data/ledger.db < ~/sentinel/scripts/verify_data_integrity.sql

# Expected: All checks should show "PASS"
# Any "WARNING" or "FAIL" requires investigation
```

### 4. Performance Check
```bash
# Monitor logs for slow operations
grep "Slow operation detected" ~/sentinel-data/logs/sentinel.log

# Monitor logs for slow queries
grep "Slow database query" ~/sentinel-data/logs/sentinel.log

# Should see minimal or no slow operation warnings
```

### 5. Error Monitoring
```bash
# Check for any new errors (last 30 minutes)
grep ERROR ~/sentinel-data/logs/sentinel.log | grep "$(date +%Y-%m-%d)" | tail -50

# Check for panics
grep "panic" ~/sentinel-data/logs/sentinel.log | grep "$(date +%Y-%m-%d)"

# Should see no panics, minimal errors
```

## 48-Hour Monitoring Period

### Daily Checks (Next 2 Days)

#### Morning Check
- [ ] System still running: `pgrep sentinel`
- [ ] No errors in last 24h: `grep ERROR ~/sentinel-data/logs/sentinel.log | grep "$(date +%Y-%m-%d)"`
- [ ] UI accessible: Open http://localhost:8080
- [ ] Portfolio values updating

#### Evening Check
- [ ] Memory stable: `ps aux | grep sentinel`
- [ ] No reconnection storms in browser DevTools
- [ ] Data integrity: Run spot checks on critical tables
- [ ] Logs clean: No unexpected warnings

### What to Watch For

**Critical Issues (Act Immediately):**
- Sentinel process crashes/restarts
- Database corruption errors
- Panic messages in logs
- UI completely inaccessible

**High Priority Issues (Investigate Within Hours):**
- Frequent ERROR messages
- EventSource reconnection loops
- Slow operation warnings (>30s)
- Negative cash balances

**Medium Priority Issues (Monitor):**
- Occasional warnings
- Slow queries (>5s)
- Stale price data warnings

## Rollback Procedure

If critical issues occur:

### 1. Stop New Version
```bash
kill -TERM $(pgrep sentinel)
```

### 2. Restore Previous Code
```bash
cd ~/sentinel
git checkout HEAD~1  # Or restore previous binary
```

### 3. Restore Databases (If Needed)
```bash
# Find latest backup
BACKUP=$(find ~/sentinel-backups -name "backup_*" -type d | sort -r | head -1)

# Stop Sentinel
kill -TERM $(pgrep sentinel)

# Restore databases
cp "$BACKUP"/*.db ~/sentinel-data/

# Verify integrity
for db in ~/sentinel-data/*.db; do
    sqlite3 "$db" "PRAGMA integrity_check;"
done
```

### 4. Restart Previous Version
```bash
~/sentinel/sentinel
```

### 5. Report Issue
Document what went wrong:
- What errors occurred
- When they started
- What triggered them
- Logs from the incident

## Post-Deployment Tasks

### Week 1
- [ ] Set up daily backups via cron
  ```bash
  crontab -e
  # Add: 0 2 * * * ~/sentinel/scripts/backup_databases.sh
  ```
- [ ] Run weekly integrity check
  ```bash
  # Add to cron: 0 3 * * 0 sqlite3 ~/sentinel-data/ledger.db < ~/sentinel/scripts/verify_data_integrity.sql > ~/sentinel-data/logs/integrity_check.log
  ```

### Month 1
- [ ] Performance baseline measurements
- [ ] Audit all trades vs broker
- [ ] Test recovery procedures (on backup copy)
- [ ] Review and optimize any slow operations found

## Success Criteria

After 48 hours, all should be true:

- [ ] Zero panics in logs
- [ ] Zero orphaned records in databases
- [ ] EventSource connection stable (no leaks)
- [ ] UI responsive with optimistic updates
- [ ] All API calls complete within 30s
- [ ] No database queries over 5s
- [ ] Memory usage stable (<100MB growth/day)
- [ ] All integrity checks passing

## Notes

### Changes Summary
- **27 improvements** implemented (18 bug fixes + 9 hardening)
- **Backend:** Type safety, error handling, validation, performance monitoring
- **Database:** Foreign key constraints, validation
- **Frontend:** Memory leak fixes, race condition fixes, optimistic updates, error notifications
- **Operations:** Backup automation, integrity verification, comprehensive documentation

### Risk Level
**LOW** - All changes are defensive and backward-compatible. No breaking changes to data structures or APIs.

### Recommended Deployment Time
- **Best:** During market close (after 4 PM ET, before 9:30 AM ET next day)
- **Avoid:** During market hours or active trading
- **Reason:** Minimizes impact if restart is needed

---

**Document Version:** 1.0
**Date:** 2026-01-08
**Status:** Ready for deployment
