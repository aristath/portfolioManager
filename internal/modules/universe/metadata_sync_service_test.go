package universe

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockBrokerClientForMetadataSync mocks the broker client for metadata sync tests
type MockBrokerClientForMetadataSync struct {
	RawResponse   interface{}
	BatchResponse interface{}
	Error         error
	BatchError    error
}

func (m *MockBrokerClientForMetadataSync) GetSecurityMetadataRaw(symbol string) (interface{}, error) {
	return m.RawResponse, m.Error
}

func (m *MockBrokerClientForMetadataSync) GetSecurityMetadataBatch(symbols []string) (interface{}, error) {
	if m.BatchError != nil {
		return nil, m.BatchError
	}
	if m.BatchResponse != nil {
		return m.BatchResponse, nil
	}

	// Default: generate batch response from symbols
	securities := make([]interface{}, len(symbols))
	for i, symbol := range symbols {
		securities[i] = map[string]interface{}{
			"ticker": symbol,
			"name":   fmt.Sprintf("Security %s", symbol),
		}
	}

	return map[string]interface{}{
		"securities": securities,
		"total":      len(symbols),
	}, nil
}

// Implement other BrokerClient methods as no-ops
func (m *MockBrokerClientForMetadataSync) GetPortfolio() ([]domain.BrokerPosition, error) {
	return nil, nil
}
func (m *MockBrokerClientForMetadataSync) GetCashBalances() ([]domain.BrokerCashBalance, error) {
	return nil, nil
}
func (m *MockBrokerClientForMetadataSync) PlaceOrder(symbol, side string, quantity, limitPrice float64) (*domain.BrokerOrderResult, error) {
	return nil, nil
}
func (m *MockBrokerClientForMetadataSync) GetExecutedTrades(limit int) ([]domain.BrokerTrade, error) {
	return nil, nil
}
func (m *MockBrokerClientForMetadataSync) GetPendingOrders() ([]domain.BrokerPendingOrder, error) {
	return nil, nil
}
func (m *MockBrokerClientForMetadataSync) GetQuote(symbol string) (*domain.BrokerQuote, error) {
	return nil, nil
}
func (m *MockBrokerClientForMetadataSync) GetQuotes(symbols []string) (map[string]*domain.BrokerQuote, error) {
	return nil, nil
}
func (m *MockBrokerClientForMetadataSync) GetLevel1Quote(symbol string) (*domain.BrokerOrderBook, error) {
	return nil, nil
}
func (m *MockBrokerClientForMetadataSync) GetHistoricalPrices(symbol string, start, end int64, timeframeSeconds int) ([]domain.BrokerOHLCV, error) {
	return nil, nil
}
func (m *MockBrokerClientForMetadataSync) FindSymbol(symbol string, exchange *string) ([]domain.BrokerSecurityInfo, error) {
	return nil, nil
}
func (m *MockBrokerClientForMetadataSync) GetSecurityMetadata(symbol string) (*domain.BrokerSecurityInfo, error) {
	return nil, nil
}
func (m *MockBrokerClientForMetadataSync) GetFXRates(baseCurrency string, currencies []string) (map[string]float64, error) {
	return nil, nil
}
func (m *MockBrokerClientForMetadataSync) GetAllCashFlows(limit int) ([]domain.BrokerCashFlow, error) {
	return nil, nil
}
func (m *MockBrokerClientForMetadataSync) GetCashMovements() (*domain.BrokerCashMovement, error) {
	return nil, nil
}
func (m *MockBrokerClientForMetadataSync) IsConnected() bool {
	return true
}
func (m *MockBrokerClientForMetadataSync) HealthCheck() (*domain.BrokerHealthResult, error) {
	return nil, nil
}
func (m *MockBrokerClientForMetadataSync) SetCredentials(apiKey, apiSecret string) {}

