package clientdata

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSchema creates all tables needed for testing
// Note: Only Tradernet-related tables remain after removing external data clients
const testSchema = `
CREATE TABLE exchangerate (pair TEXT PRIMARY KEY, data TEXT NOT NULL, expires_at INTEGER NOT NULL);
CREATE TABLE current_prices (isin TEXT PRIMARY KEY, data TEXT NOT NULL, expires_at INTEGER NOT NULL);
CREATE TABLE symbol_to_isin (symbol TEXT PRIMARY KEY, data TEXT NOT NULL, expires_at INTEGER NOT NULL);

CREATE INDEX idx_exchangerate_expires ON exchangerate(expires_at);
CREATE INDEX idx_prices_expires ON current_prices(expires_at);
CREATE INDEX idx_symbol_to_isin_expires ON symbol_to_isin(expires_at);
`

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(testSchema)
	require.NoError(t, err)

	return db
}

func TestNewRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	assert.NotNil(t, repo)
}

func TestStore(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	// Test storing a simple struct
	data := map[string]interface{}{
		"rate":     1.0856,
		"source":   "tradernet",
		"currency": "EUR:USD",
	}

	err := repo.Store("exchangerate", "EUR:USD", data, 7*24*time.Hour)
	require.NoError(t, err)

	// Verify data was stored
	var storedData string
	var expiresAt int64
	err = db.QueryRow("SELECT data, expires_at FROM exchangerate WHERE pair = ?", "EUR:USD").Scan(&storedData, &expiresAt)
	require.NoError(t, err)

	// Verify JSON was stored correctly
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(storedData), &parsed)
	require.NoError(t, err)
	assert.Equal(t, 1.0856, parsed["rate"])
	assert.Equal(t, "tradernet", parsed["source"])

	// Verify expiration is roughly 7 days from now
	expectedExpires := time.Now().Add(7 * 24 * time.Hour).Unix()
	assert.InDelta(t, expectedExpires, expiresAt, 5) // Allow 5 second tolerance
}

func TestStoreUpsert(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	// Store initial data
	data1 := map[string]string{"version": "1"}
	err := repo.Store("exchangerate", "EUR:USD", data1, time.Hour)
	require.NoError(t, err)

	// Store updated data with same key
	data2 := map[string]string{"version": "2"}
	err = repo.Store("exchangerate", "EUR:USD", data2, time.Hour)
	require.NoError(t, err)

	// Verify only one row exists with updated data
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM exchangerate WHERE pair = ?", "EUR:USD").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify data was updated
	result, err := repo.GetIfFresh("exchangerate", "EUR:USD")
	require.NoError(t, err)
	require.NotNil(t, result)

	var parsed map[string]string
	err = json.Unmarshal(result, &parsed)
	require.NoError(t, err)
	assert.Equal(t, "2", parsed["version"])
}

func TestGetIfFresh_Fresh(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	// Store data with 1 hour TTL (fresh)
	data := map[string]string{"status": "fresh"}
	err := repo.Store("current_prices", "US0000000001", data, time.Hour)
	require.NoError(t, err)

	// Should return data
	result, err := repo.GetIfFresh("current_prices", "US0000000001")
	require.NoError(t, err)
	require.NotNil(t, result)

	var parsed map[string]string
	err = json.Unmarshal(result, &parsed)
	require.NoError(t, err)
	assert.Equal(t, "fresh", parsed["status"])
}

func TestGetIfFresh_Expired(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	// Insert expired data directly (expired 1 hour ago)
	expiredAt := time.Now().Add(-time.Hour).Unix()
	_, err := db.Exec(
		"INSERT INTO current_prices (isin, data, expires_at) VALUES (?, ?, ?)",
		"US0000000001",
		`{"status":"expired"}`,
		expiredAt,
	)
	require.NoError(t, err)

	// Should return nil for expired data
	result, err := repo.GetIfFresh("current_prices", "US0000000001")
	require.NoError(t, err)
	assert.Nil(t, result, "Expected nil for expired data")
}

