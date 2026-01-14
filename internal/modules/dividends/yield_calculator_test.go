package dividends

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Mock Repositories for Testing
// ============================================================================

// MockDividendRepo implements DividendRepositoryInterface for testing
type MockDividendRepo struct {
	dividends map[string][]DividendRecord
}

func NewMockDividendRepo() *MockDividendRepo {
	return &MockDividendRepo{
		dividends: make(map[string][]DividendRecord),
	}
}

func (m *MockDividendRepo) GetByISIN(isin string) ([]DividendRecord, error) {
	if divs, ok := m.dividends[isin]; ok {
		return divs, nil
	}
	return []DividendRecord{}, nil
}

// Stub implementations for interface compliance
func (m *MockDividendRepo) Create(dividend *DividendRecord) error   { return nil }
func (m *MockDividendRepo) GetByID(id int) (*DividendRecord, error) { return nil, nil }
func (m *MockDividendRepo) GetByCashFlowID(cashFlowID int) (*DividendRecord, error) {
	return nil, nil
}
func (m *MockDividendRepo) ExistsForCashFlow(cashFlowID int) (bool, error)      { return false, nil }
func (m *MockDividendRepo) GetBySymbol(symbol string) ([]DividendRecord, error) { return nil, nil }
func (m *MockDividendRepo) GetByIdentifier(identifier string) ([]DividendRecord, error) {
	return nil, nil
}
func (m *MockDividendRepo) GetAll(limit int) ([]DividendRecord, error)          { return nil, nil }
func (m *MockDividendRepo) GetPendingBonuses() (map[string]float64, error)      { return nil, nil }
func (m *MockDividendRepo) GetPendingBonus(symbol string) (float64, error)      { return 0, nil }
func (m *MockDividendRepo) MarkReinvested(dividendID int, quantity int) error   { return nil }
func (m *MockDividendRepo) SetPendingBonus(dividendID int, bonus float64) error { return nil }
func (m *MockDividendRepo) ClearBonus(symbol string) (int, error)               { return 0, nil }
func (m *MockDividendRepo) GetUnreinvestedDividends(minAmountEUR float64) ([]DividendRecord, error) {
	return nil, nil
}
func (m *MockDividendRepo) GetTotalDividendsBySymbol() (map[string]float64, error) { return nil, nil }
func (m *MockDividendRepo) GetTotalReinvested() (float64, error)                   { return 0, nil }
func (m *MockDividendRepo) GetReinvestmentRate() (float64, error)                  { return 0, nil }

// MockPositionRepo provides market value for yield calculation
type MockPositionRepo struct {
	positions map[string]MockPosition
}

type MockPosition struct {
	MarketValueEUR float64
}

func NewMockPositionRepo() *MockPositionRepo {
	return &MockPositionRepo{
		positions: make(map[string]MockPosition),
	}
}

func (m *MockPositionRepo) GetMarketValueByISIN(isin string) (float64, error) {
	if pos, ok := m.positions[isin]; ok {
		return pos.MarketValueEUR, nil
	}
	return 0, nil
}

// ============================================================================
// YieldCalculator.CalculateYield Tests
// ============================================================================

func TestYieldCalculator_CalculateYield_QuarterlyDividends(t *testing.T) {
	divRepo := NewMockDividendRepo()
	posRepo := NewMockPositionRepo()
	log := zerolog.Nop()

	calculator := NewDividendYieldCalculator(divRepo, posRepo, log)

	// Setup: 4 quarterly dividends over last year (€25 each = €100 total)
	now := time.Now()
	isin := "US0378331005" // Apple ISIN
	divRepo.dividends[isin] = []DividendRecord{
		makeDividend(now.AddDate(0, -3, 0), 25.0),
		makeDividend(now.AddDate(0, -6, 0), 25.0),
		makeDividend(now.AddDate(0, -9, 0), 25.0),
		makeDividend(now.AddDate(0, -12, 0), 25.0),
	}
	posRepo.positions[isin] = MockPosition{MarketValueEUR: 2500.0}

	result, err := calculator.CalculateYield(isin)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Yield = 100 / 2500 = 0.04 = 4%
	assert.InDelta(t, 0.04, result.CurrentYield, 0.001)
}

