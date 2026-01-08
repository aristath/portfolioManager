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

func setupTestDBWithISINPrimaryKey(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create securities table with ISIN as PRIMARY KEY (post-migration schema)
	_, err = db.Exec(`
		CREATE TABLE securities (
			isin TEXT PRIMARY KEY,
			symbol TEXT NOT NULL,
			yahoo_symbol TEXT,
			name TEXT NOT NULL,
			product_type TEXT,
			industry TEXT,
			country TEXT,
			fullExchangeName TEXT,
			priority_multiplier REAL DEFAULT 1.0,
			min_lot INTEGER DEFAULT 1,
			active INTEGER DEFAULT 1,
			allow_buy INTEGER DEFAULT 1,
			allow_sell INTEGER DEFAULT 1,
			currency TEXT,
			last_synced TEXT,
			min_portfolio_target REAL,
			max_portfolio_target REAL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	require.NoError(t, err)

	// Create index on symbol for lookups
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_securities_symbol ON securities(symbol)`)
	require.NoError(t, err)

	return db
}

func TestGetByISIN_PrimaryMethod(t *testing.T) {
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Insert test data
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := db.Exec(`
		INSERT INTO securities (isin, symbol, name, created_at, updated_at)
		VALUES ('US0378331005', 'AAPL.US', 'Apple Inc.', ?, ?)
	`, testDate.Unix(), testDate.Unix())
	require.NoError(t, err)

	// Execute
	security, err := repo.GetByISIN("US0378331005")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, security)
	assert.Equal(t, "US0378331005", security.ISIN)
	assert.Equal(t, "AAPL.US", security.Symbol)
	assert.Equal(t, "Apple Inc.", security.Name)
}

func TestGetByISIN_NotFound(t *testing.T) {
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Execute
	security, err := repo.GetByISIN("US0000000000")

	// Assert
	require.NoError(t, err)
	assert.Nil(t, security)
}

func TestGetBySymbol_HelperMethod_LooksUpISINFirst(t *testing.T) {
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Insert test data
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := db.Exec(`
		INSERT INTO securities (isin, symbol, name, created_at, updated_at)
		VALUES ('US0378331005', 'AAPL.US', 'Apple Inc.', ?, ?)
	`, testDate.Unix(), testDate.Unix())
	require.NoError(t, err)

	// Execute - GetBySymbol should lookup ISIN first, then query by ISIN
	security, err := repo.GetBySymbol("AAPL.US")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, security)
	assert.Equal(t, "US0378331005", security.ISIN)
	assert.Equal(t, "AAPL.US", security.Symbol)
}

func TestUpdate_ByISIN(t *testing.T) {
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Insert test data
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := db.Exec(`
		INSERT INTO securities (isin, symbol, name, active, created_at, updated_at)
		VALUES ('US0378331005', 'AAPL.US', 'Apple Inc.', 1, ?, ?)
	`, testDate.Unix(), testDate.Unix())
	require.NoError(t, err)

	// Execute - Update should use ISIN
	err = repo.Update("US0378331005", map[string]interface{}{
		"name":   "Apple Inc. Updated",
		"active": false,
	})
	require.NoError(t, err)

	// Verify update
	security, err := repo.GetByISIN("US0378331005")
	require.NoError(t, err)
	require.NotNil(t, security)
	assert.Equal(t, "Apple Inc. Updated", security.Name)
	assert.False(t, security.Active)
}

func TestDelete_ByISIN(t *testing.T) {
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Insert test data
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := db.Exec(`
		INSERT INTO securities (isin, symbol, name, active, created_at, updated_at)
		VALUES ('US0378331005', 'AAPL.US', 'Apple Inc.', 1, ?, ?)
	`, testDate.Unix(), testDate.Unix())
	require.NoError(t, err)

	// Execute - Delete should use ISIN
	err = repo.Delete("US0378331005")
	require.NoError(t, err)

	// Verify soft delete (active = 0)
	security, err := repo.GetByISIN("US0378331005")
	require.NoError(t, err)
	require.NotNil(t, security)
	assert.False(t, security.Active)
}

func TestCreate_WithISINAsPrimaryKey(t *testing.T) {
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Execute - Create should use ISIN as PRIMARY KEY
	security := Security{
		ISIN:      "US0378331005",
		Symbol:    "AAPL.US",
		Name:      "Apple Inc.",
		Active:    true,
		AllowBuy:  true,
		AllowSell: true,
	}

	err := repo.Create(security)
	require.NoError(t, err)

	// Verify creation
	created, err := repo.GetByISIN("US0378331005")
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "US0378331005", created.ISIN)
	assert.Equal(t, "AAPL.US", created.Symbol)
	assert.Equal(t, "Apple Inc.", created.Name)
}

func TestGetBySymbol_FallbackToSymbolLookup(t *testing.T) {
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Insert test data
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := db.Exec(`
		INSERT INTO securities (isin, symbol, name, created_at, updated_at)
		VALUES ('US0378331005', 'AAPL.US', 'Apple Inc.', ?, ?)
	`, testDate.Unix(), testDate.Unix())
	require.NoError(t, err)

	// GetBySymbol should lookup by symbol column (indexed)
	security, err := repo.GetBySymbol("AAPL.US")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, security)
	assert.Equal(t, "US0378331005", security.ISIN)
	assert.Equal(t, "AAPL.US", security.Symbol)
}

func TestGetByIdentifier_PrioritizesISIN(t *testing.T) {
	db := setupTestDBWithISINPrimaryKey(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewSecurityRepository(db, log)

	// Insert test data
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := db.Exec(`
		INSERT INTO securities (isin, symbol, name, created_at, updated_at)
		VALUES ('US0378331005', 'AAPL.US', 'Apple Inc.', ?, ?)
	`, testDate.Unix(), testDate.Unix())
	require.NoError(t, err)

	// Test with ISIN
	security, err := repo.GetByIdentifier("US0378331005")
	require.NoError(t, err)
	require.NotNil(t, security)
	assert.Equal(t, "US0378331005", security.ISIN)

	// Test with symbol
	security, err = repo.GetByIdentifier("AAPL.US")
	require.NoError(t, err)
	require.NotNil(t, security)
	assert.Equal(t, "AAPL.US", security.Symbol)
}
