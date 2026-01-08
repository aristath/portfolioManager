package scheduler

import (
	"errors"
	"testing"

	"github.com/aristath/sentinel/internal/modules/portfolio"
	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPositionRepoForHash is a mock for position repository for hash generation
type MockPositionRepoForHash struct {
	mock.Mock
}

func (m *MockPositionRepoForHash) GetAll() ([]portfolio.Position, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]portfolio.Position), args.Error(1)
}

// MockSecurityRepoForHash is a mock for security repository for hash generation
type MockSecurityRepoForHash struct {
	mock.Mock
}

func (m *MockSecurityRepoForHash) GetAllActive() ([]universe.Security, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]universe.Security), args.Error(1)
}

// MockAllocRepoForHash is a mock for allocation repository
type MockAllocRepoForHash struct {
	mock.Mock
}

func (m *MockAllocRepoForHash) GetAll() (map[string]float64, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]float64), args.Error(1)
}

// MockCashManagerForHash is a mock for cash manager
type MockCashManagerForHash struct {
	mock.Mock
}

func (m *MockCashManagerForHash) GetAllCashBalances() (map[string]float64, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *MockCashManagerForHash) UpdateCashPosition(currency string, balance float64) error {
	args := m.Called(currency, balance)
	return args.Error(0)
}

func (m *MockCashManagerForHash) GetCashBalance(currency string) (float64, error) {
	args := m.Called(currency)
	return args.Get(0).(float64), args.Error(1)
}

func TestGeneratePortfolioHashJob_Name(t *testing.T) {
	job := &GeneratePortfolioHashJob{
		log: zerolog.Nop(),
	}

	assert.Equal(t, "generate_portfolio_hash", job.Name())
}

func TestGeneratePortfolioHashJob_Run_Success(t *testing.T) {
	// Setup
	mockPositionRepo := new(MockPositionRepoForHash)
	mockSecurityRepo := new(MockSecurityRepoForHash)
	mockCashManager := new(MockCashManagerForHash)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewGeneratePortfolioHashJob(GeneratePortfolioHashConfig{
		Log:          log,
		PositionRepo: mockPositionRepo,
		SecurityRepo: mockSecurityRepo,
		CashManager:  mockCashManager,
	})

	// Mock data
	positions := []portfolio.Position{
		{Symbol: "AAPL", Quantity: 10},
		{Symbol: "MSFT", Quantity: 5},
	}
	securities := []universe.Security{
		{Symbol: "AAPL", Country: "US", Industry: "Technology"},
		{Symbol: "MSFT", Country: "US", Industry: "Technology"},
	}
	cashBalances := map[string]float64{
		"EUR": 1000.0,
	}

	mockPositionRepo.On("GetAll").Return(positions, nil)
	mockSecurityRepo.On("GetAllActive").Return(securities, nil)
	mockCashManager.On("GetAllCashBalances").Return(cashBalances, nil)

	// Execute
	err := job.Run()

	// Assert
	assert.NoError(t, err)
	mockPositionRepo.AssertExpectations(t)
	mockSecurityRepo.AssertExpectations(t)
	mockCashManager.AssertExpectations(t)
	assert.NotEmpty(t, job.lastPortfolioHash, "Portfolio hash should be generated")
}

func TestGeneratePortfolioHashJob_Run_NoPositions(t *testing.T) {
	// Setup
	mockPositionRepo := new(MockPositionRepoForHash)
	mockSecurityRepo := new(MockSecurityRepoForHash)
	mockCashManager := new(MockCashManagerForHash)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewGeneratePortfolioHashJob(GeneratePortfolioHashConfig{
		Log:          log,
		PositionRepo: mockPositionRepo,
		SecurityRepo: mockSecurityRepo,
		CashManager:  mockCashManager,
	})

	// Mock data - no positions
	positions := []portfolio.Position{}
	securities := []universe.Security{}
	cashBalances := map[string]float64{"EUR": 1000.0}

	mockPositionRepo.On("GetAll").Return(positions, nil)
	mockSecurityRepo.On("GetAllActive").Return(securities, nil)
	mockCashManager.On("GetAllCashBalances").Return(cashBalances, nil)

	// Execute
	err := job.Run()

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, job.lastPortfolioHash, "Portfolio hash should be generated even with no positions")
}

func TestGeneratePortfolioHashJob_Run_RepositoryError(t *testing.T) {
	// Setup
	mockPositionRepo := new(MockPositionRepoForHash)
	mockSecurityRepo := new(MockSecurityRepoForHash)
	mockCashManager := new(MockCashManagerForHash)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewGeneratePortfolioHashJob(GeneratePortfolioHashConfig{
		Log:          log,
		PositionRepo: mockPositionRepo,
		SecurityRepo: mockSecurityRepo,
		CashManager:  mockCashManager,
	})

	// Mock expectations - position repo returns error
	mockPositionRepo.On("GetAll").Return(nil, errors.New("database error"))

	// Execute
	err := job.Run()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get positions")
	mockPositionRepo.AssertExpectations(t)
}

func TestGeneratePortfolioHashJob_Run_HashUnchanged(t *testing.T) {
	// Setup
	mockPositionRepo := new(MockPositionRepoForHash)
	mockSecurityRepo := new(MockSecurityRepoForHash)
	mockCashManager := new(MockCashManagerForHash)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewGeneratePortfolioHashJob(GeneratePortfolioHashConfig{
		Log:          log,
		PositionRepo: mockPositionRepo,
		SecurityRepo: mockSecurityRepo,
		CashManager:  mockCashManager,
	})

	// Mock data
	positions := []portfolio.Position{
		{Symbol: "AAPL", Quantity: 10},
	}
	securities := []universe.Security{
		{Symbol: "AAPL", Country: "US", Industry: "Technology"},
	}
	cashBalances := map[string]float64{"EUR": 1000.0}

	mockPositionRepo.On("GetAll").Return(positions, nil).Times(2)
	mockSecurityRepo.On("GetAllActive").Return(securities, nil).Times(2)
	mockCashManager.On("GetAllCashBalances").Return(cashBalances, nil).Times(2)

	// Execute first time
	err1 := job.Run()
	assert.NoError(t, err1)
	firstHash := job.lastPortfolioHash

	// Execute second time - hash should be the same
	err2 := job.Run()
	assert.NoError(t, err2)
	secondHash := job.lastPortfolioHash

	// Assert
	assert.Equal(t, firstHash, secondHash, "Hash should be the same for unchanged portfolio")
	mockPositionRepo.AssertExpectations(t)
	mockSecurityRepo.AssertExpectations(t)
	mockCashManager.AssertExpectations(t)
}