func TestGet_ReturnsStaleData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	// Insert expired data directly (expired 1 hour ago)
	expiredAt := time.Now().Add(-time.Hour).Unix()
	_, err := db.Exec(
		"INSERT INTO current_prices (isin, data, expires_at) VALUES (?, ?, ?)",
		"US0000000001",
		`{"status":"stale_but_useful"}`,
		expiredAt,
	)
	require.NoError(t, err)

	// GetIfFresh should return nil
	result, err := repo.GetIfFresh("current_prices", "US0000000001")
	require.NoError(t, err)
	assert.Nil(t, result, "GetIfFresh should return nil for expired data")

	// Get should return the stale data (useful when API fails)
	result, err = repo.Get("current_prices", "US0000000001")
	require.NoError(t, err)
	require.NotNil(t, result, "Get should return stale data")

	var parsed map[string]string
	err = json.Unmarshal(result, &parsed)
	require.NoError(t, err)
	assert.Equal(t, "stale_but_useful", parsed["status"])
}

func TestGet_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	// Get should return nil for non-existent key
	result, err := repo.Get("current_prices", "NONEXISTENT")
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestGetIfFresh_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	// Should return nil for non-existent key
	result, err := repo.GetIfFresh("current_prices", "NONEXISTENT")
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	// Store data
	data := map[string]string{"to_delete": "true"}
	err := repo.Store("symbol_to_isin", "AAPL.US", data, time.Hour)
	require.NoError(t, err)

	// Verify it exists
	result, err := repo.GetIfFresh("symbol_to_isin", "AAPL.US")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Delete it
	err = repo.Delete("symbol_to_isin", "AAPL.US")
	require.NoError(t, err)

	// Verify it's gone
	result, err = repo.GetIfFresh("symbol_to_isin", "AAPL.US")
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestDeleteNonExistent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	// Deleting non-existent key should not error
	err := repo.Delete("symbol_to_isin", "NONEXISTENT")
	require.NoError(t, err)
}

func TestDeleteExpired(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	now := time.Now()

	// Insert 3 expired entries and 2 fresh entries
	expiredAt := now.Add(-time.Hour).Unix()
	freshAt := now.Add(time.Hour).Unix()

	_, err := db.Exec("INSERT INTO exchangerate (pair, data, expires_at) VALUES (?, ?, ?)", "EUR:USD", `{}`, expiredAt)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO exchangerate (pair, data, expires_at) VALUES (?, ?, ?)", "GBP:USD", `{}`, expiredAt)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO exchangerate (pair, data, expires_at) VALUES (?, ?, ?)", "JPY:USD", `{}`, expiredAt)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO exchangerate (pair, data, expires_at) VALUES (?, ?, ?)", "CHF:USD", `{}`, freshAt)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO exchangerate (pair, data, expires_at) VALUES (?, ?, ?)", "AUD:USD", `{}`, freshAt)
	require.NoError(t, err)

	// Delete expired
	deleted, err := repo.DeleteExpired("exchangerate")
	require.NoError(t, err)
	assert.Equal(t, int64(3), deleted)

	// Verify only 2 remain
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM exchangerate").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestDeleteExpiredEmptyTable(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	// Delete from empty table should return 0
	deleted, err := repo.DeleteExpired("exchangerate")
	require.NoError(t, err)
	assert.Equal(t, int64(0), deleted)
}

