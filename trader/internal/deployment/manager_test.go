package deployment

import (
	"testing"
)

// TestDeploy_PassesRunIDToDeployGoService documents expected behavior:
// Deploy() should call CheckForNewBuild() once, store the runID, and pass it to deployGoService().
// This eliminates the duplicate CheckForNewBuild() call.
// This test will be implemented after the runID parameter is added to deployGoService().
func TestDeploy_PassesRunIDToDeployGoService(t *testing.T) {
	t.Skip("Test will be implemented after runID parameter is added to deployGoService()")
}

// TestDeployGoService_MarksDeployedAfterSuccess documents expected behavior:
// deployGoService() should call MarkDeployed() with the runID ONLY after:
// 1. Binary deployment succeeds
// 2. Service restart succeeds
// 3. Health check passes (if applicable)
// This test will be implemented after MarkDeployed is moved to after health check.
func TestDeployGoService_MarksDeployedAfterSuccess(t *testing.T) {
	t.Skip("Test will be implemented after MarkDeployed is moved to after health check")
}

// TestDeployGoService_DoesNotMarkDeployedOnFailure documents expected behavior:
// deployGoService() should NOT call MarkDeployed() if any step fails:
// - Binary deployment failure
// - Service restart failure
// - Health check failure
// This test will be implemented after MarkDeployed is moved to after health check.
func TestDeployGoService_DoesNotMarkDeployedOnFailure(t *testing.T) {
	t.Skip("Test will be implemented after MarkDeployed is moved to after health check")
}

// TestDeployGoService_WithEmptyRunID_StillWorks documents expected behavior:
// When runID is empty, deployGoService() should still work (backward compatibility).
// DeployLatest() will call CheckForNewBuild() when runID is empty.
// This test will be implemented after runID parameter is added.
func TestDeployGoService_WithEmptyRunID_StillWorks(t *testing.T) {
	t.Skip("Test will be implemented after runID parameter is added")
}
