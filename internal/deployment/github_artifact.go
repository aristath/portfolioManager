package deployment

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ArtifactTracker tracks the last deployed GitHub Actions run ID
type ArtifactTracker struct {
	trackerFile string
	log         Logger
}

// NewArtifactTracker creates a new artifact tracker
func NewArtifactTracker(trackerFile string, log Logger) *ArtifactTracker {
	return &ArtifactTracker{
		trackerFile: trackerFile,
		log:         log,
	}
}

// GetLastDeployedRunID returns the last deployed run ID, or empty string if none
func (t *ArtifactTracker) GetLastDeployedRunID() (string, error) {
	data, err := os.ReadFile(t.trackerFile)
	if os.IsNotExist(err) {
		return "", nil // No previous deployment
	}
	if err != nil {
		return "", fmt.Errorf("failed to read tracker file: %w", err)
	}

	runID := strings.TrimSpace(string(data))
	if runID == "" {
		return "", nil
	}

	return runID, nil
}

// MarkDeployed records the run ID as deployed
func (t *ArtifactTracker) MarkDeployed(runID string) error {
	// Ensure directory exists
	dir := filepath.Dir(t.trackerFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create tracker directory: %w", err)
	}

	// Write run ID to file
	if err := os.WriteFile(t.trackerFile, []byte(runID+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write tracker file: %w", err)
	}

	t.log.Debug().
		Str("run_id", runID).
		Str("file", t.trackerFile).
		Msg("Marked artifact as deployed")

	return nil
}

// GitHubArtifactDeployer handles downloading and deploying artifacts from GitHub Actions using gh CLI
type GitHubArtifactDeployer struct {
	log          Logger
	workflowName string
	artifactName string
	branch       string
	githubRepo   string // GitHub repository in format "owner/repo"
	tracker      *ArtifactTracker
}

// NewGitHubArtifactDeployer creates a new GitHub artifact deployer
// Note: githubToken parameter is ignored - gh CLI uses its own authentication
func NewGitHubArtifactDeployer(workflowName, artifactName, branch, githubRepo, githubToken string, tracker *ArtifactTracker, log Logger) *GitHubArtifactDeployer {
	return &GitHubArtifactDeployer{
		log:          log,
		workflowName: workflowName,
		artifactName: artifactName,
		branch:       branch,
		githubRepo:   githubRepo,
		tracker:      tracker,
	}
}

// CheckForNewBuild checks if a new successful build is available using gh CLI
// Returns the run ID if a new build is available, empty string otherwise
func (g *GitHubArtifactDeployer) CheckForNewBuild() (string, error) {
	// Get last deployed run ID
	lastRunID, err := g.tracker.GetLastDeployedRunID()
	if err != nil {
		return "", fmt.Errorf("failed to get last deployed run ID: %w", err)
	}

	// Use gh CLI to get latest successful run
	cmd := exec.Command("gh", "run", "list",
		"--repo", g.githubRepo,
		"--workflow", g.workflowName,
		"--branch", g.branch,
		"--status", "success",
		"--limit", "1",
		"--json", "databaseId,headSha,conclusion")

	output, err := cmd.Output()
	if err != nil {
		var stderr string
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		return "", fmt.Errorf("gh run list failed: %w (stderr: %s)", err, stderr)
	}

	// Parse JSON output
	var runs []struct {
		DatabaseID int64  `json:"databaseId"`
		HeadSHA    string `json:"headSha"`
		Conclusion string `json:"conclusion"`
	}

	if err := json.Unmarshal(output, &runs); err != nil {
		return "", fmt.Errorf("failed to parse gh output: %w", err)
	}

	if len(runs) == 0 {
		g.log.Debug().Msg("No successful workflow runs found")
		return "", nil
	}

	latestRun := runs[0]
	latestRunID := fmt.Sprintf("%d", latestRun.DatabaseID)

	// Check if this is a new build
	if lastRunID == "" {
		g.log.Info().
			Str("run_id", latestRunID).
			Str("commit", latestRun.HeadSHA).
			Msg("No previous deployment found, new build available")
		return latestRunID, nil
	}

	if latestRunID != lastRunID {
		g.log.Info().
			Str("last_run_id", lastRunID).
			Str("latest_run_id", latestRunID).
			Str("commit", latestRun.HeadSHA).
			Msg("New build available")
		return latestRunID, nil
	}

	g.log.Debug().
		Str("run_id", latestRunID).
		Msg("No new build available (already deployed)")

	return "", nil
}


// VerifyBinaryArchitecture verifies that a binary is built for linux/arm64
// Uses `file` command to check the ELF architecture
func (g *GitHubArtifactDeployer) VerifyBinaryArchitecture(binaryPath string) error {
	// Use `file` command to check binary architecture
	cmd := exec.Command("file", binaryPath)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check binary architecture: %w", err)
	}

	fileOutput := strings.ToLower(string(output))

	// Check for ELF (Linux binary format) and arm64/aarch64
	// ELF binaries don't contain the word "linux" in file output - ELF itself indicates Linux
	hasELF := strings.Contains(fileOutput, "elf")
	hasARM64 := strings.Contains(fileOutput, "arm64") || strings.Contains(fileOutput, "aarch64")

	if !hasELF {
		return fmt.Errorf("binary is not an ELF binary (Linux format) (detected: %s)", strings.TrimSpace(string(output)))
	}

	if !hasARM64 {
		return fmt.Errorf("binary is not built for ARM64 (detected: %s)", strings.TrimSpace(string(output)))
	}

	g.log.Debug().
		Str("binary", binaryPath).
		Str("file_output", strings.TrimSpace(string(output))).
		Msg("Verified binary architecture: linux/arm64")

	return nil
}