func TestYieldCalculator_CalculateYield_AnnualDividend(t *testing.T) {
	divRepo := NewMockDividendRepo()
	posRepo := NewMockPositionRepo()
	log := zerolog.Nop()

	calculator := NewDividendYieldCalculator(divRepo, posRepo, log)

	// Setup: Single annual dividend of €50
	now := time.Now()
	isin := "DE0007164600" // SAP ISIN
	divRepo.dividends[isin] = []DividendRecord{
		makeDividend(now.AddDate(0, -6, 0), 50.0),
	}
	posRepo.positions[isin] = MockPosition{MarketValueEUR: 1000.0}

	result, err := calculator.CalculateYield(isin)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Yield = 50 / 1000 = 0.05 = 5%
	assert.InDelta(t, 0.05, result.CurrentYield, 0.001)
}

func TestYieldCalculator_CalculateYield_NoDividends(t *testing.T) {
	divRepo := NewMockDividendRepo()
	posRepo := NewMockPositionRepo()
	log := zerolog.Nop()

	calculator := NewDividendYieldCalculator(divRepo, posRepo, log)

	isin := "US5949181045" // Microsoft ISIN (no dividends in mock)
	posRepo.positions[isin] = MockPosition{MarketValueEUR: 5000.0}

	result, err := calculator.CalculateYield(isin)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0.0, result.CurrentYield)
	assert.Equal(t, 0.0, result.FiveYearAvgYield)
	assert.Equal(t, 0.0, result.GrowthRate)
}

func TestYieldCalculator_CalculateYield_NoPosition(t *testing.T) {
	divRepo := NewMockDividendRepo()
	posRepo := NewMockPositionRepo()
	log := zerolog.Nop()

	calculator := NewDividendYieldCalculator(divRepo, posRepo, log)

	// Dividends exist but no position (market value = 0)
	now := time.Now()
	isin := "GB00B03MLX29"
	divRepo.dividends[isin] = []DividendRecord{
		makeDividend(now.AddDate(0, -3, 0), 100.0),
	}
	// No position set = market value 0

	result, err := calculator.CalculateYield(isin)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Cannot calculate yield without market value
	assert.Equal(t, 0.0, result.CurrentYield)
}

func TestYieldCalculator_CalculateYield_PartialYearData(t *testing.T) {
	divRepo := NewMockDividendRepo()
	posRepo := NewMockPositionRepo()
	log := zerolog.Nop()

	calculator := NewDividendYieldCalculator(divRepo, posRepo, log)

	// Setup: Only 6 months of dividend data (2 quarters @ €30 each)
	now := time.Now()
	isin := "FR0000120578" // Sanofi ISIN
	divRepo.dividends[isin] = []DividendRecord{
		makeDividend(now.AddDate(0, -2, 0), 30.0),
		makeDividend(now.AddDate(0, -5, 0), 30.0),
	}
	posRepo.positions[isin] = MockPosition{MarketValueEUR: 1500.0}

	result, err := calculator.CalculateYield(isin)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Yield based on available data: 60 / 1500 = 0.04 = 4%
	// The calculator should use actual dividends, not annualized
	assert.InDelta(t, 0.04, result.CurrentYield, 0.001)
}

