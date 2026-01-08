package reliability

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

// TestNewRestoreService tests creation of restore service
func TestNewRestoreService(t *testing.T) {
	log := zerolog.New(io.Discard)

	r2Client, _ := NewR2Client("test-account", "test-key", "test-secret", "test-bucket", log)
	dataDir := "/tmp/test"

	service := NewRestoreService(r2Client, dataDir, log)

	if service == nil {
		t.Fatal("expected service, got nil")
	}

	if service.r2Client != r2Client {
		t.Error("r2Client not set correctly")
	}

	if service.dataDir != dataDir {
		t.Error("dataDir not set correctly")
	}
}

// TestRestoreFlagJSON tests JSON serialization of restore flag
func TestRestoreFlagJSON(t *testing.T) {
	flag := RestoreFlag{
		BackupFilename: "sentinel-backup-2026-01-08-143022.tar.gz",
		StagedAt:       time.Date(2026, 1, 8, 14, 30, 0, 0, time.UTC),
		Databases:      []string{"universe", "config", "ledger", "portfolio", "agents", "history", "cache"},
	}

	// Marshal to JSON
	data, err := json.Marshal(flag)
	if err != nil {
		t.Fatalf("failed to marshal flag: %v", err)
	}

	// Unmarshal from JSON
	var decoded RestoreFlag
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal flag: %v", err)
	}

	// Verify fields
	if decoded.BackupFilename != flag.BackupFilename {
		t.Errorf("expected filename %s, got %s", flag.BackupFilename, decoded.BackupFilename)
	}

	if len(decoded.Databases) != len(flag.Databases) {
		t.Errorf("expected %d databases, got %d", len(flag.Databases), len(decoded.Databases))
	}

	if decoded.Databases[0] != "universe" {
		t.Errorf("expected first database to be universe, got %s", decoded.Databases[0])
	}
}

// TestCheckPendingRestore tests checking for pending restore flag
func TestCheckPendingRestore(t *testing.T) {
	log := zerolog.New(io.Discard)

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	service := NewRestoreService(nil, tempDir, log)

	// Test no pending restore
	hasPending, err := service.CheckPendingRestore()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if hasPending {
		t.Error("expected no pending restore, got pending")
	}

	// Create flag file
	flagPath := filepath.Join(tempDir, ".pending-restore")
	if err := os.WriteFile(flagPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create flag file: %v", err)
	}

	// Test with pending restore
	hasPending, err = service.CheckPendingRestore()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !hasPending {
		t.Error("expected pending restore, got no pending")
	}
}

// TestCancelStagedRestore tests canceling a staged restore
func TestCancelStagedRestore(t *testing.T) {
	log := zerolog.New(io.Discard)

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	service := NewRestoreService(nil, tempDir, log)

	// Create flag file
	flagPath := filepath.Join(tempDir, ".pending-restore")
	if err := os.WriteFile(flagPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create flag file: %v", err)
	}

	// Create staging directory
	stagingDir := filepath.Join(tempDir, "restore-staging")
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		t.Fatalf("failed to create staging dir: %v", err)
	}

	// Cancel restore
	if err := service.CancelStagedRestore(); err != nil {
		t.Errorf("failed to cancel restore: %v", err)
	}

	// Verify flag deleted
	if _, err := os.Stat(flagPath); !os.IsNotExist(err) {
		t.Error("expected flag file to be deleted")
	}

	// Verify staging dir deleted
	if _, err := os.Stat(stagingDir); !os.IsNotExist(err) {
		t.Error("expected staging directory to be deleted")
	}
}

