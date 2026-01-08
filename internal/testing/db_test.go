package testing

import (
	"testing"

	"github.com/aristath/sentinel/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTestDB_CreatesInMemoryDatabase tests that NewTestDB creates an in-memory SQLite database
func TestNewTestDB_CreatesInMemoryDatabase(t *testing.T) {
	db, cleanup := NewTestDB(t, "test_db")
	defer cleanup()

	require.NotNil(t, db)
	assert.NotNil(t, db.Conn())

	// Verify it's a valid database connection
	err := db.Conn().Ping()
	require.NoError(t, err)
}

// TestNewTestDB_WithSchemaMigration tests that NewTestDB applies schema migration
func TestNewTestDB_WithSchemaMigration(t *testing.T) {
	db, cleanup := NewTestDB(t, "universe")
	defer cleanup()

	// Verify universe schema was applied
	var exists bool
	err := db.Conn().QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='securities')").Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists, "securities table should exist after migration")

	// Verify other universe tables exist
	err = db.Conn().QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='tags')").Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists, "tags table should exist after migration")
}

// TestNewTestDB_WithMultipleDatabases tests creating multiple test databases
func TestNewTestDB_WithMultipleDatabases(t *testing.T) {
	db1, cleanup1 := NewTestDB(t, "universe")
	defer cleanup1()

	db2, cleanup2 := NewTestDB(t, "portfolio")
	defer cleanup2()

	// Both should be valid independent databases
	err := db1.Conn().Ping()
	require.NoError(t, err)

	err = db2.Conn().Ping()
	require.NoError(t, err)

	// Verify each has its own schema
	var exists bool
	err = db1.Conn().QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='securities')").Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists, "universe db should have securities table")

	err = db2.Conn().QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='positions')").Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists, "portfolio db should have positions table")
}

// TestNewTestDB_WithCustomSchema tests that NewTestDB can use a custom schema
func TestNewTestDB_WithCustomSchema(t *testing.T) {
	customSchema := `
		CREATE TABLE custom_table (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
	`

	db, cleanup := NewTestDBWithSchema(t, "custom_db", customSchema)
	defer cleanup()

	// Verify custom schema was applied
	var exists bool
	err := db.Conn().QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='custom_table')").Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists, "custom_table should exist after custom schema migration")
}

// TestCleanup_ClosesDatabase tests that cleanup function closes the database connection
func TestCleanup_ClosesDatabase(t *testing.T) {
	db, cleanup := NewTestDB(t, "test_db")

	// Verify database is open
	err := db.Conn().Ping()
	require.NoError(t, err)

	// Call cleanup
	cleanup()

	// Verify database is closed
	err = db.Conn().Ping()
	require.Error(t, err, "database should be closed after cleanup")
}

// TestCleanup_CanBeCalledMultipleTimes tests that cleanup is idempotent
func TestCleanup_CanBeCalledMultipleTimes(t *testing.T) {
	db, cleanup := NewTestDB(t, "test_db")

	// Verify database is open
	err := db.Conn().Ping()
	require.NoError(t, err)

	// Call cleanup multiple times
	cleanup()
	cleanup()
	cleanup()

	// Should not panic, database should be closed
	err = db.Conn().Ping()
	require.Error(t, err, "database should be closed after cleanup")
}

// TestNewTestDB_WithInvalidSchemaName tests that NewTestDB handles unknown schema names gracefully
func TestNewTestDB_WithInvalidSchemaName(t *testing.T) {
	// Unknown schema name should create database but not apply any schema
	db, cleanup := NewTestDB(t, "unknown_schema")
	defer cleanup()

	// Database should still be valid
	err := db.Conn().Ping()
	require.NoError(t, err)

	// But no tables should exist
	var count int
	err = db.Conn().QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "unknown schema should not create any tables")
}

