package deployment

import (
	"fmt"
	"os"
	"path/filepath"
)

// DisplayAppDeployer handles Python display app deployment (copy-only, no building)
type DisplayAppDeployer struct {
	log Logger
}

// NewDisplayAppDeployer creates a new display app deployer
func NewDisplayAppDeployer(log Logger) *DisplayAppDeployer {
	return &DisplayAppDeployer{
		log: log,
	}
}

// DeployDisplayApp copies the Python app files from display/app/ to ArduinoApps/trader-display/python/
// The Python app files are managed by Arduino App Framework and do not require building.
func (d *DisplayAppDeployer) DeployDisplayApp(repoDir string) error {
	sourceDir := filepath.Join(repoDir, "display/app")
	targetDir := "/home/arduino/ArduinoApps/trader-display/python"

	// Check if source exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		d.log.Warn().
			Str("source", sourceDir).
			Msg("Display app directory not found in repo")
		return nil // Non-fatal - just log warning
	}

	d.log.Info().
		Str("source", sourceDir).
		Str("target", targetDir).
		Msg("Deploying display app (Python files)")

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Copy files recursively
	if err := d.copyDirectory(sourceDir, targetDir); err != nil {
		return fmt.Errorf("failed to copy display app files: %w", err)
	}

	d.log.Info().
		Str("target", targetDir).
		Msg("Successfully deployed display app")

	return nil
}

// copyDirectory recursively copies a directory
func (d *DisplayAppDeployer) copyDirectory(sourceDir string, targetDir string) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, relPath)

		// Re-stat to get current file info
		fileInfo, err := os.Stat(path)
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			// Create directory
			return os.MkdirAll(targetPath, fileInfo.Mode())
		}

		// Copy file
		return d.copyFile(path, targetPath, fileInfo.Mode())
	})
}

// copyFile copies a single file
func (d *DisplayAppDeployer) copyFile(sourcePath string, targetPath string, mode os.FileMode) error {
	// Read source file
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	// Create target directory if needed
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	// Write target file
	if err := os.WriteFile(targetPath, data, mode); err != nil {
		return err
	}

	return nil
}