// TestWriteAndReadRestoreFlag tests flag file I/O
func TestWriteAndReadRestoreFlag(t *testing.T) {
	log := zerolog.New(io.Discard)

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	service := NewRestoreService(nil, tempDir, log)

	// Create flag
	flag := RestoreFlag{
		BackupFilename: "test-backup.tar.gz",
		StagedAt:       time.Now().UTC(),
		Databases:      []string{"universe", "config", "ledger"},
	}

	flagPath := filepath.Join(tempDir, "test-flag.json")

	// Write flag
	if err := service.writeRestoreFlag(flagPath, flag); err != nil {
		t.Fatalf("failed to write flag: %v", err)
	}

	// Read flag
	readFlag, err := service.readRestoreFlag(flagPath)
	if err != nil {
		t.Fatalf("failed to read flag: %v", err)
	}

	// Verify
	if readFlag.BackupFilename != flag.BackupFilename {
		t.Errorf("expected filename %s, got %s", flag.BackupFilename, readFlag.BackupFilename)
	}

	if len(readFlag.Databases) != len(flag.Databases) {
		t.Errorf("expected %d databases, got %d", len(flag.Databases), len(readFlag.Databases))
	}
}

// TestRestoreServiceCopyFile tests file copying
func TestRestoreServiceCopyFile(t *testing.T) {
	log := zerolog.New(io.Discard)

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	service := NewRestoreService(nil, tempDir, log)

	// Create source file
	srcPath := filepath.Join(tempDir, "source.txt")
	testData := []byte("test data for copying")
	if err := os.WriteFile(srcPath, testData, 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Copy file
	dstPath := filepath.Join(tempDir, "destination.txt")
	if err := service.copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("failed to copy file: %v", err)
	}

	// Verify destination exists
	if _, err := os.Stat(dstPath); err != nil {
		t.Errorf("destination file not created: %v", err)
	}

	// Verify content matches
	copiedData, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}

	if string(copiedData) != string(testData) {
		t.Errorf("content mismatch: expected %q, got %q", testData, copiedData)
	}
}

// TestFileWriterAt tests the FileWriterAt wrapper
func TestFileWriterAt(t *testing.T) {
	// Create temp file
	tempDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "test.dat")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()

	// Create FileWriterAt
	writer := &FileWriterAt{File: file, Offset: 0}

	// Write sequential data
	data1 := []byte("hello")
	n, err := writer.WriteAt(data1, 0)
	if err != nil {
		t.Errorf("failed to write at offset 0: %v", err)
	}
	if n != len(data1) {
		t.Errorf("expected to write %d bytes, wrote %d", len(data1), n)
	}

	data2 := []byte(" world")
	n, err = writer.WriteAt(data2, 5)
	if err != nil {
		t.Errorf("failed to write at offset 5: %v", err)
	}
	if n != len(data2) {
		t.Errorf("expected to write %d bytes, wrote %d", len(data2), n)
	}

	// Try non-sequential write (should fail)
	data3 := []byte("!")
	_, err = writer.WriteAt(data3, 0) // Try to write at offset 0 again
	if err == nil {
		t.Error("expected error for non-sequential write, got nil")
	}

	// Verify file content
	file.Close()
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	expected := "hello world"
	if string(content) != expected {
		t.Errorf("expected content %q, got %q", expected, string(content))
	}
}

// TestValidateStagedBackup tests backup validation logic
func TestValidateStagedBackup(t *testing.T) {
	log := zerolog.New(io.Discard)

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	service := NewRestoreService(nil, tempDir, log)

	// Test with missing metadata
	err = service.validateStagedBackup(tempDir)
	if err == nil {
		t.Error("expected error for missing metadata, got nil")
	}

	// Create minimal metadata file (validation will still fail due to missing DB files)
	metadata := BackupMetadata{
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Databases: []DatabaseMetadata{
			{
				Name:      "test",
				Filename:  "test.db",
				SizeBytes: 100,
				Checksum:  "sha256:abc123",
			},
		},
	}

	metadataPath := filepath.Join(tempDir, "backup-metadata.json")
	file, err := os.Create(metadataPath)
	if err != nil {
		t.Fatalf("failed to create metadata file: %v", err)
	}
	json.NewEncoder(file).Encode(metadata)
	file.Close()

	// Validation should fail (database file doesn't exist)
	err = service.validateStagedBackup(tempDir)
	if err == nil {
		t.Error("expected error for missing database file, got nil")
	}
}
