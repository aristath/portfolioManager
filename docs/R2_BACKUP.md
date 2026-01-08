# Cloudflare R2 Backup System

## Overview

Sentinel includes automatic cloud backup to Cloudflare R2, providing off-site disaster recovery for all 7 SQLite databases. This complements the existing local backup system with cloud storage for enhanced data safety.

## Features

- **Automatic Backups**: Schedule daily, weekly, or monthly backups to R2
- **Manual Backups**: Trigger immediate backup via UI or API
- **Two-Phase Restore**: Safe restore process with automatic service restart
- **Safety Backups**: Pre-restore backup created automatically
- **Automatic Rotation**: Old backups deleted based on retention policy
- **Integrity Validation**: SHA256 checksums and SQLite integrity checks
- **Compressed Archives**: All 7 databases in single `.tar.gz` file

## Configuration

### 1. Create Cloudflare R2 Bucket

1. Log in to Cloudflare dashboard
2. Navigate to R2 Object Storage
3. Create a new bucket (e.g., `sentinel-backups`)
4. Generate API credentials (Account ID, Access Key ID, Secret Access Key)

### 2. Configure in Sentinel

**Via Settings UI** (Recommended):
1. Open Settings → Backup tab
2. Enable R2 backups
3. Enter R2 credentials:
   - Account ID
   - Access Key ID
   - Secret Access Key
   - Bucket Name
4. Select backup schedule (daily/weekly/monthly)
5. Set retention days (0 = forever, default = 90)
6. Click "Test Connection" to verify
7. Click "Backup Now" to create first backup

**Via Settings API**:
```bash
curl -X PUT http://localhost:8080/api/settings/r2_backup_enabled \
  -H "Content-Type: application/json" \
  -d '{"value": 1}'

curl -X PUT http://localhost:8080/api/settings/r2_account_id \
  -H "Content-Type: application/json" \
  -d '{"value": "your_account_id"}'

# ... (repeat for other settings)
```

## Backup Schedule

Backups run according to your configured schedule:

- **Daily**: Every day at 3:00 AM (after local backups at 1:00 AM)
- **Weekly**: Every Sunday at 3:00 AM
- **Monthly**: 1st of each month at 3:00 AM

Schedule can be changed via Settings UI → Backup tab → Backup Schedule dropdown.

## Backup Format

Archives are named: `sentinel-backup-YYYY-MM-DD-HHMMSS.tar.gz`

Contents:
```
sentinel-backup-2026-01-08-143022.tar.gz
├── backup-metadata.json    # Metadata with checksums
├── universe.db             # Investment universe
├── config.db               # Configuration
├── ledger.db               # Audit trail
├── portfolio.db            # Portfolio state
├── agents.db               # Strategy agents
├── history.db              # Historical data
└── cache.db                # Ephemeral cache
```

### Metadata Format

`backup-metadata.json` includes:
```json
{
  "timestamp": "2026-01-08T14:30:22Z",
  "version": "1.0.0",
  "sentinel_version": "0.1.0",
  "databases": [
    {
      "name": "universe",
      "filename": "universe.db",
      "size_bytes": 1234567,
      "checksum": "sha256:abc123..."
    }
  ]
}
```

## API Endpoints

### List Backups
```bash
GET /api/backups/r2
```
Returns list of all backups with metadata.

### Create Backup
```bash
POST /api/backups/r2
```
Triggers immediate backup job.

### Test Connection
```bash
POST /api/backups/r2/test
```
Tests R2 credentials and connectivity.

### Download Backup
```bash
GET /api/backups/r2/{filename}/download
```
Downloads backup archive.

### Delete Backup
```bash
DELETE /api/backups/r2/{filename}
```
Deletes specific backup from R2.

### Restore Backup
```bash
POST /api/backups/r2/restore
Body: {"filename": "sentinel-backup-2026-01-08-143022.tar.gz"}
```
Stages restore and restarts service.

### Cancel Staged Restore
```bash
DELETE /api/backups/r2/restore/staged
```
Cancels pending restore operation.

## Two-Phase Restore Process

Sentinel uses a two-phase restore to safely replace databases without downtime:

### Phase 1: Stage Restore (User-Triggered)
1. User selects backup in UI and clicks "Restore"
2. Backend downloads archive from R2
3. Extracts and validates all databases
4. Runs SQLite `PRAGMA integrity_check` on each
5. Creates flag file `.pending-restore`
6. Returns success and triggers system restart

