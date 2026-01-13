package universe

import (
	"database/sql"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDBForClientSymbols(t *testing.T) *sql.DB {
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

	// Create client_symbols table
	_, err = db.Exec(`
		CREATE TABLE client_symbols (
			isin TEXT NOT NULL,
			client_name TEXT NOT NULL,
			client_symbol TEXT NOT NULL,
			PRIMARY KEY (isin, client_name),
			FOREIGN KEY (isin) REFERENCES securities(isin) ON DELETE CASCADE
		)
	`)
	require.NoError(t, err)

	// Create indexes
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_client_symbols_isin ON client_symbols(isin)`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_client_symbols_client ON client_symbols(client_name)`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_client_symbols_symbol ON client_symbols(client_symbol)`)
	require.NoError(t, err)

	// Insert a test security for foreign key constraint
	_, err = db.Exec(`
		INSERT INTO securities (isin, symbol, name, created_at, updated_at)
		VALUES ('US0378331005', 'AAPL.US', 'Apple Inc.', 1704067200, 1704067200)
	`)
	require.NoError(t, err)

	return db
}

func TestGetClientSymbol_Success(t *testing.T) {
	db := setupTestDBForClientSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewClientSymbolRepository(db, log)

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO client_symbols (isin, client_name, client_symbol)
		VALUES ('US0378331005', 'tradernet', 'AAPL.US')
	`)
	require.NoError(t, err)

	// Execute
	symbol, err := repo.GetClientSymbol("US0378331005", "tradernet")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "AAPL.US", symbol)
}

func TestGetClientSymbol_NotFound(t *testing.T) {
	db := setupTestDBForClientSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewClientSymbolRepository(db, log)

	// Execute
	symbol, err := repo.GetClientSymbol("US0378331005", "tradernet")

	// Assert
	require.Error(t, err)
	assert.Empty(t, symbol)
	assert.Contains(t, err.Error(), "not found")
}

func TestSetClientSymbol_Create(t *testing.T) {
	db := setupTestDBForClientSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewClientSymbolRepository(db, log)

	// Execute
	err := repo.SetClientSymbol("US0378331005", "tradernet", "AAPL.US")
	require.NoError(t, err)

	// Verify
	symbol, err := repo.GetClientSymbol("US0378331005", "tradernet")
	require.NoError(t, err)
	assert.Equal(t, "AAPL.US", symbol)
}

func TestSetClientSymbol_Update(t *testing.T) {
	db := setupTestDBForClientSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewClientSymbolRepository(db, log)

	// Insert initial mapping
	_, err := db.Exec(`
		INSERT INTO client_symbols (isin, client_name, client_symbol)
		VALUES ('US0378331005', 'tradernet', 'AAPL.US')
	`)
	require.NoError(t, err)

	// Execute - update existing
	err = repo.SetClientSymbol("US0378331005", "tradernet", "AAPL")
	require.NoError(t, err)

	// Verify update
	symbol, err := repo.GetClientSymbol("US0378331005", "tradernet")
	require.NoError(t, err)
	assert.Equal(t, "AAPL", symbol)
}

func TestGetAllClientSymbols(t *testing.T) {
	db := setupTestDBForClientSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewClientSymbolRepository(db, log)

	// Insert multiple client symbols for same ISIN
	_, err := db.Exec(`
		INSERT INTO client_symbols (isin, client_name, client_symbol)
		VALUES
			('US0378331005', 'tradernet', 'AAPL.US'),
			('US0378331005', 'ibkr', 'AAPL')
	`)
	require.NoError(t, err)

	// Execute
	symbols, err := repo.GetAllClientSymbols("US0378331005")

	// Assert
	require.NoError(t, err)
	assert.Len(t, symbols, 2)
	assert.Equal(t, "AAPL.US", symbols["tradernet"])
	assert.Equal(t, "AAPL", symbols["ibkr"])
}

func TestGetAllClientSymbols_Empty(t *testing.T) {
	db := setupTestDBForClientSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewClientSymbolRepository(db, log)

	// Execute
	symbols, err := repo.GetAllClientSymbols("US0378331005")

	// Assert
	require.NoError(t, err)
	assert.Empty(t, symbols)
}

func TestGetISINByClientSymbol(t *testing.T) {
	db := setupTestDBForClientSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewClientSymbolRepository(db, log)

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO client_symbols (isin, client_name, client_symbol)
		VALUES ('US0378331005', 'tradernet', 'AAPL.US')
	`)
	require.NoError(t, err)

	// Execute
	isin, err := repo.GetISINByClientSymbol("tradernet", "AAPL.US")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "US0378331005", isin)
}

func TestGetISINByClientSymbol_NotFound(t *testing.T) {
	db := setupTestDBForClientSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewClientSymbolRepository(db, log)

	// Execute
	isin, err := repo.GetISINByClientSymbol("tradernet", "AAPL.US")

	// Assert
	require.Error(t, err)
	assert.Empty(t, isin)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteClientSymbol(t *testing.T) {
	db := setupTestDBForClientSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewClientSymbolRepository(db, log)

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO client_symbols (isin, client_name, client_symbol)
		VALUES ('US0378331005', 'tradernet', 'AAPL.US')
	`)
	require.NoError(t, err)

	// Execute
	err = repo.DeleteClientSymbol("US0378331005", "tradernet")
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetClientSymbol("US0378331005", "tradernet")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteClientSymbol_NotFound(t *testing.T) {
	db := setupTestDBForClientSymbols(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewClientSymbolRepository(db, log)

	// Execute - delete non-existent mapping
	err := repo.DeleteClientSymbol("US0378331005", "tradernet")

	// Assert - should not error (idempotent operation)
	require.NoError(t, err)
}
