package calculations

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Create tables matching calculations_schema.sql
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS technical_cache (
			isin TEXT NOT NULL,
			metric TEXT NOT NULL,
			period INTEGER NOT NULL DEFAULT 0,
			value REAL NOT NULL,
			expires_at INTEGER NOT NULL,
			PRIMARY KEY (isin, metric, period)
		);

		CREATE TABLE IF NOT EXISTS optimizer_cache (
			cache_type TEXT NOT NULL,
			isin_hash TEXT NOT NULL,
			value BLOB NOT NULL,
			expires_at INTEGER NOT NULL,
			PRIMARY KEY (cache_type, isin_hash)
		);

		CREATE INDEX IF NOT EXISTS idx_tech_expires ON technical_cache(expires_at);
		CREATE INDEX IF NOT EXISTS idx_opt_expires ON optimizer_cache(expires_at);
	`)
	require.NoError(t, err)

	return db
}

func setupTestCache(t *testing.T) *Cache {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	return NewCache(db)
}

func TestTechnicalCache_SetAndGet(t *testing.T) {
	cache := setupTestCache(t)

	err := cache.SetTechnical("ISIN1", "ema", 200, 150.5, time.Hour)
	require.NoError(t, err)

	value, ok := cache.GetTechnical("ISIN1", "ema", 200)
	assert.True(t, ok)
	assert.Equal(t, 150.5, value)
}

func TestTechnicalCache_GetExpired(t *testing.T) {
	cache := setupTestCache(t)

	err := cache.SetTechnical("ISIN1", "ema", 200, 150.5, -time.Hour)
	require.NoError(t, err)

	_, ok := cache.GetTechnical("ISIN1", "ema", 200)
	assert.False(t, ok, "Expired entries should not be returned")
}

func TestTechnicalCache_GetMissing(t *testing.T) {
	cache := setupTestCache(t)

	_, ok := cache.GetTechnical("NONEXISTENT", "ema", 200)
	assert.False(t, ok, "Non-existent entries should return false")
}

func TestTechnicalCache_Upsert(t *testing.T) {
	cache := setupTestCache(t)

	err := cache.SetTechnical("ISIN1", "ema", 200, 100.0, time.Hour)
	require.NoError(t, err)

	err = cache.SetTechnical("ISIN1", "ema", 200, 200.0, time.Hour)
	require.NoError(t, err)

	value, ok := cache.GetTechnical("ISIN1", "ema", 200)
	assert.True(t, ok)
	assert.Equal(t, 200.0, value, "Upsert should update the value")
}

func TestTechnicalCache_DifferentPeriods(t *testing.T) {
	cache := setupTestCache(t)

	err := cache.SetTechnical("ISIN1", "ema", 50, 100.0, time.Hour)
	require.NoError(t, err)

	err = cache.SetTechnical("ISIN1", "ema", 200, 150.0, time.Hour)
	require.NoError(t, err)

	value50, ok := cache.GetTechnical("ISIN1", "ema", 50)
	assert.True(t, ok)
	assert.Equal(t, 100.0, value50)

	value200, ok := cache.GetTechnical("ISIN1", "ema", 200)
	assert.True(t, ok)
	assert.Equal(t, 150.0, value200)
}

func TestTechnicalCache_DifferentMetrics(t *testing.T) {
	cache := setupTestCache(t)

	err := cache.SetTechnical("ISIN1", "ema", 200, 150.0, time.Hour)
	require.NoError(t, err)

	err = cache.SetTechnical("ISIN1", "rsi", 14, 55.0, time.Hour)
	require.NoError(t, err)

	err = cache.SetTechnical("ISIN1", "sharpe", 0, 1.5, time.Hour)
	require.NoError(t, err)

	ema, ok := cache.GetTechnical("ISIN1", "ema", 200)
	assert.True(t, ok)
	assert.Equal(t, 150.0, ema)

	rsi, ok := cache.GetTechnical("ISIN1", "rsi", 14)
	assert.True(t, ok)
	assert.Equal(t, 55.0, rsi)

	sharpe, ok := cache.GetTechnical("ISIN1", "sharpe", 0)
	assert.True(t, ok)
	assert.Equal(t, 1.5, sharpe)
}

func TestOptimizerCache_SetAndGet(t *testing.T) {
	cache := setupTestCache(t)

	data := []byte(`{"matrix":[[1,2],[3,4]]}`)
	err := cache.SetOptimizer("covariance", "abc123", data, time.Hour)
	require.NoError(t, err)

	value, ok := cache.GetOptimizer("covariance", "abc123")
	assert.True(t, ok)
	assert.Equal(t, data, value)
}

func TestOptimizerCache_GetExpired(t *testing.T) {
	cache := setupTestCache(t)

	err := cache.SetOptimizer("covariance", "abc123", []byte("data"), -time.Hour)
	require.NoError(t, err)

	_, ok := cache.GetOptimizer("covariance", "abc123")
	assert.False(t, ok, "Expired entries should not be returned")
}

func TestOptimizerCache_GetMissing(t *testing.T) {
	cache := setupTestCache(t)

	_, ok := cache.GetOptimizer("covariance", "nonexistent")
	assert.False(t, ok)
}

func TestOptimizerCache_Upsert(t *testing.T) {
	cache := setupTestCache(t)

	err := cache.SetOptimizer("covariance", "hash1", []byte("data1"), time.Hour)
	require.NoError(t, err)

	err = cache.SetOptimizer("covariance", "hash1", []byte("data2"), time.Hour)
	require.NoError(t, err)

	value, ok := cache.GetOptimizer("covariance", "hash1")
	assert.True(t, ok)
	assert.Equal(t, []byte("data2"), value)
}

func TestOptimizerCache_DifferentTypes(t *testing.T) {
	cache := setupTestCache(t)

	err := cache.SetOptimizer("covariance", "hash1", []byte("cov_data"), time.Hour)
	require.NoError(t, err)

	err = cache.SetOptimizer("hrp", "hash1", []byte("hrp_data"), time.Hour)
	require.NoError(t, err)

	cov, ok := cache.GetOptimizer("covariance", "hash1")
	assert.True(t, ok)
	assert.Equal(t, []byte("cov_data"), cov)

	hrp, ok := cache.GetOptimizer("hrp", "hash1")
	assert.True(t, ok)
	assert.Equal(t, []byte("hrp_data"), hrp)
}

func TestCache_Cleanup(t *testing.T) {
	cache := setupTestCache(t)

	// Add expired technical entries
	err := cache.SetTechnical("ISIN1", "ema", 200, 1.0, -time.Hour)
	require.NoError(t, err)
	err = cache.SetTechnical("ISIN2", "ema", 200, 2.0, -time.Hour)
	require.NoError(t, err)

	// Add fresh technical entry
	err = cache.SetTechnical("ISIN3", "ema", 200, 3.0, time.Hour)
	require.NoError(t, err)

	// Add expired optimizer entry
	err = cache.SetOptimizer("cov", "hash1", []byte("x"), -time.Hour)
	require.NoError(t, err)

	// Add fresh optimizer entry
	err = cache.SetOptimizer("cov", "hash2", []byte("y"), time.Hour)
	require.NoError(t, err)

	count, err := cache.Cleanup()
	require.NoError(t, err)
	assert.Equal(t, int64(3), count, "Should remove 3 expired entries")

	// Fresh entries should still exist
	_, ok := cache.GetTechnical("ISIN3", "ema", 200)
	assert.True(t, ok, "Fresh technical entry should still exist")

	_, ok = cache.GetOptimizer("cov", "hash2")
	assert.True(t, ok, "Fresh optimizer entry should still exist")

	// Expired entries should be gone
	_, ok = cache.GetTechnical("ISIN1", "ema", 200)
	assert.False(t, ok, "Expired entry should be cleaned up")
}

func TestCache_GetISINsNeedingCalculation(t *testing.T) {
	cache := setupTestCache(t)

	isins := []string{"ISIN1", "ISIN2", "ISIN3"}

	// ISIN1 has fresh cache
	err := cache.SetTechnical("ISIN1", "ema", 200, 1.0, time.Hour)
	require.NoError(t, err)

	// ISIN2 has expired cache
	err = cache.SetTechnical("ISIN2", "ema", 200, 2.0, -time.Hour)
	require.NoError(t, err)

	// ISIN3 has no cache (never calculated)

	needsCalc := cache.GetISINsNeedingCalculation(isins, "ema", 200)

	assert.Contains(t, needsCalc, "ISIN2", "Expired ISIN should need calculation")
	assert.Contains(t, needsCalc, "ISIN3", "Missing ISIN should need calculation")
	assert.NotContains(t, needsCalc, "ISIN1", "Fresh ISIN should not need calculation")
	assert.Len(t, needsCalc, 2)
}

func TestCache_GetISINsNeedingCalculation_DifferentMetrics(t *testing.T) {
	cache := setupTestCache(t)

	isins := []string{"ISIN1"}

	// ISIN1 has fresh EMA-200 but no RSI
	err := cache.SetTechnical("ISIN1", "ema", 200, 1.0, time.Hour)
	require.NoError(t, err)

	needsEMA := cache.GetISINsNeedingCalculation(isins, "ema", 200)
	assert.Empty(t, needsEMA, "Should not need EMA calculation")

	needsRSI := cache.GetISINsNeedingCalculation(isins, "rsi", 14)
	assert.Contains(t, needsRSI, "ISIN1", "Should need RSI calculation")
}

func TestCache_DeleteTechnicalForISIN(t *testing.T) {
	cache := setupTestCache(t)

	// Add multiple metrics for ISIN1
	err := cache.SetTechnical("ISIN1", "ema", 200, 1.0, time.Hour)
	require.NoError(t, err)
	err = cache.SetTechnical("ISIN1", "ema", 50, 2.0, time.Hour)
	require.NoError(t, err)
	err = cache.SetTechnical("ISIN1", "rsi", 14, 55.0, time.Hour)
	require.NoError(t, err)

	// Add metrics for ISIN2
	err = cache.SetTechnical("ISIN2", "ema", 200, 3.0, time.Hour)
	require.NoError(t, err)

	// Delete all for ISIN1
	err = cache.DeleteTechnicalForISIN("ISIN1")
	require.NoError(t, err)

	// ISIN1 entries should be gone
	_, ok := cache.GetTechnical("ISIN1", "ema", 200)
	assert.False(t, ok)
	_, ok = cache.GetTechnical("ISIN1", "ema", 50)
	assert.False(t, ok)
	_, ok = cache.GetTechnical("ISIN1", "rsi", 14)
	assert.False(t, ok)

	// ISIN2 should still exist
	_, ok = cache.GetTechnical("ISIN2", "ema", 200)
	assert.True(t, ok)
}

func TestCache_DeleteOptimizerByType(t *testing.T) {
	cache := setupTestCache(t)

	// Add multiple hashes for covariance
	err := cache.SetOptimizer("covariance", "hash1", []byte("data1"), time.Hour)
	require.NoError(t, err)
	err = cache.SetOptimizer("covariance", "hash2", []byte("data2"), time.Hour)
	require.NoError(t, err)

	// Add HRP entry
	err = cache.SetOptimizer("hrp", "hash1", []byte("hrp"), time.Hour)
	require.NoError(t, err)

	// Delete all covariance entries
	err = cache.DeleteOptimizerByType("covariance")
	require.NoError(t, err)

	// Covariance entries should be gone
	_, ok := cache.GetOptimizer("covariance", "hash1")
	assert.False(t, ok)
	_, ok = cache.GetOptimizer("covariance", "hash2")
	assert.False(t, ok)

	// HRP should still exist
	_, ok = cache.GetOptimizer("hrp", "hash1")
	assert.True(t, ok)
}

func TestCache_ExpiresAtPrecision(t *testing.T) {
	cache := setupTestCache(t)

	// Set entry to expire in exactly 1 second
	err := cache.SetTechnical("ISIN1", "ema", 200, 100.0, time.Second)
	require.NoError(t, err)

	// Should still be valid
	_, ok := cache.GetTechnical("ISIN1", "ema", 200)
	assert.True(t, ok, "Entry should be valid before expiration")

	// Wait for expiration
	time.Sleep(1100 * time.Millisecond)

	// Should now be expired
	_, ok = cache.GetTechnical("ISIN1", "ema", 200)
	assert.False(t, ok, "Entry should be expired after TTL")
}

func TestCache_MultipleISINs(t *testing.T) {
	cache := setupTestCache(t)

	// Add different values for different ISINs
	for i := 1; i <= 10; i++ {
		isin := "ISIN" + string(rune('0'+i))
		err := cache.SetTechnical(isin, "ema", 200, float64(i*100), time.Hour)
		require.NoError(t, err)
	}

	// Verify each ISIN has correct value
	for i := 1; i <= 10; i++ {
		isin := "ISIN" + string(rune('0'+i))
		value, ok := cache.GetTechnical(isin, "ema", 200)
		assert.True(t, ok)
		assert.Equal(t, float64(i*100), value)
	}
}
