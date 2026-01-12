package scheduler

import (
	"testing"
	"time"

	"github.com/aristath/sentinel/internal/modules/portfolio"
	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockScoresRepoForContext is a mock implementation of ScoresRepositoryInterface
type MockScoresRepoForContext struct {
	GetCAGRsFunc         func(isinList []string) (map[string]float64, error)
	GetQualityScoresFunc func(isinList []string) (map[string]float64, map[string]float64, error)
	GetValueTrapDataFunc func(isinList []string) (map[string]float64, map[string]float64, map[string]float64, error)
	GetTotalScoresFunc   func(isinList []string) (map[string]float64, error)
	GetRiskMetricsFunc   func(isinList []string) (map[string]float64, map[string]float64, error)
}

func (m *MockScoresRepoForContext) GetCAGRs(isinList []string) (map[string]float64, error) {
	if m.GetCAGRsFunc != nil {
		return m.GetCAGRsFunc(isinList)
	}
	return map[string]float64{}, nil
}

func (m *MockScoresRepoForContext) GetQualityScores(isinList []string) (map[string]float64, map[string]float64, error) {
	if m.GetQualityScoresFunc != nil {
		return m.GetQualityScoresFunc(isinList)
	}
	return map[string]float64{}, map[string]float64{}, nil
}

func (m *MockScoresRepoForContext) GetValueTrapData(isinList []string) (map[string]float64, map[string]float64, map[string]float64, error) {
	if m.GetValueTrapDataFunc != nil {
		return m.GetValueTrapDataFunc(isinList)
	}
	return map[string]float64{}, map[string]float64{}, map[string]float64{}, nil
}

func (m *MockScoresRepoForContext) GetTotalScores(isinList []string) (map[string]float64, error) {
	if m.GetTotalScoresFunc != nil {
		return m.GetTotalScoresFunc(isinList)
	}
	return map[string]float64{}, nil
}

func (m *MockScoresRepoForContext) GetRiskMetrics(isinList []string) (map[string]float64, map[string]float64, error) {
	if m.GetRiskMetricsFunc != nil {
		return m.GetRiskMetricsFunc(isinList)
	}
	return map[string]float64{}, map[string]float64{}, nil
}

// MockSettingsRepoForContext is a mock implementation of SettingsRepositoryInterface
type MockSettingsRepoForContext struct {
	GetTargetReturnSettingsFunc func() (float64, float64, error)
	GetVirtualTestCashFunc      func() (float64, error)
}

func (m *MockSettingsRepoForContext) GetTargetReturnSettings() (float64, float64, error) {
	if m.GetTargetReturnSettingsFunc != nil {
		return m.GetTargetReturnSettingsFunc()
	}
	return 0.11, 0.80, nil
}

func (m *MockSettingsRepoForContext) GetVirtualTestCash() (float64, error) {
	if m.GetVirtualTestCashFunc != nil {
		return m.GetVirtualTestCashFunc()
	}
	return 0.0, nil
}

// MockRegimeRepoForContext is a mock implementation of RegimeRepositoryInterface
type MockRegimeRepoForContext struct {
	GetCurrentRegimeScoreFunc func() (float64, error)
}

func (m *MockRegimeRepoForContext) GetCurrentRegimeScore() (float64, error) {
	if m.GetCurrentRegimeScoreFunc != nil {
		return m.GetCurrentRegimeScoreFunc()
	}
	return 0.0, nil
}

// MockGroupingRepoForContext is a mock implementation of GroupingRepositoryInterface
type MockGroupingRepoForContext struct {
	GetCountryGroupsFunc  func() (map[string][]string, error)
	GetIndustryGroupsFunc func() (map[string][]string, error)
}

func (m *MockGroupingRepoForContext) GetCountryGroups() (map[string][]string, error) {
	if m.GetCountryGroupsFunc != nil {
		return m.GetCountryGroupsFunc()
	}
	return map[string][]string{}, nil
}

func (m *MockGroupingRepoForContext) GetIndustryGroups() (map[string][]string, error) {
	if m.GetIndustryGroupsFunc != nil {
		return m.GetIndustryGroupsFunc()
	}
	return map[string][]string{}, nil
}

func TestBuildOpportunityContextJob_Name(t *testing.T) {
	job := NewBuildOpportunityContextJob(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	assert.Equal(t, "build_opportunity_context", job.Name())
}

func TestBuildOpportunityContextJob_Run_Success(t *testing.T) {
	// Mock position repo
	positions := []portfolio.Position{
		{Symbol: "AAPL", ISIN: "US0378331005", Quantity: 10, Currency: "USD", AvgPrice: 150.0, CostBasisEUR: 1363.64, CurrencyRate: 1.1},
	}
	positionsInterface := make([]interface{}, len(positions))
	for i := range positions {
		positionsInterface[i] = positions[i]
	}

	mockPositionRepo := &MockPositionRepoForOptimizer{
		GetAllFunc: func() ([]interface{}, error) {
			return positionsInterface, nil
		},
	}

	// Mock security repo
	securities := []universe.Security{
		{Symbol: "AAPL", ISIN: "US0378331005", YahooSymbol: "AAPL", Country: "US", Industry: "Technology", Active: true, Name: "Apple Inc."},
	}
	securitiesInterface := make([]interface{}, len(securities))
	for i := range securities {
		securitiesInterface[i] = securities[i]
	}

	mockSecurityRepo := &MockSecurityRepoForOptimizer{
		GetAllActiveFunc: func() ([]interface{}, error) {
			return securitiesInterface, nil
		},
	}

	// Mock allocation repo
	mockAllocRepo := &MockAllocationRepoForOptimizer{
		GetAllFunc: func() (map[string]float64, error) {
			return map[string]float64{
				"country_group:US":          0.5,
				"industry_group:Technology": 0.3,
			}, nil
		},
	}

	// Mock cash manager
	mockCashManager := &MockCashManagerForOptimizer{
		GetAllCashBalancesFunc: func() (map[string]float64, error) {
			return map[string]float64{"EUR": 1000.0}, nil
		},
	}

	// Mock price client
	mockPriceClient := &MockPriceClient{
		GetBatchQuotesFunc: func(symbolMap map[string]*string) (map[string]*float64, error) {
			return map[string]*float64{
				"AAPL": floatPtr(150.0),
			}, nil
		},
	}

	// Mock scores repo
	mockScoresRepo := &MockScoresRepoForContext{
		GetCAGRsFunc: func(isinList []string) (map[string]float64, error) {
			return map[string]float64{
				"US0378331005": 0.12,
				"AAPL":         0.12,
			}, nil
		},
		GetQualityScoresFunc: func(isinList []string) (map[string]float64, map[string]float64, error) {
			return map[string]float64{
					"US0378331005": 0.8,
					"AAPL":         0.8,
				}, map[string]float64{
					"US0378331005": 0.75,
					"AAPL":         0.75,
				}, nil
		},
		GetValueTrapDataFunc: func(isinList []string) (map[string]float64, map[string]float64, map[string]float64, error) {
			return map[string]float64{
					"US0378331005": 0.7,
					"AAPL":         0.7,
				}, map[string]float64{
					"US0378331005": 0.5,
					"AAPL":         0.5,
				}, map[string]float64{
					"US0378331005": 0.25,
					"AAPL":         0.25,
				}, nil
		},
	}

	// Mock settings repo
	mockSettingsRepo := &MockSettingsRepoForContext{
		GetTargetReturnSettingsFunc: func() (float64, float64, error) {
			return 0.11, 0.80, nil
		},
		GetVirtualTestCashFunc: func() (float64, error) {
			return 0.0, nil
		},
	}

	// Mock regime repo
	mockRegimeRepo := &MockRegimeRepoForContext{
		GetCurrentRegimeScoreFunc: func() (float64, error) {
			return 0.5, nil
		},
	}

	mockGroupingRepo := &MockGroupingRepoForContext{}
	job := NewBuildOpportunityContextJob(
		mockPositionRepo,
		mockSecurityRepo,
		mockAllocRepo,
		mockGroupingRepo,
		mockCashManager,
		mockPriceClient,
		nil, // priceConversionService
		mockScoresRepo,
		mockSettingsRepo,
		mockRegimeRepo,
	)

	err := job.Run()
	require.NoError(t, err)

	ctx := job.GetOpportunityContext()
	require.NotNil(t, ctx)
	assert.Equal(t, 2500.0, ctx.TotalPortfolioValueEUR) // 1000 EUR + (10 * 150 USD)
	assert.Equal(t, 1000.0, ctx.AvailableCashEUR)
	assert.Equal(t, 1, len(ctx.EnrichedPositions))
	assert.Equal(t, 1, len(ctx.Securities))
}

func TestBuildOpportunityContextJob_Run_PositionRepoError(t *testing.T) {
	mockPositionRepo := &MockPositionRepoForOptimizer{
		GetAllFunc: func() ([]interface{}, error) {
			return nil, assert.AnError
		},
	}

	mockGroupingRepo := &MockGroupingRepoForContext{}
	job := NewBuildOpportunityContextJob(
		mockPositionRepo,
		nil,
		nil,
		mockGroupingRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	err := job.Run()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get positions")
}

func TestBuildOpportunityContextJob_Run_SecurityRepoError(t *testing.T) {
	mockPositionRepo := &MockPositionRepoForOptimizer{
		GetAllFunc: func() ([]interface{}, error) {
			return []interface{}{}, nil
		},
	}

	mockSecurityRepo := &MockSecurityRepoForOptimizer{
		GetAllActiveFunc: func() ([]interface{}, error) {
			return nil, assert.AnError
		},
	}

	mockGroupingRepo := &MockGroupingRepoForContext{}
	job := NewBuildOpportunityContextJob(
		mockPositionRepo,
		mockSecurityRepo,
		nil,
		mockGroupingRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	err := job.Run()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get securities")
}

func TestBuildOpportunityContextJob_Run_AllocationRepoError(t *testing.T) {
	mockPositionRepo := &MockPositionRepoForOptimizer{
		GetAllFunc: func() ([]interface{}, error) {
			return []interface{}{}, nil
		},
	}

	mockSecurityRepo := &MockSecurityRepoForOptimizer{
		GetAllActiveFunc: func() ([]interface{}, error) {
			return []interface{}{}, nil
		},
	}

	mockAllocRepo := &MockAllocationRepoForOptimizer{
		GetAllFunc: func() (map[string]float64, error) {
			return nil, assert.AnError
		},
	}

	mockGroupingRepo := &MockGroupingRepoForContext{}
	job := NewBuildOpportunityContextJob(
		mockPositionRepo,
		mockSecurityRepo,
		mockAllocRepo,
		mockGroupingRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	err := job.Run()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get allocations")
}

// ============================================================================
// fetchCurrentPrices() EUR Conversion Tests
// ============================================================================

// MockPriceClientForConversion is a mock implementation for price fetching tests
type MockPriceClientForConversion struct {
	quotes map[string]*float64
	err    error
}

func (m *MockPriceClientForConversion) GetBatchQuotes(symbolMap map[string]*string) (map[string]*float64, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.quotes, nil
}

// MockPriceConversionServiceForScheduler is a mock implementation of PriceConversionServiceInterface
type MockPriceConversionServiceForScheduler struct {
	convertFunc func(prices map[string]float64, securities []universe.Security) map[string]float64
}

func (m *MockPriceConversionServiceForScheduler) ConvertPricesToEUR(
	prices map[string]float64,
	securities []universe.Security,
) map[string]float64 {
	if m.convertFunc != nil {
		return m.convertFunc(prices, securities)
	}
	return prices
}

// TestFetchCurrentPrices_HKD_Conversion verifies HKD prices are converted to EUR
func TestFetchCurrentPrices_HKD_Conversion(t *testing.T) {
	// Setup: HKD security with price 497.4 HKD
	isin := "KYG2108Y1052" // CAT ISIN
	securities := []universe.Security{
		{Symbol: "CAT.3750.AS", ISIN: isin, Currency: "HKD"},
	}

	// Mock Yahoo returns 497.4 HKD
	hkdPrice := 497.4
	priceClient := &MockPriceClientForConversion{
		quotes: map[string]*float64{
			"CAT.3750.AS": &hkdPrice,
		},
	}

	// Mock price conversion service: converts HKD to EUR (1 HKD = 0.11 EUR)
	priceConversionService := &MockPriceConversionServiceForScheduler{
		convertFunc: func(prices map[string]float64, secs []universe.Security) map[string]float64 {
			converted := make(map[string]float64)
			for symbol, price := range prices {
				if symbol == "CAT.3750.AS" {
					converted[symbol] = price * 0.11 // Convert HKD to EUR
				} else {
					converted[symbol] = price
				}
			}
			return converted
		},
	}

	job := &BuildOpportunityContextJob{
		priceClient:            priceClient,
		priceConversionService: priceConversionService,
	}

	// Execute
	prices := job.fetchCurrentPrices(securities)

	// Verify: Price converted to EUR (prices map uses ISIN keys)
	eurPrice := prices[isin]
	expected := 497.4 * 0.11 // = 54.714
	assert.InDelta(t, expected, eurPrice, 0.01, "HKD price should be converted to EUR")
}

// TestFetchCurrentPrices_EUR_NoConversion verifies EUR prices are not converted
func TestFetchCurrentPrices_EUR_NoConversion(t *testing.T) {
	isin := "NL0000009082" // VWS ISIN
	securities := []universe.Security{
		{Symbol: "VWS.AS", ISIN: isin, Currency: "EUR"},
	}

	eurPrice := 42.5
	priceClient := &MockPriceClientForConversion{
		quotes: map[string]*float64{
			"VWS.AS": &eurPrice,
		},
	}

	// Mock conversion service that doesn't change EUR prices
	priceConversionService := &MockPriceConversionServiceForScheduler{
		convertFunc: func(prices map[string]float64, secs []universe.Security) map[string]float64 {
			return prices // No conversion for EUR
		},
	}

	job := &BuildOpportunityContextJob{
		priceClient:            priceClient,
		priceConversionService: priceConversionService,
	}

	prices := job.fetchCurrentPrices(securities)

	// Verify: EUR price unchanged (prices map uses ISIN keys)
	assert.Equal(t, 42.5, prices[isin], "EUR price should not be converted")
}

// TestFetchCurrentPrices_USD_Conversion verifies USD prices are converted to EUR
func TestFetchCurrentPrices_USD_Conversion(t *testing.T) {
	isin := "US0378331005" // AAPL ISIN
	securities := []universe.Security{
		{Symbol: "AAPL", ISIN: isin, Currency: "USD"},
	}

	usdPrice := 150.0
	priceClient := &MockPriceClientForConversion{
		quotes: map[string]*float64{
			"AAPL": &usdPrice,
		},
	}

	// Mock conversion: 1 USD = 0.93 EUR
	priceConversionService := &MockPriceConversionServiceForScheduler{
		convertFunc: func(prices map[string]float64, secs []universe.Security) map[string]float64 {
			converted := make(map[string]float64)
			for symbol, price := range prices {
				if symbol == "AAPL" {
					converted[symbol] = price * 0.93 // Convert USD to EUR
				} else {
					converted[symbol] = price
				}
			}
			return converted
		},
	}

	job := &BuildOpportunityContextJob{
		priceClient:            priceClient,
		priceConversionService: priceConversionService,
	}

	prices := job.fetchCurrentPrices(securities)

	expected := 150.0 * 0.93 // = 139.5 EUR
	assert.InDelta(t, expected, prices[isin], 0.01, "USD price should be converted to EUR")
}

// TestFetchCurrentPrices_MissingRate_FallbackToNative verifies graceful fallback when rate unavailable
func TestFetchCurrentPrices_MissingRate_FallbackToNative(t *testing.T) {
	isin := "HK0000093390" // Tencent ISIN
	securities := []universe.Security{
		{Symbol: "0700.HK", ISIN: isin, Currency: "HKD"},
	}

	hkdPrice := 90.0
	priceClient := &MockPriceClientForConversion{
		quotes: map[string]*float64{
			"0700.HK": &hkdPrice,
		},
	}

	// Mock conversion service that returns native price when conversion fails
	priceConversionService := &MockPriceConversionServiceForScheduler{
		convertFunc: func(prices map[string]float64, secs []universe.Security) map[string]float64 {
			// Simulate missing rate - return native prices
			return prices
		},
	}

	job := &BuildOpportunityContextJob{
		priceClient:            priceClient,
		priceConversionService: priceConversionService,
	}

	prices := job.fetchCurrentPrices(securities)

	// Should fallback to native price (logged warning in real implementation)
	assert.Equal(t, 90.0, prices[isin], "Should use native price when rate unavailable")
}

// TestFetchCurrentPrices_NilPriceConversionService verifies graceful handling of nil service
func TestFetchCurrentPrices_NilPriceConversionService(t *testing.T) {
	isin := "KYG2108Y1052" // CAT ISIN
	securities := []universe.Security{
		{Symbol: "CAT.3750.AS", ISIN: isin, Currency: "HKD"},
	}

	hkdPrice := 497.4
	priceClient := &MockPriceClientForConversion{
		quotes: map[string]*float64{
			"CAT.3750.AS": &hkdPrice,
		},
	}

	job := &BuildOpportunityContextJob{
		priceClient:            priceClient,
		priceConversionService: nil, // Not injected
	}

	prices := job.fetchCurrentPrices(securities)

	// Should use native price when conversion service is nil (logged warning in real implementation)
	assert.Equal(t, 497.4, prices[isin], "Should use native price when conversion service nil")
}

// TestFetchCurrentPrices_MultipleCurrencies verifies multiple currencies in single batch
func TestFetchCurrentPrices_MultipleCurrencies(t *testing.T) {
	vwsISIN := "NL0000009082"     // VWS ISIN
	aaplISIN := "US0378331005"    // AAPL ISIN
	tencentISIN := "HK0000093390" // Tencent ISIN
	barcISIN := "GB0031348658"    // Barclays ISIN

	securities := []universe.Security{
		{Symbol: "VWS.AS", ISIN: vwsISIN, Currency: "EUR"},
		{Symbol: "AAPL", ISIN: aaplISIN, Currency: "USD"},
		{Symbol: "0700.HK", ISIN: tencentISIN, Currency: "HKD"},
		{Symbol: "BARC.L", ISIN: barcISIN, Currency: "GBP"},
	}

	eurPrice := 42.5
	usdPrice := 150.0
	hkdPrice := 90.0
	gbpPrice := 25.0

	priceClient := &MockPriceClientForConversion{
		quotes: map[string]*float64{
			"VWS.AS":  &eurPrice,
			"AAPL":    &usdPrice,
			"0700.HK": &hkdPrice,
			"BARC.L":  &gbpPrice,
		},
	}

	// Mock conversion with multiple rates
	priceConversionService := &MockPriceConversionServiceForScheduler{
		convertFunc: func(prices map[string]float64, secs []universe.Security) map[string]float64 {
			converted := make(map[string]float64)
			for symbol, price := range prices {
				switch symbol {
				case "VWS.AS":
					converted[symbol] = price // EUR unchanged
				case "AAPL":
					converted[symbol] = price * 0.93 // USD to EUR
				case "0700.HK":
					converted[symbol] = price * 0.11 // HKD to EUR
				case "BARC.L":
					converted[symbol] = price * 1.17 // GBP to EUR
				default:
					converted[symbol] = price
				}
			}
			return converted
		},
	}

	job := &BuildOpportunityContextJob{
		priceClient:            priceClient,
		priceConversionService: priceConversionService,
	}

	prices := job.fetchCurrentPrices(securities)

	// Verify all conversions (prices map uses ISIN keys)
	assert.Equal(t, 42.5, prices[vwsISIN], "EUR unchanged")
	assert.InDelta(t, 139.5, prices[aaplISIN], 0.01, "USD converted")
	assert.InDelta(t, 9.9, prices[tencentISIN], 0.01, "HKD converted")
	assert.InDelta(t, 29.25, prices[barcISIN], 0.01, "GBP converted")
}

// ============================================================================
// EnrichedPosition Enrichment Tests (Phase 5)
// ============================================================================

// TestBuildOpportunityContext_EnrichedPositions_AllFieldsPopulated verifies all 24 fields are populated
func TestBuildOpportunityContext_EnrichedPositions_AllFieldsPopulated(t *testing.T) {
	// Setup: Position with all 14 database fields
	firstBought := int64(1640000000) // 2021-12-20 00:00:00 UTC
	lastSold := int64(1672531200)    // 2023-01-01 00:00:00 UTC
	lastUpdated := int64(1704067200) // 2024-01-01 00:00:00 UTC

	positions := []portfolio.Position{
		{
			ISIN:             "US0378331005",
			Symbol:           "AAPL",
			Quantity:         100.0,
			AvgPrice:         150.0, // Maps to AverageCost
			Currency:         "USD",
			CurrencyRate:     1.1,
			MarketValueEUR:   13636.36,
			CostBasisEUR:     13636.36,
			UnrealizedPnL:    0.0,
			UnrealizedPnLPct: 0.0,
			LastUpdated:      &lastUpdated,
			FirstBoughtAt:    &firstBought,
			LastSoldAt:       &lastSold,
			CurrentPrice:     150.0,
		},
	}
	positionsInterface := make([]interface{}, len(positions))
	for i := range positions {
		positionsInterface[i] = positions[i]
	}

	// Setup: Security with metadata (7 fields)
	securities := []universe.Security{
		{
			ISIN:             "US0378331005",
			Symbol:           "AAPL",
			Name:             "Apple Inc.",
			Country:          "US",
			FullExchangeName: "NASDAQ",
			Active:           true,
			AllowBuy:         true,
			AllowSell:        true,
			MinLot:           1,
			Currency:         "USD",
			YahooSymbol:      "AAPL",
		},
	}
	securitiesInterface := make([]interface{}, len(securities))
	for i := range securities {
		securitiesInterface[i] = securities[i]
	}

	// Setup: Market data
	currentPrice := 160.0
	currentPriceEUR := 145.45 // 160 / 1.1

	mockPositionRepo := &MockPositionRepoForOptimizer{
		GetAllFunc: func() ([]interface{}, error) {
			return positionsInterface, nil
		},
	}

	mockSecurityRepo := &MockSecurityRepoForOptimizer{
		GetAllActiveFunc: func() ([]interface{}, error) {
			return securitiesInterface, nil
		},
	}

	mockAllocRepo := &MockAllocationRepoForOptimizer{
		GetAllFunc: func() (map[string]float64, error) {
			return map[string]float64{"country_group:US": 1.0}, nil
		},
	}

	mockCashManager := &MockCashManagerForOptimizer{
		GetAllCashBalancesFunc: func() (map[string]float64, error) {
			return map[string]float64{"EUR": 1000.0}, nil
		},
	}

	mockPriceClient := &MockPriceClient{
		GetBatchQuotesFunc: func(symbolMap map[string]*string) (map[string]*float64, error) {
			return map[string]*float64{
				"AAPL": &currentPrice,
			}, nil
		},
	}

	mockPriceConversionService := &MockPriceConversionServiceForScheduler{
		convertFunc: func(prices map[string]float64, secs []universe.Security) map[string]float64 {
			// Convert USD to EUR (1 USD = 0.909 EUR, or 1 EUR = 1.1 USD)
			converted := make(map[string]float64)
			for symbol, price := range prices {
				if symbol == "AAPL" {
					converted[symbol] = price / 1.1 // Convert USD to EUR
				} else {
					converted[symbol] = price
				}
			}
			return converted
		},
	}

	mockScoresRepo := &MockScoresRepoForContext{}
	mockSettingsRepo := &MockSettingsRepoForContext{}
	mockRegimeRepo := &MockRegimeRepoForContext{}
	mockGroupingRepo := &MockGroupingRepoForContext{}

	job := NewBuildOpportunityContextJob(
		mockPositionRepo,
		mockSecurityRepo,
		mockAllocRepo,
		mockGroupingRepo,
		mockCashManager,
		mockPriceClient,
		mockPriceConversionService,
		mockScoresRepo,
		mockSettingsRepo,
		mockRegimeRepo,
	)

	err := job.Run()
	require.NoError(t, err)

	ctx := job.GetOpportunityContext()
	require.NotNil(t, ctx)

	// Verify EnrichedPositions field exists and has data
	require.Len(t, ctx.EnrichedPositions, 1, "Should have 1 enriched position")

	pos := ctx.EnrichedPositions[0]

	// Verify all 14 core position fields from database
	assert.Equal(t, "US0378331005", pos.ISIN, "ISIN should be populated")
	assert.Equal(t, "AAPL", pos.Symbol, "Symbol should be populated")
	assert.Equal(t, 100.0, pos.Quantity, "Quantity should be populated")
	// AverageCost should be in EUR (converted from native avg_price)
	// CostBasisEUR = 13636.36, Quantity = 100 → EUR avg cost = 136.3636
	assert.InDelta(t, 136.3636, pos.AverageCost, 0.01, "AverageCost should be in EUR (CostBasisEUR/Quantity)")
	assert.Equal(t, "USD", pos.Currency, "Currency should be populated")
	assert.Equal(t, 1.1, pos.CurrencyRate, "CurrencyRate should be populated")
	assert.Equal(t, 13636.36, pos.MarketValueEUR, "MarketValueEUR should be populated")
	assert.Equal(t, 13636.36, pos.CostBasisEUR, "CostBasisEUR should be populated")
	assert.Equal(t, 0.0, pos.UnrealizedPnL, "UnrealizedPnL should be populated")
	assert.Equal(t, 0.0, pos.UnrealizedPnLPct, "UnrealizedPnLPct should be populated")

	// Verify timestamp conversion (Unix → time.Time)
	require.NotNil(t, pos.LastUpdated, "LastUpdated should be populated")
	require.NotNil(t, pos.FirstBoughtAt, "FirstBoughtAt should be populated")
	require.NotNil(t, pos.LastSoldAt, "LastSoldAt should be populated")
	assert.Equal(t, lastUpdated, pos.LastUpdated.Unix(), "LastUpdated timestamp should match")
	assert.Equal(t, firstBought, pos.FirstBoughtAt.Unix(), "FirstBoughtAt timestamp should match")
	assert.Equal(t, lastSold, pos.LastSoldAt.Unix(), "LastSoldAt timestamp should match")

	// Verify all 7 security metadata fields
	assert.Equal(t, "Apple Inc.", pos.SecurityName, "SecurityName should be populated from security")
	assert.Equal(t, "US", pos.Country, "Country should be populated from security")
	assert.Equal(t, "NASDAQ", pos.Exchange, "Exchange should be populated from security")
	assert.True(t, pos.Active, "Active should be populated from security")
	assert.True(t, pos.AllowBuy, "AllowBuy should be populated from security")
	assert.True(t, pos.AllowSell, "AllowSell should be populated from security")
	assert.Equal(t, 1, pos.MinLot, "MinLot should be populated from security")

	// Verify market data (1 field)
	assert.InDelta(t, currentPriceEUR, pos.CurrentPrice, 0.01, "CurrentPrice should be populated from current prices (in EUR)")

	// Verify calculated fields (2 fields)
	require.NotNil(t, pos.DaysHeld, "DaysHeld should be calculated")
	assert.Greater(t, *pos.DaysHeld, 0, "DaysHeld should be positive")

	assert.Greater(t, pos.WeightInPortfolio, 0.0, "WeightInPortfolio should be calculated")
	assert.LessOrEqual(t, pos.WeightInPortfolio, 1.0, "WeightInPortfolio should be <= 1.0")

	// Total: 14 + 7 + 1 + 2 = 24 fields verified
}

// TestBuildOpportunityContext_EnrichedPositions_TimestampConversion verifies timestamp conversion
func TestBuildOpportunityContext_EnrichedPositions_TimestampConversion(t *testing.T) {
	firstBought := int64(1609459200) // 2021-01-01 00:00:00 UTC
	lastSold := int64(1640995200)    // 2022-01-01 00:00:00 UTC
	lastUpdated := int64(1672531200) // 2023-01-01 00:00:00 UTC

	positions := []portfolio.Position{
		{
			ISIN:          "US0378331005",
			Symbol:        "AAPL",
			Quantity:      100.0,
			AvgPrice:      150.0,
			Currency:      "USD",
			CostBasisEUR:  13636.36,
			CurrencyRate:  1.1,
			FirstBoughtAt: &firstBought,
			LastSoldAt:    &lastSold,
			LastUpdated:   &lastUpdated,
		},
	}
	positionsInterface := make([]interface{}, len(positions))
	for i := range positions {
		positionsInterface[i] = positions[i]
	}

	securities := []universe.Security{
		{ISIN: "US0378331005", Symbol: "AAPL", Name: "Apple Inc.", Active: true},
	}
	securitiesInterface := make([]interface{}, len(securities))
	for i := range securities {
		securitiesInterface[i] = securities[i]
	}

	mockPositionRepo := &MockPositionRepoForOptimizer{
		GetAllFunc: func() ([]interface{}, error) {
			return positionsInterface, nil
		},
	}

	mockSecurityRepo := &MockSecurityRepoForOptimizer{
		GetAllActiveFunc: func() ([]interface{}, error) {
			return securitiesInterface, nil
		},
	}

	mockAllocRepo := &MockAllocationRepoForOptimizer{
		GetAllFunc: func() (map[string]float64, error) {
			return map[string]float64{}, nil
		},
	}

	mockCashManager := &MockCashManagerForOptimizer{
		GetAllCashBalancesFunc: func() (map[string]float64, error) {
			return map[string]float64{"EUR": 1000.0}, nil
		},
	}

	price := 150.0
	mockPriceClient := &MockPriceClient{
		GetBatchQuotesFunc: func(symbolMap map[string]*string) (map[string]*float64, error) {
			return map[string]*float64{"AAPL": &price}, nil
		},
	}

	job := NewBuildOpportunityContextJob(
		mockPositionRepo,
		mockSecurityRepo,
		mockAllocRepo,
		&MockGroupingRepoForContext{},
		mockCashManager,
		mockPriceClient,
		nil,
		&MockScoresRepoForContext{},
		&MockSettingsRepoForContext{},
		&MockRegimeRepoForContext{},
	)

	err := job.Run()
	require.NoError(t, err)

	ctx := job.GetOpportunityContext()
	require.Len(t, ctx.EnrichedPositions, 1)

	pos := ctx.EnrichedPositions[0]

	// Verify Unix timestamps converted to time.Time
	require.NotNil(t, pos.FirstBoughtAt)
	require.NotNil(t, pos.LastSoldAt)
	require.NotNil(t, pos.LastUpdated)

	assert.Equal(t, firstBought, pos.FirstBoughtAt.Unix())
	assert.Equal(t, lastSold, pos.LastSoldAt.Unix())
	assert.Equal(t, lastUpdated, pos.LastUpdated.Unix())

	// Verify times are in UTC
	assert.Equal(t, "UTC", pos.FirstBoughtAt.Location().String())
	assert.Equal(t, "UTC", pos.LastSoldAt.Location().String())
	assert.Equal(t, "UTC", pos.LastUpdated.Location().String())
}

// TestBuildOpportunityContext_EnrichedPositions_NilTimestamps verifies nil timestamps handled correctly
func TestBuildOpportunityContext_EnrichedPositions_NilTimestamps(t *testing.T) {
	positions := []portfolio.Position{
		{
			ISIN:          "US0378331005",
			Symbol:        "AAPL",
			Quantity:      100.0,
			AvgPrice:      150.0,
			Currency:      "USD",
			CostBasisEUR:  13636.36,
			CurrencyRate:  1.1,
			FirstBoughtAt: nil, // Nil timestamp
			LastSoldAt:    nil,
			LastUpdated:   nil,
		},
	}
	positionsInterface := make([]interface{}, len(positions))
	for i := range positions {
		positionsInterface[i] = positions[i]
	}

	securities := []universe.Security{
		{ISIN: "US0378331005", Symbol: "AAPL", Name: "Apple Inc.", Active: true},
	}
	securitiesInterface := make([]interface{}, len(securities))
	for i := range securities {
		securitiesInterface[i] = securities[i]
	}

	mockPositionRepo := &MockPositionRepoForOptimizer{
		GetAllFunc: func() ([]interface{}, error) {
			return positionsInterface, nil
		},
	}

	mockSecurityRepo := &MockSecurityRepoForOptimizer{
		GetAllActiveFunc: func() ([]interface{}, error) {
			return securitiesInterface, nil
		},
	}

	mockAllocRepo := &MockAllocationRepoForOptimizer{
		GetAllFunc: func() (map[string]float64, error) {
			return map[string]float64{}, nil
		},
	}

	mockCashManager := &MockCashManagerForOptimizer{
		GetAllCashBalancesFunc: func() (map[string]float64, error) {
			return map[string]float64{"EUR": 1000.0}, nil
		},
	}

	price := 150.0
	mockPriceClient := &MockPriceClient{
		GetBatchQuotesFunc: func(symbolMap map[string]*string) (map[string]*float64, error) {
			return map[string]*float64{"AAPL": &price}, nil
		},
	}

	job := NewBuildOpportunityContextJob(
		mockPositionRepo,
		mockSecurityRepo,
		mockAllocRepo,
		&MockGroupingRepoForContext{},
		mockCashManager,
		mockPriceClient,
		nil,
		&MockScoresRepoForContext{},
		&MockSettingsRepoForContext{},
		&MockRegimeRepoForContext{},
	)

	err := job.Run()
	require.NoError(t, err)

	ctx := job.GetOpportunityContext()
	require.Len(t, ctx.EnrichedPositions, 1)

	pos := ctx.EnrichedPositions[0]

	// Verify nil timestamps remain nil (no crash)
	assert.Nil(t, pos.FirstBoughtAt, "FirstBoughtAt should be nil")
	assert.Nil(t, pos.LastSoldAt, "LastSoldAt should be nil")
	assert.Nil(t, pos.LastUpdated, "LastUpdated should be nil")

	// Verify DaysHeld is nil when FirstBoughtAt is nil
	assert.Nil(t, pos.DaysHeld, "DaysHeld should be nil when FirstBoughtAt is nil")

	// Verify other fields still work
	assert.Equal(t, "US0378331005", pos.ISIN)
	// AverageCost should be in EUR: CostBasisEUR / Quantity = 13636.36 / 100 = 136.36
	assert.InDelta(t, 136.36, pos.AverageCost, 0.01, "AverageCost should be in EUR")
}

// TestBuildOpportunityContext_EnrichedPositions_CalculatedFields verifies DaysHeld and WeightInPortfolio calculation
func TestBuildOpportunityContext_EnrichedPositions_CalculatedFields(t *testing.T) {
	// Position bought 365 days ago from NOW
	daysAgo := int64(365)
	firstBought := time.Now().Unix() - (daysAgo * 24 * 60 * 60) // 365 days before now

	positions := []portfolio.Position{
		{
			ISIN:           "US0378331005",
			Symbol:         "AAPL",
			Quantity:       100.0,
			AvgPrice:       150.0,
			Currency:       "USD",
			CostBasisEUR:   13636.36,
			CurrencyRate:   1.1,
			MarketValueEUR: 15000.0,
			FirstBoughtAt:  &firstBought,
		},
	}
	positionsInterface := make([]interface{}, len(positions))
	for i := range positions {
		positionsInterface[i] = positions[i]
	}

	securities := []universe.Security{
		{ISIN: "US0378331005", Symbol: "AAPL", Name: "Apple Inc.", Active: true},
	}
	securitiesInterface := make([]interface{}, len(securities))
	for i := range securities {
		securitiesInterface[i] = securities[i]
	}

	mockPositionRepo := &MockPositionRepoForOptimizer{
		GetAllFunc: func() ([]interface{}, error) {
			return positionsInterface, nil
		},
	}

	mockSecurityRepo := &MockSecurityRepoForOptimizer{
		GetAllActiveFunc: func() ([]interface{}, error) {
			return securitiesInterface, nil
		},
	}

	mockAllocRepo := &MockAllocationRepoForOptimizer{
		GetAllFunc: func() (map[string]float64, error) {
			return map[string]float64{}, nil
		},
	}

	// Cash: 5000 EUR, Position: 15000 EUR, Total: 20000 EUR
	// Expected WeightInPortfolio = 15000 / 20000 = 0.75
	mockCashManager := &MockCashManagerForOptimizer{
		GetAllCashBalancesFunc: func() (map[string]float64, error) {
			return map[string]float64{"EUR": 5000.0}, nil
		},
	}

	price := 150.0
	mockPriceClient := &MockPriceClient{
		GetBatchQuotesFunc: func(symbolMap map[string]*string) (map[string]*float64, error) {
			return map[string]*float64{"AAPL": &price}, nil
		},
	}

	job := NewBuildOpportunityContextJob(
		mockPositionRepo,
		mockSecurityRepo,
		mockAllocRepo,
		&MockGroupingRepoForContext{},
		mockCashManager,
		mockPriceClient,
		nil,
		&MockScoresRepoForContext{},
		&MockSettingsRepoForContext{},
		&MockRegimeRepoForContext{},
	)

	err := job.Run()
	require.NoError(t, err)

	ctx := job.GetOpportunityContext()
	require.Len(t, ctx.EnrichedPositions, 1)

	pos := ctx.EnrichedPositions[0]

	// Verify DaysHeld calculation
	require.NotNil(t, pos.DaysHeld, "DaysHeld should be calculated")
	assert.Greater(t, *pos.DaysHeld, 360, "DaysHeld should be ~365 days")
	assert.Less(t, *pos.DaysHeld, 370, "DaysHeld should be ~365 days")

	// Verify WeightInPortfolio calculation
	expectedWeight := 15000.0 / 20000.0 // 0.75
	assert.InDelta(t, expectedWeight, pos.WeightInPortfolio, 0.01, "WeightInPortfolio should be 0.75")
}

// TestBuildOpportunityContext_EnrichedPositions_MissingISIN_Skipped verifies positions without ISIN are skipped
func TestBuildOpportunityContext_EnrichedPositions_MissingISIN_Skipped(t *testing.T) {
	positions := []portfolio.Position{
		{ISIN: "US0378331005", Symbol: "AAPL", Quantity: 100.0, AvgPrice: 150.0}, // Valid
		{ISIN: "", Symbol: "INVALID", Quantity: 50.0, AvgPrice: 100.0},           // Missing ISIN
		{ISIN: "US5949181045", Symbol: "MSFT", Quantity: 75.0, AvgPrice: 300.0},  // Valid
	}
	positionsInterface := make([]interface{}, len(positions))
	for i := range positions {
		positionsInterface[i] = positions[i]
	}

	securities := []universe.Security{
		{ISIN: "US0378331005", Symbol: "AAPL", Name: "Apple Inc.", Active: true},
		{ISIN: "US5949181045", Symbol: "MSFT", Name: "Microsoft Corp.", Active: true},
	}
	securitiesInterface := make([]interface{}, len(securities))
	for i := range securities {
		securitiesInterface[i] = securities[i]
	}

	mockPositionRepo := &MockPositionRepoForOptimizer{
		GetAllFunc: func() ([]interface{}, error) {
			return positionsInterface, nil
		},
	}

	mockSecurityRepo := &MockSecurityRepoForOptimizer{
		GetAllActiveFunc: func() ([]interface{}, error) {
			return securitiesInterface, nil
		},
	}

	mockAllocRepo := &MockAllocationRepoForOptimizer{
		GetAllFunc: func() (map[string]float64, error) {
			return map[string]float64{}, nil
		},
	}

	mockCashManager := &MockCashManagerForOptimizer{
		GetAllCashBalancesFunc: func() (map[string]float64, error) {
			return map[string]float64{"EUR": 1000.0}, nil
		},
	}

	price := 150.0
	mockPriceClient := &MockPriceClient{
		GetBatchQuotesFunc: func(symbolMap map[string]*string) (map[string]*float64, error) {
			return map[string]*float64{
				"AAPL": &price,
				"MSFT": &price,
			}, nil
		},
	}

	job := NewBuildOpportunityContextJob(
		mockPositionRepo,
		mockSecurityRepo,
		mockAllocRepo,
		&MockGroupingRepoForContext{},
		mockCashManager,
		mockPriceClient,
		nil,
		&MockScoresRepoForContext{},
		&MockSettingsRepoForContext{},
		&MockRegimeRepoForContext{},
	)

	err := job.Run()
	require.NoError(t, err)

	ctx := job.GetOpportunityContext()

	// Should only have 2 enriched positions (INVALID skipped)
	require.Len(t, ctx.EnrichedPositions, 2, "Should skip position without ISIN")

	// Verify correct positions included
	isins := make([]string, len(ctx.EnrichedPositions))
	for i, pos := range ctx.EnrichedPositions {
		isins[i] = pos.ISIN
	}
	assert.Contains(t, isins, "US0378331005", "AAPL should be included")
	assert.Contains(t, isins, "US5949181045", "MSFT should be included")
	assert.NotContains(t, isins, "", "INVALID should NOT be included")
}

// TestBuildOpportunityContext_EnrichedPositions_MissingSecurity_Skipped verifies positions with unknown ISIN are skipped
func TestBuildOpportunityContext_EnrichedPositions_MissingSecurity_Skipped(t *testing.T) {
	positions := []portfolio.Position{
		{ISIN: "US0378331005", Symbol: "AAPL", Quantity: 100.0, AvgPrice: 150.0}, // Valid security
		{ISIN: "UNKNOWN12345", Symbol: "UNKN", Quantity: 50.0, AvgPrice: 100.0},  // Unknown security
		{ISIN: "US5949181045", Symbol: "MSFT", Quantity: 75.0, AvgPrice: 300.0},  // Valid security
	}
	positionsInterface := make([]interface{}, len(positions))
	for i := range positions {
		positionsInterface[i] = positions[i]
	}

	securities := []universe.Security{
		{ISIN: "US0378331005", Symbol: "AAPL", Name: "Apple Inc.", Active: true},
		{ISIN: "US5949181045", Symbol: "MSFT", Name: "Microsoft Corp.", Active: true},
		// UNKNOWN12345 not in securities list
	}
	securitiesInterface := make([]interface{}, len(securities))
	for i := range securities {
		securitiesInterface[i] = securities[i]
	}

	mockPositionRepo := &MockPositionRepoForOptimizer{
		GetAllFunc: func() ([]interface{}, error) {
			return positionsInterface, nil
		},
	}

	mockSecurityRepo := &MockSecurityRepoForOptimizer{
		GetAllActiveFunc: func() ([]interface{}, error) {
			return securitiesInterface, nil
		},
	}

	mockAllocRepo := &MockAllocationRepoForOptimizer{
		GetAllFunc: func() (map[string]float64, error) {
			return map[string]float64{}, nil
		},
	}

	mockCashManager := &MockCashManagerForOptimizer{
		GetAllCashBalancesFunc: func() (map[string]float64, error) {
			return map[string]float64{"EUR": 1000.0}, nil
		},
	}

	price := 150.0
	mockPriceClient := &MockPriceClient{
		GetBatchQuotesFunc: func(symbolMap map[string]*string) (map[string]*float64, error) {
			return map[string]*float64{
				"AAPL": &price,
				"MSFT": &price,
			}, nil
		},
	}

	job := NewBuildOpportunityContextJob(
		mockPositionRepo,
		mockSecurityRepo,
		mockAllocRepo,
		&MockGroupingRepoForContext{},
		mockCashManager,
		mockPriceClient,
		nil,
		&MockScoresRepoForContext{},
		&MockSettingsRepoForContext{},
		&MockRegimeRepoForContext{},
	)

	err := job.Run()
	require.NoError(t, err)

	ctx := job.GetOpportunityContext()

	// Should only have 2 enriched positions (UNKNOWN skipped)
	require.Len(t, ctx.EnrichedPositions, 2, "Should skip position with unknown security")

	// Verify correct positions included
	isins := make([]string, len(ctx.EnrichedPositions))
	for i, pos := range ctx.EnrichedPositions {
		isins[i] = pos.ISIN
	}
	assert.Contains(t, isins, "US0378331005", "AAPL should be included")
	assert.Contains(t, isins, "US5949181045", "MSFT should be included")
	assert.NotContains(t, isins, "UNKNOWN12345", "Unknown security should NOT be included")
}