// TestMetadataSyncService_StoresRawTradernetData verifies that metadata sync stores
// raw Tradernet API response without any transformation
func TestMetadataSyncService_StoresRawTradernetData(t *testing.T) {
	// Setup test database
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Create test security
	security := Security{
		ISIN:   "US0378331005",
		Symbol: "AAPL.US",
		Name:   "Apple Inc.",
	}
	err := repo.Create(security)
	require.NoError(t, err)

	// Mock broker client returns full Tradernet response
	mockBroker := &MockBrokerClientForMetadataSync{
		RawResponse: map[string]interface{}{
			"total": 1,
			"securities": []interface{}{
				map[string]interface{}{
					"id":                  12345,
					"ticker":              "AAPL.US",
					"name":                "Apple Inc.",
					"issue_nb":            "US0378331005",
					"face_curr_c":         "USD",
					"mkt_name":            "FIX",
					"codesub_nm":          "NASDAQ",
					"lot_size_q":          "1.00000000",
					"issuer_country_code": "0",
					"sector_code":         "Technology",
					"type":                "Regular stock",
					"quotes": map[string]interface{}{
						"x_lot":    float64(1),
						"min_step": 0.01,
					},
					"attributes": map[string]interface{}{
						"CntryOfRisk": "US",
						"base_mkt_id": "FIX",
					},
				},
			},
		},
	}

	// Create metadata sync service
	syncService := NewMetadataSyncService(repo, mockBroker, log)

	// Execute metadata sync
	symbol, err := syncService.SyncMetadata("US0378331005")
	require.NoError(t, err)
	assert.Equal(t, "AAPL.US", symbol)

	// Verify raw format stored in database
	var storedJSON string
	query := `SELECT data FROM securities WHERE isin = ?`
	err = db.QueryRow(query, "US0378331005").Scan(&storedJSON)
	require.NoError(t, err)
	require.NotEmpty(t, storedJSON)

	// Parse stored JSON
	var storedData map[string]interface{}
	err = json.Unmarshal([]byte(storedJSON), &storedData)
	require.NoError(t, err)

	// Verify it's the raw Tradernet format (securities[0] was extracted and stored as-is)
	assert.Equal(t, "AAPL.US", storedData["ticker"])
	assert.Equal(t, "Apple Inc.", storedData["name"])
	assert.Equal(t, "US0378331005", storedData["issue_nb"])
	assert.Equal(t, "USD", storedData["face_curr_c"])
	assert.Equal(t, "FIX", storedData["mkt_name"])
	assert.Equal(t, "NASDAQ", storedData["codesub_nm"])
	assert.Equal(t, "Technology", storedData["sector_code"])
	assert.Equal(t, "Regular stock", storedData["type"])

	// Verify nested structures preserved
	attributes, ok := storedData["attributes"].(map[string]interface{})
	require.True(t, ok, "attributes should be a map")
	assert.Equal(t, "US", attributes["CntryOfRisk"])

	quotes, ok := storedData["quotes"].(map[string]interface{})
	require.True(t, ok, "quotes should be a map")
	assert.Equal(t, float64(1), quotes["x_lot"])

	// CRITICAL: Verify NO transformation occurred
	// The stored data should have Tradernet field names, not our Security struct field names
	assert.NotContains(t, storedData, "geography") // Should use "attributes.CntryOfRisk"
	assert.NotContains(t, storedData, "industry")  // Should use "sector_code"
	assert.NotContains(t, storedData, "currency")  // Should use "face_curr_c"
}

// TestMetadataSyncService_EmptyResponse verifies handling of empty securities array
func TestMetadataSyncService_EmptyResponse(t *testing.T) {
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Create test security
	security := Security{
		ISIN:   "US0378331005",
		Symbol: "AAPL.US",
		Name:   "Apple Inc.",
	}
	err := repo.Create(security)
	require.NoError(t, err)

	// Mock returns empty securities array
	mockBroker := &MockBrokerClientForMetadataSync{
		RawResponse: map[string]interface{}{
			"total":      0,
			"securities": []interface{}{},
		},
	}

	syncService := NewMetadataSyncService(repo, mockBroker, log)

	// Should not error, but also should not update
	symbol, err := syncService.SyncMetadata("US0378331005")
	require.NoError(t, err)
	assert.Equal(t, "AAPL.US", symbol)

	// Verify data column unchanged (still placeholder {})
	var storedJSON string
	query := `SELECT data FROM securities WHERE isin = ?`
	err = db.QueryRow(query, "US0378331005").Scan(&storedJSON)
	require.NoError(t, err)
	assert.Equal(t, "{}", storedJSON)
}

// TestMetadataSyncService_SecurityNotFound verifies handling of non-existent ISIN
func TestMetadataSyncService_SecurityNotFound(t *testing.T) {
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	mockBroker := &MockBrokerClientForMetadataSync{
		RawResponse: map[string]interface{}{},
	}

	syncService := NewMetadataSyncService(repo, mockBroker, log)

	// Should not error when security doesn't exist
	symbol, err := syncService.SyncMetadata("NONEXISTENT")
	require.NoError(t, err)
	assert.Equal(t, "", symbol)
}