// DownloadArtifact downloads the artifact for a specific run ID using gh CLI
// Returns the path to the downloaded binary
// Verifies that the binary is built for linux/arm64
func (g *GitHubArtifactDeployer) DownloadArtifact(runID string, outputDir string) (string, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	g.log.Info().
		Str("run_id", runID).
		Str("artifact", g.artifactName).
		Str("output_dir", outputDir).
		Msg("Downloading artifact from GitHub Actions using gh CLI")

	// Use gh CLI to download artifact
	// gh automatically handles authentication, retries, and extraction
	cmd := exec.Command("gh", "run", "download", runID,
		"--repo", g.githubRepo,
		"--name", g.artifactName,
		"--dir", outputDir)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gh run download failed: %w (stderr: %s, stdout: %s)",
			err, stderr.String(), stdout.String())
	}

	g.log.Debug().
		Str("output", stdout.String()).
		Msg("gh run download completed")

	// Find the downloaded binary in output directory
	// gh extracts the artifact contents directly into the output directory
	var binaryPath string

	// First check if artifact name exists as a file
	artifactPath := filepath.Join(outputDir, g.artifactName)
	g.log.Debug().
		Str("checking_path", artifactPath).
		Msg("Looking for downloaded artifact")

	if info, err := os.Stat(artifactPath); err == nil && !info.IsDir() {
		binaryPath = artifactPath
		g.log.Debug().
			Str("found", binaryPath).
			Msg("Found artifact at expected path")
	} else {
		// Log why the direct path check failed
		if err != nil {
			g.log.Debug().
				Err(err).
				Str("path", artifactPath).
				Msg("Artifact not at expected path, searching directory")
		} else {
			g.log.Debug().
				Str("path", artifactPath).
				Msg("Path exists but is a directory, searching for file")
		}

		// Try to find any file matching artifact name pattern
		entries, err := os.ReadDir(outputDir)
		if err != nil {
			return "", fmt.Errorf("failed to read output directory: %w", err)
		}

		g.log.Debug().
			Int("file_count", len(entries)).
			Str("directory", outputDir).
			Msg("Scanning directory for artifact")

		for _, entry := range entries {
			g.log.Debug().
				Str("name", entry.Name()).
				Bool("is_dir", entry.IsDir()).
				Msg("Found entry in output directory")

			if !entry.IsDir() {
				// Match files containing artifact name or ending with -arm64
				if strings.Contains(entry.Name(), g.artifactName) || strings.HasSuffix(entry.Name(), "-arm64") {
					binaryPath = filepath.Join(outputDir, entry.Name())
					g.log.Debug().
						Str("matched", binaryPath).
						Msg("Found matching artifact file")
					break
				}
			}
		}
	}

	if binaryPath == "" {
		// List all files for debugging
		entries, _ := os.ReadDir(outputDir)
		fileList := []string{}
		for _, entry := range entries {
			fileList = append(fileList, entry.Name())
		}
		g.log.Error().
			Str("directory", outputDir).
			Str("files_found", strings.Join(fileList, ", ")).
			Str("looking_for", g.artifactName).
			Msg("Downloaded artifact not found in directory")
		return "", fmt.Errorf("downloaded artifact not found in %s", outputDir)
	}

	// Verify binary architecture (CRITICAL: must be linux/arm64)
	g.log.Debug().
		Str("binary", binaryPath).
		Msg("Verifying binary architecture (must be linux/arm64)")

	if err := g.VerifyBinaryArchitecture(binaryPath); err != nil {
		// Remove the invalid binary
		g.log.Error().
			Err(err).
			Str("binary", binaryPath).
			Msg("Binary architecture verification failed - removing invalid binary")
		os.Remove(binaryPath)
		return "", fmt.Errorf("binary architecture verification failed: %w", err)
	}

	g.log.Info().
		Str("binary", binaryPath).
		Msg("Downloaded and verified linux/arm64 binary using gh CLI")

	return binaryPath, nil
}

// DeployLatest checks for a new build and deploys it if available
// Returns the path to the deployed binary, or empty string if no new build
// If runID is provided (non-empty), skips CheckForNewBuild() and uses the provided runID.
// If runID is empty, calls CheckForNewBuild() as before (backward compatibility).
// Note: MarkDeployed() is NOT called here - it should be called after successful deployment.
func (g *GitHubArtifactDeployer) DeployLatest(outputDir string, runID string) (string, error) {
	// If runID is not provided, check for new build
	if runID == "" {
		var err error
		runID, err = g.CheckForNewBuild()
		if err != nil {
			return "", fmt.Errorf("failed to check for new build: %w", err)
		}

		if runID == "" {
			return "", nil // No new build
		}
	}

	// Download artifact
	binaryPath, err := g.DownloadArtifact(runID, outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to download artifact: %w", err)
	}

	// Note: MarkDeployed() is NOT called here - it should be called after successful deployment
	// This prevents marking as deployed if deployment fails later

	g.log.Info().
		Str("run_id", runID).
		Str("binary", binaryPath).
		Msg("Successfully downloaded artifact")

	return binaryPath, nil
}
