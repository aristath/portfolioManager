package deployment

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeployLatest_WithRunID_SkipsCheckForNewBuild documents expected behavior:
// When runID is provided, CheckForNewBuild should not be called.
// This test will be implemented after DeployLatest accepts runID parameter.
func TestDeployLatest_WithRunID_SkipsCheckForNewBuild(t *testing.T) {
	t.Skip("Test will be implemented after DeployLatest accepts runID parameter")
}

// TestDeployLatest_WithoutRunID_CallsCheckForNewBuild documents expected behavior:
// When runID is empty, CheckForNewBuild should be called (backward compatibility).
// This test will be implemented after DeployLatest accepts runID parameter.
func TestDeployLatest_WithoutRunID_CallsCheckForNewBuild(t *testing.T) {
	t.Skip("Test will be implemented after DeployLatest accepts runID parameter")
}

// TestDeployLatest_DoesNotMarkDeployed documents expected behavior:
// MarkDeployed should NOT be called in DeployLatest.
// Marking should happen only after successful deployment.
// This test will be implemented after MarkDeployed is removed from DeployLatest.
func TestDeployLatest_DoesNotMarkDeployed(t *testing.T) {
	t.Skip("Test will be implemented after MarkDeployed is removed from DeployLatest")
}

func TestArtifactTracker_MarkDeployed(t *testing.T) {
	tempDir := t.TempDir()
	trackerFile := filepath.Join(tempDir, "github-artifact-id.txt")
	log := &mockLogger{}

	tracker := NewArtifactTracker(trackerFile, log)

	// Mark as deployed
	runID := "12345"
	err := tracker.MarkDeployed(runID)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(trackerFile)
	require.NoError(t, err, "Tracker file should be created")

	// Verify content
	data, err := os.ReadFile(trackerFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), runID, "Tracker file should contain runID")

	// Verify GetLastDeployedRunID returns it
	lastRunID, err := tracker.GetLastDeployedRunID()
	require.NoError(t, err)
	assert.Equal(t, runID, lastRunID, "GetLastDeployedRunID should return the marked runID")
}

func TestArtifactTracker_GetLastDeployedRunID_NoFile(t *testing.T) {
	tempDir := t.TempDir()
	trackerFile := filepath.Join(tempDir, "non-existent.txt")
	log := &mockLogger{}

	tracker := NewArtifactTracker(trackerFile, log)

	// Get last deployed when file doesn't exist
	runID, err := tracker.GetLastDeployedRunID()
	require.NoError(t, err)
	assert.Empty(t, runID, "Should return empty string when tracker file doesn't exist")
}
