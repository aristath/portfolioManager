package charts

import (
	"database/sql"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestUniverseDB creates an in-memory SQLite database with securities
func setupTestUniverseDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE securities (
			isin TEXT PRIMARY KEY,
			symbol TEXT NOT NULL,
			name TEXT NOT NULL,
			product_type TEXT,
			active INTEGER DEFAULT 1
		)
	`)
	require.NoError(t, err)

	return db
}

// setupTestHistoryDB creates an in-memory SQLite database for historical prices
func setupTestHistoryDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create daily_prices table with unix timestamp for date (matches production schema)
	_, err = db.Exec(`
		CREATE TABLE daily_prices (
			isin TEXT NOT NULL,
			date INTEGER NOT NULL,
			close REAL NOT NULL,
			PRIMARY KEY (isin, date)
		)
	`)
	require.NoError(t, err)

	return db
}

func TestGetSparklinesAggregated_ExcludesIndices(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	universeDB := setupTestUniverseDB(t)
	defer universeDB.Close()

	historyDB := setupTestHistoryDB(t)
	defer historyDB.Close()

	// Insert regular securities
	_, err := universeDB.Exec(`
		INSERT INTO securities (isin, symbol, name, product_type, active)
		VALUES
			('US0378331005', 'AAPL', 'Apple Inc.', 'EQUITY', 1),
			('US5949181045', 'MSFT', 'Microsoft Corp', 'EQUITY', 1)
	`)
	require.NoError(t, err)

	// Insert market indices (should be excluded from sparklines)
	_, err = universeDB.Exec(`
		INSERT INTO securities (isin, symbol, name, product_type, active)
		VALUES
			('INDEX-SP500.IDX', 'SP500.IDX', 'S&P 500', 'INDEX', 1),
			('INDEX-NASDAQ.IDX', 'NASDAQ.IDX', 'NASDAQ Composite', 'INDEX', 1)
	`)
	require.NoError(t, err)

	// Insert historical prices for all securities using Unix timestamps
	// Use dates within the past year so they're included in 1Y range
	now := time.Now()
	week1 := now.AddDate(0, -6, 0).Unix() // 6 months ago
	week2 := now.AddDate(0, -5, 0).Unix() // 5 months ago

	_, err = historyDB.Exec(`
		INSERT INTO daily_prices (isin, date, close)
		VALUES
			('US0378331005', ?, 150.0),
			('US0378331005', ?, 155.0),
			('US5949181045', ?, 300.0),
			('US5949181045', ?, 310.0),
			('INDEX-SP500.IDX', ?, 4500.0),
			('INDEX-SP500.IDX', ?, 4550.0),
			('INDEX-NASDAQ.IDX', ?, 14000.0),
			('INDEX-NASDAQ.IDX', ?, 14100.0)
	`, week1, week2, week1, week2, week1, week2, week1, week2)
	require.NoError(t, err)

	service := NewService(historyDB, nil, universeDB, log)

	// Execute
	sparklines, err := service.GetSparklinesAggregated("1Y")
	require.NoError(t, err)

	// Verify only regular securities are included
	assert.Contains(t, sparklines, "AAPL", "AAPL should be included")
	assert.Contains(t, sparklines, "MSFT", "MSFT should be included")
	assert.NotContains(t, sparklines, "SP500.IDX", "Index SP500.IDX should be excluded")
	assert.NotContains(t, sparklines, "NASDAQ.IDX", "Index NASDAQ.IDX should be excluded")
	assert.Len(t, sparklines, 2, "Should have exactly 2 securities")
}

func TestGetSparklinesAggregated_IncludesNullProductType(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	universeDB := setupTestUniverseDB(t)
	defer universeDB.Close()

	historyDB := setupTestHistoryDB(t)
	defer historyDB.Close()

	// Insert security with NULL product_type (should be included)
	_, err := universeDB.Exec(`
		INSERT INTO securities (isin, symbol, name, product_type, active)
		VALUES
			('US0378331005', 'AAPL', 'Apple Inc.', NULL, 1)
	`)
	require.NoError(t, err)

	// Insert index (should be excluded)
	_, err = universeDB.Exec(`
		INSERT INTO securities (isin, symbol, name, product_type, active)
		VALUES
			('INDEX-SP500.IDX', 'SP500.IDX', 'S&P 500', 'INDEX', 1)
	`)
	require.NoError(t, err)

	// Insert historical prices using Unix timestamps within the past year
	now := time.Now()
	recentDate := now.AddDate(0, -3, 0).Unix() // 3 months ago

	_, err = historyDB.Exec(`
		INSERT INTO daily_prices (isin, date, close)
		VALUES
			('US0378331005', ?, 150.0),
			('INDEX-SP500.IDX', ?, 4500.0)
	`, recentDate, recentDate)
	require.NoError(t, err)

	service := NewService(historyDB, nil, universeDB, log)

	// Execute
	sparklines, err := service.GetSparklinesAggregated("1Y")
	require.NoError(t, err)

	// Verify NULL product_type is included
	assert.Contains(t, sparklines, "AAPL", "AAPL with NULL product_type should be included")
	assert.NotContains(t, sparklines, "SP500.IDX", "Index should be excluded")
	assert.Len(t, sparklines, 1, "Should have exactly 1 security")
}
