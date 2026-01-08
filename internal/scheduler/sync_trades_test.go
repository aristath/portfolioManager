package scheduler

import (
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTradingService is a mock for TradingServiceInterface
type MockTradingService struct {
	mock.Mock
}

func (m *MockTradingService) SyncFromTradernet() error {
	args := m.Called()
	return args.Error(0)
}

func TestSyncTradesJob_Name(t *testing.T) {
	job := &SyncTradesJob{
		log: zerolog.Nop(),
	}

	assert.Equal(t, "sync_trades", job.Name())
}

func TestSyncTradesJob_Run_Success(t *testing.T) {
	// Setup
	mockTradingService := new(MockTradingService)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewSyncTradesJob(SyncTradesConfig{
		Log:            log,
		TradingService: mockTradingService,
	})

	// Mock expectations
	mockTradingService.On("SyncFromTradernet").Return(nil)

	// Execute
	err := job.Run()

	// Assert
	assert.NoError(t, err)
	mockTradingService.AssertExpectations(t)
	mockTradingService.AssertCalled(t, "SyncFromTradernet")
}

func TestSyncTradesJob_Run_ServiceError(t *testing.T) {
	// Setup
	mockTradingService := new(MockTradingService)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewSyncTradesJob(SyncTradesConfig{
		Log:            log,
		TradingService: mockTradingService,
	})

	// Mock expectations - service returns error
	mockTradingService.On("SyncFromTradernet").Return(errors.New("tradernet connection failed"))

	// Execute
	err := job.Run()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sync trades failed")
	mockTradingService.AssertExpectations(t)
}

func TestSyncTradesJob_Run_NoService(t *testing.T) {
	// Setup
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewSyncTradesJob(SyncTradesConfig{
		Log:            log,
		TradingService: nil,
	})

	// Execute - should not panic, just log warning and return nil
	err := job.Run()

	// Assert
	assert.NoError(t, err) // Non-critical, don't fail
}