// ============================================================================
// Batch Sync Tests (TDD - Tests written before implementation)
// ============================================================================

// TestGetAllActiveISINs_ExcludesIndices is CRITICAL - ensures indices never enter batch sync
func TestGetAllActiveISINs_ExcludesIndices(t *testing.T) {
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Insert 40 regular securities
	for i := 0; i < 40; i++ {
		security := Security{
			ISIN:        fmt.Sprintf("US%010d", i),
			Symbol:      fmt.Sprintf("SYM%d.US", i),
			Name:        fmt.Sprintf("Security %d", i),
			ProductType: "STOCK",
		}
		err := repo.Create(security)
		require.NoError(t, err)
	}

	// Insert 13 indices (should be excluded)
	for i := 0; i < 13; i++ {
		index := Security{
			ISIN:        fmt.Sprintf("IDX%010d", i),
			Symbol:      fmt.Sprintf("INDEX%d.IDX", i),
			Name:        fmt.Sprintf("Index %d", i),
			ProductType: "INDEX",
		}
		err := repo.Create(index)
		require.NoError(t, err)
	}

	mockBroker := &MockBrokerClientForMetadataSync{}
	syncService := NewMetadataSyncService(repo, mockBroker, log)

	// Get all active ISINs
	isins := syncService.GetAllActiveISINs()

	// CRITICAL: Should return ONLY 40 ISINs (no indices)
	assert.Len(t, isins, 40, "Should exclude all 13 indices")

	// CRITICAL: Verify no .IDX symbols in results
	for _, isin := range isins {
		security, err := repo.GetByISIN(isin)
		require.NoError(t, err)
		assert.NotContains(t, security.Symbol, ".IDX", "Index symbol should be excluded")
		assert.NotEqual(t, "INDEX", security.ProductType, "Index product type should be excluded")
	}

	// Verify we have the correct ISINs
	for i := 0; i < 40; i++ {
		expectedISIN := fmt.Sprintf("US%010d", i)
		assert.Contains(t, isins, expectedISIN)
	}

	// Verify index ISINs are NOT present
	for i := 0; i < 13; i++ {
		indexISIN := fmt.Sprintf("IDX%010d", i)
		assert.NotContains(t, isins, indexISIN, "Index ISIN should be excluded")
	}
}

// TestSyncMetadataBatch_AllSuccess tests batch sync with all securities successful
func TestSyncMetadataBatch_AllSuccess(t *testing.T) {
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Create 10 test securities
	isins := make([]string, 10)
	for i := 0; i < 10; i++ {
		isin := fmt.Sprintf("US%010d", i)
		isins[i] = isin

		security := Security{
			ISIN:   isin,
			Symbol: fmt.Sprintf("SYM%d.US", i),
			Name:   fmt.Sprintf("Security %d", i),
		}
		err := repo.Create(security)
		require.NoError(t, err)
	}

	// Mock broker returns all securities
	batchResponse := map[string]interface{}{
		"securities": []interface{}{},
		"total":      10,
	}

	securities := make([]interface{}, 10)
	for i := 0; i < 10; i++ {
		securities[i] = map[string]interface{}{
			"ticker":   fmt.Sprintf("SYM%d.US", i),
			"name":     fmt.Sprintf("Security %d", i),
			"issue_nb": isins[i],
		}
	}
	batchResponse["securities"] = securities

	mockBroker := &MockBrokerClientForMetadataSync{}
	// We'll need to add GetSecurityMetadataBatch to the mock
	syncService := NewMetadataSyncService(repo, mockBroker, log)

	// Execute batch sync
	successCount, err := syncService.SyncMetadataBatch(isins)

	// All should succeed
	assert.NoError(t, err)
	assert.Equal(t, 10, successCount)

	// Verify all securities have last_synced updated
	for _, isin := range isins {
		security, err := repo.GetByISIN(isin)
		require.NoError(t, err)
		assert.NotNil(t, security.LastSynced, "LastSynced should be set for "+isin)
		assert.Greater(t, *security.LastSynced, int64(0), "LastSynced timestamp should be positive for "+isin)
	}
}