### Phase 2: Apply Restore (On Startup)
1. Service starts and checks for `.pending-restore` flag
2. Creates safety backup in `pre-restore-backup-{timestamp}/`
3. Copies staged databases to production location
4. Deletes flag and staging directory
5. Continues normal startup with restored data

**Safety**: If Phase 2 fails, original databases remain intact in pre-restore-backup.

## Backup Rotation

Old backups are automatically rotated daily at 3:30 AM based on:

- **Retention Policy**: Backups older than N days are deleted
- **Minimum Kept**: Always keep at least 3 backups, regardless of age
- **Retention = 0**: Keep all backups (unlimited)
- **Default**: 90 days

Configure via Settings UI → Backup tab → Retention Days.

## Security

- **Credentials**: Stored encrypted in `config.db`
- **Access**: R2 credentials require S3-compatible permissions
- **Transfer**: HTTPS encryption for all uploads/downloads
- **Validation**: SHA256 checksums prevent corrupted restores

## Monitoring

### Via Logs
```bash
# Check backup status
grep "R2 backup" /var/log/sentinel/sentinel.log

# Check rotation status
grep "R2 backup rotation" /var/log/sentinel/sentinel.log
```

### Via UI
1. Settings → Backup tab → "View Backups"
2. Shows all backups with age and size
3. Actions: Download, Delete, Restore

### Via API
```bash
curl http://localhost:8080/api/backups/r2 | jq
```

## Troubleshooting

### Backup Fails
- Check R2 credentials in Settings
- Click "Test Connection" to verify
- Check logs for detailed error messages
- Verify bucket exists and is accessible

### Restore Fails
- Check disk space (need 2x largest DB size)
- Verify backup integrity via logs
- Pre-restore backup available for recovery
- Contact support if data loss occurs

### Connection Timeout
- Large databases may take time to upload/download
- Default timeout: 30 minutes per operation
- Consider monthly schedule for very large databases

### Credentials Invalid
- Regenerate R2 API credentials in Cloudflare dashboard
- Update in Settings UI → Backup tab
- Click "Test Connection" to verify

## Cost Considerations

Cloudflare R2 pricing (as of 2026):
- **Storage**: $0.015/GB/month
- **Class A Operations** (writes): $4.50 per million
- **Class B Operations** (reads): $0.36 per million
- **Egress**: Free (no data transfer fees)

### Example Costs

7 databases totaling 500MB:
- **Storage**: ~$0.01/month
- **Daily Backups**: ~$0.14/month (1 write/day)
- **Monthly Total**: ~$0.15/month

## Migration from .env

If you previously configured R2 via `.env` file:

1. Open Settings UI → Backup tab
2. Enter your existing credentials
3. Save settings
4. Remove R2 variables from `.env` file
5. Restart service

Settings database takes precedence over `.env`.

## Architecture

### Components

1. **R2Client** (`internal/reliability/r2_client.go`)
   - AWS SDK wrapper for R2 operations
   - Methods: Upload, Download, List, Delete, TestConnection

2. **R2BackupService** (`internal/reliability/r2_backup_service.go`)
   - Orchestrates backup creation and upload
   - Handles rotation and listing
   - Creates compressed archives with metadata

3. **RestoreService** (`internal/reliability/restore_service.go`)
   - Manages two-phase restore process
   - Validates archives and database integrity
   - Creates safety backups

4. **R2BackupJob** (`internal/scheduler/r2_backup.go`)
   - Scheduled job for automatic backups
   - Checks enabled flag and schedule setting
   - Prevents duplicate backups within time window

5. **R2BackupRotationJob** (`internal/scheduler/r2_backup_rotation.go`)
   - Scheduled job for deleting old backups
   - Enforces retention policy
   - Always keeps minimum 3 backups

### Database Schema

Settings stored in `config.db`:
```
r2_backup_enabled          (0/1)
r2_account_id              (string)
r2_access_key_id           (string)
r2_secret_access_key       (string)
r2_bucket_name             (string)
r2_backup_schedule         (daily/weekly/monthly)
r2_backup_retention_days   (integer, 0 = forever)
```

## Testing

Unit tests available in:
- `internal/reliability/r2_client_test.go`
- `internal/reliability/r2_backup_service_test.go`
- `internal/reliability/restore_service_test.go`
- `internal/scheduler/r2_backup_test.go`

Run tests:
```bash
go test ./internal/reliability/... -v
go test ./internal/scheduler/... -run R2 -v
```

## Future Enhancements

Potential improvements:
- Multi-region replication
- Incremental backups
- Encryption at rest
- Backup verification job
- Email notifications
- Slack/Discord webhooks
- Backup size trending
- Cost analytics
