package database

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDBForValidation(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create securities table with current schema (symbol as PRIMARY KEY)
	_, err = db.Exec(`
		CREATE TABLE securities (
			symbol TEXT PRIMARY KEY,
			yahoo_symbol TEXT,
			isin TEXT,
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
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	// Create scores table
	_, err = db.Exec(`
		CREATE TABLE scores (
			symbol TEXT PRIMARY KEY,
			total_score REAL NOT NULL,
			last_updated TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	// Create positions table
	_, err = db.Exec(`
		CREATE TABLE positions (
			symbol TEXT PRIMARY KEY,
			quantity REAL NOT NULL,
			isin TEXT
		)
	`)
	require.NoError(t, err)

	return db
}

func TestValidateAllSecuritiesHaveISIN_AllHaveISIN(t *testing.T) {
	db := setupTestDBForValidation(t)
	defer db.Close()

	// Insert securities with ISINs
	_, err := db.Exec(`
		INSERT INTO securities (symbol, isin, name, created_at, updated_at) VALUES
		('AAPL.US', 'US0378331005', 'Apple Inc.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z'),
		('MSFT.US', 'US5949181045', 'Microsoft Corp.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z')
	`)
	require.NoError(t, err)

	validator := NewISINValidator(db)
	errors, err := validator.ValidateAllSecuritiesHaveISIN()
	require.NoError(t, err)
	assert.Empty(t, errors, "Should have no errors when all securities have ISIN")
}

func TestValidateAllSecuritiesHaveISIN_MissingISIN(t *testing.T) {
	db := setupTestDBForValidation(t)
	defer db.Close()

	// Insert securities - one with ISIN, one without
	_, err := db.Exec(`
		INSERT INTO securities (symbol, isin, name, created_at, updated_at) VALUES
		('AAPL.US', 'US0378331005', 'Apple Inc.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z'),
		('MSFT.US', NULL, 'Microsoft Corp.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z'),
		('GOOGL.US', '', 'Google Inc.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z')
	`)
	require.NoError(t, err)

	validator := NewISINValidator(db)
	errors, err := validator.ValidateAllSecuritiesHaveISIN()
	require.NoError(t, err)
	assert.Len(t, errors, 2, "Should find 2 securities without ISIN")
	assert.Contains(t, errors, "MSFT.US")
	assert.Contains(t, errors, "GOOGL.US")
}

func TestValidateNoDuplicateISINs_NoDuplicates(t *testing.T) {
	db := setupTestDBForValidation(t)
	defer db.Close()

	// Insert securities with unique ISINs
	_, err := db.Exec(`
		INSERT INTO securities (symbol, isin, name, created_at, updated_at) VALUES
		('AAPL.US', 'US0378331005', 'Apple Inc.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z'),
		('MSFT.US', 'US5949181045', 'Microsoft Corp.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z')
	`)
	require.NoError(t, err)

	validator := NewISINValidator(db)
	duplicates, err := validator.ValidateNoDuplicateISINs()
	require.NoError(t, err)
	assert.Empty(t, duplicates, "Should have no duplicates when all ISINs are unique")
}

func TestValidateNoDuplicateISINs_HasDuplicates(t *testing.T) {
	db := setupTestDBForValidation(t)
	defer db.Close()

	// Insert securities with duplicate ISIN
	_, err := db.Exec(`
		INSERT INTO securities (symbol, isin, name, created_at, updated_at) VALUES
		('AAPL.US', 'US0378331005', 'Apple Inc.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z'),
		('AAPL.NASDAQ', 'US0378331005', 'Apple Inc. (NASDAQ)', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z')
	`)
	require.NoError(t, err)

	validator := NewISINValidator(db)
	duplicates, err := validator.ValidateNoDuplicateISINs()
	require.NoError(t, err)
	assert.Len(t, duplicates, 1, "Should find 1 duplicate ISIN")
	assert.Contains(t, duplicates, "US0378331005")
}

func TestValidateNoDuplicateISINs_IgnoresNullAndEmpty(t *testing.T) {
	db := setupTestDBForValidation(t)
	defer db.Close()

	// Insert securities - some with ISIN, some without (NULL or empty)
	_, err := db.Exec(`
		INSERT INTO securities (symbol, isin, name, created_at, updated_at) VALUES
		('AAPL.US', 'US0378331005', 'Apple Inc.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z'),
		('MSFT.US', NULL, 'Microsoft Corp.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z'),
		('GOOGL.US', '', 'Google Inc.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z')
	`)
	require.NoError(t, err)

	validator := NewISINValidator(db)
	duplicates, err := validator.ValidateNoDuplicateISINs()
	require.NoError(t, err)
	assert.Empty(t, duplicates, "Should ignore NULL and empty ISINs when checking duplicates")
}

func TestValidateForeignKeys_AllValid(t *testing.T) {
	db := setupTestDBForValidation(t)
	defer db.Close()

	// Insert securities with ISINs
	_, err := db.Exec(`
		INSERT INTO securities (symbol, isin, name, created_at, updated_at) VALUES
		('AAPL.US', 'US0378331005', 'Apple Inc.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z'),
		('MSFT.US', 'US5949181045', 'Microsoft Corp.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z')
	`)
	require.NoError(t, err)

	// Insert scores referencing securities by symbol
	_, err = db.Exec(`
		INSERT INTO scores (symbol, total_score, last_updated) VALUES
		('AAPL.US', 85.5, '2024-01-01T00:00:00Z'),
		('MSFT.US', 90.0, '2024-01-01T00:00:00Z')
	`)
	require.NoError(t, err)

	// Insert positions referencing securities by symbol
	_, err = db.Exec(`
		INSERT INTO positions (symbol, quantity, isin) VALUES
		('AAPL.US', 10.0, 'US0378331005'),
		('MSFT.US', 5.0, 'US5949181045')
	`)
	require.NoError(t, err)

	validator := NewISINValidator(db)
	errors, err := validator.ValidateForeignKeys()
	require.NoError(t, err)
	assert.Empty(t, errors, "Should have no errors when all foreign keys are valid")
}

func TestValidateForeignKeys_OrphanedReferences(t *testing.T) {
	db := setupTestDBForValidation(t)
	defer db.Close()

	// Insert only one security
	_, err := db.Exec(`
		INSERT INTO securities (symbol, isin, name, created_at, updated_at) VALUES
		('AAPL.US', 'US0378331005', 'Apple Inc.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z')
	`)
	require.NoError(t, err)

	// Insert scores/positions referencing non-existent securities
	_, err = db.Exec(`
		INSERT INTO scores (symbol, total_score, last_updated) VALUES
		('AAPL.US', 85.5, '2024-01-01T00:00:00Z'),
		('ORPHAN.US', 50.0, '2024-01-01T00:00:00Z')
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO positions (symbol, quantity, isin) VALUES
		('AAPL.US', 10.0, 'US0378331005'),
		('ORPHAN.US', 5.0, NULL)
	`)
	require.NoError(t, err)

	validator := NewISINValidator(db)
	errors, err := validator.ValidateForeignKeys()
	require.NoError(t, err)
	assert.NotEmpty(t, errors, "Should find orphaned references")
}

func TestValidateAll_Comprehensive(t *testing.T) {
	db := setupTestDBForValidation(t)
	defer db.Close()

	// Insert valid data
	_, err := db.Exec(`
		INSERT INTO securities (symbol, isin, name, created_at, updated_at) VALUES
		('AAPL.US', 'US0378331005', 'Apple Inc.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z'),
		('MSFT.US', 'US5949181045', 'Microsoft Corp.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z')
	`)
	require.NoError(t, err)

	validator := NewISINValidator(db)
	result, err := validator.ValidateAll()
	require.NoError(t, err)
	assert.True(t, result.IsValid, "Should be valid when all checks pass")
	assert.Empty(t, result.MissingISINs)
	assert.Empty(t, result.DuplicateISINs)
	assert.Empty(t, result.OrphanedReferences)
}

func TestValidateAll_FailsOnMissingISIN(t *testing.T) {
	db := setupTestDBForValidation(t)
	defer db.Close()

	// Insert security without ISIN
	_, err := db.Exec(`
		INSERT INTO securities (symbol, isin, name, created_at, updated_at) VALUES
		('AAPL.US', NULL, 'Apple Inc.', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z')
	`)
	require.NoError(t, err)

	validator := NewISINValidator(db)
	result, err := validator.ValidateAll()
	require.NoError(t, err)
	assert.False(t, result.IsValid, "Should be invalid when securities missing ISIN")
	assert.NotEmpty(t, result.MissingISINs)
}