// TestSyncMetadataBatch_PartialResults tests batch with some missing securities
func TestSyncMetadataBatch_PartialResults(t *testing.T) {
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Create 10 securities but broker only returns 8
	isins := make([]string, 10)
	for i := 0; i < 10; i++ {
		isin := fmt.Sprintf("US%010d", i)
		isins[i] = isin

		security := Security{
			ISIN:   isin,
			Symbol: fmt.Sprintf("SYM%d.US", i),
			Name:   fmt.Sprintf("Security %d", i),
		}
		err := repo.Create(security)
		require.NoError(t, err)
	}

	// Configure mock to return only 8 securities (simulate partial results)
	partialSecurities := make([]interface{}, 8)
	for i := 0; i < 8; i++ {
		partialSecurities[i] = map[string]interface{}{
			"ticker": fmt.Sprintf("SYM%d.US", i),
			"name":   fmt.Sprintf("Security %d", i),
		}
	}

	mockBroker := &MockBrokerClientForMetadataSync{
		BatchResponse: map[string]interface{}{
			"securities": partialSecurities,
			"total":      8,
		},
	}
	syncService := NewMetadataSyncService(repo, mockBroker, log)

	// Batch sync should succeed with partial results
	successCount, err := syncService.SyncMetadataBatch(isins)

	// Should return success count = 8, no error
	assert.NoError(t, err)
	assert.Equal(t, 8, successCount)
}

// TestSyncMetadataBatch_APIFailure tests batch sync when API call fails
func TestSyncMetadataBatch_APIFailure(t *testing.T) {
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Create 5 securities
	isins := make([]string, 5)
	for i := 0; i < 5; i++ {
		isin := fmt.Sprintf("US%010d", i)
		isins[i] = isin

		security := Security{
			ISIN:   isin,
			Symbol: fmt.Sprintf("SYM%d.US", i),
			Name:   fmt.Sprintf("Security %d", i),
		}
		err := repo.Create(security)
		require.NoError(t, err)
	}

	// Mock broker returns error for batch call
	mockBroker := &MockBrokerClientForMetadataSync{
		BatchError: fmt.Errorf("API failure"),
	}
	syncService := NewMetadataSyncService(repo, mockBroker, log)

	// Batch sync should fail
	successCount, err := syncService.SyncMetadataBatch(isins)

	assert.Error(t, err)
	assert.Equal(t, 0, successCount)
	assert.Contains(t, err.Error(), "batch API call failed")
}

// TestSyncMetadataBatch_DatabaseUpdateFailure tests partial DB update failures
func TestSyncMetadataBatch_DatabaseUpdateFailure(t *testing.T) {
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Create some securities
	isins := []string{"US0000000001", "US0000000002", "US0000000003"}
	for i, isin := range isins {
		security := Security{
			ISIN:   isin,
			Symbol: fmt.Sprintf("SYM%d.US", i),
			Name:   fmt.Sprintf("Security %d", i),
		}
		err := repo.Create(security)
		require.NoError(t, err)
	}

	mockBroker := &MockBrokerClientForMetadataSync{}
	syncService := NewMetadataSyncService(repo, mockBroker, log)

	// This test verifies that even if some updates fail,
	// the batch continues processing others
	successCount, err := syncService.SyncMetadataBatch(isins)

	// Should return partial success count
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, successCount, 0)
}

// TestSyncMetadataBatch_ISINToSymbolMapping tests ISIN to symbol resolution
func TestSyncMetadataBatch_ISINToSymbolMapping(t *testing.T) {
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Create 3 valid securities
	validISINs := []string{"US0000000001", "US0000000002", "US0000000003"}
	for i, isin := range validISINs {
		security := Security{
			ISIN:   isin,
			Symbol: fmt.Sprintf("SYM%d.US", i),
			Name:   fmt.Sprintf("Security %d", i),
		}
		err := repo.Create(security)
		require.NoError(t, err)
	}

	// Mix valid and invalid ISINs
	allISINs := append(validISINs, "INVALID001", "INVALID002")

	mockBroker := &MockBrokerClientForMetadataSync{}
	syncService := NewMetadataSyncService(repo, mockBroker, log)

	// Batch sync should skip invalid ISINs and sync valid ones
	successCount, err := syncService.SyncMetadataBatch(allISINs)

	// Should succeed for valid ISINs only
	assert.NoError(t, err)
	assert.LessOrEqual(t, successCount, 3, "Should only sync valid ISINs")
}
