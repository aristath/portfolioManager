package reliability

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	_ "modernc.org/sqlite" // SQLite driver
)

// RestoreService manages database restore operations from R2
type RestoreService struct {
	r2Client *R2Client
	dataDir  string
	log      zerolog.Logger
}

// RestoreFlag contains information about a pending restore
type RestoreFlag struct {
	BackupFilename string    `json:"backup_filename"`
	StagedAt       time.Time `json:"staged_at"`
	Databases      []string  `json:"databases"`
}

// NewRestoreService creates a new restore service
func NewRestoreService(r2Client *R2Client, dataDir string, log zerolog.Logger) *RestoreService {
	return &RestoreService{
		r2Client: r2Client,
		dataDir:  dataDir,
		log:      log.With().Str("service", "restore").Logger(),
	}
}

// CheckPendingRestore checks if there is a pending restore operation
func (s *RestoreService) CheckPendingRestore() (bool, error) {
	flagPath := filepath.Join(s.dataDir, ".pending-restore")
	_, err := os.Stat(flagPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check pending restore flag: %w", err)
	}
	return true, nil
}

// StageRestoreFromR2 downloads a backup from R2, validates it, and stages it for restore
// Phase 1 of the two-phase restore process
func (s *RestoreService) StageRestoreFromR2(ctx context.Context, filename string) error {
	s.log.Info().Str("filename", filename).Msg("Starting restore staging from R2")
	startTime := time.Now()

	// Create staging directory
	stagingDir := filepath.Join(s.dataDir, "restore-staging")
	if err := os.RemoveAll(stagingDir); err != nil {
		return fmt.Errorf("failed to clean staging directory: %w", err)
	}
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}

	// Download archive from R2
	archivePath := filepath.Join(stagingDir, filename)
	s.log.Info().Str("filename", filename).Msg("Downloading backup from R2")

	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}

	// Download using R2Client
	writerAt := &FileWriterAt{File: archiveFile, Offset: 0}
	bytesDownloaded, err := s.r2Client.Download(ctx, filename, writerAt)
	archiveFile.Close() // Close before checking error
	if err != nil {
		os.RemoveAll(stagingDir) // Clean up on error
		return fmt.Errorf("failed to download from r2: %w", err)
	}

	s.log.Info().
		Str("filename", filename).
		Int64("bytes", bytesDownloaded).
		Msg("Successfully downloaded backup")

	// Extract archive
	s.log.Info().Msg("Extracting backup archive")
	if err := s.extractArchive(archivePath, stagingDir); err != nil {
		os.RemoveAll(stagingDir) // Clean up on error
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	// Validate extracted files
	s.log.Info().Msg("Validating extracted databases")
	if err := s.validateStagedBackup(stagingDir); err != nil {
		os.RemoveAll(stagingDir) // Clean up on error
		return fmt.Errorf("backup validation failed: %w", err)
	}

	// Read metadata
	metadataPath := filepath.Join(stagingDir, "backup-metadata.json")
	metadata, err := s.readMetadata(metadataPath)
	if err != nil {
		os.RemoveAll(stagingDir) // Clean up on error
		return fmt.Errorf("failed to read metadata: %w", err)
	}

	// Create restore flag
	dbNames := make([]string, len(metadata.Databases))
	for i, db := range metadata.Databases {
		dbNames[i] = db.Name
	}

	flag := RestoreFlag{
		BackupFilename: filename,
		StagedAt:       time.Now().UTC(),
		Databases:      dbNames,
	}

	flagPath := filepath.Join(s.dataDir, ".pending-restore")
	if err := s.writeRestoreFlag(flagPath, flag); err != nil {
		os.RemoveAll(stagingDir) // Clean up on error
		return fmt.Errorf("failed to write restore flag: %w", err)
	}

	duration := time.Since(startTime)
	s.log.Info().
		Dur("duration_ms", duration).
		Str("filename", filename).
		Int("databases", len(dbNames)).
		Msg("Restore staged successfully - restart service to apply")

	return nil
}

