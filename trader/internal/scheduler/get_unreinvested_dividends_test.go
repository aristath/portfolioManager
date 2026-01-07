package scheduler

import (
	"errors"
	"testing"

	"github.com/aristath/sentinel/internal/modules/dividends"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDividendRepositoryForUnreinvested is a mock for DividendRepositoryInterface
type MockDividendRepositoryForUnreinvested struct {
	mock.Mock
}

func (m *MockDividendRepositoryForUnreinvested) GetUnreinvestedDividends(minAmountEUR float64) ([]interface{}, error) {
	args := m.Called(minAmountEUR)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockDividendRepositoryForUnreinvested) SetPendingBonus(dividendID int, bonus float64) error {
	args := m.Called(dividendID, bonus)
	return args.Error(0)
}

func (m *MockDividendRepositoryForUnreinvested) MarkReinvested(dividendID int, quantity int) error {
	args := m.Called(dividendID, quantity)
	return args.Error(0)
}

func TestGetUnreinvestedDividendsJob_Name(t *testing.T) {
	job := &GetUnreinvestedDividendsJob{
		log: zerolog.Nop(),
	}
	assert.Equal(t, "get_unreinvested_dividends", job.Name())
}

func TestGetUnreinvestedDividendsJob_Run_Success(t *testing.T) {
	mockRepo := new(MockDividendRepositoryForUnreinvested)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewGetUnreinvestedDividendsJob(mockRepo, 0.0)
	job.SetLogger(log)

	dividends := []dividends.DividendRecord{
		{ID: 1, Symbol: "AAPL", AmountEUR: 10.0},
		{ID: 2, Symbol: "MSFT", AmountEUR: 20.0},
	}
	dividendsInterface := make([]interface{}, len(dividends))
	for i := range dividends {
		dividendsInterface[i] = dividends[i]
	}

	mockRepo.On("GetUnreinvestedDividends", 0.0).Return(dividendsInterface, nil)

	err := job.Run()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(job.GetDividends()))
	mockRepo.AssertExpectations(t)
}

func TestGetUnreinvestedDividendsJob_Run_NoRepository(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	job := NewGetUnreinvestedDividendsJob(nil, 0.0)
	job.SetLogger(log)

	err := job.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dividend repository not available")
}

func TestGetUnreinvestedDividendsJob_Run_RepositoryError(t *testing.T) {
	mockRepo := new(MockDividendRepositoryForUnreinvested)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewGetUnreinvestedDividendsJob(mockRepo, 0.0)
	job.SetLogger(log)

	mockRepo.On("GetUnreinvestedDividends", 0.0).Return(nil, errors.New("db error"))

	err := job.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get unreinvested dividends")
	mockRepo.AssertExpectations(t)
}
