package deployment

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/aristath/sentinel/pkg/embedded"
)

// SketchDeployer handles Arduino sketch compilation and upload
type SketchDeployer struct {
	log Logger
}

// NewSketchDeployer creates a new sketch deployer
func NewSketchDeployer(log Logger) *SketchDeployer {
	return &SketchDeployer{
		log: log,
	}
}

// DeploySketch extracts sketch from embedded files and deploys to ArduinoApps directory
// The Arduino App Framework will automatically rebuild and upload the sketch when the app restarts
// sketchPath is the relative path within display/sketch (e.g., "display/sketch/sketch.ino")
func (d *SketchDeployer) DeploySketch(sketchPath string) error {
	// Target directory where Arduino App Framework expects sketch files
	targetDir := "/home/arduino/ArduinoApps/trader-display/sketch"

	d.log.Info().
		Str("sketch", sketchPath).
		Str("target", targetDir).
		Msg("Deploying sketch files to ArduinoApps (framework will auto-rebuild on app restart)")

	// Extract sketch directory from embedded files
	// sketchPath is like "display/sketch/sketch.ino", we need "display/sketch"
	// The embed path is display/sketch relative to the embedded package
	sketchDirPath := filepath.Dir(sketchPath)
	sketchFS, err := fs.Sub(embedded.Files, sketchDirPath)
	if err != nil {
		return fmt.Errorf("failed to get sketch directory from embedded files: %w", err)
	}

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Extract all sketch files to target directory
	if err := d.extractSketchFiles(sketchFS, targetDir); err != nil {
		return fmt.Errorf("failed to extract sketch files: %w", err)
	}

	// Verify sketch.ino exists in target directory
	sketchFile := filepath.Join(targetDir, "sketch.ino")
	if _, err := os.Stat(sketchFile); os.IsNotExist(err) {
		return fmt.Errorf("sketch file not found after extraction: %s", sketchFile)
	}

	d.log.Info().
		Str("sketch", sketchPath).
		Str("target", targetDir).
		Msg("Sketch files deployed successfully - restart app to trigger rebuild/upload")

	return nil
}

// extractSketchFiles extracts all files from embed.FS to target directory
func (d *SketchDeployer) extractSketchFiles(sourceFS fs.FS, targetDir string) error {
	return fs.WalkDir(sourceFS, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself if it's just "."
		if path == "." && entry.IsDir() {
			return nil
		}

		targetPath := filepath.Join(targetDir, path)

		if entry.IsDir() {
			// Create directory
			return os.MkdirAll(targetPath, 0755)
		}

		// Extract file
		return d.extractSketchFile(sourceFS, path, targetPath)
	})
}

// extractSketchFile extracts a single file from embed.FS to target path
func (d *SketchDeployer) extractSketchFile(sourceFS fs.FS, sourcePath string, targetPath string) error {
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

	// Set file permissions
	if err := os.Chmod(targetPath, 0644); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}
