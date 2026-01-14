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

// MockDividendYieldCalculator is a mock for dividend yield calculation
type MockDividendYieldCalculator struct {
	mock.Mock
}

func (m *MockDividendYieldCalculator) CalculateYield(isin string) (*dividends.YieldResult, error) {
	args := m.Called(isin)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dividends.YieldResult), args.Error(1)
}

func TestCheckDividendYieldsJob_Name(t *testing.T) {
	job := &CheckDividendYieldsJob{
		log: zerolog.Nop(),
	}
	assert.Equal(t, "check_dividend_yields", job.Name())
}

func TestCheckDividendYieldsJob_Run_Success(t *testing.T) {
	mockSecurityRepo := new(MockSecurityRepoForDividendsYields)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	// Since we can't easily mock the DividendYieldCalculator (it's a concrete type),
	// we test the job with nil yieldCalculator which should handle gracefully
	job := NewCheckDividendYieldsJob(mockSecurityRepo, nil)
	job.SetLogger(log)

	dividendRecords := []dividends.DividendRecord{
		{ID: 1, Symbol: "AAPL", AmountEUR: 10.0},
	}
	grouped := map[string]SymbolDividendInfoForGroup{
		"AAPL": {
			Dividends:     dividendRecords,
			DividendIDs:   []int{1},
			TotalAmount:   10.0,
			DividendCount: 1,
		},
	}
	job.SetGroupedDividends(grouped)

	err := job.Run()
	assert.NoError(t, err)

	// With nil yieldCalculator, yields should be -1.0 (unavailable)
	results := job.GetYieldResults()
	assert.Equal(t, -1.0, results["AAPL"].Yield)
	assert.False(t, results["AAPL"].IsHighYield)
	assert.False(t, results["AAPL"].IsAvailable)
}

func TestCheckDividendYieldsJob_Run_NoGroupedDividends(t *testing.T) {
	mockSecurityRepo := new(MockSecurityRepoForDividendsYields)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewCheckDividendYieldsJob(mockSecurityRepo, nil)
	job.SetLogger(log)

	err := job.Run()
	assert.NoError(t, err)
}

func TestCheckDividendYieldsJob_GetHighLowYieldSymbols(t *testing.T) {
	mockSecurityRepo := new(MockSecurityRepoForDividendsYields)
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewCheckDividendYieldsJob(mockSecurityRepo, nil)
	job.SetLogger(log)

	// Initialize empty maps
	assert.Empty(t, job.GetHighYieldSymbols())
	assert.Empty(t, job.GetLowYieldSymbols())
}
