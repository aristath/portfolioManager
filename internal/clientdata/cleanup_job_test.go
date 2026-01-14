package clientdata

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCleanupJob(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	job := NewCleanupJob(repo, zerolog.Nop())

	assert.NotNil(t, job)
}

func TestCleanupJobName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	job := NewCleanupJob(repo, zerolog.Nop())

	assert.Equal(t, "client_data_cleanup", job.Name())
}

func TestCleanupJobRun(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	job := NewCleanupJob(repo, zerolog.Nop())

	now := time.Now()
	expiredAt := now.Add(-time.Hour).Unix()
	freshAt := now.Add(time.Hour).Unix()

	// Insert expired and fresh entries across multiple tables
	insertExpiredAndFresh(t, db, "exchangerate", "pair", expiredAt, freshAt)
	insertExpiredAndFresh(t, db, "current_prices", "isin", expiredAt, freshAt)
	insertExpiredAndFresh(t, db, "symbol_to_isin", "symbol", expiredAt, freshAt)

	// Count before cleanup
	var countBefore int
	db.QueryRow("SELECT (SELECT COUNT(*) FROM exchangerate) + (SELECT COUNT(*) FROM current_prices) + (SELECT COUNT(*) FROM symbol_to_isin)").Scan(&countBefore)
	assert.Equal(t, 6, countBefore) // 2 per table (1 expired + 1 fresh)

	// Run cleanup
	err := job.Run()
	require.NoError(t, err)

	// Count after cleanup - should only have fresh entries
	var countAfter int
	db.QueryRow("SELECT (SELECT COUNT(*) FROM exchangerate) + (SELECT COUNT(*) FROM current_prices) + (SELECT COUNT(*) FROM symbol_to_isin)").Scan(&countAfter)
	assert.Equal(t, 3, countAfter) // 1 fresh per table
}

func TestCleanupJobRunEmptyTables(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	job := NewCleanupJob(repo, zerolog.Nop())

	// Run cleanup on empty tables - should not error
	err := job.Run()
	require.NoError(t, err)
}

func TestCleanupJobRunAllExpired(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	job := NewCleanupJob(repo, zerolog.Nop())

	expiredAt := time.Now().Add(-time.Hour).Unix()

	// Insert only expired entries
	_, err := db.Exec("INSERT INTO exchangerate (pair, data, expires_at) VALUES (?, ?, ?)", "EUR:USD", `{}`, expiredAt)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO exchangerate (pair, data, expires_at) VALUES (?, ?, ?)", "GBP:USD", `{}`, expiredAt)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO current_prices (isin, data, expires_at) VALUES (?, ?, ?)", "US003", `{}`, expiredAt)
	require.NoError(t, err)

	// Run cleanup
	err = job.Run()
	require.NoError(t, err)

	// Verify all entries removed
	var count int
	db.QueryRow("SELECT COUNT(*) FROM exchangerate").Scan(&count)
	assert.Equal(t, 0, count)
	db.QueryRow("SELECT COUNT(*) FROM current_prices").Scan(&count)
	assert.Equal(t, 0, count)
}

func TestCleanupJobRunAllFresh(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	job := NewCleanupJob(repo, zerolog.Nop())

	freshAt := time.Now().Add(time.Hour).Unix()

	// Insert only fresh entries
	_, err := db.Exec("INSERT INTO exchangerate (pair, data, expires_at) VALUES (?, ?, ?)", "EUR:USD", `{}`, freshAt)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO exchangerate (pair, data, expires_at) VALUES (?, ?, ?)", "GBP:USD", `{}`, freshAt)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO current_prices (isin, data, expires_at) VALUES (?, ?, ?)", "US003", `{}`, freshAt)
	require.NoError(t, err)

	// Run cleanup
	err = job.Run()
	require.NoError(t, err)

	// Verify no entries removed
	var count int
	db.QueryRow("SELECT COUNT(*) FROM exchangerate").Scan(&count)
	assert.Equal(t, 2, count)
	db.QueryRow("SELECT COUNT(*) FROM current_prices").Scan(&count)
	assert.Equal(t, 1, count)
}

func TestCleanupJobSetJob(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	job := NewCleanupJob(repo, zerolog.Nop())

	// SetJob should not panic
	job.SetJob(nil)
	job.SetJob(struct{}{})
}

// Helper function to insert one expired and one fresh entry per table
func insertExpiredAndFresh(t *testing.T, db *sql.DB, table, keyCol string, expiredAt, freshAt int64) {
	t.Helper()

	var key1, key2 string
	switch keyCol {
	case "pair":
		key1 = "EUR:USD_" + table
		key2 = "GBP:USD_" + table
	case "symbol":
		key1 = "AAPL.US_" + table
		key2 = "MSFT.US_" + table
	default:
		key1 = "US_EXPIRED_" + table
		key2 = "US_FRESH_" + table
	}

	_, err := db.Exec(
		"INSERT INTO "+table+" ("+keyCol+", data, expires_at) VALUES (?, ?, ?)",
		key1, `{"status":"expired"}`, expiredAt,
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO "+table+" ("+keyCol+", data, expires_at) VALUES (?, ?, ?)",
		key2, `{"status":"fresh"}`, freshAt,
	)
	require.NoError(t, err)
}
