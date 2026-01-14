package market_regime

import (
	"database/sql"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// setupIndexTestDB creates a test database with market_indices table
func setupIndexTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE market_indices (
			symbol TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			market_code TEXT NOT NULL,
			region TEXT NOT NULL,
			index_type TEXT NOT NULL DEFAULT 'PRICE',
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE INDEX idx_market_indices_region ON market_indices(region)`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE INDEX idx_market_indices_enabled ON market_indices(enabled)`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE INDEX idx_market_indices_type ON market_indices(index_type)`)
	require.NoError(t, err)

	return db
}

func TestIndexRepository_Upsert(t *testing.T) {
	db := setupIndexTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewIndexRepository(db, log)

	index := MarketIndex{
		Symbol:     "SP500.IDX",
		Name:       "S&P 500",
		MarketCode: "FIX",
		Region:     RegionUS,
		IndexType:  IndexTypePrice,
		Enabled:    true,
	}

	err := repo.Upsert(index)
	require.NoError(t, err)

	// Verify it was inserted
	retrieved, err := repo.GetBySymbol("SP500.IDX")
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, "SP500.IDX", retrieved.Symbol)
	assert.Equal(t, "S&P 500", retrieved.Name)
	assert.Equal(t, RegionUS, retrieved.Region)
	assert.Equal(t, IndexTypePrice, retrieved.IndexType)
	assert.True(t, retrieved.Enabled)
}

func TestIndexRepository_Upsert_Update(t *testing.T) {
	db := setupIndexTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewIndexRepository(db, log)

	// Insert initial
	index := MarketIndex{
		Symbol:     "SP500.IDX",
		Name:       "S&P 500",
		MarketCode: "FIX",
		Region:     RegionUS,
		IndexType:  IndexTypePrice,
		Enabled:    true,
	}
	err := repo.Upsert(index)
	require.NoError(t, err)

	// Update with new name
	index.Name = "Standard & Poor's 500"
	index.Enabled = false
	err = repo.Upsert(index)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetBySymbol("SP500.IDX")
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, "Standard & Poor's 500", retrieved.Name)
	assert.False(t, retrieved.Enabled)
}

func TestIndexRepository_GetEnabledByRegion(t *testing.T) {
	db := setupIndexTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewIndexRepository(db, log)

	// Seed test data
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO market_indices (symbol, name, market_code, region, index_type, enabled, created_at, updated_at)
		VALUES
			('SP500.IDX', 'S&P 500', 'FIX', 'US', 'PRICE', 1, ?, ?),
			('VIX.IDX', 'VIX', 'FIX', 'US', 'VOLATILITY', 1, ?, ?),
			('NASDAQ.IDX', 'NASDAQ', 'FIX', 'US', 'PRICE', 0, ?, ?),
			('DAX.IDX', 'DAX', 'EU', 'EU', 'PRICE', 1, ?, ?)
	`, now, now, now, now, now, now, now, now)
	require.NoError(t, err)

	// Get enabled US indices
	usIndices, err := repo.GetEnabledByRegion(RegionUS)
	require.NoError(t, err)
	assert.Len(t, usIndices, 2) // SP500 and VIX (both enabled)

	// Get enabled EU indices
	euIndices, err := repo.GetEnabledByRegion(RegionEU)
	require.NoError(t, err)
	assert.Len(t, euIndices, 1) // DAX
}

func TestIndexRepository_GetEnabledPriceByRegion(t *testing.T) {
	db := setupIndexTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewIndexRepository(db, log)

	// Seed test data
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO market_indices (symbol, name, market_code, region, index_type, enabled, created_at, updated_at)
		VALUES
			('SP500.IDX', 'S&P 500', 'FIX', 'US', 'PRICE', 1, ?, ?),
			('VIX.IDX', 'VIX', 'FIX', 'US', 'VOLATILITY', 1, ?, ?),
			('NASDAQ.IDX', 'NASDAQ', 'FIX', 'US', 'PRICE', 1, ?, ?),
			('DAX.IDX', 'DAX', 'EU', 'EU', 'PRICE', 1, ?, ?)
	`, now, now, now, now, now, now, now, now)
	require.NoError(t, err)

	// Get enabled PRICE US indices (should exclude VIX)
	usIndices, err := repo.GetEnabledPriceByRegion(RegionUS)
	require.NoError(t, err)
	assert.Len(t, usIndices, 2) // SP500 and NASDAQ (not VIX)

	for _, idx := range usIndices {
		assert.Equal(t, IndexTypePrice, idx.IndexType)
		assert.NotEqual(t, "VIX.IDX", idx.Symbol)
	}
}

func TestIndexRepository_GetAllEnabled(t *testing.T) {
	db := setupIndexTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewIndexRepository(db, log)

	// Seed test data
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO market_indices (symbol, name, market_code, region, index_type, enabled, created_at, updated_at)
		VALUES
			('SP500.IDX', 'S&P 500', 'FIX', 'US', 'PRICE', 1, ?, ?),
			('VIX.IDX', 'VIX', 'FIX', 'US', 'VOLATILITY', 1, ?, ?),
			('DISABLED.IDX', 'Disabled', 'FIX', 'US', 'PRICE', 0, ?, ?),
			('DAX.IDX', 'DAX', 'EU', 'EU', 'PRICE', 1, ?, ?)
	`, now, now, now, now, now, now, now, now)
	require.NoError(t, err)

	// Get all enabled
	indices, err := repo.GetAllEnabled()
	require.NoError(t, err)
	assert.Len(t, indices, 3) // SP500, VIX, DAX (not DISABLED)
}

func TestIndexRepository_SyncFromKnownIndices(t *testing.T) {
	db := setupIndexTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewIndexRepository(db, log)

	// Sync all known indices
	err := repo.SyncFromKnownIndices()
	require.NoError(t, err)

	// Verify all known indices are in DB
	allIndices, err := repo.GetAllEnabled()
	require.NoError(t, err)

	knownIndices := GetKnownIndices()
	assert.Len(t, allIndices, len(knownIndices))
}

func TestIndexRepository_Delete(t *testing.T) {
	db := setupIndexTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewIndexRepository(db, log)

	// Insert
	index := MarketIndex{
		Symbol:     "TEST.IDX",
		Name:       "Test Index",
		MarketCode: "FIX",
		Region:     RegionUS,
		IndexType:  IndexTypePrice,
		Enabled:    true,
	}
	err := repo.Upsert(index)
	require.NoError(t, err)

	// Verify exists
	retrieved, err := repo.GetBySymbol("TEST.IDX")
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	// Delete
	err = repo.Delete("TEST.IDX")
	require.NoError(t, err)

	// Verify deleted
	retrieved, err = repo.GetBySymbol("TEST.IDX")
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}
