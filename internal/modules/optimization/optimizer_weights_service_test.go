package optimization

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/aristath/sentinel/internal/modules/portfolio"
	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

type mockPositionRepoForWeights struct {
	GetAllFunc func() ([]portfolio.Position, error)
}

func (m *mockPositionRepoForWeights) GetAll() ([]portfolio.Position, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	return []portfolio.Position{}, nil
}

// Stub methods to satisfy PositionRepositoryInterface
func (m *mockPositionRepoForWeights) GetWithSecurityInfo() ([]portfolio.PositionWithSecurity, error) {
	return nil, nil
}
func (m *mockPositionRepoForWeights) GetBySymbol(symbol string) (*portfolio.Position, error) {
	return nil, nil
}
func (m *mockPositionRepoForWeights) GetByISIN(isin string) (*portfolio.Position, error) {
	return nil, nil
}
func (m *mockPositionRepoForWeights) GetByIdentifier(identifier string) (*portfolio.Position, error) {
	return nil, nil
}
func (m *mockPositionRepoForWeights) GetCount() (int, error) {
	return 0, nil
}
func (m *mockPositionRepoForWeights) GetTotalValue() (float64, error) {
	return 0, nil
}
func (m *mockPositionRepoForWeights) Upsert(position portfolio.Position) error {
	return nil
}
func (m *mockPositionRepoForWeights) Delete(isin string) error {
	return nil
}
func (m *mockPositionRepoForWeights) DeleteAll() error {
	return nil
}
func (m *mockPositionRepoForWeights) UpdatePrice(isin string, price float64, currencyRate float64) error {
	return nil
}
func (m *mockPositionRepoForWeights) UpdateLastSoldAt(isin string) error {
	return nil
}

type mockSecurityRepoForWeights struct {
	GetAllActiveFunc func() ([]universe.Security, error)
}

func (m *mockSecurityRepoForWeights) GetAllActive() ([]universe.Security, error) {
	if m.GetAllActiveFunc != nil {
		return m.GetAllActiveFunc()
	}
	return []universe.Security{}, nil
}

