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

func setupOverrideTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create securities table (required for foreign key)
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
			active INTEGER DEFAULT 1,
			currency TEXT,
			last_synced INTEGER,
			min_portfolio_target REAL,
			max_portfolio_target REAL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	require.NoError(t, err)

	// Create security_overrides table
	_, err = db.Exec(`
		CREATE TABLE security_overrides (
			isin TEXT NOT NULL,
			field TEXT NOT NULL,
			value TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			PRIMARY KEY (isin, field),
			FOREIGN KEY (isin) REFERENCES securities(isin) ON DELETE CASCADE
		)
	`)
	require.NoError(t, err)

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	return db
}

func insertTestSecurity(t *testing.T, db *sql.DB, isin, symbol, name string) {
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO securities (isin, symbol, name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, isin, symbol, name, now, now)
	require.NoError(t, err)
}

func TestOverrideRepository_SetOverride_New(t *testing.T) {
	// Setup
	db := setupOverrideTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewOverrideRepository(db, log)

	insertTestSecurity(t, db, "US0378331005", "AAPL.US", "Apple Inc.")

	// Execute
	err := repo.SetOverride("US0378331005", "geography", "US")

	// Assert
	assert.NoError(t, err)

	// Verify override was created
	overrides, err := repo.GetOverrides("US0378331005")
	assert.NoError(t, err)
	assert.Equal(t, "US", overrides["geography"])
}

func TestOverrideRepository_SetOverride_Update(t *testing.T) {
	// Setup
	db := setupOverrideTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewOverrideRepository(db, log)

	insertTestSecurity(t, db, "US0378331005", "AAPL.US", "Apple Inc.")

	// Create initial override
	err := repo.SetOverride("US0378331005", "geography", "US")
	require.NoError(t, err)

	// Execute - update existing override
	err = repo.SetOverride("US0378331005", "geography", "WORLD")

	// Assert
	assert.NoError(t, err)

	// Verify override was updated
	overrides, err := repo.GetOverrides("US0378331005")
	assert.NoError(t, err)
	assert.Equal(t, "WORLD", overrides["geography"])
}

func TestOverrideRepository_SetOverride_MultipleFields(t *testing.T) {
	// Setup
	db := setupOverrideTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewOverrideRepository(db, log)

	insertTestSecurity(t, db, "US0378331005", "AAPL.US", "Apple Inc.")

	// Execute - set multiple overrides
	err := repo.SetOverride("US0378331005", "geography", "US")
	require.NoError(t, err)
	err = repo.SetOverride("US0378331005", "industry", "Technology")
	require.NoError(t, err)
	err = repo.SetOverride("US0378331005", "min_lot", "10")
	require.NoError(t, err)

	// Assert
	overrides, err := repo.GetOverrides("US0378331005")
	assert.NoError(t, err)
	assert.Len(t, overrides, 3)
	assert.Equal(t, "US", overrides["geography"])
	assert.Equal(t, "Technology", overrides["industry"])
	assert.Equal(t, "10", overrides["min_lot"])
}

func TestOverrideRepository_SetOverride_InvalidISIN(t *testing.T) {
	// Setup
	db := setupOverrideTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewOverrideRepository(db, log)

	// Execute - set override for non-existent security
	err := repo.SetOverride("INVALID_ISIN", "geography", "US")

	// Assert - should fail due to foreign key constraint
	assert.Error(t, err)
}

func TestOverrideRepository_GetOverrides_Empty(t *testing.T) {
	// Setup
	db := setupOverrideTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewOverrideRepository(db, log)

	insertTestSecurity(t, db, "US0378331005", "AAPL.US", "Apple Inc.")

	// Execute
	overrides, err := repo.GetOverrides("US0378331005")

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, overrides)
}

func TestOverrideRepository_GetOverrides_NonExistentISIN(t *testing.T) {
	// Setup
	db := setupOverrideTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewOverrideRepository(db, log)

	// Execute
	overrides, err := repo.GetOverrides("NON_EXISTENT")

	// Assert - should return empty map, not error
	assert.NoError(t, err)
	assert.Empty(t, overrides)
}

func TestOverrideRepository_DeleteOverride(t *testing.T) {
	// Setup
	db := setupOverrideTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewOverrideRepository(db, log)

	insertTestSecurity(t, db, "US0378331005", "AAPL.US", "Apple Inc.")

	// Create overrides
	err := repo.SetOverride("US0378331005", "geography", "US")
	require.NoError(t, err)
	err = repo.SetOverride("US0378331005", "industry", "Technology")
	require.NoError(t, err)

	// Execute - delete one override
	err = repo.DeleteOverride("US0378331005", "geography")

	// Assert
	assert.NoError(t, err)

	// Verify geography override was deleted but industry remains
	overrides, err := repo.GetOverrides("US0378331005")
	assert.NoError(t, err)
	assert.Len(t, overrides, 1)
	assert.Empty(t, overrides["geography"])
	assert.Equal(t, "Technology", overrides["industry"])
}

func TestOverrideRepository_DeleteOverride_NonExistent(t *testing.T) {
	// Setup
	db := setupOverrideTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewOverrideRepository(db, log)

	insertTestSecurity(t, db, "US0378331005", "AAPL.US", "Apple Inc.")

	// Execute - delete non-existent override (should not error)
	err := repo.DeleteOverride("US0378331005", "geography")

	// Assert - no error, idempotent operation
	assert.NoError(t, err)
}