// ExecuteStagedRestore applies a staged restore operation
// Phase 2 of the two-phase restore process - called on startup
func (s *RestoreService) ExecuteStagedRestore() error {
	s.log.Warn().Msg("Executing staged restore")
	startTime := time.Now()

	// Read restore flag
	flagPath := filepath.Join(s.dataDir, ".pending-restore")
	flag, err := s.readRestoreFlag(flagPath)
	if err != nil {
		return fmt.Errorf("failed to read restore flag: %w", err)
	}

	stagingDir := filepath.Join(s.dataDir, "restore-staging")

	// Validate staging directory still exists
	if _, err := os.Stat(stagingDir); err != nil {
		return fmt.Errorf("staging directory not found: %w", err)
	}

	// Validate staged files again
	if err := s.validateStagedBackup(stagingDir); err != nil {
		return fmt.Errorf("staged backup validation failed: %w", err)
	}

	// Create pre-restore safety backup
	safetyBackupDir := filepath.Join(s.dataDir, fmt.Sprintf("pre-restore-backup-%s", time.Now().Format("20060102-150405")))
	if err := os.MkdirAll(safetyBackupDir, 0755); err != nil {
		return fmt.Errorf("failed to create safety backup directory: %w", err)
	}

	s.log.Info().Str("backup_dir", safetyBackupDir).Msg("Creating safety backup of current databases")

	// Copy current databases to safety backup
	for _, dbName := range flag.Databases {
		currentPath := filepath.Join(s.dataDir, dbName+".db")
		if _, err := os.Stat(currentPath); err == nil {
			safetyPath := filepath.Join(safetyBackupDir, dbName+".db")
			if err := s.copyFile(currentPath, safetyPath); err != nil {
				s.log.Error().Err(err).Str("database", dbName).Msg("Failed to create safety backup")
				// Continue anyway - restoration is more important
			} else {
				s.log.Debug().Str("database", dbName).Msg("Safety backup created")
			}
		}
	}

	// Apply restore - copy staged files to production location
	s.log.Info().Msg("Applying restore - copying staged databases")

	for _, dbName := range flag.Databases {
		stagedPath := filepath.Join(stagingDir, dbName+".db")
		productionPath := filepath.Join(s.dataDir, dbName+".db")

		// Remove existing database and WAL files
		os.Remove(productionPath)
		os.Remove(productionPath + "-wal")
		os.Remove(productionPath + "-shm")

		// Copy staged database to production
		if err := s.copyFile(stagedPath, productionPath); err != nil {
			return fmt.Errorf("failed to copy %s to production: %w", dbName, err)
		}

		s.log.Info().Str("database", dbName).Msg("Database restored")
	}

	// Delete flag file
	if err := os.Remove(flagPath); err != nil {
		s.log.Error().Err(err).Msg("Failed to delete restore flag")
		// Continue - restore succeeded
	}

	// Delete staging directory
	if err := os.RemoveAll(stagingDir); err != nil {
		s.log.Error().Err(err).Msg("Failed to delete staging directory")
		// Continue - restore succeeded
	}

	duration := time.Since(startTime)
	s.log.Info().
		Dur("duration_ms", duration).
		Int("databases", len(flag.Databases)).
		Str("safety_backup", safetyBackupDir).
		Msg("Restore completed successfully")

	return nil
}

// CancelStagedRestore cancels a pending restore operation
func (s *RestoreService) CancelStagedRestore() error {
	s.log.Info().Msg("Canceling staged restore")

	// Remove flag file
	flagPath := filepath.Join(s.dataDir, ".pending-restore")
	if err := os.Remove(flagPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete restore flag: %w", err)
	}

	// Remove staging directory
	stagingDir := filepath.Join(s.dataDir, "restore-staging")
	if err := os.RemoveAll(stagingDir); err != nil {
		return fmt.Errorf("failed to delete staging directory: %w", err)
	}

	s.log.Info().Msg("Staged restore canceled")
	return nil
}

// validateStagedBackup validates all staged database files
func (s *RestoreService) validateStagedBackup(stagingDir string) error {
	// Check for metadata file
	metadataPath := filepath.Join(stagingDir, "backup-metadata.json")
	metadata, err := s.readMetadata(metadataPath)
	if err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	// Validate each database
	for _, dbInfo := range metadata.Databases {
		dbPath := filepath.Join(stagingDir, dbInfo.Filename)

		// Check file exists
		info, err := os.Stat(dbPath)
		if err != nil {
			return fmt.Errorf("database %s not found: %w", dbInfo.Name, err)
		}

		// Check size matches
		if info.Size() != dbInfo.SizeBytes {
			return fmt.Errorf("database %s size mismatch: expected %d, got %d",
				dbInfo.Name, dbInfo.SizeBytes, info.Size())
		}

		// Run SQLite integrity check
		if err := s.checkIntegrity(dbPath); err != nil {
			return fmt.Errorf("database %s integrity check failed: %w", dbInfo.Name, err)
		}

		s.log.Debug().Str("database", dbInfo.Name).Msg("Database validated")
	}

	return nil
}

// checkIntegrity runs SQLite PRAGMA integrity_check on a database
func (s *RestoreService) checkIntegrity(dbPath string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	var result string
	err = db.QueryRow("PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return fmt.Errorf("integrity check query failed: %w", err)
	}

	if result != "ok" {
		return fmt.Errorf("integrity check failed: %s", result)
	}

	return nil
}

// extractArchive extracts a tar.gz archive
func (s *RestoreService) extractArchive(archivePath, destDir string) error {
	archiveFile, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer archiveFile.Close()

	gzipReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		targetPath := filepath.Join(destDir, header.Name)

		// Security: prevent path traversal
		if !filepath.HasPrefix(targetPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in archive: %s", header.Name)
		}

		if header.Typeflag == tar.TypeReg {
			outFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", header.Name, err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file %s: %w", header.Name, err)
			}

			outFile.Close()
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func (s *RestoreService) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return destFile.Sync()
}

// readMetadata reads backup metadata from a JSON file
func (s *RestoreService) readMetadata(path string) (*BackupMetadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var metadata BackupMetadata
	if err := json.NewDecoder(file).Decode(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// readRestoreFlag reads the restore flag file
func (s *RestoreService) readRestoreFlag(path string) (*RestoreFlag, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var flag RestoreFlag
	if err := json.NewDecoder(file).Decode(&flag); err != nil {
		return nil, err
	}

	return &flag, nil
}

// writeRestoreFlag writes the restore flag file
func (s *RestoreService) writeRestoreFlag(path string, flag RestoreFlag) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(flag)
}

// FileWriterAt wraps a file to implement io.WriterAt for sequential writes
type FileWriterAt struct {
	File   *os.File
	Offset int64
}

func (f *FileWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	if off != f.Offset {
		return 0, fmt.Errorf("FileWriterAt only supports sequential writes")
	}
	n, err = f.File.Write(p)
	f.Offset += int64(n)
	return n, err
}
