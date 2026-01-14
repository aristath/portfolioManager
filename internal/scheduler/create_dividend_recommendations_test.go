package scheduler

import (
	"errors"
	"testing"

	"github.com/aristath/sentinel/internal/modules/dividends"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCurrentPriceProvider is a mock for CurrentPriceProviderInterface
type MockCurrentPriceProvider struct {
	mock.Mock
}

func (m *MockCurrentPriceProvider) GetCurrentPrice(symbol string) (*float64, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*float64), args.Error(1)
}

func TestCreateDividendRecommendationsJob_Name(t *testing.T) {
	job := &CreateDividendRecommendationsJob{
		log: zerolog.Nop(),
	}
	assert.Equal(t, "create_dividend_recommendations", job.Name())
}

func TestCreateDividendRecommendationsJob_Run_NoHighYieldSymbols(t *testing.T) {
	mockSecurityRepo := new(MockSecurityRepoForDividendsYields)
	mockPriceProvider := new(MockCurrentPriceProvider)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewCreateDividendRecommendationsJob(mockSecurityRepo, mockPriceProvider, nil, 200.0)
	job.SetLogger(log)

	err := job.Run()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(job.GetRecommendations()))
}

func TestCreateDividendRecommendationsJob_Run_Success(t *testing.T) {
	mockSecurityRepo := new(MockSecurityRepoForDividendsYields)
	mockPriceProvider := new(MockCurrentPriceProvider)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewCreateDividendRecommendationsJob(mockSecurityRepo, mockPriceProvider, nil, 200.0)
	job.SetLogger(log)

	dividends := []dividends.DividendRecord{
		{ID: 1, Symbol: "AAPL", AmountEUR: 10.0},
	}
	highYieldSymbols := map[string]SymbolDividendInfoForGroup{
		"AAPL": {
			Dividends:     dividends,
			DividendIDs:   []int{1},
			TotalAmount:   1000.0, // Above min trade size
			DividendCount: 1,
		},
	}
	job.SetHighYieldSymbols(highYieldSymbols)

	security := &SecurityForDividends{
		ISIN:     "US0378331005",
		Symbol:   "AAPL",
		Name:     "Apple Inc.",
		Currency: "USD",
		MinLot:   1,
	}
	price := 150.0

	mockSecurityRepo.On("GetBySymbol", "AAPL").Return(security, nil)
	mockPriceProvider.On("GetCurrentPrice", "AAPL").Return(&price, nil)

	err := job.Run()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(job.GetRecommendations()))
	rec := job.GetRecommendations()[0]
	assert.Equal(t, "AAPL", rec.Symbol)
	assert.Equal(t, "BUY", rec.Side)
	mockSecurityRepo.AssertExpectations(t)
	mockPriceProvider.AssertExpectations(t)
}

func TestCreateDividendRecommendationsJob_Run_BelowMinTradeSize(t *testing.T) {
	mockSecurityRepo := new(MockSecurityRepoForDividendsYields)
	mockPriceProvider := new(MockCurrentPriceProvider)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewCreateDividendRecommendationsJob(mockSecurityRepo, mockPriceProvider, nil, 200.0)
	job.SetLogger(log)

	dividends := []dividends.DividendRecord{
		{ID: 1, Symbol: "AAPL", AmountEUR: 10.0},
	}
	highYieldSymbols := map[string]SymbolDividendInfoForGroup{
		"AAPL": {
			Dividends:     dividends,
			DividendIDs:   []int{1},
			TotalAmount:   100.0, // Below min trade size
			DividendCount: 1,
		},
	}
	job.SetHighYieldSymbols(highYieldSymbols)

	err := job.Run()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(job.GetRecommendations()))
}

func TestCreateDividendRecommendationsJob_Run_SecurityNotFound(t *testing.T) {
	mockSecurityRepo := new(MockSecurityRepoForDividendsYields)
	mockPriceProvider := new(MockCurrentPriceProvider)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewCreateDividendRecommendationsJob(mockSecurityRepo, mockPriceProvider, nil, 200.0)
	job.SetLogger(log)

	dividends := []dividends.DividendRecord{
		{ID: 1, Symbol: "AAPL", AmountEUR: 10.0},
	}
	highYieldSymbols := map[string]SymbolDividendInfoForGroup{
		"AAPL": {
			Dividends:     dividends,
			DividendIDs:   []int{1},
			TotalAmount:   1000.0,
			DividendCount: 1,
		},
	}
	job.SetHighYieldSymbols(highYieldSymbols)

	mockSecurityRepo.On("GetBySymbol", "AAPL").Return(nil, errors.New("not found"))

	err := job.Run()
	assert.NoError(t, err) // Job continues even if one fails
	assert.Equal(t, 0, len(job.GetRecommendations()))
	mockSecurityRepo.AssertExpectations(t)
}
