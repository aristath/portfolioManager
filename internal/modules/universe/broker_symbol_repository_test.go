package universe

import (
	"database/sql"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDBForBrokerSymbols(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create securities table (required for foreign key)
	_, err = db.Exec(`
		CREATE TABLE securities (
			isin TEXT PRIMARY KEY,
			symbol TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	require.NoError(t, err)

	// Create broker_symbols table
	_, err = db.Exec(`
		CREATE TABLE broker_symbols (
			isin TEXT NOT NULL,
			broker_name TEXT NOT NULL,
			broker_symbol TEXT NOT NULL,
			PRIMARY KEY (isin, broker_name),
			FOREIGN KEY (isin) REFERENCES securities(isin) ON DELETE CASCADE
		)
	`)
	require.NoError(t, err)

	// Create indexes
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_broker_symbols_isin ON broker_symbols(isin)`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_broker_symbols_broker ON broker_symbols(broker_name)`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_broker_symbols_symbol ON broker_symbols(broker_symbol)`)
	require.NoError(t, err)

	// Insert a test security for foreign key constraint
	_, err = db.Exec(`
		INSERT INTO securities (isin, symbol, name, created_at, updated_at)
		VALUES ('US0378331005', 'AAPL.US', 'Apple Inc.', 1704067200, 1704067200)
	`)
	require.NoError(t, err)

	return db
}

func TestGetBrokerSymbol_Success(t *testing.T) {
	db := setupTestDBForBrokerSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewBrokerSymbolRepository(db, log)

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO broker_symbols (isin, broker_name, broker_symbol)
		VALUES ('US0378331005', 'tradernet', 'AAPL.US')
	`)
	require.NoError(t, err)

	// Execute
	symbol, err := repo.GetBrokerSymbol("US0378331005", "tradernet")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "AAPL.US", symbol)
}

func TestGetBrokerSymbol_NotFound(t *testing.T) {
	db := setupTestDBForBrokerSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewBrokerSymbolRepository(db, log)

	// Execute
	symbol, err := repo.GetBrokerSymbol("US0378331005", "tradernet")

	// Assert
	require.Error(t, err)
	assert.Empty(t, symbol)
	assert.Contains(t, err.Error(), "not found")
}

func TestSetBrokerSymbol_Create(t *testing.T) {
	db := setupTestDBForBrokerSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewBrokerSymbolRepository(db, log)

	// Execute
	err := repo.SetBrokerSymbol("US0378331005", "tradernet", "AAPL.US")
	require.NoError(t, err)

	// Verify
	symbol, err := repo.GetBrokerSymbol("US0378331005", "tradernet")
	require.NoError(t, err)
	assert.Equal(t, "AAPL.US", symbol)
}

func TestSetBrokerSymbol_Update(t *testing.T) {
	db := setupTestDBForBrokerSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewBrokerSymbolRepository(db, log)

	// Insert initial mapping
	_, err := db.Exec(`
		INSERT INTO broker_symbols (isin, broker_name, broker_symbol)
		VALUES ('US0378331005', 'tradernet', 'AAPL.US')
	`)
	require.NoError(t, err)

	// Execute - update existing
	err = repo.SetBrokerSymbol("US0378331005", "tradernet", "AAPL")
	require.NoError(t, err)

	// Verify update
	symbol, err := repo.GetBrokerSymbol("US0378331005", "tradernet")
	require.NoError(t, err)
	assert.Equal(t, "AAPL", symbol)
}

func TestGetAllBrokerSymbols(t *testing.T) {
	db := setupTestDBForBrokerSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewBrokerSymbolRepository(db, log)

	// Insert multiple broker symbols for same ISIN
	_, err := db.Exec(`
		INSERT INTO broker_symbols (isin, broker_name, broker_symbol)
		VALUES
			('US0378331005', 'tradernet', 'AAPL.US'),
			('US0378331005', 'ibkr', 'AAPL')
	`)
	require.NoError(t, err)

	// Execute
	symbols, err := repo.GetAllBrokerSymbols("US0378331005")

	// Assert
	require.NoError(t, err)
	assert.Len(t, symbols, 2)
	assert.Equal(t, "AAPL.US", symbols["tradernet"])
	assert.Equal(t, "AAPL", symbols["ibkr"])
}

func TestGetAllBrokerSymbols_Empty(t *testing.T) {
	db := setupTestDBForBrokerSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewBrokerSymbolRepository(db, log)

	// Execute
	symbols, err := repo.GetAllBrokerSymbols("US0378331005")

	// Assert
	require.NoError(t, err)
	assert.Empty(t, symbols)
}

func TestGetISINByBrokerSymbol(t *testing.T) {
	db := setupTestDBForBrokerSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewBrokerSymbolRepository(db, log)

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO broker_symbols (isin, broker_name, broker_symbol)
		VALUES ('US0378331005', 'tradernet', 'AAPL.US')
	`)
	require.NoError(t, err)

	// Execute
	isin, err := repo.GetISINByBrokerSymbol("tradernet", "AAPL.US")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "US0378331005", isin)
}

func TestGetISINByBrokerSymbol_NotFound(t *testing.T) {
	db := setupTestDBForBrokerSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewBrokerSymbolRepository(db, log)

	// Execute
	isin, err := repo.GetISINByBrokerSymbol("tradernet", "AAPL.US")

	// Assert
	require.Error(t, err)
	assert.Empty(t, isin)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteBrokerSymbol(t *testing.T) {
	db := setupTestDBForBrokerSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewBrokerSymbolRepository(db, log)

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO broker_symbols (isin, broker_name, broker_symbol)
		VALUES ('US0378331005', 'tradernet', 'AAPL.US')
	`)
	require.NoError(t, err)

	// Execute
	err = repo.DeleteBrokerSymbol("US0378331005", "tradernet")
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetBrokerSymbol("US0378331005", "tradernet")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteBrokerSymbol_NotFound(t *testing.T) {
	db := setupTestDBForBrokerSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewBrokerSymbolRepository(db, log)

	// Execute - delete non-existent mapping
	err := repo.DeleteBrokerSymbol("US0378331005", "tradernet")

	// Assert - should not error (idempotent operation)
	require.NoError(t, err)
}
