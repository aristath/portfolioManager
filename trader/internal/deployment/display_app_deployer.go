package deployment

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/aristath/sentinel/pkg/embedded"
)

// DisplayAppDeployer handles Python display app deployment (extracts from embedded files)
type DisplayAppDeployer struct {
	log Logger
}

// NewDisplayAppDeployer creates a new display app deployer
func NewDisplayAppDeployer(log Logger) *DisplayAppDeployer {
	return &DisplayAppDeployer{
		log: log,
	}
}

// DeployDisplayApp extracts the Python app files from embedded filesystem to ArduinoApps/trader-display/python/
// The Python app files are managed by Arduino App Framework and do not require building.
func (d *DisplayAppDeployer) DeployDisplayApp() error {
	targetDir := "/home/arduino/ArduinoApps/trader-display/python"

	d.log.Info().
		Str("target", targetDir).
		Msg("Deploying display app (extracting from embedded files)")

	// Get display/app subdirectory from embedded files
	// The embed path is display/app relative to the embedded package
	displayAppFS, err := fs.Sub(embedded.Files, "display/app")
	if err != nil {
		return fmt.Errorf("failed to get display/app from embedded files: %w", err)
	}

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Extract files from embedded filesystem to target directory
	if err := d.extractDirectory(displayAppFS, ".", targetDir); err != nil {
		return fmt.Errorf("failed to extract display app files: %w", err)
	}

	d.log.Info().
		Str("target", targetDir).
		Msg("Successfully deployed display app")

	return nil
}

// extractDirectory recursively extracts files from embed.FS to target directory
func (d *DisplayAppDeployer) extractDirectory(sourceFS fs.FS, sourcePath string, targetDir string) error {
	return fs.WalkDir(sourceFS, sourcePath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself if it's just "."
		if path == "." && entry.IsDir() {
			return nil
		}

		// Calculate relative path from source root
		relPath := path
		if sourcePath != "." {
			relPath, err = filepath.Rel(sourcePath, path)
			if err != nil {
				return err
			}
		}

		// Remove leading "./" if present
		relPath = strings.TrimPrefix(relPath, "./")
		if relPath == "" {
			return nil
		}

		targetPath := filepath.Join(targetDir, relPath)

		if entry.IsDir() {
			// Create directory
			return os.MkdirAll(targetPath, 0755)
		}

		// Extract file
		return d.extractFile(sourceFS, path, targetPath)
	})
}

// extractFile extracts a single file from embed.FS to target path
func (d *DisplayAppDeployer) extractFile(sourceFS fs.FS, sourcePath string, targetPath string) error {
	// Open source file from embedded filesystem
	sourceFile, err := sourceFS.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open embedded file %s: %w", sourcePath, err)
	}
	defer sourceFile.Close()

	// Create target directory if needed
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	// Create target file
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target file %s: %w", targetPath, err)
	}
	defer targetFile.Close()

	// Copy file contents
	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Set file permissions (executable for Python files)
	if err := os.Chmod(targetPath, 0644); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}
