package work

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDeploymentService is a mock implementation for testing
type mockDeploymentService struct {
	checkErr      error
	checkInterval time.Duration
	checkCalls    int
}

func (m *mockDeploymentService) CheckForDeployment() error {
	m.checkCalls++
	return m.checkErr
}

func (m *mockDeploymentService) GetCheckInterval() time.Duration {
	return m.checkInterval
}

func TestRegisterDeploymentWorkTypes(t *testing.T) {
	registry := NewRegistry()

	deps := &DeploymentDeps{
		DeploymentService: &mockDeploymentService{checkInterval: 1 * time.Hour},
	}

	RegisterDeploymentWorkTypes(registry, deps)

	// Verify work type is registered
	wt := registry.Get("deployment:check")
	require.NotNil(t, wt)

	assert.Equal(t, "deployment:check", wt.ID)
	assert.Equal(t, PriorityLow, wt.Priority)
	assert.Equal(t, AnyTime, wt.MarketTiming)
	assert.Equal(t, 1*time.Hour, wt.Interval)
}

func TestDeploymentCheck_Execute(t *testing.T) {
	registry := NewRegistry()

	mock := &mockDeploymentService{checkInterval: 1 * time.Hour}
	deps := &DeploymentDeps{
		DeploymentService: mock,
	}

	RegisterDeploymentWorkTypes(registry, deps)

	wt := registry.Get("deployment:check")
	require.NotNil(t, wt)

	err := wt.Execute(context.Background(), "", nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, mock.checkCalls)
}

func TestDeploymentCheck_FindSubjects(t *testing.T) {
	registry := NewRegistry()

	deps := &DeploymentDeps{
		DeploymentService: &mockDeploymentService{checkInterval: 1 * time.Hour},
	}

	RegisterDeploymentWorkTypes(registry, deps)

	wt := registry.Get("deployment:check")
	require.NotNil(t, wt)

	// Deployment check always returns global work
	subjects := wt.FindSubjects()
	assert.Equal(t, []string{""}, subjects)
}

func TestDeploymentCheck_NoDependencies(t *testing.T) {
	registry := NewRegistry()

	deps := &DeploymentDeps{
		DeploymentService: &mockDeploymentService{checkInterval: 30 * time.Minute},
	}

	RegisterDeploymentWorkTypes(registry, deps)

	wt := registry.Get("deployment:check")
	require.NotNil(t, wt)

	// Deployment check has no dependencies
	assert.Empty(t, wt.DependsOn)
}

func TestDeploymentCheck_CustomInterval(t *testing.T) {
	registry := NewRegistry()

	// Test with custom interval from settings
	deps := &DeploymentDeps{
		DeploymentService: &mockDeploymentService{checkInterval: 15 * time.Minute},
	}

	RegisterDeploymentWorkTypes(registry, deps)

	wt := registry.Get("deployment:check")
	require.NotNil(t, wt)

	assert.Equal(t, 15*time.Minute, wt.Interval)
}