func TestYieldCalculator_CalculateYield_FiveYearAverage(t *testing.T) {
	divRepo := NewMockDividendRepo()
	posRepo := NewMockPositionRepo()
	log := zerolog.Nop()

	calculator := NewDividendYieldCalculator(divRepo, posRepo, log)

	// Setup: 5 years of dividends with increasing amounts
	// Year 1: €80, Year 2: €90, Year 3: €100, Year 4: €110, Year 5: €120
	now := time.Now()
	isin := "US0378331005"
	divRepo.dividends[isin] = []DividendRecord{
		// Year 5 (most recent)
		makeDividend(now.AddDate(0, -3, 0), 30.0),
		makeDividend(now.AddDate(0, -6, 0), 30.0),
		makeDividend(now.AddDate(0, -9, 0), 30.0),
		makeDividend(now.AddDate(0, -12, 0), 30.0),
		// Year 4
		makeDividend(now.AddDate(-1, -3, 0), 27.5),
		makeDividend(now.AddDate(-1, -6, 0), 27.5),
		makeDividend(now.AddDate(-1, -9, 0), 27.5),
		makeDividend(now.AddDate(-1, -12, 0), 27.5),
		// Year 3
		makeDividend(now.AddDate(-2, -3, 0), 25.0),
		makeDividend(now.AddDate(-2, -6, 0), 25.0),
		makeDividend(now.AddDate(-2, -9, 0), 25.0),
		makeDividend(now.AddDate(-2, -12, 0), 25.0),
		// Year 2
		makeDividend(now.AddDate(-3, -3, 0), 22.5),
		makeDividend(now.AddDate(-3, -6, 0), 22.5),
		makeDividend(now.AddDate(-3, -9, 0), 22.5),
		makeDividend(now.AddDate(-3, -12, 0), 22.5),
		// Year 1 (oldest)
		makeDividend(now.AddDate(-4, -3, 0), 20.0),
		makeDividend(now.AddDate(-4, -6, 0), 20.0),
		makeDividend(now.AddDate(-4, -9, 0), 20.0),
		makeDividend(now.AddDate(-4, -12, 0), 20.0),
	}
	posRepo.positions[isin] = MockPosition{MarketValueEUR: 3000.0}

	result, err := calculator.CalculateYield(isin)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Current yield = 120 / 3000 = 4%
	assert.InDelta(t, 0.04, result.CurrentYield, 0.001)
	// 5-year average yield = (80+90+100+110+120) / 5 / 3000 = 100 / 3000 = 3.33%
	assert.InDelta(t, 0.0333, result.FiveYearAvgYield, 0.002)
}

func TestYieldCalculator_CalculateYield_GrowthRate(t *testing.T) {
	divRepo := NewMockDividendRepo()
	posRepo := NewMockPositionRepo()
	log := zerolog.Nop()

	calculator := NewDividendYieldCalculator(divRepo, posRepo, log)

	// Setup: 2 years of dividends showing 20% growth
	// Year 0 (0-12 months): €120, Year 1 (12-24 months): €100
	now := time.Now()
	isin := "US0378331005"
	divRepo.dividends[isin] = []DividendRecord{
		// Current year (year 0): €120 - within last 12 months
		makeDividend(now.AddDate(0, -2, 0), 30.0),  // 2 months ago
		makeDividend(now.AddDate(0, -5, 0), 30.0),  // 5 months ago
		makeDividend(now.AddDate(0, -8, 0), 30.0),  // 8 months ago
		makeDividend(now.AddDate(0, -11, 0), 30.0), // 11 months ago
		// Previous year (year 1): €100 - 12-24 months ago
		makeDividend(now.AddDate(-1, -2, 0), 25.0),  // 14 months ago
		makeDividend(now.AddDate(-1, -5, 0), 25.0),  // 17 months ago
		makeDividend(now.AddDate(-1, -8, 0), 25.0),  // 20 months ago
		makeDividend(now.AddDate(-1, -11, 0), 25.0), // 23 months ago
	}
	posRepo.positions[isin] = MockPosition{MarketValueEUR: 3000.0}

	result, err := calculator.CalculateYield(isin)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Growth rate = (120 - 100) / 100 = 0.20 = 20%
	assert.InDelta(t, 0.20, result.GrowthRate, 0.01)
}

