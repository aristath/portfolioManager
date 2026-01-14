package universe

import (
	"database/sql"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDBForIndexFiltering creates a test database with securities including indices
func setupTestDBForIndexFiltering(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create securities table with full schema
	_, err = db.Exec(`
		CREATE TABLE securities (
			isin TEXT PRIMARY KEY,
			symbol TEXT NOT NULL,
			name TEXT NOT NULL,
			product_type TEXT,
			industry TEXT,
			geography TEXT,
			fullExchangeName TEXT,
			market_code TEXT,
			priority_multiplier REAL DEFAULT 1.0,
			min_lot INTEGER DEFAULT 1,
			active INTEGER DEFAULT 1,
			allow_buy INTEGER DEFAULT 1,
			allow_sell INTEGER DEFAULT 1,
			currency TEXT,
			last_synced INTEGER,
			min_portfolio_target REAL,
			max_portfolio_target REAL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	require.NoError(t, err)

	// Create indices
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_securities_symbol ON securities(symbol)`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_securities_market_code ON securities(market_code)`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_securities_product_type ON securities(product_type)`)
	require.NoError(t, err)

	// Create tags tables (needed for scanSecurity)
	_, err = db.Exec(`
		CREATE TABLE tags (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE security_tags (
			isin TEXT NOT NULL,
			tag_id TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			PRIMARY KEY (isin, tag_id)
		)
	`)
	require.NoError(t, err)

	return db
}

// insertTestSecurities inserts test securities including indices
func insertTestSecurities(t *testing.T, db *sql.DB) {
	now := time.Now().Unix()

	// Insert regular securities (EQUITY, ETF)
	_, err := db.Exec(`
		INSERT INTO securities (isin, symbol, name, product_type, market_code, fullExchangeName, active, created_at, updated_at) VALUES
		('US0378331005', 'AAPL.US', 'Apple Inc.', 'EQUITY', 'FIX', 'NASDAQ', 1, ?, ?),
		('US5949181045', 'MSFT.US', 'Microsoft Corp', 'EQUITY', 'FIX', 'NASDAQ', 1, ?, ?),
		('IE00B3XXRP09', 'VUSA.EU', 'Vanguard S&P 500 ETF', 'ETF', 'EU', 'LSE', 1, ?, ?),
		('US0000000001', 'NULL_TYPE.US', 'Security with NULL type', NULL, 'FIX', 'NYSE', 1, ?, ?)
	`, now, now, now, now, now, now, now, now)
	require.NoError(t, err)

	// Insert market indices (should be excluded from tradable queries)
	_, err = db.Exec(`
		INSERT INTO securities (isin, symbol, name, product_type, market_code, fullExchangeName, active, allow_buy, allow_sell, created_at, updated_at) VALUES
		('INDEX-SP500.IDX', 'SP500.IDX', 'S&P 500', 'INDEX', 'FIX', NULL, 1, 0, 0, ?, ?),
		('INDEX-NASDAQ.IDX', 'NASDAQ.IDX', 'NASDAQ Composite', 'INDEX', 'FIX', NULL, 1, 0, 0, ?, ?),
		('INDEX-DAX.IDX', 'DAX.IDX', 'DAX (Germany)', 'INDEX', 'EU', NULL, 1, 0, 0, ?, ?)
	`, now, now, now, now, now, now)
	require.NoError(t, err)

	// Insert inactive security (should be excluded regardless of type)
	_, err = db.Exec(`
		INSERT INTO securities (isin, symbol, name, product_type, market_code, active, created_at, updated_at)
		VALUES ('US0000000002', 'INACTIVE.US', 'Inactive Security', 'EQUITY', 'FIX', 0, ?, ?)
	`, now, now)
	require.NoError(t, err)
}

func TestGetAllActive_ExcludesIndices(t *testing.T) {
	db := setupTestDBForIndexFiltering(t)
	defer db.Close()
	insertTestSecurities(t, db)

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	securities, err := repo.GetAllActive()
	require.NoError(t, err)

	// Should return 4 securities (3 tradable + 1 NULL type), excluding 3 indices and 1 inactive
	assert.Len(t, securities, 4)

	// Verify no indices in result
	for _, sec := range securities {
		assert.NotEqual(t, "INDEX", sec.ProductType, "Index %s should be excluded", sec.Symbol)
		assert.NotContains(t, sec.Symbol, ".IDX", "Index symbol %s should be excluded", sec.Symbol)
	}
}

func TestGetAllActive_IncludesNullProductType(t *testing.T) {
	db := setupTestDBForIndexFiltering(t)
	defer db.Close()
	insertTestSecurities(t, db)

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	securities, err := repo.GetAllActive()
	require.NoError(t, err)

	// Verify NULL product_type security is included
	var foundNullType bool
	for _, sec := range securities {
		if sec.Symbol == "NULL_TYPE.US" {
			foundNullType = true
			assert.Equal(t, "", sec.ProductType) // NULL scans as empty string
			break
		}
	}
	assert.True(t, foundNullType, "Security with NULL product_type should be included")
}

func TestGetAllActiveTradable_ExcludesIndices(t *testing.T) {
	db := setupTestDBForIndexFiltering(t)
	defer db.Close()
	insertTestSecurities(t, db)

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	securities, err := repo.GetAllActiveTradable()
	require.NoError(t, err)

	// Should return 4 tradable securities, excluding indices
	assert.Len(t, securities, 4)

	// Verify no indices in result
	for _, sec := range securities {
		assert.NotEqual(t, "INDEX", sec.ProductType, "Index %s should be excluded", sec.Symbol)
	}
}

func TestGetByMarketCode_ExcludesIndices(t *testing.T) {
	db := setupTestDBForIndexFiltering(t)
	defer db.Close()
	insertTestSecurities(t, db)

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Test FIX market code - should return AAPL, MSFT, NULL_TYPE but not SP500.IDX, NASDAQ.IDX
	fixSecurities, err := repo.GetByMarketCode("FIX")
	require.NoError(t, err)
	assert.Len(t, fixSecurities, 3) // AAPL, MSFT, NULL_TYPE

	for _, sec := range fixSecurities {
		assert.NotEqual(t, "INDEX", sec.ProductType, "Index %s should be excluded", sec.Symbol)
	}

	// Test EU market code - should return VUSA but not DAX.IDX
	euSecurities, err := repo.GetByMarketCode("EU")
	require.NoError(t, err)
	assert.Len(t, euSecurities, 1) // VUSA only
	assert.Equal(t, "VUSA.EU", euSecurities[0].Symbol)
}

func TestGetDistinctExchanges_ExcludesIndices(t *testing.T) {
	db := setupTestDBForIndexFiltering(t)
	defer db.Close()
	insertTestSecurities(t, db)

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	exchanges, err := repo.GetDistinctExchanges()
	require.NoError(t, err)

	// Should return NASDAQ, LSE, NYSE (from tradable securities)
	// Indices have NULL fullExchangeName, so they wouldn't appear anyway
	assert.Contains(t, exchanges, "NASDAQ")
	assert.Contains(t, exchanges, "LSE")
	assert.Contains(t, exchanges, "NYSE")
}

func TestGetByTags_ExcludesIndices(t *testing.T) {
	db := setupTestDBForIndexFiltering(t)
	defer db.Close()
	insertTestSecurities(t, db)

	now := time.Now().Unix()

	// Create a tag
	_, err := db.Exec(`INSERT INTO tags (id, name, created_at, updated_at) VALUES ('test-tag', 'Test Tag', ?, ?)`, now, now)
	require.NoError(t, err)

	// Associate tag with both a regular security and an index
	_, err = db.Exec(`INSERT INTO security_tags (isin, tag_id, created_at, updated_at) VALUES
		('US0378331005', 'test-tag', ?, ?),
		('INDEX-SP500.IDX', 'test-tag', ?, ?)
	`, now, now, now, now)
	require.NoError(t, err)

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	securities, err := repo.GetByTags([]string{"test-tag"})
	require.NoError(t, err)

	// Should return only AAPL (regular security), not SP500.IDX (index)
	assert.Len(t, securities, 1)
	assert.Equal(t, "AAPL.US", securities[0].Symbol)
}