// Stub methods to satisfy SecurityRepositoryInterface
func (m *mockSecurityRepoForWeights) GetBySymbol(symbol string) (*universe.Security, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetByISIN(isin string) (*universe.Security, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetByIdentifier(identifier string) (*universe.Security, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetAll() ([]universe.Security, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetByISINs(isins []string) ([]universe.Security, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetBySymbols(symbols []string) ([]universe.Security, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetTradable() ([]universe.Security, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetByMarketCode(marketCode string) ([]universe.Security, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetByGeography(geography string) ([]universe.Security, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetByIndustry(industry string) ([]universe.Security, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetByTags(tagIDs []string) ([]universe.Security, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetPositionsByTags(positionSymbols []string, tagIDs []string) ([]universe.Security, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetDistinctGeographies() ([]string, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetDistinctIndustries() ([]string, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetDistinctExchanges() ([]string, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetGeographiesAndIndustries() (map[string][]string, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetSecuritiesForOptimization() ([]universe.SecurityOptimizationData, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetSecuritiesForCharts() ([]universe.SecurityChartData, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetISINBySymbol(symbol string) (string, error) {
	return "", nil
}
func (m *mockSecurityRepoForWeights) GetSymbolByISIN(isin string) (string, error) {
	return "", nil
}
func (m *mockSecurityRepoForWeights) BatchGetISINsBySymbols(symbols []string) (map[string]string, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) Exists(isin string) (bool, error) {
	return false, nil
}
func (m *mockSecurityRepoForWeights) ExistsBySymbol(symbol string) (bool, error) {
	return false, nil
}
func (m *mockSecurityRepoForWeights) CountTradable() (int, error) {
	return 0, nil
}
func (m *mockSecurityRepoForWeights) GetWithScores(portfolioDB *sql.DB) ([]universe.SecurityWithScore, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) Create(security universe.Security) error {
	return nil
}
func (m *mockSecurityRepoForWeights) Update(isin string, updates map[string]interface{}) error {
	return nil
}
func (m *mockSecurityRepoForWeights) Delete(isin string) error {
	return nil
}
func (m *mockSecurityRepoForWeights) HardDelete(isin string) error {
	return nil
}
func (m *mockSecurityRepoForWeights) SetTagsForSecurity(symbol string, tagIDs []string) error {
	return nil
}
func (m *mockSecurityRepoForWeights) GetTagsForSecurity(symbol string) ([]string, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) GetTagsWithUpdateTimes(symbol string) (map[string]time.Time, error) {
	return nil, nil
}
func (m *mockSecurityRepoForWeights) UpdateSpecificTags(symbol string, tagIDs []string) error {
	return nil
}
func (m *mockSecurityRepoForWeights) GetAllActiveTradable() ([]universe.Security, error) {
	return nil, nil
}

type mockAllocationRepoForWeights struct {
	GetAllFunc func() (map[string]float64, error)
}

func (m *mockAllocationRepoForWeights) GetAll() (map[string]float64, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	return map[string]float64{}, nil
}

type mockCashManagerForWeights struct {
	GetAllCashBalancesFunc func() (map[string]float64, error)
}

func (m *mockCashManagerForWeights) GetAllCashBalances() (map[string]float64, error) {
	if m.GetAllCashBalancesFunc != nil {
		return m.GetAllCashBalancesFunc()
	}
	return map[string]float64{}, nil
}

// Stub methods to satisfy domain.CashManager interface
func (m *mockCashManagerForWeights) UpdateCashPosition(currency string, balance float64) error {
	return nil
}
func (m *mockCashManagerForWeights) GetCashBalance(currency string) (float64, error) {
	return 0, nil
}

type mockPriceClientForWeights struct {
	GetBatchQuotesFunc func(symbolMap map[string]*string) (map[string]*float64, error)
}

func (m *mockPriceClientForWeights) GetBatchQuotes(symbolMap map[string]*string) (map[string]*float64, error) {
	if m.GetBatchQuotesFunc != nil {
		return m.GetBatchQuotesFunc(symbolMap)
	}
	return make(map[string]*float64), nil
}

type mockOptimizerServiceForWeights struct {
	OptimizeFunc func(state PortfolioState, settings Settings) (*Result, error)
}

func (m *mockOptimizerServiceForWeights) Optimize(state PortfolioState, settings Settings) (*Result, error) {
	if m.OptimizeFunc != nil {
		return m.OptimizeFunc(state, settings)
	}
	return nil, nil
}

type mockPriceConversionServiceForWeights struct {
	ConvertPricesToEURFunc func(prices map[string]float64, securities []universe.Security) map[string]float64
}

func (m *mockPriceConversionServiceForWeights) ConvertPricesToEUR(prices map[string]float64, securities []universe.Security) map[string]float64 {
	if m.ConvertPricesToEURFunc != nil {
		return m.ConvertPricesToEURFunc(prices, securities)
	}
	return prices
}

type mockClientDataRepoForWeights struct {
	GetIfFreshFunc func(table, key string) (json.RawMessage, error)
	GetFunc        func(table, key string) (json.RawMessage, error)
	StoreFunc      func(table, key string, data interface{}, ttl time.Duration) error
	StoredData     map[string]json.RawMessage
	StoredTTLs     map[string]time.Duration
}

func (m *mockClientDataRepoForWeights) GetIfFresh(table, key string) (json.RawMessage, error) {
	if m.GetIfFreshFunc != nil {
		return m.GetIfFreshFunc(table, key)
	}
	return nil, nil
}

func (m *mockClientDataRepoForWeights) Get(table, key string) (json.RawMessage, error) {
	if m.GetFunc != nil {
		return m.GetFunc(table, key)
	}
	return nil, nil
}

func (m *mockClientDataRepoForWeights) Store(table, key string, data interface{}, ttl time.Duration) error {
	if m.StoreFunc != nil {
		return m.StoreFunc(table, key, data, ttl)
	}
	if m.StoredData == nil {
		m.StoredData = make(map[string]json.RawMessage)
	}
	if m.StoredTTLs == nil {
		m.StoredTTLs = make(map[string]time.Duration)
	}
	jsonData, _ := json.Marshal(data)
	m.StoredData[table+":"+key] = jsonData
	m.StoredTTLs[table+":"+key] = ttl
	return nil
}

type mockMarketHoursServiceForWeights struct {
	AnyMajorMarketOpenFunc func(t time.Time) bool
}

func (m *mockMarketHoursServiceForWeights) AnyMajorMarketOpen(t time.Time) bool {
	if m.AnyMajorMarketOpenFunc != nil {
		return m.AnyMajorMarketOpenFunc(t)
	}
	return false
}

// Helper function
func floatPtr(f float64) *float64 {
	return &f
}

func TestOptimizerWeightsService_CalculateWeights_Success(t *testing.T) {
	// Mock position repo
	positions := []portfolio.Position{
		{Symbol: "AAPL", ISIN: "US0378331005", Quantity: 10},
	}

	mockPositionRepo := &mockPositionRepoForWeights{
		GetAllFunc: func() ([]portfolio.Position, error) {
			return positions, nil
		},
	}

	// Mock security repo
	securities := []universe.Security{
		{ISIN: "US0378331005", Symbol: "AAPL", Geography: "US", Industry: "Technology"},
	}

	mockSecurityRepo := &mockSecurityRepoForWeights{
		GetAllActiveFunc: func() ([]universe.Security, error) {
			return securities, nil
		},
	}

	// Mock allocation repo
	mockAllocRepo := &mockAllocationRepoForWeights{
		GetAllFunc: func() (map[string]float64, error) {
			return map[string]float64{
				"geography:US":        0.5,
				"industry:Technology": 0.3,
			}, nil
		},
	}

	// Mock cash manager
	mockCashManager := &mockCashManagerForWeights{
		GetAllCashBalancesFunc: func() (map[string]float64, error) {
			return map[string]float64{"EUR": 1000.0}, nil
		},
	}

	// Mock price client
	mockPriceClient := &mockPriceClientForWeights{
		GetBatchQuotesFunc: func(symbolMap map[string]*string) (map[string]*float64, error) {
			return map[string]*float64{
				"AAPL": floatPtr(150.0),
			}, nil
		},
	}

	// Mock optimizer service
	optimizeCalled := false
	mockOptimizerService := &mockOptimizerServiceForWeights{
		OptimizeFunc: func(state PortfolioState, settings Settings) (*Result, error) {
			optimizeCalled = true
			return &Result{
				Success:       true,
				TargetWeights: map[string]float64{"US0378331005": 0.5},
			}, nil
		},
	}

	// Create service - this will fail initially (TDD)
	service := NewOptimizerWeightsService(
		mockPositionRepo,
		mockSecurityRepo,
		mockAllocRepo,
		mockCashManager,
		nil, // brokerClient - not used in this test
		mockOptimizerService,
		nil, // priceConversionService
		nil, // plannerConfigRepo
		nil, // clientDataRepo
		nil, // marketHoursService
	)
	// Set priceClient directly for testing (service will use adapter in production)
	service.priceClient = mockPriceClient

	weights, err := service.CalculateWeights(context.Background())
	require.NoError(t, err)
	assert.True(t, optimizeCalled, "Optimize should have been called")
	assert.Equal(t, 0.5, weights["US0378331005"])
}

func TestOptimizerWeightsService_CalculateWeights_NoOptimizerService(t *testing.T) {
	service := NewOptimizerWeightsService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	_, err := service.CalculateWeights(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "optimizer service not available")
}

func TestOptimizerWeightsService_CalculateWeights_PositionRepoError(t *testing.T) {
	mockPositionRepo := &mockPositionRepoForWeights{
		GetAllFunc: func() ([]portfolio.Position, error) {
			return nil, assert.AnError
		},
	}

	mockOptimizerService := &mockOptimizerServiceForWeights{}

	service := NewOptimizerWeightsService(
		mockPositionRepo,
		nil,
		nil,
		nil,
		nil,
		mockOptimizerService,
		nil,
		nil,
		nil,
		nil,
	)

	_, err := service.CalculateWeights(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get positions")
}

func TestOptimizerWeightsService_CalculateWeights_SecurityRepoError(t *testing.T) {
	mockPositionRepo := &mockPositionRepoForWeights{
		GetAllFunc: func() ([]portfolio.Position, error) {
			return []portfolio.Position{}, nil
		},
	}

	mockSecurityRepo := &mockSecurityRepoForWeights{
		GetAllActiveFunc: func() ([]universe.Security, error) {
			return nil, assert.AnError
		},
	}

	mockOptimizerService := &mockOptimizerServiceForWeights{}

	service := NewOptimizerWeightsService(
		mockPositionRepo,
		mockSecurityRepo,
		nil,
		nil,
		nil,
		mockOptimizerService,
		nil,
		nil,
		nil,
		nil,
	)

	_, err := service.CalculateWeights(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get securities")
}

func TestOptimizerWeightsService_CalculateWeights_AllocationRepoError(t *testing.T) {
	mockPositionRepo := &mockPositionRepoForWeights{
		GetAllFunc: func() ([]portfolio.Position, error) {
			return []portfolio.Position{}, nil
		},
	}

	mockSecurityRepo := &mockSecurityRepoForWeights{
		GetAllActiveFunc: func() ([]universe.Security, error) {
			return []universe.Security{}, nil
		},
	}

	mockAllocRepo := &mockAllocationRepoForWeights{
		GetAllFunc: func() (map[string]float64, error) {
			return nil, assert.AnError
		},
	}

	mockOptimizerService := &mockOptimizerServiceForWeights{}

	service := NewOptimizerWeightsService(
		mockPositionRepo,
		mockSecurityRepo,
		mockAllocRepo,
		nil,
		nil,
		mockOptimizerService,
		nil,
		nil,
		nil,
		nil,
	)

	_, err := service.CalculateWeights(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get allocations")
}

func TestOptimizerWeightsService_CalculateWeights_OptimizerError(t *testing.T) {
	mockPositionRepo := &mockPositionRepoForWeights{
		GetAllFunc: func() ([]portfolio.Position, error) {
			return []portfolio.Position{}, nil
		},
	}

	mockSecurityRepo := &mockSecurityRepoForWeights{
		GetAllActiveFunc: func() ([]universe.Security, error) {
			return []universe.Security{}, nil
		},
	}

	mockAllocRepo := &mockAllocationRepoForWeights{
		GetAllFunc: func() (map[string]float64, error) {
			return map[string]float64{}, nil
		},
	}

	mockCashManager := &mockCashManagerForWeights{
		GetAllCashBalancesFunc: func() (map[string]float64, error) {
			return map[string]float64{"EUR": 1000.0}, nil
		},
	}

	mockPriceClient := &mockPriceClientForWeights{
		GetBatchQuotesFunc: func(symbolMap map[string]*string) (map[string]*float64, error) {
			return map[string]*float64{}, nil
		},
	}

	mockOptimizerService := &mockOptimizerServiceForWeights{
		OptimizeFunc: func(state PortfolioState, settings Settings) (*Result, error) {
			return nil, assert.AnError
		},
	}

	service := NewOptimizerWeightsService(
		mockPositionRepo,
		mockSecurityRepo,
		mockAllocRepo,
		mockCashManager,
		nil, // brokerClient
		mockOptimizerService,
		nil,
		nil,
		nil,
		nil,
	)
	service.priceClient = mockPriceClient

	_, err := service.CalculateWeights(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "optimizer failed")
}

func TestOptimizerWeightsService_FetchCurrentPrices_CacheHit(t *testing.T) {
	// Setup: All prices are cached
	cachedPrice := 150.0
	mockClientDataRepo := &mockClientDataRepoForWeights{
		GetIfFreshFunc: func(table, key string) (json.RawMessage, error) {
			if table == "current_prices" {
				data, _ := json.Marshal(cachedPrice)
				return data, nil
			}
			return nil, nil
		},
	}

	mockMarketHours := &mockMarketHoursServiceForWeights{
		AnyMajorMarketOpenFunc: func(t time.Time) bool {
			return true // Markets open
		},
	}

	// Price client should NOT be called when cache hit
	priceClientCalled := false
	mockPriceClient := &mockPriceClientForWeights{
		GetBatchQuotesFunc: func(symbolMap map[string]*string) (map[string]*float64, error) {
			priceClientCalled = true
			return map[string]*float64{}, nil
		},
	}

	service := NewOptimizerWeightsService(
		nil, nil, nil, nil,
		nil, // brokerClient
		nil, nil, nil,
		mockClientDataRepo,
		mockMarketHours,
	)
	service.priceClient = mockPriceClient

	securities := []universe.Security{
		{ISIN: "US0378331005", Symbol: "AAPL", Currency: "USD"},
	}

	// This will fail initially - fetchCurrentPrices is a private method
	// We'll test it indirectly through CalculateWeights or make it public for testing
	// For now, assuming we'll test via CalculateWeights or export it
	_ = securities
	_ = priceClientCalled
	_ = cachedPrice

	// Test via CalculateWeights when service is implemented
	// This test compiles but will fail until service is implemented
	t.Skip("Will test via CalculateWeights once service is implemented")
}

func TestOptimizerWeightsService_FetchCurrentPrices_CacheMiss(t *testing.T) {
	// Setup: No cached prices
	mockClientDataRepo := &mockClientDataRepoForWeights{
		GetIfFreshFunc: func(table, key string) (json.RawMessage, error) {
			return nil, nil // Cache miss
		},
		StoredData: make(map[string]json.RawMessage),
		StoredTTLs: make(map[string]time.Duration),
	}

	mockMarketHours := &mockMarketHoursServiceForWeights{
		AnyMajorMarketOpenFunc: func(t time.Time) bool {
			return true // Markets open
		},
	}

	// Price client should be called
	fetchedPrice := 150.0
	mockPriceClient := &mockPriceClientForWeights{
		GetBatchQuotesFunc: func(symbolMap map[string]*string) (map[string]*float64, error) {
			return map[string]*float64{
				"AAPL": &fetchedPrice,
			}, nil
		},
	}

	// Mock price conversion service
	mockPriceConversion := &mockPriceConversionServiceForWeights{
		ConvertPricesToEURFunc: func(prices map[string]float64, securities []universe.Security) map[string]float64 {
			return prices
		},
	}

	service := NewOptimizerWeightsService(
		nil, nil, nil, nil,
		nil, // brokerClient
		nil,
		mockPriceConversion,
		nil,
		mockClientDataRepo,
		mockMarketHours,
	)
	service.priceClient = mockPriceClient

	_ = service

	// Test via CalculateWeights when service is implemented
	t.Skip("Will test via CalculateWeights once service is implemented")
}

func TestOptimizerWeightsService_FetchCurrentPrices_PartialCache(t *testing.T) {
	// Setup: One price cached, one not
	cachedPrice := 150.0
	fetchedPrice := 200.0

	mockClientDataRepo := &mockClientDataRepoForWeights{
		GetIfFreshFunc: func(table, key string) (json.RawMessage, error) {
			if key == "US0378331005" { // AAPL cached
				data, _ := json.Marshal(cachedPrice)
				return data, nil
			}
			return nil, nil // GOOGL not cached
		},
		StoredData: make(map[string]json.RawMessage),
		StoredTTLs: make(map[string]time.Duration),
	}

	mockMarketHours := &mockMarketHoursServiceForWeights{
		AnyMajorMarketOpenFunc: func(t time.Time) bool {
			return true
		},
	}

	// Price client should only be called for missing prices
	var fetchedSymbols []string
	mockPriceClient := &mockPriceClientForWeights{
		GetBatchQuotesFunc: func(symbolMap map[string]*string) (map[string]*float64, error) {
			for symbol := range symbolMap {
				fetchedSymbols = append(fetchedSymbols, symbol)
			}
			return map[string]*float64{
				"GOOGL": &fetchedPrice,
			}, nil
		},
	}

	mockPriceConversion := &mockPriceConversionServiceForWeights{
		ConvertPricesToEURFunc: func(prices map[string]float64, securities []universe.Security) map[string]float64 {
			return prices
		},
	}

	service := NewOptimizerWeightsService(
		nil, nil, nil, nil,
		nil, // brokerClient
		nil,
		mockPriceConversion,
		nil,
		mockClientDataRepo,
		mockMarketHours,
	)
	service.priceClient = mockPriceClient

	_ = service
	_ = fetchedSymbols

	// Test via CalculateWeights when service is implemented
	t.Skip("Will test via CalculateWeights once service is implemented")
}

func TestOptimizerWeightsService_FetchCurrentPrices_MarketOpenTTL(t *testing.T) {
	mockClientDataRepo := &mockClientDataRepoForWeights{
		GetIfFreshFunc: func(table, key string) (json.RawMessage, error) {
			return nil, nil // Cache miss
		},
		StoredData: make(map[string]json.RawMessage),
		StoredTTLs: make(map[string]time.Duration),
	}

	mockMarketHours := &mockMarketHoursServiceForWeights{
		AnyMajorMarketOpenFunc: func(t time.Time) bool {
			return true // Markets OPEN
		},
	}

	fetchedPrice := 150.0
	mockPriceClient := &mockPriceClientForWeights{
		GetBatchQuotesFunc: func(symbolMap map[string]*string) (map[string]*float64, error) {
			return map[string]*float64{"AAPL": &fetchedPrice}, nil
		},
	}

	mockPriceConversion := &mockPriceConversionServiceForWeights{
		ConvertPricesToEURFunc: func(prices map[string]float64, securities []universe.Security) map[string]float64 {
			return prices
		},
	}

	service := NewOptimizerWeightsService(
		nil, nil, nil, nil,
		nil, // brokerClient
		nil,
		mockPriceConversion,
		nil,
		mockClientDataRepo,
		mockMarketHours,
	)
	service.priceClient = mockPriceClient

	_ = service

	// Test via CalculateWeights when service is implemented
	// Verify TTL is 30 minutes when markets open
	t.Skip("Will test via CalculateWeights once service is implemented")
}

func TestOptimizerWeightsService_FetchCurrentPrices_MarketClosedTTL(t *testing.T) {
	mockClientDataRepo := &mockClientDataRepoForWeights{
		GetIfFreshFunc: func(table, key string) (json.RawMessage, error) {
			return nil, nil // Cache miss
		},
		StoredData: make(map[string]json.RawMessage),
		StoredTTLs: make(map[string]time.Duration),
	}

	mockMarketHours := &mockMarketHoursServiceForWeights{
		AnyMajorMarketOpenFunc: func(t time.Time) bool {
			return false // Markets CLOSED
		},
	}

	fetchedPrice := 150.0
	mockPriceClient := &mockPriceClientForWeights{
		GetBatchQuotesFunc: func(symbolMap map[string]*string) (map[string]*float64, error) {
			return map[string]*float64{"AAPL": &fetchedPrice}, nil
		},
	}

	mockPriceConversion := &mockPriceConversionServiceForWeights{
		ConvertPricesToEURFunc: func(prices map[string]float64, securities []universe.Security) map[string]float64 {
			return prices
		},
	}

	service := NewOptimizerWeightsService(
		nil, nil, nil, nil,
		nil, // brokerClient
		nil,
		mockPriceConversion,
		nil,
		mockClientDataRepo,
		mockMarketHours,
	)
	service.priceClient = mockPriceClient

	_ = service

	// Test via CalculateWeights when service is implemented
	// Verify TTL is 24 hours when markets closed
	t.Skip("Will test via CalculateWeights once service is implemented")
}

func TestOptimizerWeightsService_FetchCurrentPrices_APIFallbackToStale(t *testing.T) {
	stalePrice := 145.0 // Stale cached price

	mockClientDataRepo := &mockClientDataRepoForWeights{
		GetIfFreshFunc: func(table, key string) (json.RawMessage, error) {
			return nil, nil // No fresh cache
		},
		GetFunc: func(table, key string) (json.RawMessage, error) {
			// Return stale data
			data, _ := json.Marshal(stalePrice)
			return data, nil
		},
	}

	mockMarketHours := &mockMarketHoursServiceForWeights{
		AnyMajorMarketOpenFunc: func(t time.Time) bool {
			return true
		},
	}

	// API fails
	mockPriceClient := &mockPriceClientForWeights{
		GetBatchQuotesFunc: func(symbolMap map[string]*string) (map[string]*float64, error) {
			return nil, assert.AnError // API error
		},
	}

	service := NewOptimizerWeightsService(
		nil, nil, nil, nil,
		nil, // brokerClient
		nil, nil, nil,
		mockClientDataRepo,
		mockMarketHours,
	)
	service.priceClient = mockPriceClient

	_ = service

	// Test via CalculateWeights when service is implemented
	// Should fallback to stale cache when API fails
	t.Skip("Will test via CalculateWeights once service is implemented")
}