func TestYieldCalculator_CalculateYield_NegativeGrowthRate(t *testing.T) {
	divRepo := NewMockDividendRepo()
	posRepo := NewMockPositionRepo()
	log := zerolog.Nop()

	calculator := NewDividendYieldCalculator(divRepo, posRepo, log)

	// Setup: Dividend cut - previous year €100, current year €80
	now := time.Now()
	isin := "US0378331005"
	divRepo.dividends[isin] = []DividendRecord{
		// Current year (year 0): €80 - within last 12 months
		makeDividend(now.AddDate(0, -2, 0), 20.0),  // 2 months ago
		makeDividend(now.AddDate(0, -5, 0), 20.0),  // 5 months ago
		makeDividend(now.AddDate(0, -8, 0), 20.0),  // 8 months ago
		makeDividend(now.AddDate(0, -11, 0), 20.0), // 11 months ago
		// Previous year (year 1): €100 - 12-24 months ago
		makeDividend(now.AddDate(-1, -2, 0), 25.0),  // 14 months ago
		makeDividend(now.AddDate(-1, -5, 0), 25.0),  // 17 months ago
		makeDividend(now.AddDate(-1, -8, 0), 25.0),  // 20 months ago
		makeDividend(now.AddDate(-1, -11, 0), 25.0), // 23 months ago
	}
	posRepo.positions[isin] = MockPosition{MarketValueEUR: 2000.0}

	result, err := calculator.CalculateYield(isin)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Growth rate = (80 - 100) / 100 = -0.20 = -20%
	assert.InDelta(t, -0.20, result.GrowthRate, 0.01)
}

func TestYieldCalculator_CalculateYield_SingleYearNoGrowth(t *testing.T) {
	divRepo := NewMockDividendRepo()
	posRepo := NewMockPositionRepo()
	log := zerolog.Nop()

	calculator := NewDividendYieldCalculator(divRepo, posRepo, log)

	// Setup: Only 1 year of data - cannot calculate growth
	now := time.Now()
	isin := "US0378331005"
	divRepo.dividends[isin] = []DividendRecord{
		makeDividend(now.AddDate(0, -3, 0), 25.0),
		makeDividend(now.AddDate(0, -6, 0), 25.0),
		makeDividend(now.AddDate(0, -9, 0), 25.0),
		makeDividend(now.AddDate(0, -12, 0), 25.0),
	}
	posRepo.positions[isin] = MockPosition{MarketValueEUR: 2000.0}

	result, err := calculator.CalculateYield(isin)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Cannot calculate growth with only 1 year
	assert.Equal(t, 0.0, result.GrowthRate)
}

func TestYieldCalculator_CalculateYield_OldDividendsExcluded(t *testing.T) {
	divRepo := NewMockDividendRepo()
	posRepo := NewMockPositionRepo()
	log := zerolog.Nop()

	calculator := NewDividendYieldCalculator(divRepo, posRepo, log)

	// Setup: Mix of recent and very old dividends
	now := time.Now()
	isin := "US0378331005"
	divRepo.dividends[isin] = []DividendRecord{
		// Recent (last 12 months): €100
		makeDividend(now.AddDate(0, -3, 0), 50.0),
		makeDividend(now.AddDate(0, -9, 0), 50.0),
		// Very old (6+ years ago) - should NOT be included in 5-year calc
		makeDividend(now.AddDate(-7, 0, 0), 1000.0),
		makeDividend(now.AddDate(-8, 0, 0), 1000.0),
	}
	posRepo.positions[isin] = MockPosition{MarketValueEUR: 2500.0}

	result, err := calculator.CalculateYield(isin)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Current yield only uses last 12 months: 100 / 2500 = 4%
	assert.InDelta(t, 0.04, result.CurrentYield, 0.001)
}

// ============================================================================
// Helper Functions
// ============================================================================

// makeDividend creates a DividendRecord with the given payment date and amount
func makeDividend(paymentDate time.Time, amountEUR float64) DividendRecord {
	ts := paymentDate.Unix()
	return DividendRecord{
		PaymentDate: &ts,
		AmountEUR:   amountEUR,
		Amount:      amountEUR,
		Currency:    "EUR",
		Symbol:      "TEST",
	}
}
