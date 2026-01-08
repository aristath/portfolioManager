package scheduler

import (
	"errors"
	"testing"

	"github.com/aristath/sentinel/internal/modules/dividends"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetPendingBonusesJob_Name(t *testing.T) {
	job := &SetPendingBonusesJob{
		log: zerolog.Nop(),
	}
	assert.Equal(t, "set_pending_bonuses", job.Name())
}

func TestSetPendingBonusesJob_Run_Success(t *testing.T) {
	mockRepo := new(MockDividendRepositoryForUnreinvested)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewSetPendingBonusesJob(mockRepo)
	job.SetLogger(log)

	dividends := []dividends.DividendRecord{
		{ID: 1, Symbol: "AAPL", AmountEUR: 10.0},
		{ID: 2, Symbol: "MSFT", AmountEUR: 20.0},
	}
	job.SetDividends(dividends)

	mockRepo.On("SetPendingBonus", 1, 10.0).Return(nil)
	mockRepo.On("SetPendingBonus", 2, 20.0).Return(nil)

	err := job.Run()
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSetPendingBonusesJob_Run_NoRepository(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	job := NewSetPendingBonusesJob(nil)
	job.SetLogger(log)

	err := job.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dividend repository not available")
}

func TestSetPendingBonusesJob_Run_Empty(t *testing.T) {
	mockRepo := new(MockDividendRepositoryForUnreinvested)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewSetPendingBonusesJob(mockRepo)
	job.SetLogger(log)

	err := job.Run()
	assert.NoError(t, err)
	mockRepo.AssertNotCalled(t, "SetPendingBonus", mock.Anything, mock.Anything)
}

func TestSetPendingBonusesJob_Run_RepositoryError(t *testing.T) {
	mockRepo := new(MockDividendRepositoryForUnreinvested)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewSetPendingBonusesJob(mockRepo)
	job.SetLogger(log)

	dividends := []dividends.DividendRecord{
		{ID: 1, Symbol: "AAPL", AmountEUR: 10.0},
	}
	job.SetDividends(dividends)

	mockRepo.On("SetPendingBonus", 1, 10.0).Return(errors.New("db error"))

	err := job.Run()
	assert.NoError(t, err) // Job continues even if one fails
	mockRepo.AssertExpectations(t)
}
