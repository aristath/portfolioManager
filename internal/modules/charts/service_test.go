package charts

import (
	"database/sql"
	"testing"
	"time"

	"github.com/aristath/sentinel/internal/modules/universe"
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
// and returns it wrapped as a HistoryDB with nil filter (tests handle their own data)
func setupTestHistoryDB(t *testing.T) (*sql.DB, *universe.HistoryDB) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create daily_prices table with full schema (matches production)
	_, err = db.Exec(`
		CREATE TABLE daily_prices (
			isin TEXT NOT NULL,
			date INTEGER NOT NULL,
			open REAL,
			high REAL,
			low REAL,
			close REAL NOT NULL,
			volume INTEGER,
			adjusted_close REAL,
			PRIMARY KEY (isin, date)
		)
	`)
	require.NoError(t, err)

	// Create monthly_prices table
	_, err = db.Exec(`
		CREATE TABLE monthly_prices (
			isin TEXT NOT NULL,
			year_month TEXT NOT NULL,
			avg_close REAL,
			avg_adj_close REAL,
			source TEXT,
			created_at INTEGER,
			PRIMARY KEY (isin, year_month)
		)
	`)
	require.NoError(t, err)

	log := zerolog.New(nil).Level(zerolog.Disabled)
	historyDBClient := universe.NewHistoryDB(db, nil, log) // nil filter = no filtering for tests

	return db, historyDBClient
}

func TestGetSparklinesAggregated_ExcludesIndices(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	universeDB := setupTestUniverseDB(t)
	defer universeDB.Close()

	historyDB, historyDBClient := setupTestHistoryDB(t)
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
		INSERT INTO daily_prices (isin, date, close, open, high, low)
		VALUES
			('US0378331005', ?, 150.0, 149.0, 151.0, 148.0),
			('US0378331005', ?, 155.0, 154.0, 156.0, 153.0),
			('US5949181045', ?, 300.0, 299.0, 301.0, 298.0),
			('US5949181045', ?, 310.0, 309.0, 311.0, 308.0),
			('INDEX-SP500.IDX', ?, 4500.0, 4490.0, 4510.0, 4480.0),
			('INDEX-SP500.IDX', ?, 4550.0, 4540.0, 4560.0, 4530.0),
			('INDEX-NASDAQ.IDX', ?, 14000.0, 13990.0, 14010.0, 13980.0),
			('INDEX-NASDAQ.IDX', ?, 14100.0, 14090.0, 14110.0, 14080.0)
	`, week1, week2, week1, week2, week1, week2, week1, week2)
	require.NoError(t, err)

	service := NewService(historyDBClient, nil, universeDB, log)

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

	historyDB, historyDBClient := setupTestHistoryDB(t)
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
		INSERT INTO daily_prices (isin, date, close, open, high, low)
		VALUES
			('US0378331005', ?, 150.0, 149.0, 151.0, 148.0),
			('INDEX-SP500.IDX', ?, 4500.0, 4490.0, 4510.0, 4480.0)
	`, recentDate, recentDate)
	require.NoError(t, err)

	service := NewService(historyDBClient, nil, universeDB, log)

	// Execute
	sparklines, err := service.GetSparklinesAggregated("1Y")
	require.NoError(t, err)

	// Verify NULL product_type is included
	assert.Contains(t, sparklines, "AAPL", "AAPL with NULL product_type should be included")
	assert.NotContains(t, sparklines, "SP500.IDX", "Index should be excluded")
	assert.Len(t, sparklines, 1, "Should have exactly 1 security")
}
