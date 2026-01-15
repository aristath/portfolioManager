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

func setupHistoryTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create daily_prices table (consolidated schema)
	_, err = db.Exec(`
		CREATE TABLE daily_prices (
			isin TEXT NOT NULL,
			date INTEGER NOT NULL,
			open REAL NOT NULL,
			high REAL NOT NULL,
			low REAL NOT NULL,
			close REAL NOT NULL,
			volume INTEGER,
			adjusted_close REAL,
			PRIMARY KEY (isin, date)
		) STRICT
	`)
	require.NoError(t, err)

	// Create monthly_prices table
	_, err = db.Exec(`
		CREATE TABLE monthly_prices (
			isin TEXT NOT NULL,
			year_month TEXT NOT NULL,
			avg_close REAL NOT NULL,
			avg_adj_close REAL NOT NULL,
			source TEXT,
			created_at INTEGER,
			PRIMARY KEY (isin, year_month)
		) STRICT
	`)
	require.NoError(t, err)

	// Create indexes
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_prices_isin ON daily_prices(isin);
		CREATE INDEX IF NOT EXISTS idx_prices_date ON daily_prices(date DESC);
		CREATE INDEX IF NOT EXISTS idx_monthly_isin ON monthly_prices(isin);
		CREATE INDEX IF NOT EXISTS idx_monthly_year_month ON monthly_prices(year_month DESC);
	`)
	require.NoError(t, err)

	return db
}

func TestNewHistoryDB(t *testing.T) {
	db := setupHistoryTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	historyDB := NewHistoryDB(db, nil, log) // nil filter for basic tests

	require.NotNil(t, historyDB)
	assert.NotNil(t, historyDB.db)
}

func TestGetDailyPrices_WithISIN(t *testing.T) {
	db := setupHistoryTestDB(t)
	defer db.Close()

	// Insert test data with ISIN (using Unix timestamps for dates)
	date1 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC).Unix()
	date2 := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC).Unix()
	date3 := time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC).Unix()
	date4 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC).Unix() // Same date for different ISIN
	_, err := db.Exec(`
		INSERT INTO daily_prices (isin, date, open, high, low, close, volume, adjusted_close)
		VALUES
			('US0378331005', ?, 185.0, 186.5, 184.0, 185.5, 50000000, 185.5),
			('US0378331005', ?, 185.5, 187.0, 185.0, 186.0, 45000000, 186.0),
			('US0378331005', ?, 186.0, 188.0, 185.5, 187.5, 55000000, 187.5),
			('NL0010273215', ?, 800.0, 810.0, 795.0, 805.0, 1000000, 805.0)
	`, date1, date2, date3, date4)
	require.NoError(t, err)

	log := zerolog.New(nil).Level(zerolog.Disabled)
	historyDB := NewHistoryDB(db, nil, log)

	// Test: Get daily prices for ISIN US0378331005 (AAPL)
	prices, err := historyDB.GetDailyPrices("US0378331005", 10)

	assert.NoError(t, err)
	assert.Len(t, prices, 3)
	assert.Equal(t, "2024-01-04", prices[0].Date) // Most recent first
	assert.Equal(t, 187.5, prices[0].Close)
	assert.Equal(t, 188.0, prices[0].High)
	assert.Equal(t, 185.5, prices[0].Low)
	assert.Equal(t, 186.0, prices[0].Open)
	require.NotNil(t, prices[0].Volume)
	assert.Equal(t, int64(55000000), *prices[0].Volume)
}

func TestGetDailyPrices_NoData(t *testing.T) {
	db := setupHistoryTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	historyDB := NewHistoryDB(db, nil, log)

	// Test: Get daily prices for ISIN with no data
	prices, err := historyDB.GetDailyPrices("US0000000000", 10)

	assert.NoError(t, err)
	assert.Empty(t, prices)
}

func TestGetDailyPrices_Limit(t *testing.T) {
	db := setupHistoryTestDB(t)
	defer db.Close()

	// Insert 5 days of data
	for i := 1; i <= 5; i++ {
		dateUnix := time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC).Unix()
		_, err := db.Exec(`
			INSERT INTO daily_prices (isin, date, open, high, low, close, volume, adjusted_close)
			VALUES (?, ?, 100.0, 105.0, 95.0, 102.0, 1000000, 102.0)
		`, "US0378331005", dateUnix)
		require.NoError(t, err)
	}

	log := zerolog.New(nil).Level(zerolog.Disabled)
	historyDB := NewHistoryDB(db, nil, log)

	// Test: Limit to 3
	prices, err := historyDB.GetDailyPrices("US0378331005", 3)

	assert.NoError(t, err)
	assert.Len(t, prices, 3)
}

func TestGetMonthlyPrices_WithISIN(t *testing.T) {
	db := setupHistoryTestDB(t)
	defer db.Close()

	// Insert monthly data with ISIN
	_, err := db.Exec(`
		INSERT INTO monthly_prices (isin, year_month, avg_close, avg_adj_close, source, created_at)
		VALUES
			('US0378331005', '2024-01', 185.0, 185.0, 'calculated', strftime('%s', 'now')),
			('US0378331005', '2024-02', 186.5, 186.5, 'calculated', strftime('%s', 'now')),
			('US0378331005', '2024-03', 188.0, 188.0, 'calculated', strftime('%s', 'now')),
			('NL0010273215', '2024-01', 800.0, 800.0, 'calculated', strftime('%s', 'now'))
	`)
	require.NoError(t, err)

	log := zerolog.New(nil).Level(zerolog.Disabled)
	historyDB := NewHistoryDB(db, nil, log)

	// Test: Get monthly prices for ISIN US0378331005
	prices, err := historyDB.GetMonthlyPrices("US0378331005", 10)

	assert.NoError(t, err)
	assert.Len(t, prices, 3)
	assert.Equal(t, "2024-03", prices[0].YearMonth) // Most recent first
	assert.Equal(t, 188.0, prices[0].AvgAdjClose)
}

func TestGetMonthlyPrices_NoData(t *testing.T) {
	db := setupHistoryTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	historyDB := NewHistoryDB(db, nil, log)

	// Test: Get monthly prices for ISIN with no data
	prices, err := historyDB.GetMonthlyPrices("US0000000000", 10)

	assert.NoError(t, err)
	assert.Empty(t, prices)
}

func TestHasMonthlyData_WithISIN(t *testing.T) {
	db := setupHistoryTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	historyDB := NewHistoryDB(db, nil, log)

	// Test: No monthly data initially
	hasData, err := historyDB.HasMonthlyData("US0378331005")
	assert.NoError(t, err)
	assert.False(t, hasData)

	// Insert monthly data
	_, err = db.Exec(`
		INSERT INTO monthly_prices (isin, year_month, avg_close, avg_adj_close, source, created_at)
		VALUES ('US0378331005', '2024-01', 185.0, 185.0, 'calculated', strftime('%s', 'now'))
	`)
	require.NoError(t, err)

	// Test: Has monthly data now
	hasData, err = historyDB.HasMonthlyData("US0378331005")
	assert.NoError(t, err)
	assert.True(t, hasData)

	// Test: Different ISIN still has no data
	hasData, err = historyDB.HasMonthlyData("NL0010273215")
	assert.NoError(t, err)
	assert.False(t, hasData)
}

func TestSyncHistoricalPrices_WithISIN(t *testing.T) {
	db := setupHistoryTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	historyDB := NewHistoryDB(db, nil, log)

	// Test data
	isin := "US0378331005"
	prices := []DailyPrice{
		{Date: "2024-01-02", Open: 185.0, High: 186.5, Low: 184.0, Close: 185.5, Volume: intPtr(50000000)},
		{Date: "2024-01-03", Open: 185.5, High: 187.0, Low: 185.0, Close: 186.0, Volume: intPtr(45000000)},
		{Date: "2024-01-04", Open: 186.0, High: 188.0, Low: 185.5, Close: 187.5, Volume: intPtr(55000000)},
	}

	// Test: Sync historical prices
	err := historyDB.SyncHistoricalPrices(isin, prices)
	assert.NoError(t, err)

	// Verify daily prices were inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM daily_prices WHERE isin = ?", isin).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	// Verify monthly prices were aggregated
	err = db.QueryRow("SELECT COUNT(*) FROM monthly_prices WHERE isin = ?", isin).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count) // All 3 days are in January 2024

	// Verify monthly price value
	var avgClose float64
	err = db.QueryRow("SELECT avg_close FROM monthly_prices WHERE isin = ? AND year_month = '2024-01'", isin).Scan(&avgClose)
	assert.NoError(t, err)
	expectedAvg := (185.5 + 186.0 + 187.5) / 3.0
	assert.InDelta(t, expectedAvg, avgClose, 0.01)
}

func TestSyncHistoricalPrices_MultipleISINs(t *testing.T) {
	db := setupHistoryTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	historyDB := NewHistoryDB(db, nil, log)

	// Sync prices for first ISIN
	isin1 := "US0378331005"
	prices1 := []DailyPrice{
		{Date: "2024-01-02", Open: 185.0, High: 186.5, Low: 184.0, Close: 185.5, Volume: intPtr(50000000)},
	}
	err := historyDB.SyncHistoricalPrices(isin1, prices1)
	assert.NoError(t, err)

	// Sync prices for second ISIN
	isin2 := "NL0010273215"
	prices2 := []DailyPrice{
		{Date: "2024-01-02", Open: 800.0, High: 810.0, Low: 795.0, Close: 805.0, Volume: intPtr(1000000)},
	}
	err = historyDB.SyncHistoricalPrices(isin2, prices2)
	assert.NoError(t, err)

	// Verify ISIN isolation - each ISIN has its own data
	var count1, count2 int
	err = db.QueryRow("SELECT COUNT(*) FROM daily_prices WHERE isin = ?", isin1).Scan(&count1)
	assert.NoError(t, err)
	err = db.QueryRow("SELECT COUNT(*) FROM daily_prices WHERE isin = ?", isin2).Scan(&count2)
	assert.NoError(t, err)

	assert.Equal(t, 1, count1)
	assert.Equal(t, 1, count2)

	// Verify data is correct for each ISIN
	var close1, close2 float64
	err = db.QueryRow("SELECT close FROM daily_prices WHERE isin = ?", isin1).Scan(&close1)
	assert.NoError(t, err)
	err = db.QueryRow("SELECT close FROM daily_prices WHERE isin = ?", isin2).Scan(&close2)
	assert.NoError(t, err)

	assert.Equal(t, 185.5, close1)
	assert.Equal(t, 805.0, close2)
}

func TestSyncHistoricalPrices_ReplaceExisting(t *testing.T) {
	db := setupHistoryTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	historyDB := NewHistoryDB(db, nil, log)

	isin := "US0378331005"

	// First sync
	prices1 := []DailyPrice{
		{Date: "2024-01-02", Open: 185.0, High: 186.5, Low: 184.0, Close: 185.5, Volume: intPtr(50000000)},
	}
	err := historyDB.SyncHistoricalPrices(isin, prices1)
	assert.NoError(t, err)

	// Second sync with updated price for same date
	prices2 := []DailyPrice{
		{Date: "2024-01-02", Open: 186.0, High: 187.0, Low: 185.0, Close: 186.5, Volume: intPtr(51000000)},
	}
	err = historyDB.SyncHistoricalPrices(isin, prices2)
	assert.NoError(t, err)

	// Verify only one row exists (replaced, not duplicated)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM daily_prices WHERE isin = ?", isin).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify updated price
	// Convert date string to Unix timestamp for query
	testDate, _ := time.Parse("2006-01-02", "2024-01-02")
	testDateUnix := time.Date(testDate.Year(), testDate.Month(), testDate.Day(), 0, 0, 0, 0, time.UTC).Unix()
	var close float64
	err = db.QueryRow("SELECT close FROM daily_prices WHERE isin = ? AND date = ?", isin, testDateUnix).Scan(&close)
	assert.NoError(t, err)
	assert.Equal(t, 186.5, close) // Updated value, not original
}

func TestSyncHistoricalPrices_MonthlyAggregation(t *testing.T) {
	db := setupHistoryTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	historyDB := NewHistoryDB(db, nil, log)

	isin := "US0378331005"
	// Prices spanning two months
	prices := []DailyPrice{
		{Date: "2024-01-30", Open: 185.0, High: 186.0, Low: 184.0, Close: 185.5, Volume: intPtr(50000000)},
		{Date: "2024-01-31", Open: 185.5, High: 186.5, Low: 185.0, Close: 186.0, Volume: intPtr(45000000)},
		{Date: "2024-02-01", Open: 186.0, High: 187.0, Low: 185.5, Close: 186.5, Volume: intPtr(55000000)},
		{Date: "2024-02-02", Open: 186.5, High: 188.0, Low: 186.0, Close: 187.5, Volume: intPtr(60000000)},
	}

	err := historyDB.SyncHistoricalPrices(isin, prices)
	assert.NoError(t, err)

	// Verify two monthly aggregates were created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM monthly_prices WHERE isin = ?", isin).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)

	// Verify January average
	var janAvg float64
	err = db.QueryRow("SELECT avg_close FROM monthly_prices WHERE isin = ? AND year_month = '2024-01'", isin).Scan(&janAvg)
	assert.NoError(t, err)
	expectedJanAvg := (185.5 + 186.0) / 2.0
	assert.InDelta(t, expectedJanAvg, janAvg, 0.01)

	// Verify February average
	var febAvg float64
	err = db.QueryRow("SELECT avg_close FROM monthly_prices WHERE isin = ? AND year_month = '2024-02'", isin).Scan(&febAvg)
	assert.NoError(t, err)
	expectedFebAvg := (186.5 + 187.5) / 2.0
	assert.InDelta(t, expectedFebAvg, febAvg, 0.01)
}

// Helper function
func intPtr(i int64) *int64 {
	return &i
}

// ==========================================
// Cache and Filtering Tests
// ==========================================

func setupHistoryDBWithFilter(t *testing.T) (*sql.DB, *HistoryDB) {
	db := setupHistoryTestDB(t)
	log := zerolog.New(nil).Level(zerolog.Disabled)
	priceFilter := NewPriceFilter(log)
	historyDB := NewHistoryDB(db, priceFilter, log)
	return db, historyDB
}

func TestHistoryDB_GetDailyPrices_FiltersAnomalies(t *testing.T) {
	db, historyDB := setupHistoryDBWithFilter(t)
	defer db.Close()

	// Insert data including an anomaly (extreme high)
	date1 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC).Unix()
	date2 := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC).Unix()
	date3 := time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC).Unix()
	_, err := db.Exec(`
		INSERT INTO daily_prices (isin, date, open, high, low, close, volume, adjusted_close)
		VALUES
			('US0378331005', ?, 185.0, 186.5, 184.0, 185.5, 50000000, 185.5),
			('US0378331005', ?, 185.5, 50000.0, 185.0, 186.0, 45000000, 186.0),
			('US0378331005', ?, 186.0, 188.0, 185.5, 187.5, 55000000, 187.5)
	`, date1, date2, date3)
	require.NoError(t, err)

	// The anomaly (High=50000 while Close=186) should be filtered out
	prices, err := historyDB.GetDailyPrices("US0378331005", 10)
	assert.NoError(t, err)
	assert.Len(t, prices, 2, "Anomaly should be filtered out")
	// Results are ordered by date descending (most recent first)
	assert.Equal(t, 187.5, prices[0].Close) // Jan 4 (most recent)
	assert.Equal(t, 185.5, prices[1].Close) // Jan 2 (oldest)
}

func TestHistoryDB_GetDailyPrices_CachesResult(t *testing.T) {
	db, historyDB := setupHistoryDBWithFilter(t)
	defer db.Close()

	// Insert data
	date1 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC).Unix()
	_, err := db.Exec(`
		INSERT INTO daily_prices (isin, date, open, high, low, close, volume, adjusted_close)
		VALUES ('US0378331005', ?, 185.0, 186.5, 184.0, 185.5, 50000000, 185.5)
	`, date1)
	require.NoError(t, err)

	// First call
	prices1, err := historyDB.GetDailyPrices("US0378331005", 10)
	assert.NoError(t, err)
	assert.Len(t, prices1, 1)

	// Delete data from DB directly (simulating stale cache scenario)
	_, err = db.Exec("DELETE FROM daily_prices WHERE isin = 'US0378331005'")
	require.NoError(t, err)

	// Second call should return cached data (DB is now empty)
	prices2, err := historyDB.GetDailyPrices("US0378331005", 10)
	assert.NoError(t, err)
	assert.Len(t, prices2, 1, "Should return cached data")
	assert.Equal(t, prices1[0].Close, prices2[0].Close)
}

func TestHistoryDB_SyncHistoricalPrices_InvalidatesCache(t *testing.T) {
	db, historyDB := setupHistoryDBWithFilter(t)
	defer db.Close()

	isin := "US0378331005"

	// First sync and read (populates cache)
	prices1 := []DailyPrice{
		{Date: "2024-01-02", Open: 185.0, High: 186.5, Low: 184.0, Close: 185.5, Volume: intPtr(50000000)},
	}
	err := historyDB.SyncHistoricalPrices(isin, prices1)
	require.NoError(t, err)

	cached1, err := historyDB.GetDailyPrices(isin, 10)
	require.NoError(t, err)
	require.Len(t, cached1, 1)
	assert.Equal(t, 185.5, cached1[0].Close)

	// Second sync with different data (should invalidate cache)
	prices2 := []DailyPrice{
		{Date: "2024-01-02", Open: 200.0, High: 205.0, Low: 195.0, Close: 200.0, Volume: intPtr(60000000)},
	}
	err = historyDB.SyncHistoricalPrices(isin, prices2)
	require.NoError(t, err)

	// Should get new data, not cached data
	cached2, err := historyDB.GetDailyPrices(isin, 10)
	require.NoError(t, err)
	require.Len(t, cached2, 1)
	assert.Equal(t, 200.0, cached2[0].Close, "Cache should be invalidated, showing new data")
}

func TestHistoryDB_Cache_IndependentPerISIN(t *testing.T) {
	db, historyDB := setupHistoryDBWithFilter(t)
	defer db.Close()

	isin1 := "US0378331005"
	isin2 := "NL0010273215"

	// Sync and cache first ISIN
	err := historyDB.SyncHistoricalPrices(isin1, []DailyPrice{
		{Date: "2024-01-02", Open: 185.0, High: 186.5, Low: 184.0, Close: 185.5},
	})
	require.NoError(t, err)
	_, err = historyDB.GetDailyPrices(isin1, 10)
	require.NoError(t, err)

	// Sync and cache second ISIN
	err = historyDB.SyncHistoricalPrices(isin2, []DailyPrice{
		{Date: "2024-01-02", Open: 800.0, High: 810.0, Low: 795.0, Close: 805.0},
	})
	require.NoError(t, err)
	_, err = historyDB.GetDailyPrices(isin2, 10)
	require.NoError(t, err)

	// Update first ISIN (should only invalidate its cache)
	err = historyDB.SyncHistoricalPrices(isin1, []DailyPrice{
		{Date: "2024-01-02", Open: 200.0, High: 205.0, Low: 195.0, Close: 200.0},
	})
	require.NoError(t, err)

	// Delete second ISIN from DB (to verify its cache is still intact)
	_, err = db.Exec("DELETE FROM daily_prices WHERE isin = ?", isin2)
	require.NoError(t, err)

	// Second ISIN should still return cached data
	prices2, err := historyDB.GetDailyPrices(isin2, 10)
	assert.NoError(t, err)
	assert.Len(t, prices2, 1, "Second ISIN cache should be unaffected")
	assert.Equal(t, 805.0, prices2[0].Close)

	// First ISIN should return fresh data
	prices1, err := historyDB.GetDailyPrices(isin1, 10)
	assert.NoError(t, err)
	assert.Len(t, prices1, 1)
	assert.Equal(t, 200.0, prices1[0].Close, "First ISIN should have fresh data")
}

func TestHistoryDB_GetRecentPrices_UsesFilterAndCache(t *testing.T) {
	db, historyDB := setupHistoryDBWithFilter(t)
	defer db.Close()

	// Insert recent data including an anomaly
	now := time.Now()
	date1 := time.Date(now.Year(), now.Month(), now.Day()-2, 0, 0, 0, 0, time.UTC).Unix()
	date2 := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC).Unix()

	_, err := db.Exec(`
		INSERT INTO daily_prices (isin, date, open, high, low, close, volume, adjusted_close)
		VALUES
			('US0378331005', ?, 185.0, 186.5, 184.0, 185.5, 50000000, 185.5),
			('US0378331005', ?, 185.5, 50000.0, 185.0, 186.0, 45000000, 186.0)
	`, date1, date2)
	require.NoError(t, err)

	// The anomaly should be filtered out
	prices, err := historyDB.GetRecentPrices("US0378331005", 30)
	assert.NoError(t, err)
	assert.Len(t, prices, 1, "Anomaly should be filtered out")
	assert.Equal(t, 185.5, prices[0].Close)
}

func TestHistoryDB_InvalidateCache(t *testing.T) {
	db, historyDB := setupHistoryDBWithFilter(t)
	defer db.Close()

	isin := "US0378331005"

	// Sync and cache
	err := historyDB.SyncHistoricalPrices(isin, []DailyPrice{
		{Date: "2024-01-02", Open: 185.0, High: 186.5, Low: 184.0, Close: 185.5},
	})
	require.NoError(t, err)
	_, err = historyDB.GetDailyPrices(isin, 10)
	require.NoError(t, err)

	// Delete from DB
	_, err = db.Exec("DELETE FROM daily_prices WHERE isin = ?", isin)
	require.NoError(t, err)

	// Should still return cached data
	prices1, err := historyDB.GetDailyPrices(isin, 10)
	assert.NoError(t, err)
	assert.Len(t, prices1, 1)

	// Invalidate cache
	historyDB.InvalidateCache(isin)

	// Now should return empty (DB is empty)
	prices2, err := historyDB.GetDailyPrices(isin, 10)
	assert.NoError(t, err)
	assert.Empty(t, prices2, "After invalidation, should fetch from empty DB")
}

func TestHistoryDB_InvalidateAllCaches(t *testing.T) {
	db, historyDB := setupHistoryDBWithFilter(t)
	defer db.Close()

	isin1 := "US0378331005"
	isin2 := "NL0010273215"

	// Sync and cache both ISINs
	err := historyDB.SyncHistoricalPrices(isin1, []DailyPrice{
		{Date: "2024-01-02", Open: 185.0, High: 186.5, Low: 184.0, Close: 185.5},
	})
	require.NoError(t, err)
	_, err = historyDB.GetDailyPrices(isin1, 10)
	require.NoError(t, err)

	err = historyDB.SyncHistoricalPrices(isin2, []DailyPrice{
		{Date: "2024-01-02", Open: 800.0, High: 810.0, Low: 795.0, Close: 805.0},
	})
	require.NoError(t, err)
	_, err = historyDB.GetDailyPrices(isin2, 10)
	require.NoError(t, err)

	// Delete all from DB
	_, err = db.Exec("DELETE FROM daily_prices")
	require.NoError(t, err)

	// Should still return cached data for both
	prices1, _ := historyDB.GetDailyPrices(isin1, 10)
	assert.Len(t, prices1, 1)
	prices2, _ := historyDB.GetDailyPrices(isin2, 10)
	assert.Len(t, prices2, 1)

	// Invalidate all caches
	historyDB.InvalidateAllCaches()

	// Now both should return empty
	prices1, _ = historyDB.GetDailyPrices(isin1, 10)
	assert.Empty(t, prices1)
	prices2, _ = historyDB.GetDailyPrices(isin2, 10)
	assert.Empty(t, prices2)
}

func TestHistoryDB_GetDailyPrices_LimitWorksWithCache(t *testing.T) {
	db, historyDB := setupHistoryDBWithFilter(t)
	defer db.Close()

	// Insert 5 days of data
	for i := 1; i <= 5; i++ {
		dateUnix := time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC).Unix()
		_, err := db.Exec(`
			INSERT INTO daily_prices (isin, date, open, high, low, close, volume, adjusted_close)
			VALUES (?, ?, 100.0, 105.0, 95.0, ?, 1000000, ?)
		`, "US0378331005", dateUnix, float64(100+i), float64(100+i))
		require.NoError(t, err)
	}

	// First call with limit 3
	prices1, err := historyDB.GetDailyPrices("US0378331005", 3)
	assert.NoError(t, err)
	assert.Len(t, prices1, 3)

	// Second call with limit 5 (should still work from cache)
	prices2, err := historyDB.GetDailyPrices("US0378331005", 5)
	assert.NoError(t, err)
	assert.Len(t, prices2, 5)

	// Third call with limit 2
	prices3, err := historyDB.GetDailyPrices("US0378331005", 2)
	assert.NoError(t, err)
	assert.Len(t, prices3, 2)
}

func TestSyncHistoricalPrices_MonthlyAggregationUsesFilteredData(t *testing.T) {
	db, historyDB := setupHistoryDBWithFilter(t)
	defer db.Close()

	isin := "US0378331005"

	// Include an anomaly (extreme high that should be filtered)
	// High > Close × 100 threshold means High=50000 with Close=200 is an anomaly
	// All prices have valid OHLC relationships (High >= Close, Low <= Close, etc.)
	prices := []DailyPrice{
		{Date: "2024-02-01", Open: 100.0, High: 105.0, Low: 95.0, Close: 100.0, Volume: intPtr(1000000)},
		{Date: "2024-02-02", Open: 200.0, High: 50000.0, Low: 195.0, Close: 200.0, Volume: intPtr(1000000)}, // Anomaly: High (50000) > Close×100 (20000)
		{Date: "2024-02-03", Open: 148.0, High: 155.0, Low: 145.0, Close: 150.0, Volume: intPtr(1000000)},   // Valid OHLC
	}

	err := historyDB.SyncHistoricalPrices(isin, prices)
	require.NoError(t, err)

	// Verify all 3 daily prices were stored (raw data)
	var dailyCount int
	err = db.QueryRow("SELECT COUNT(*) FROM daily_prices WHERE isin = ?", isin).Scan(&dailyCount)
	require.NoError(t, err)
	assert.Equal(t, 3, dailyCount, "All 3 daily prices should be stored (raw)")

	// Verify monthly average for February
	// If anomaly included: avg = (100 + 200 + 150) / 3 = 150
	// With anomaly filtered: avg = (100 + 150) / 2 = 125
	var febAvg float64
	err = db.QueryRow("SELECT avg_close FROM monthly_prices WHERE isin = ? AND year_month = '2024-02'", isin).Scan(&febAvg)
	require.NoError(t, err)
	assert.InDelta(t, 125.0, febAvg, 0.01, "Monthly average should be from filtered data (125), not raw data (150)")
}

func TestSyncHistoricalPrices_MonthlyAggregationExcludesCrash(t *testing.T) {
	db, historyDB := setupHistoryDBWithFilter(t)
	defer db.Close()

	isin := "US0378331005"

	// First establish context with valid prices (needed for day-over-day crash detection)
	contextPrices := []DailyPrice{
		{Date: "2024-01-01", Open: 100.0, High: 105.0, Low: 95.0, Close: 100.0, Volume: intPtr(1000000)},
	}
	err := historyDB.SyncHistoricalPrices(isin, contextPrices)
	require.NoError(t, err)

	// Now sync prices that include a crash (>90% drop in one day)
	prices := []DailyPrice{
		{Date: "2024-01-01", Open: 100.0, High: 105.0, Low: 95.0, Close: 100.0, Volume: intPtr(1000000)},
		{Date: "2024-01-02", Open: 100.0, High: 105.0, Low: 5.0, Close: 5.0, Volume: intPtr(1000000)}, // 95% crash - anomaly
		{Date: "2024-01-03", Open: 100.0, High: 105.0, Low: 95.0, Close: 102.0, Volume: intPtr(1000000)},
	}

	err = historyDB.SyncHistoricalPrices(isin, prices)
	require.NoError(t, err)

	// Verify monthly average excludes the crash price
	// If crash included: avg = (100 + 5 + 102) / 3 = 69
	// With crash filtered: avg = (100 + 102) / 2 = 101
	var janAvg float64
	err = db.QueryRow("SELECT avg_close FROM monthly_prices WHERE isin = ? AND year_month = '2024-01'", isin).Scan(&janAvg)
	require.NoError(t, err)
	assert.InDelta(t, 101.0, janAvg, 0.01, "Monthly average should exclude crash anomaly")
}
