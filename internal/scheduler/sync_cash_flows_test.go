package scheduler

import (
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCashFlowsService is a mock for CashFlowsServiceInterface
type MockCashFlowsService struct {
	mock.Mock
}

func (m *MockCashFlowsService) SyncFromTradernet() error {
	args := m.Called()
	return args.Error(0)
}

func TestSyncCashFlowsJob_Name(t *testing.T) {
	job := &SyncCashFlowsJob{
		log: zerolog.Nop(),
	}

	assert.Equal(t, "sync_cash_flows", job.Name())
}

func TestSyncCashFlowsJob_Run_Success(t *testing.T) {
	// Setup
	mockCashFlowsService := new(MockCashFlowsService)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewSyncCashFlowsJob(SyncCashFlowsConfig{
		Log:              log,
		CashFlowsService: mockCashFlowsService,
	})

	// Mock expectations
	mockCashFlowsService.On("SyncFromTradernet").Return(nil)

	// Execute
	err := job.Run()

	// Assert
	assert.NoError(t, err)
	mockCashFlowsService.AssertExpectations(t)
	mockCashFlowsService.AssertCalled(t, "SyncFromTradernet")
}

func TestSyncCashFlowsJob_Run_ServiceError(t *testing.T) {
	// Setup
	mockCashFlowsService := new(MockCashFlowsService)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewSyncCashFlowsJob(SyncCashFlowsConfig{
		Log:              log,
		CashFlowsService: mockCashFlowsService,
	})

	// Mock expectations - service returns error
	mockCashFlowsService.On("SyncFromTradernet").Return(errors.New("tradernet connection failed"))

	// Execute
	err := job.Run()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sync cash flows failed")
	mockCashFlowsService.AssertExpectations(t)
}

func TestSyncCashFlowsJob_Run_NoService(t *testing.T) {
	// Setup
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewSyncCashFlowsJob(SyncCashFlowsConfig{
		Log:              log,
		CashFlowsService: nil,
	})

	// Execute - should not panic, just log warning and return nil
	err := job.Run()

	// Assert
	assert.NoError(t, err) // Non-critical, don't fail
}
