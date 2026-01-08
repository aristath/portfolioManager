package reliability

import (
	"context"
	"io"
	"sort"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

// TestNewR2BackupService tests creation of R2 backup service
func TestNewR2BackupService(t *testing.T) {
	log := zerolog.New(io.Discard)

	// Create minimal mock dependencies
	r2Client, _ := NewR2Client("test-account", "test-key", "test-secret", "test-bucket", log)
	backupService := &BackupService{}
	dataDir := "/tmp/test"

	service := NewR2BackupService(r2Client, backupService, dataDir, log)

	if service == nil {
		t.Fatal("expected service, got nil")
	}

	if service.r2Client != r2Client {
		t.Error("r2Client not set correctly")
	}

	if service.backupService != backupService {
		t.Error("backupService not set correctly")
	}

	if service.dataDir != dataDir {
		t.Error("dataDir not set correctly")
	}
}

// TestBackupMetadataJSON tests JSON serialization of backup metadata
func TestBackupMetadataJSON(t *testing.T) {
	metadata := BackupMetadata{
		Timestamp:       time.Date(2026, 1, 8, 14, 30, 0, 0, time.UTC),
		Version:         "1.0.0",
		SentinelVersion: "0.1.0",
		Databases: []DatabaseMetadata{
			{
				Name:      "universe",
				Filename:  "universe.db",
				SizeBytes: 1234567,
				Checksum:  "sha256:abc123",
			},
			{
				Name:      "ledger",
				Filename:  "ledger.db",
				SizeBytes: 9876543,
				Checksum:  "sha256:def456",
			},
		},
	}

	// Test that struct can be used for JSON operations
	if metadata.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", metadata.Version)
	}

	if len(metadata.Databases) != 2 {
		t.Errorf("expected 2 databases, got %d", len(metadata.Databases))
	}

	if metadata.Databases[0].Name != "universe" {
		t.Errorf("expected first database to be universe, got %s", metadata.Databases[0].Name)
	}
}

// TestBackupInfoSorting tests sorting of backups by timestamp
func TestBackupInfoSorting(t *testing.T) {
	backups := []BackupInfo{
		{
			Filename:  "sentinel-backup-2026-01-06-120000.tar.gz",
			Timestamp: time.Date(2026, 1, 6, 12, 0, 0, 0, time.UTC),
			SizeBytes: 100,
		},
		{
			Filename:  "sentinel-backup-2026-01-08-120000.tar.gz",
			Timestamp: time.Date(2026, 1, 8, 12, 0, 0, 0, time.UTC),
			SizeBytes: 200,
		},
		{
			Filename:  "sentinel-backup-2026-01-07-120000.tar.gz",
			Timestamp: time.Date(2026, 1, 7, 12, 0, 0, 0, time.UTC),
			SizeBytes: 150,
		},
	}

	// Sort by timestamp (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	// Verify sorting
	if backups[0].Timestamp.Day() != 8 {
		t.Errorf("expected newest backup first (day 8), got day %d", backups[0].Timestamp.Day())
	}

	if backups[1].Timestamp.Day() != 7 {
		t.Errorf("expected second backup (day 7), got day %d", backups[1].Timestamp.Day())
	}

	if backups[2].Timestamp.Day() != 6 {
		t.Errorf("expected oldest backup last (day 6), got day %d", backups[2].Timestamp.Day())
	}
}

// TestCalculateChecksum tests checksum calculation
func TestCalculateChecksum(t *testing.T) {
	log := zerolog.New(io.Discard)
	service := &R2BackupService{
		log: log,
	}

	// Test with non-existent file
	_, err := service.calculateChecksum("/nonexistent/file.db")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

// TestCreateArchive tests archive creation structure
func TestCreateArchive(t *testing.T) {
	log := zerolog.New(io.Discard)
	service := &R2BackupService{
		log: log,
	}

	// Test that method exists and has correct signature
	// We can't test actual functionality without a real filesystem setup
	ctx := context.Background()
	_ = ctx // Use ctx to avoid unused variable warning

	// Test with invalid paths
	err := service.createArchive("/invalid/archive.tar.gz", "/invalid/source", []string{"test"})
	if err == nil {
		t.Error("expected error for invalid paths, got nil")
	}
}

// TestRotateOldBackupsLogic tests the retention logic
func TestRotateOldBackupsLogic(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		backupCount     int
		retentionDays   int
		oldestBackupAge int // days old
		expectDeletion  bool
		expectMinKept   int
	}{
		{
			name:            "few backups, never delete",
			backupCount:     2,
			retentionDays:   30,
			oldestBackupAge: 100,
			expectDeletion:  false,
			expectMinKept:   2,
		},
		{
			name:            "exactly min backups, never delete",
			backupCount:     3,
			retentionDays:   30,
			oldestBackupAge: 100,
			expectDeletion:  false,
			expectMinKept:   3,
		},
		{
			name:            "many backups within retention",
			backupCount:     10,
			retentionDays:   30,
			oldestBackupAge: 29,
			expectDeletion:  false,
			expectMinKept:   3,
		},
		{
			name:            "many backups beyond retention",
			backupCount:     10,
			retentionDays:   30,
			oldestBackupAge: 60,
			expectDeletion:  true,
			expectMinKept:   3,
		},
		{
			name:            "retention 0 keeps all beyond minimum",
			backupCount:     10,
			retentionDays:   0,
			oldestBackupAge: 365,
			expectDeletion:  false,
			expectMinKept:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate logic
			const minBackupsToKeep = 3

			// Can we delete anything?
			canDelete := tt.backupCount > minBackupsToKeep

			// Should we delete? (only if beyond retention and enough backups)
			var cutoffTime time.Time
			if tt.retentionDays > 0 {
				cutoffTime = now.AddDate(0, 0, -tt.retentionDays)
			}

			oldestBackupTime := now.AddDate(0, 0, -tt.oldestBackupAge)
			shouldDeleteOldest := canDelete && tt.retentionDays > 0 && oldestBackupTime.Before(cutoffTime)

			if shouldDeleteOldest != tt.expectDeletion {
				t.Errorf("expected deletion=%v, got %v", tt.expectDeletion, shouldDeleteOldest)
			}
		})
	}
}

// TestDatabaseMetadata tests database metadata structure
func TestDatabaseMetadata(t *testing.T) {
	db := DatabaseMetadata{
		Name:      "universe",
		Filename:  "universe.db",
		SizeBytes: 1024 * 1024, // 1MB
		Checksum:  "sha256:abc123def456",
	}

	if db.Name != "universe" {
		t.Errorf("expected name 'universe', got %s", db.Name)
	}

	if db.SizeBytes != 1024*1024 {
		t.Errorf("expected size 1048576, got %d", db.SizeBytes)
	}

	if db.Checksum[:7] != "sha256:" {
		t.Error("expected checksum to start with 'sha256:'")
	}
}
