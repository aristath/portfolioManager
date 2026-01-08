package scheduler

import (
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUniverseService is a mock for testing
type MockUniverseService struct {
	mock.Mock
}

func (m *MockUniverseService) SyncPrices() error {
	args := m.Called()
	return args.Error(0)
}

func TestSyncPricesJob_Name(t *testing.T) {
	job := &SyncPricesJob{
		log: zerolog.Nop(),
	}

	assert.Equal(t, "sync_prices", job.Name())
}

func TestSyncPricesJob_Run_Success(t *testing.T) {
	// Setup
	mockUniverseService := new(MockUniverseService)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewSyncPricesJob(SyncPricesConfig{
		Log:             log,
		UniverseService: mockUniverseService,
	})

	// Mock expectations
	mockUniverseService.On("SyncPrices").Return(nil)

	// Execute
	err := job.Run()

	// Assert
	assert.NoError(t, err)
	mockUniverseService.AssertExpectations(t)
	mockUniverseService.AssertCalled(t, "SyncPrices")
}

func TestSyncPricesJob_Run_ServiceError(t *testing.T) {
	// Setup
	mockUniverseService := new(MockUniverseService)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewSyncPricesJob(SyncPricesConfig{
		Log:             log,
		UniverseService: mockUniverseService,
	})

	// Mock expectations - sync fails
	mockUniverseService.On("SyncPrices").Return(errors.New("yahoo api error"))

	// Execute - should not panic, just log error
	err := job.Run()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sync prices failed")
	mockUniverseService.AssertExpectations(t)
}

func TestSyncPricesJob_Run_NoService(t *testing.T) {
	// Setup
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewSyncPricesJob(SyncPricesConfig{
		Log:             log,
		UniverseService: nil,
	})

	// Execute - should not panic, just log warning
	err := job.Run()

	// Assert
	assert.NoError(t, err) // Non-critical, don't fail
}