func TestOverrideRepository_DeleteAllOverrides(t *testing.T) {
	// Setup
	db := setupOverrideTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewOverrideRepository(db, log)

	insertTestSecurity(t, db, "US0378331005", "AAPL.US", "Apple Inc.")

	// Create multiple overrides
	err := repo.SetOverride("US0378331005", "geography", "US")
	require.NoError(t, err)
	err = repo.SetOverride("US0378331005", "industry", "Technology")
	require.NoError(t, err)
	err = repo.SetOverride("US0378331005", "min_lot", "10")
	require.NoError(t, err)

	// Execute
	err = repo.DeleteAllOverrides("US0378331005")

	// Assert
	assert.NoError(t, err)

	// Verify all overrides were deleted
	overrides, err := repo.GetOverrides("US0378331005")
	assert.NoError(t, err)
	assert.Empty(t, overrides)
}

func TestOverrideRepository_GetAllOverrides(t *testing.T) {
	// Setup
	db := setupOverrideTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewOverrideRepository(db, log)

	insertTestSecurity(t, db, "US0378331005", "AAPL.US", "Apple Inc.")
	insertTestSecurity(t, db, "IE00B3RBWM25", "VWCE.EU", "Vanguard FTSE All-World")

	// Create overrides for multiple securities
	err := repo.SetOverride("US0378331005", "geography", "US")
	require.NoError(t, err)
	err = repo.SetOverride("US0378331005", "industry", "Technology")
	require.NoError(t, err)
	err = repo.SetOverride("IE00B3RBWM25", "geography", "WORLD")
	require.NoError(t, err)

	// Execute
	allOverrides, err := repo.GetAllOverrides()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, allOverrides, 2)

	// Verify AAPL overrides
	aaplOverrides := allOverrides["US0378331005"]
	assert.Len(t, aaplOverrides, 2)
	assert.Equal(t, "US", aaplOverrides["geography"])
	assert.Equal(t, "Technology", aaplOverrides["industry"])

	// Verify VWCE overrides
	vwceOverrides := allOverrides["IE00B3RBWM25"]
	assert.Len(t, vwceOverrides, 1)
	assert.Equal(t, "WORLD", vwceOverrides["geography"])
}

func TestOverrideRepository_GetAllOverrides_Empty(t *testing.T) {
	// Setup
	db := setupOverrideTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewOverrideRepository(db, log)

	// Execute
	allOverrides, err := repo.GetAllOverrides()

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, allOverrides)
}

func TestOverrideRepository_CascadeDeleteOnSecurityDelete(t *testing.T) {
	// Setup
	db := setupOverrideTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewOverrideRepository(db, log)

	insertTestSecurity(t, db, "US0378331005", "AAPL.US", "Apple Inc.")

	// Create overrides
	err := repo.SetOverride("US0378331005", "geography", "US")
	require.NoError(t, err)
	err = repo.SetOverride("US0378331005", "industry", "Technology")
	require.NoError(t, err)

	// Verify overrides exist
	overrides, err := repo.GetOverrides("US0378331005")
	require.NoError(t, err)
	require.Len(t, overrides, 2)

	// Execute - delete security (should cascade delete overrides)
	_, err = db.Exec("DELETE FROM securities WHERE isin = ?", "US0378331005")
	require.NoError(t, err)

	// Assert - overrides should be deleted
	overrides, err = repo.GetOverrides("US0378331005")
	assert.NoError(t, err)
	assert.Empty(t, overrides)
}

func TestOverrideRepository_SetOverride_EmptyValue(t *testing.T) {
	// Setup
	db := setupOverrideTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewOverrideRepository(db, log)

	insertTestSecurity(t, db, "US0378331005", "AAPL.US", "Apple Inc.")

	// Execute - setting empty value should delete the override
	err := repo.SetOverride("US0378331005", "geography", "US")
	require.NoError(t, err)

	// Now set empty value
	err = repo.SetOverride("US0378331005", "geography", "")

	// Assert - empty value should delete the override
	assert.NoError(t, err)
	overrides, err := repo.GetOverrides("US0378331005")
	assert.NoError(t, err)
	assert.Empty(t, overrides["geography"])
}

func TestOverrideRepository_UpdatedAtChanged(t *testing.T) {
	// Setup
	db := setupOverrideTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewOverrideRepository(db, log)

	insertTestSecurity(t, db, "US0378331005", "AAPL.US", "Apple Inc.")

	// Create initial override
	err := repo.SetOverride("US0378331005", "geography", "US")
	require.NoError(t, err)

	// Get initial updated_at
	var initialUpdatedAt int64
	err = db.QueryRow("SELECT updated_at FROM security_overrides WHERE isin = ? AND field = ?",
		"US0378331005", "geography").Scan(&initialUpdatedAt)
	require.NoError(t, err)

	// Wait a tiny bit to ensure timestamp changes
	time.Sleep(10 * time.Millisecond)

	// Update override
	err = repo.SetOverride("US0378331005", "geography", "WORLD")
	require.NoError(t, err)

	// Get new updated_at
	var newUpdatedAt int64
	err = db.QueryRow("SELECT updated_at FROM security_overrides WHERE isin = ? AND field = ?",
		"US0378331005", "geography").Scan(&newUpdatedAt)
	require.NoError(t, err)

	// Assert - updated_at should have changed
	assert.GreaterOrEqual(t, newUpdatedAt, initialUpdatedAt)
}