// TestNewTestDB_TransactionSupport tests that NewTestDB supports transactions
func TestNewTestDB_TransactionSupport(t *testing.T) {
	db, cleanup := NewTestDB(t, "test_db")
	defer cleanup()

	// Create a simple test table
	_, err := db.Conn().Exec(`
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY,
			value TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	// Test transaction
	tx, err := db.Conn().Begin()
	require.NoError(t, err)

	_, err = tx.Exec("INSERT INTO test_table (value) VALUES (?)", "test")
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	// Verify data was committed
	var count int
	err = db.Conn().QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestNewTestDB_IsolatedDatabases tests that test databases are isolated from each other
func TestNewTestDB_IsolatedDatabases(t *testing.T) {
	// Create two databases
	db1, cleanup1 := NewTestDB(t, "test_db_1")
	defer cleanup1()

	db2, cleanup2 := NewTestDB(t, "test_db_2")
	defer cleanup2()

	// Create same table in both
	_, err := db1.Conn().Exec(`CREATE TABLE test_table (id INTEGER PRIMARY KEY, value TEXT)`)
	require.NoError(t, err)

	_, err = db2.Conn().Exec(`CREATE TABLE test_table (id INTEGER PRIMARY KEY, value TEXT)`)
	require.NoError(t, err)

	// Insert data into db1
	_, err = db1.Conn().Exec("INSERT INTO test_table (value) VALUES (?)", "db1_value")
	require.NoError(t, err)

	// Insert data into db2
	_, err = db2.Conn().Exec("INSERT INTO test_table (value) VALUES (?)", "db2_value")
	require.NoError(t, err)

	// Verify isolation: db1 should only have its own data
	var value1 string
	err = db1.Conn().QueryRow("SELECT value FROM test_table WHERE id = 1").Scan(&value1)
	require.NoError(t, err)
	assert.Equal(t, "db1_value", value1)

	// Verify isolation: db2 should only have its own data
	var value2 string
	err = db2.Conn().QueryRow("SELECT value FROM test_table WHERE id = 1").Scan(&value2)
	require.NoError(t, err)
	assert.Equal(t, "db2_value", value2)
}

// TestNewTestDB_SupportsWALMode tests that test databases use WAL mode
func TestNewTestDB_SupportsWALMode(t *testing.T) {
	db, cleanup := NewTestDB(t, "test_db")
	defer cleanup()

	// Check journal mode
	var journalMode string
	err := db.Conn().QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	require.NoError(t, err)
	assert.Equal(t, "wal", journalMode, "test database should use WAL mode")
}

// TestNewTestDB_ProfileConfiguration tests that test databases use appropriate profile
func TestNewTestDB_ProfileConfiguration(t *testing.T) {
	db, cleanup := NewTestDB(t, "test_db")
	defer cleanup()

	// Test databases should use standard profile
	assert.Equal(t, database.ProfileStandard, db.Profile())
}

// TestNewTestDB_ReturnsValidName tests that test database has correct name
func TestNewTestDB_ReturnsValidName(t *testing.T) {
	db, cleanup := NewTestDB(t, "test_db_name")
	defer cleanup()

	assert.Equal(t, "test_db_name", db.Name())
}

// TestNewTestDB_CanInsertAndQuery tests basic CRUD operations
func TestNewTestDB_CanInsertAndQuery(t *testing.T) {
	db, cleanup := NewTestDB(t, "test_db")
	defer cleanup()

	// Create test table
	_, err := db.Conn().Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL
		)
	`)
	require.NoError(t, err)

	// Insert
	result, err := db.Conn().Exec("INSERT INTO users (name, email) VALUES (?, ?)", "John Doe", "john@example.com")
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Query
	var name, email string
	err = db.Conn().QueryRow("SELECT name, email FROM users WHERE id = ?", id).Scan(&name, &email)
	require.NoError(t, err)
	assert.Equal(t, "John Doe", name)
	assert.Equal(t, "john@example.com", email)
}

// TestNewTestDB_SupportsPreparedStatements tests prepared statement support
func TestNewTestDB_SupportsPreparedStatements(t *testing.T) {
	db, cleanup := NewTestDB(t, "test_db")
	defer cleanup()

	// Create test table
	_, err := db.Conn().Exec(`
		CREATE TABLE products (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			price REAL NOT NULL
		)
	`)
	require.NoError(t, err)

	// Prepare statement
	stmt, err := db.Conn().Prepare("INSERT INTO products (name, price) VALUES (?, ?)")
	require.NoError(t, err)
	defer stmt.Close()

	// Execute prepared statement multiple times
	_, err = stmt.Exec("Product 1", 10.50)
	require.NoError(t, err)

	_, err = stmt.Exec("Product 2", 20.75)
	require.NoError(t, err)

	// Verify data
	var count int
	err = db.Conn().QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}
