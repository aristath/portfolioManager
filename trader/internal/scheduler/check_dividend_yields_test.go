package scheduler

import (
	"testing"

	"github.com/aristath/sentinel/internal/modules/dividends"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSecurityRepoForDividendsYields is a mock for SecurityRepositoryForDividendsInterface
type MockSecurityRepoForDividendsYields struct {
	mock.Mock
}

func (m *MockSecurityRepoForDividendsYields) GetBySymbol(symbol string) (*SecurityForDividends, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SecurityForDividends), args.Error(1)
}

// MockYahooClientForDividendsYields is a mock for YahooClientForDividendsInterface
type MockYahooClientForDividendsYields struct {
	mock.Mock
}

func (m *MockYahooClientForDividendsYields) GetCurrentPrice(symbol string, yahooSymbolOverride *string, maxRetries int) (*float64, error) {
	args := m.Called(symbol, yahooSymbolOverride, maxRetries)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*float64), args.Error(1)
}

func (m *MockYahooClientForDividendsYields) GetFundamentalData(symbol string, yahooSymbolOverride *string) (*FundamentalDataForDividends, error) {
	args := m.Called(symbol, yahooSymbolOverride)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*FundamentalDataForDividends), args.Error(1)
}

func TestCheckDividendYieldsJob_Name(t *testing.T) {
	job := &CheckDividendYieldsJob{
		log: zerolog.Nop(),
	}
	assert.Equal(t, "check_dividend_yields", job.Name())
}

func TestCheckDividendYieldsJob_Run_Success(t *testing.T) {
	mockSecurityRepo := new(MockSecurityRepoForDividendsYields)
	mockYahooClient := new(MockYahooClientForDividendsYields)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewCheckDividendYieldsJob(mockSecurityRepo, mockYahooClient)
	job.SetLogger(log)

	dividends := []dividends.DividendRecord{
		{ID: 1, Symbol: "AAPL", AmountEUR: 10.0},
	}
	grouped := map[string]SymbolDividendInfoForGroup{
		"AAPL": {
			Dividends:     dividends,
			DividendIDs:   []int{1},
			TotalAmount:   10.0,
			DividendCount: 1,
		},
	}
	job.SetGroupedDividends(grouped)

	security := &SecurityForDividends{
		Symbol:      "AAPL",
		YahooSymbol: "AAPL",
		Name:        "Apple Inc.",
		Currency:    "USD",
		MinLot:      1,
	}
	dividendYield := 0.04 // 4%
	fundamentals := &FundamentalDataForDividends{
		DividendYield: &dividendYield,
	}

	mockSecurityRepo.On("GetBySymbol", "AAPL").Return(security, nil)
	mockYahooClient.On("GetFundamentalData", "AAPL", (*string)(nil)).Return(fundamentals, nil)

	err := job.Run()
	assert.NoError(t, err)

	results := job.GetYieldResults()
	assert.Equal(t, 0.04, results["AAPL"].Yield)
	assert.True(t, results["AAPL"].IsHighYield) // >= 3%
	assert.True(t, results["AAPL"].IsAvailable)
}

func TestCheckDividendYieldsJob_Run_NoGroupedDividends(t *testing.T) {
	mockSecurityRepo := new(MockSecurityRepoForDividendsYields)
	mockYahooClient := new(MockYahooClientForDividendsYields)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewCheckDividendYieldsJob(mockSecurityRepo, mockYahooClient)
	job.SetLogger(log)

	err := job.Run()
	assert.NoError(t, err)
}
