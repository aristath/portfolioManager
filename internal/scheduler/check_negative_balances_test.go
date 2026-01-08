package scheduler

import (
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBalanceService is a mock for testing
type MockBalanceService struct {
	mock.Mock
}

func (m *MockBalanceService) GetAllCurrencies() ([]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockBalanceService) GetTotalByCurrency(currency string) (float64, error) {
	args := m.Called(currency)
	return args.Get(0).(float64), args.Error(1)
}

func TestCheckNegativeBalancesJob_Name(t *testing.T) {
	job := &CheckNegativeBalancesJob{
		log: zerolog.Nop(),
	}

	assert.Equal(t, "check_negative_balances", job.Name())
}

func TestCheckNegativeBalancesJob_Run_AllPositive(t *testing.T) {
	// Setup
	mockBalanceService := new(MockBalanceService)
	log := zerolog.New(nil).Level(zerolog.Disabled)
	emergencyCalled := false

	job := NewCheckNegativeBalancesJob(CheckNegativeBalancesConfig{
		Log:            log,
		BalanceService: mockBalanceService,
		EmergencyRebalance: func() error {
			emergencyCalled = true
			return nil
		},
	})

	// Mock data - all currencies positive
	currencies := []string{"USD", "RUB", "EUR"}
	mockBalanceService.On("GetAllCurrencies").Return(currencies, nil)
	mockBalanceService.On("GetTotalByCurrency", "USD").Return(10000.0, nil)
	mockBalanceService.On("GetTotalByCurrency", "RUB").Return(500000.0, nil)
	mockBalanceService.On("GetTotalByCurrency", "EUR").Return(5000.0, nil)

	// Execute
	err := job.Run()

	// Assert
	assert.NoError(t, err)
	mockBalanceService.AssertExpectations(t)
	assert.False(t, emergencyCalled, "Emergency rebalance should not have been called for positive balances")
}

func TestCheckNegativeBalancesJob_Run_OneNegative(t *testing.T) {
	// Setup
	mockBalanceService := new(MockBalanceService)
	log := zerolog.New(nil).Level(zerolog.Disabled)
	emergencyCalled := false

	job := NewCheckNegativeBalancesJob(CheckNegativeBalancesConfig{
		Log:            log,
		BalanceService: mockBalanceService,
		EmergencyRebalance: func() error {
			emergencyCalled = true
			return nil
		},
	})

	// Mock data - USD is negative
	currencies := []string{"USD", "RUB", "EUR"}
	mockBalanceService.On("GetAllCurrencies").Return(currencies, nil)
	mockBalanceService.On("GetTotalByCurrency", "USD").Return(-100.0, nil)
	mockBalanceService.On("GetTotalByCurrency", "RUB").Return(500000.0, nil)
	mockBalanceService.On("GetTotalByCurrency", "EUR").Return(5000.0, nil)

	// Execute
	err := job.Run()

	// Assert
	assert.NoError(t, err)
	mockBalanceService.AssertExpectations(t)
	assert.True(t, emergencyCalled, "Emergency rebalance should have been called for negative balance")
}

func TestCheckNegativeBalancesJob_Run_NoBalanceService(t *testing.T) {
	// Setup
	log := zerolog.New(nil).Level(zerolog.Disabled)
	emergencyCalled := false

	job := NewCheckNegativeBalancesJob(CheckNegativeBalancesConfig{
		Log:            log,
		BalanceService: nil,
		EmergencyRebalance: func() error {
			emergencyCalled = true
			return nil
		},
	})

	// Execute - should not panic
	err := job.Run()

	// Assert
	assert.NoError(t, err)
	assert.False(t, emergencyCalled, "Emergency rebalance should not be called when balance service is nil")
}

func TestCheckNegativeBalancesJob_Run_NoEmergencyCallback(t *testing.T) {
	// Setup
	mockBalanceService := new(MockBalanceService)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewCheckNegativeBalancesJob(CheckNegativeBalancesConfig{
		Log:                log,
		BalanceService:     mockBalanceService,
		EmergencyRebalance: nil,
	})

	// Mock data - USD is negative
	currencies := []string{"USD"}
	mockBalanceService.On("GetAllCurrencies").Return(currencies, nil)
	mockBalanceService.On("GetTotalByCurrency", "USD").Return(-100.0, nil)

	// Execute - should not panic
	err := job.Run()

	// Assert
	assert.NoError(t, err)
	mockBalanceService.AssertExpectations(t)
}

func TestCheckNegativeBalancesJob_Run_GetCurrenciesError(t *testing.T) {
	// Setup
	mockBalanceService := new(MockBalanceService)
	log := zerolog.New(nil).Level(zerolog.Disabled)
	emergencyCalled := false

	job := NewCheckNegativeBalancesJob(CheckNegativeBalancesConfig{
		Log:            log,
		BalanceService: mockBalanceService,
		EmergencyRebalance: func() error {
			emergencyCalled = true
			return nil
		},
	})

	// Mock data - error getting currencies
	mockBalanceService.On("GetAllCurrencies").Return(nil, errors.New("database error"))

	// Execute - should not panic
	err := job.Run()

	// Assert
	assert.NoError(t, err)
	mockBalanceService.AssertExpectations(t)
	assert.False(t, emergencyCalled, "Emergency rebalance should not be called when currency fetch fails")
}

func TestCheckNegativeBalancesJob_Run_EmergencyRebalanceFails(t *testing.T) {
	// Setup
	mockBalanceService := new(MockBalanceService)
	log := zerolog.New(nil).Level(zerolog.Disabled)
	emergencyCalled := false

	job := NewCheckNegativeBalancesJob(CheckNegativeBalancesConfig{
		Log:            log,
		BalanceService: mockBalanceService,
		EmergencyRebalance: func() error {
			emergencyCalled = true
			return errors.New("rebalance failed")
		},
	})

	// Mock data - negative balance
	currencies := []string{"USD"}
	mockBalanceService.On("GetAllCurrencies").Return(currencies, nil)
	mockBalanceService.On("GetTotalByCurrency", "USD").Return(-100.0, nil)

	// Execute - should not panic even if emergency rebalance fails
	err := job.Run()

	// Assert
	assert.NoError(t, err)
	mockBalanceService.AssertExpectations(t)
	assert.True(t, emergencyCalled, "Emergency rebalance should have been attempted")
}
