package universe

import (
	"database/sql"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// TestSecurityRetrievalPreservesLotSize verifies that lot size is preserved
// through the full data flow: JSON storage → parse → defaults → retrieval
func TestSecurityRetrievalPreservesLotSize(t *testing.T) {
	// Setup: Create test database with JSON-based schema
	db := setupTestDBForLotSizeTest(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Insert security with lot size 100 in JSON data (simulating Tradernet API response)
	jsonData := `{
		"ticker": "CAT.3750.AS",
		"issue_nb": "CNE100006WS8",
		"name": "Caterpillar Inc. 3.75% Notes",
		"quotes": {
			"x_lot": 100,
			"min_step": 0.01
		}
	}`

	_, err := db.Exec(`
		INSERT INTO securities (isin, symbol, data, last_synced)
		VALUES (?, ?, ?, NULL)
	`, "CNE100006WS8", "CAT.3750.AS", jsonData)
	require.NoError(t, err)

	// Execute: Retrieve security through repository
	security, err := repo.GetByISIN("CNE100006WS8")

	// Assert: Lot size must be preserved from JSON data
	require.NoError(t, err)
	require.NotNil(t, security)
	assert.Equal(t, "CNE100006WS8", security.ISIN)
	assert.Equal(t, "CAT.3750.AS", security.Symbol)
	assert.Equal(t, 100, security.MinLot,
		"MinLot must be preserved from JSON data through parse → defaults → retrieval")

	// Verify other defaults were applied
	assert.True(t, security.AllowBuy, "AllowBuy should be defaulted to true")
	assert.True(t, security.AllowSell, "AllowSell should be defaulted to true")
	assert.Equal(t, 1.0, security.PriorityMultiplier, "PriorityMultiplier should be defaulted to 1.0")
}

// TestSecurityRetrievalDefaultsLotSizeWhenMissing verifies that lot size
// defaults to 1 when not provided by the API
func TestSecurityRetrievalDefaultsLotSizeWhenMissing(t *testing.T) {
	// Setup
	db := setupTestDBForLotSizeTest(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Insert security WITHOUT lot size in JSON data
	jsonData := `{
		"ticker": "TEST.US",
		"issue_nb": "US1234567890",
		"name": "Test Security",
		"quotes": {
			"min_step": 0.01
		}
	}`

	_, err := db.Exec(`
		INSERT INTO securities (isin, symbol, data, last_synced)
		VALUES (?, ?, ?, NULL)
	`, "US1234567890", "TEST.US", jsonData)
	require.NoError(t, err)

	// Execute
	security, err := repo.GetByISIN("US1234567890")

	// Assert: MinLot should default to 1 when not provided
	require.NoError(t, err)
	require.NotNil(t, security)
	assert.Equal(t, 1, security.MinLot,
		"MinLot should default to 1 when not provided by API")
}

// TestSecurityRetrievalWithVariousLotSizes tests multiple securities with different lot sizes
func TestSecurityRetrievalWithVariousLotSizes(t *testing.T) {
	// Setup
	db := setupTestDBForLotSizeTest(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Test cases with different lot sizes
	testCases := []struct {
		isin     string
		symbol   string
		lotSize  int
		jsonData string
	}{
		{
			"US0001", "LOT1.US", 1,
			`{"ticker": "LOT1.US", "issue_nb": "US0001", "name": "Test Security", "quotes": {"x_lot": 1, "min_step": 0.01}}`,
		},
		{
			"US0010", "LOT10.US", 10,
			`{"ticker": "LOT10.US", "issue_nb": "US0010", "name": "Test Security", "quotes": {"x_lot": 10, "min_step": 0.01}}`,
		},
		{
			"US0050", "LOT50.US", 50,
			`{"ticker": "LOT50.US", "issue_nb": "US0050", "name": "Test Security", "quotes": {"x_lot": 50, "min_step": 0.01}}`,
		},
		{
			"US0100", "LOT100.US", 100,
			`{"ticker": "LOT100.US", "issue_nb": "US0100", "name": "Test Security", "quotes": {"x_lot": 100, "min_step": 0.01}}`,
		},
		{
			"US1000", "LOT1000.US", 1000,
			`{"ticker": "LOT1000.US", "issue_nb": "US1000", "name": "Test Security", "quotes": {"x_lot": 1000, "min_step": 0.01}}`,
		},
	}

	// Insert test securities
	for _, tc := range testCases {
		_, err := db.Exec(`
			INSERT INTO securities (isin, symbol, data, last_synced)
			VALUES (?, ?, ?, NULL)
		`, tc.isin, tc.symbol, tc.jsonData)
		require.NoError(t, err)
	}

	// Verify each security preserves its lot size
	for _, tc := range testCases {
		security, err := repo.GetByISIN(tc.isin)
		require.NoError(t, err)
		require.NotNil(t, security)
		assert.Equal(t, tc.lotSize, security.MinLot,
			"Security %s should have lot size %d", tc.symbol, tc.lotSize)
	}
}

// setupTestDBForLotSizeTest creates a test database with the JSON-based securities schema
func setupTestDBForLotSizeTest(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create securities table with JSON storage (migration 038 schema)
	_, err = db.Exec(`
		CREATE TABLE securities (
			isin TEXT PRIMARY KEY,
			symbol TEXT NOT NULL,
			data TEXT NOT NULL,
			last_synced INTEGER
		) STRICT
	`)
	require.NoError(t, err)

	// Create index on symbol for lookups
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_securities_symbol ON securities(symbol)`)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	return db
}
