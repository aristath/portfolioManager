package scheduler

import (
	"testing"

	"github.com/aristath/sentinel/internal/modules/planning/domain"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTradeExecutionServiceForDividends is a mock for TradeExecutionServiceInterface
type MockTradeExecutionServiceForDividends struct {
	mock.Mock
}

func (m *MockTradeExecutionServiceForDividends) ExecuteTrades(recommendations []TradeRecommendationForDividends) []TradeResultForDividends {
	args := m.Called(recommendations)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]TradeResultForDividends)
}

func TestExecuteDividendTradesJob_Name(t *testing.T) {
	job := &ExecuteDividendTradesJob{
		log: zerolog.Nop(),
	}
	assert.Equal(t, "execute_dividend_trades", job.Name())
}

func TestExecuteDividendTradesJob_Run_NoRecommendations(t *testing.T) {
	mockDividendRepo := new(MockDividendRepositoryForUnreinvested)
	mockTradeService := new(MockTradeExecutionServiceForDividends)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewExecuteDividendTradesJob(mockDividendRepo, mockTradeService)
	job.SetLogger(log)

	err := job.Run()
	assert.NoError(t, err)
	assert.Equal(t, 0, job.GetExecutedCount())
}

func TestExecuteDividendTradesJob_Run_Success(t *testing.T) {
	mockDividendRepo := new(MockDividendRepositoryForUnreinvested)
	mockTradeService := new(MockTradeExecutionServiceForDividends)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewExecuteDividendTradesJob(mockDividendRepo, mockTradeService)
	job.SetLogger(log)

	recommendations := []domain.HolisticStep{
		{Symbol: "AAPL", Side: "BUY", Quantity: 10, EstimatedPrice: 150.0, EstimatedValue: 1500.0, Currency: "USD", Reason: "test"},
	}
	dividendsToMark := map[string][]int{
		"AAPL": {1, 2},
	}
	job.SetRecommendations(recommendations, dividendsToMark)

	// The job converts domain.HolisticStep to TradeRecommendationForDividends internally
	// Use mock.MatchedBy to match the actual call
	tradeResults := []TradeResultForDividends{
		{Symbol: "AAPL", Status: "success", Error: nil},
	}

	mockTradeService.On("ExecuteTrades", mock.MatchedBy(func(recs []TradeRecommendationForDividends) bool {
		return len(recs) == 1 && recs[0].Symbol == "AAPL" && recs[0].Side == "BUY"
	})).Return(tradeResults)
	mockDividendRepo.On("MarkReinvested", 1, 10).Return(nil)
	mockDividendRepo.On("MarkReinvested", 2, 10).Return(nil)

	err := job.Run()
	assert.NoError(t, err)
	assert.Equal(t, 1, job.GetExecutedCount())
	mockTradeService.AssertExpectations(t)
	mockDividendRepo.AssertExpectations(t)
}

func TestExecuteDividendTradesJob_Run_NoTradeService(t *testing.T) {
	mockDividendRepo := new(MockDividendRepositoryForUnreinvested)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewExecuteDividendTradesJob(mockDividendRepo, nil)
	job.SetLogger(log)

	recommendations := []domain.HolisticStep{
		{Symbol: "AAPL", Side: "BUY", Quantity: 10},
	}
	job.SetRecommendations(recommendations, map[string][]int{})

	err := job.Run()
	assert.NoError(t, err) // Job handles missing service gracefully
}
