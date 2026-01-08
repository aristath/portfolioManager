package scheduler

import (
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPortfolioService is a mock for PortfolioServiceInterface
type MockPortfolioService struct {
	mock.Mock
}

func (m *MockPortfolioService) SyncFromTradernet() error {
	args := m.Called()
	return args.Error(0)
}

func TestSyncPortfolioJob_Name(t *testing.T) {
	job := &SyncPortfolioJob{
		log: zerolog.Nop(),
	}

	assert.Equal(t, "sync_portfolio", job.Name())
}

func TestSyncPortfolioJob_Run_Success(t *testing.T) {
	// Setup
	mockPortfolioService := new(MockPortfolioService)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewSyncPortfolioJob(SyncPortfolioConfig{
		Log:              log,
		PortfolioService: mockPortfolioService,
	})

	// Mock expectations
	mockPortfolioService.On("SyncFromTradernet").Return(nil)

	// Execute
	err := job.Run()

	// Assert
	assert.NoError(t, err)
	mockPortfolioService.AssertExpectations(t)
	mockPortfolioService.AssertCalled(t, "SyncFromTradernet")
}

func TestSyncPortfolioJob_Run_ServiceError(t *testing.T) {
	// Setup
	mockPortfolioService := new(MockPortfolioService)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewSyncPortfolioJob(SyncPortfolioConfig{
		Log:              log,
		PortfolioService: mockPortfolioService,
	})

	// Mock expectations - service returns error
	mockPortfolioService.On("SyncFromTradernet").Return(errors.New("tradernet connection failed"))

	// Execute
	err := job.Run()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sync portfolio failed")
	mockPortfolioService.AssertExpectations(t)
}

func TestSyncPortfolioJob_Run_NoService(t *testing.T) {
	// Setup
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewSyncPortfolioJob(SyncPortfolioConfig{
		Log:              log,
		PortfolioService: nil,
	})

	// Execute - should fail for critical job
	err := job.Run()

	// Assert
	assert.Error(t, err) // Critical job, should fail if service not available
	assert.Contains(t, err.Error(), "portfolio service not available")
}