func TestDeleteAllExpired(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	now := time.Now()
	expiredAt := now.Add(-time.Hour).Unix()
	freshAt := now.Add(time.Hour).Unix()

	// Insert expired entries in multiple tables
	_, err := db.Exec("INSERT INTO exchangerate (pair, data, expires_at) VALUES (?, ?, ?)", "EUR:USD", `{}`, expiredAt)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO exchangerate (pair, data, expires_at) VALUES (?, ?, ?)", "GBP:USD", `{}`, freshAt)
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO current_prices (isin, data, expires_at) VALUES (?, ?, ?)", "US003", `{}`, expiredAt)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO current_prices (isin, data, expires_at) VALUES (?, ?, ?)", "US004", `{}`, expiredAt)
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO symbol_to_isin (symbol, data, expires_at) VALUES (?, ?, ?)", "AAPL.US", `{}`, freshAt)
	require.NoError(t, err)

	// Delete all expired
	results, err := repo.DeleteAllExpired()
	require.NoError(t, err)

	// Verify counts
	assert.Equal(t, int64(1), results["exchangerate"])
	assert.Equal(t, int64(2), results["current_prices"])
	assert.Equal(t, int64(0), results["symbol_to_isin"])

	// Verify total remaining
	var count int
	db.QueryRow("SELECT COUNT(*) FROM exchangerate").Scan(&count)
	assert.Equal(t, 1, count) // 1 fresh entry

	db.QueryRow("SELECT COUNT(*) FROM current_prices").Scan(&count)
	assert.Equal(t, 0, count) // All expired

	db.QueryRow("SELECT COUNT(*) FROM symbol_to_isin").Scan(&count)
	assert.Equal(t, 1, count) // 1 fresh entry
}

func TestStoreWithDifferentTables(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	// Test storing to different tables
	tables := []struct {
		table string
		key   string
	}{
		{"exchangerate", "EUR:USD"},
		{"current_prices", "US0000000001"},
		{"symbol_to_isin", "AAPL.US"},
	}

	for _, tc := range tables {
		t.Run(tc.table, func(t *testing.T) {
			data := map[string]string{"table": tc.table}
			err := repo.Store(tc.table, tc.key, data, time.Hour)
			require.NoError(t, err)

			result, err := repo.GetIfFresh(tc.table, tc.key)
			require.NoError(t, err)
			require.NotNil(t, result)

			var parsed map[string]string
			json.Unmarshal(result, &parsed)
			assert.Equal(t, tc.table, parsed["table"])
		})
	}
}

func TestStoreComplexJSON(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	// Test with complex nested structure
	data := map[string]interface{}{
		"isin":     "US0378331005",
		"symbol":   "AAPL.US",
		"price":    185.50,
		"currency": "USD",
		"metadata": map[string]interface{}{
			"source":    "tradernet",
			"timestamp": time.Now().Unix(),
		},
	}

	err := repo.Store("current_prices", "US0378331005", data, 7*24*time.Hour)
	require.NoError(t, err)

	result, err := repo.GetIfFresh("current_prices", "US0378331005")
	require.NoError(t, err)
	require.NotNil(t, result)

	var parsed map[string]interface{}
	err = json.Unmarshal(result, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "US0378331005", parsed["isin"])
	assert.Equal(t, "AAPL.US", parsed["symbol"])
	assert.Equal(t, float64(185.50), parsed["price"])

	// Verify nested object
	metadata, ok := parsed["metadata"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "tradernet", metadata["source"])
}

func TestGetKeyColumn(t *testing.T) {
	// Test the key column mapping
	tests := []struct {
		table    string
		expected string
	}{
		{"exchangerate", "pair"},
		{"current_prices", "isin"},
		{"symbol_to_isin", "symbol"},
	}

	for _, tc := range tests {
		t.Run(tc.table, func(t *testing.T) {
			result := getKeyColumn(tc.table)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestInvalidTableName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	// All methods should reject invalid table names
	t.Run("Store", func(t *testing.T) {
		err := repo.Store("invalid_table; DROP TABLE current_prices;--", "key", map[string]string{}, time.Hour)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid table name")
	})

	t.Run("GetIfFresh", func(t *testing.T) {
		_, err := repo.GetIfFresh("users", "key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid table name")
	})

	t.Run("Get", func(t *testing.T) {
		_, err := repo.Get("passwords", "key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid table name")
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete("secrets", "key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid table name")
	})

	t.Run("DeleteExpired", func(t *testing.T) {
		_, err := repo.DeleteExpired("nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid table name")
	})
}

func TestValidateTable(t *testing.T) {
	// All tables in AllTables should be valid
	for _, table := range AllTables {
		t.Run(table, func(t *testing.T) {
			err := validateTable(table)
			assert.NoError(t, err)
		})
	}
}
