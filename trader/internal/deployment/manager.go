package deployment

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

// Status represents the deployment status
type Status struct {
	Version         string    `json:"version"`
	DeployedAt      time.Time `json:"deployed_at"`
	GitCommit       string    `json:"git_commit,omitempty"`
	GitBranch       string    `json:"git_branch,omitempty"`
	LastChecked     time.Time `json:"last_checked"`
	UpdateAvailable bool      `json:"update_available"`
}

// Manager handles deployment status tracking
type Manager struct {
	log        zerolog.Logger
	statusFile string
	version    string
	gitCommit  string
	gitBranch  string
}

// NewManager creates a new deployment manager
func NewManager(dataDir string, version string, log zerolog.Logger) *Manager {
	return &Manager{
		log:        log.With().Str("component", "deployment").Logger(),
		statusFile: filepath.Join(dataDir, "deployment_status.json"),
		version:    version,
		gitCommit:  getEnv("GIT_COMMIT", "unknown"),
		gitBranch:  getEnv("GIT_BRANCH", "main"),
	}
}

// GetStatus returns the current deployment status
func (m *Manager) GetStatus() (*Status, error) {
	// Try to read existing status
	status, err := m.loadStatus()
	if err != nil {
		// If status file doesn't exist, create a new one
		if os.IsNotExist(err) {
			status = &Status{
				Version:         m.version,
				DeployedAt:      time.Now(),
				GitCommit:       m.gitCommit,
				GitBranch:       m.gitBranch,
				LastChecked:     time.Now(),
				UpdateAvailable: false,
			}

			if err := m.saveStatus(status); err != nil {
				m.log.Warn().Err(err).Msg("Failed to save initial deployment status")
			}

			return status, nil
		}
		return nil, fmt.Errorf("failed to load status: %w", err)
	}

	// Update last checked time
	status.LastChecked = time.Now()

	return status, nil
}

// MarkDeployed marks a new deployment
func (m *Manager) MarkDeployed() error {
	status := &Status{
		Version:         m.version,
		DeployedAt:      time.Now(),
		GitCommit:       m.gitCommit,
		GitBranch:       m.gitBranch,
		LastChecked:     time.Now(),
		UpdateAvailable: false,
	}

	if err := m.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save deployment status: %w", err)
	}

	m.log.Info().
		Str("version", m.version).
		Str("commit", m.gitCommit).
		Str("branch", m.gitBranch).
		Msg("Deployment marked")

	return nil
}

// CheckForUpdates checks if updates are available
// This is a placeholder - full implementation would check GitHub releases
func (m *Manager) CheckForUpdates() (bool, error) {
	status, err := m.loadStatus()
	if err != nil {
		return false, err
	}

	// Placeholder logic - in production this would:
	// 1. Check GitHub releases API
	// 2. Compare current version with latest release
	// 3. Update status.UpdateAvailable

	// For now, always return false
	status.UpdateAvailable = false
	status.LastChecked = time.Now()

	if err := m.saveStatus(status); err != nil {
		m.log.Warn().Err(err).Msg("Failed to save status after update check")
	}

	m.log.Debug().
		Str("current_version", m.version).
		Bool("update_available", status.UpdateAvailable).
		Msg("Checked for updates")

	return status.UpdateAvailable, nil
}

// TriggerDeployment triggers a deployment workflow
// This is a placeholder - full implementation would:
// 1. Pull latest code from git
// 2. Run build
// 3. Restart service
// For safety, this should only work in development mode
func (m *Manager) TriggerDeployment() error {
	m.log.Warn().Msg("Deployment trigger requested - not implemented for safety")

	// For safety, do not implement automatic deployment in production
	// This should be handled by external CI/CD systems or manual process

	return fmt.Errorf("automatic deployment not implemented - use manual deployment process")
}

// GetUptime returns the time since deployment
func (m *Manager) GetUptime() (time.Duration, error) {
	status, err := m.loadStatus()
	if err != nil {
		return 0, err
	}

	return time.Since(status.DeployedAt), nil
}

// loadStatus loads the deployment status from file
func (m *Manager) loadStatus() (*Status, error) {
	data, err := os.ReadFile(m.statusFile)
	if err != nil {
		return nil, err
	}

	var status Status
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	return &status, nil
}

// saveStatus saves the deployment status to file
func (m *Manager) saveStatus(status *Status) error {
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	if err := os.WriteFile(m.statusFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write status file: %w", err)
	}

	return nil
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
