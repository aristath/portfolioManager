package scheduler

import (
	"testing"

	planningdomain "github.com/aristath/sentinel/internal/modules/planning/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockPlannerService is a mock implementation of PlannerServiceInterface
type MockPlannerService struct {
	CreatePlanFunc func(ctx interface{}, config interface{}) (interface{}, error)
}

func (m *MockPlannerService) CreatePlan(ctx interface{}, config interface{}) (interface{}, error) {
	if m.CreatePlanFunc != nil {
		return m.CreatePlanFunc(ctx, config)
	}
	return nil, nil
}

// MockConfigRepoForPlan is a mock implementation of ConfigRepositoryInterface
type MockConfigRepoForPlan struct {
	GetDefaultConfigFunc func() (interface{}, error)
}

func (m *MockConfigRepoForPlan) GetDefaultConfig() (interface{}, error) {
	if m.GetDefaultConfigFunc != nil {
		return m.GetDefaultConfigFunc()
	}
	return nil, nil
}

func TestCreateTradePlanJob_Name(t *testing.T) {
	job := NewCreateTradePlanJob(nil, nil)
	assert.Equal(t, "create_trade_plan", job.Name())
}

func TestCreateTradePlanJob_Run_Success(t *testing.T) {
	createPlanCalled := false
	var calledContext interface{}
	var calledConfig interface{}

	mockPlannerService := &MockPlannerService{
		CreatePlanFunc: func(ctx interface{}, config interface{}) (interface{}, error) {
			createPlanCalled = true
			calledContext = ctx
			calledConfig = config
			return &planningdomain.HolisticPlan{
				Steps: []planningdomain.HolisticStep{},
				Feasible: true,
			}, nil
		},
	}

	mockConfigRepo := &MockConfigRepoForPlan{
		GetDefaultConfigFunc: func() (interface{}, error) {
			return &planningdomain.PlannerConfiguration{
				Name: "default",
			}, nil
		},
	}

	opportunityContext := &planningdomain.OpportunityContext{
		Positions: []planningdomain.Position{},
	}

	job := NewCreateTradePlanJob(mockPlannerService, mockConfigRepo)
	job.SetOpportunityContext(opportunityContext)

	err := job.Run()
	require.NoError(t, err)
	assert.True(t, createPlanCalled, "CreatePlan should have been called")
	assert.Equal(t, opportunityContext, calledContext)
	assert.NotNil(t, calledConfig)
	
	plan := job.GetPlan()
	require.NotNil(t, plan)
	assert.NotNil(t, plan.Steps)
}

func TestCreateTradePlanJob_Run_NoPlannerService(t *testing.T) {
	job := NewCreateTradePlanJob(nil, nil)
	job.SetOpportunityContext(&planningdomain.OpportunityContext{
		Positions: []planningdomain.Position{},
	})

	err := job.Run()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "planner service not available")
}

func TestCreateTradePlanJob_Run_NoOpportunityContext(t *testing.T) {
	mockPlannerService := &MockPlannerService{}

	job := NewCreateTradePlanJob(mockPlannerService, nil)
	// Don't set opportunity context

	err := job.Run()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "opportunity context not set")
}

func TestCreateTradePlanJob_Run_PlannerServiceError(t *testing.T) {
	mockPlannerService := &MockPlannerService{
		CreatePlanFunc: func(ctx interface{}, config interface{}) (interface{}, error) {
			return nil, assert.AnError
		},
	}

	job := NewCreateTradePlanJob(mockPlannerService, nil)
	job.SetOpportunityContext(&planningdomain.OpportunityContext{
		Positions: []planningdomain.Position{},
	})

	err := job.Run()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create plan")
}

func TestCreateTradePlanJob_Run_ConfigRepoError(t *testing.T) {
	mockPlannerService := &MockPlannerService{
		CreatePlanFunc: func(ctx interface{}, config interface{}) (interface{}, error) {
			return &planningdomain.HolisticPlan{
				Steps: []planningdomain.HolisticStep{},
				Feasible: true,
			}, nil
		},
	}

	mockConfigRepo := &MockConfigRepoForPlan{
		GetDefaultConfigFunc: func() (interface{}, error) {
			return nil, assert.AnError
		},
	}

	job := NewCreateTradePlanJob(mockPlannerService, mockConfigRepo)
	job.SetOpportunityContext(&planningdomain.OpportunityContext{
		Positions: []planningdomain.Position{},
	})

	// Should use default config when repo fails
	err := job.Run()
	require.NoError(t, err)
}

func TestCreateTradePlanJob_Run_InvalidPlanType(t *testing.T) {
	mockPlannerService := &MockPlannerService{
		CreatePlanFunc: func(ctx interface{}, config interface{}) (interface{}, error) {
			// Return wrong type
			return map[string]interface{}{}, nil
		},
	}

	job := NewCreateTradePlanJob(mockPlannerService, nil)
	job.SetOpportunityContext(&planningdomain.OpportunityContext{
		Positions: []planningdomain.Position{},
	})

	err := job.Run()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan has invalid type")
}
