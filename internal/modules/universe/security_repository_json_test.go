package universe_test

import (
	"database/sql"
	"testing"

	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// setupTestRepo creates a temporary in-memory database with JSON schema for testing
func setupTestRepo(t *testing.T) universe.SecurityProvider {
	t.Helper()

	// Create in-memory database
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Create JSON-based schema (migration 038)
	_, err = db.Exec(`
		CREATE TABLE securities (
			isin TEXT PRIMARY KEY,
			symbol TEXT NOT NULL,
			data TEXT NOT NULL,
			last_synced INTEGER
		) STRICT;

		CREATE INDEX idx_securities_symbol ON securities(symbol);
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	// Return repository implementation
	// Note: This will need to be updated when we refactor the concrete implementation
	return nil // TODO: Return actual repo once implemented
}

// Test suite for JSON-based security repository
func TestSecurityRepository_JSON(t *testing.T) {
	t.Skip("TODO: Implement after refactoring concrete repository")

	repo := setupTestRepo(t)
	if repo == nil {
		t.Skip("Repository not yet implemented")
	}

	t.Run("Create stores data as JSON", func(t *testing.T) {
		_ = universe.Security{
			ISIN:               "US0378331005",
			Symbol:             "AAPL.US",
			Name:               "Apple Inc.",
			Geography:          "US",
			Industry:           "Technology",
			Currency:           "USD",
			MinLot:             1,
			ProductType:        "EQUITY",
			MinPortfolioTarget: 0.0,
			MaxPortfolioTarget: 0.15,
		}

		// Create security - should serialize to JSON
		// err := repo.Create(security)
		// require.NoError(t, err)

		// Verify JSON stored correctly
		retrieved, err := repo.GetByISIN("US0378331005")
		require.NoError(t, err)
		assert.Equal(t, "Apple Inc.", retrieved.Name)
		assert.Equal(t, "US", retrieved.Geography)
		assert.Equal(t, "Technology", retrieved.Industry)
		assert.Equal(t, 1, retrieved.MinLot)
	})

	t.Run("GetBySymbol returns correct security", func(t *testing.T) {
		security, err := repo.GetBySymbol("AAPL.US")
		require.NoError(t, err)
		assert.Equal(t, "US0378331005", security.ISIN)
		assert.Equal(t, "Apple Inc.", security.Name)
	})

	t.Run("GetISINBySymbol returns ISIN only", func(t *testing.T) {
		isin, err := repo.GetISINBySymbol("AAPL.US")
		require.NoError(t, err)
		assert.Equal(t, "US0378331005", isin)
	})

	t.Run("BatchGetISINsBySymbols returns map", func(t *testing.T) {
		symbols := []string{"AAPL.US", "MSFT.US", "GOOGL.US"}

		// First create test securities
		// ... (TODO: Create MSFT and GOOGL)

		mapping, err := repo.BatchGetISINsBySymbols(symbols)
		require.NoError(t, err)
		assert.Len(t, mapping, 3)
		assert.Equal(t, "US0378331005", mapping["AAPL.US"])
	})

	t.Run("GetTradable excludes indices", func(t *testing.T) {
		// Create an index security
		_ = universe.Security{
			ISIN:        "US1234567890",
			Symbol:      "SPX.US",
			Name:        "S&P 500 Index",
			ProductType: "INDEX",
			Geography:   "US",
			Currency:    "USD",
		}
		// err := repo.Create(indexSecurity)
		// require.NoError(t, err)

		securities, err := repo.GetTradable()
		require.NoError(t, err)

		// Should not include the INDEX
		for _, sec := range securities {
			assert.NotEqual(t, "INDEX", sec.ProductType)
		}
	})

	t.Run("GetDistinctGeographies returns unique list", func(t *testing.T) {
		geographies, err := repo.GetDistinctGeographies()
		require.NoError(t, err)
		assert.Contains(t, geographies, "US")

		// Should not contain empty strings or "0"
		assert.NotContains(t, geographies, "")
		assert.NotContains(t, geographies, "0")

		// Should be sorted and unique
		seenMap := make(map[string]bool)
		for _, geo := range geographies {
			assert.False(t, seenMap[geo], "Geography %s appears multiple times", geo)
			seenMap[geo] = true
		}
	})

	t.Run("GetDistinctIndustries returns unique list", func(t *testing.T) {
		industries, err := repo.GetDistinctIndustries()
		require.NoError(t, err)
		assert.Contains(t, industries, "Technology")

		// Should not contain empty strings
		assert.NotContains(t, industries, "")
	})

	t.Run("GetSecuritiesForOptimization returns optimization data", func(t *testing.T) {
		data, err := repo.GetSecuritiesForOptimization()
		require.NoError(t, err)
		assert.Greater(t, len(data), 0)

		for _, sec := range data {
			assert.NotEmpty(t, sec.Symbol)
			assert.NotEmpty(t, sec.ISIN)
			// Ensure numeric fields are present
			assert.GreaterOrEqual(t, sec.MinPortfolioTarget, 0.0)
			assert.GreaterOrEqual(t, sec.MaxPortfolioTarget, 0.0)
		}
	})

	t.Run("GetSecuritiesForCharts returns chart data", func(t *testing.T) {
		data, err := repo.GetSecuritiesForCharts()
		require.NoError(t, err)
		assert.Greater(t, len(data), 0)

		for _, sec := range data {
			assert.NotEmpty(t, sec.Symbol)
			assert.NotEmpty(t, sec.ISIN)
		}
	})

	t.Run("Update modifies JSON data", func(t *testing.T) {
		// This test requires SecurityRepositoryInterface (write operations)
		t.Skip("Requires write interface")
	})

	t.Run("Delete removes security when no positions", func(t *testing.T) {
		t.Skip("Requires write interface")
	})

	t.Run("Delete fails when positions exist", func(t *testing.T) {
		t.Skip("Requires write interface and positions table")
	})

	t.Run("Exists returns true for existing security", func(t *testing.T) {
		exists, err := repo.Exists("US0378331005")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Exists returns false for non-existing security", func(t *testing.T) {
		exists, err := repo.Exists("NONEXISTENT")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("ExistsBySymbol works correctly", func(t *testing.T) {
		exists, err := repo.ExistsBySymbol("AAPL.US")
		require.NoError(t, err)
		assert.True(t, exists)

		exists, err = repo.ExistsBySymbol("NONEXISTENT.US")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("CountTradable returns correct count", func(t *testing.T) {
		count, err := repo.CountTradable()
		require.NoError(t, err)
		assert.Greater(t, count, 0)
	})

	t.Run("GetByIdentifier works with ISIN", func(t *testing.T) {
		security, err := repo.GetByIdentifier("US0378331005")
		require.NoError(t, err)
		assert.Equal(t, "AAPL.US", security.Symbol)
	})

	t.Run("GetByIdentifier works with symbol", func(t *testing.T) {
		security, err := repo.GetByIdentifier("AAPL.US")
		require.NoError(t, err)
		assert.Equal(t, "US0378331005", security.ISIN)
	})
}

// TestJSONHelpers - Obsolete test removed
// SecurityToJSON no longer exists - only raw Tradernet data is stored in the data column
// See json_helpers_tradernet_test.go for raw Tradernet format parsing tests

// TestSymbolISINConversion tests the symbol/ISIN conversion methods
func TestSymbolISINConversion(t *testing.T) {
	t.Skip("TODO: Implement after refactoring concrete repository")

	repo := setupTestRepo(t)
	if repo == nil {
		t.Skip("Repository not yet implemented")
	}

	t.Run("GetISINBySymbol for single symbol", func(t *testing.T) {
		isin, err := repo.GetISINBySymbol("AAPL.US")
		require.NoError(t, err)
		assert.Equal(t, "US0378331005", isin)
	})

	t.Run("GetSymbolByISIN for single ISIN", func(t *testing.T) {
		symbol, err := repo.GetSymbolByISIN("US0378331005")
		require.NoError(t, err)
		assert.Equal(t, "AAPL.US", symbol)
	})

	t.Run("BatchGetISINsBySymbols for multiple symbols", func(t *testing.T) {
		symbols := []string{"AAPL.US", "MSFT.US"}
		mapping, err := repo.BatchGetISINsBySymbols(symbols)
		require.NoError(t, err)

		assert.Len(t, mapping, 2)
		assert.Equal(t, "US0378331005", mapping["AAPL.US"])
		assert.Contains(t, mapping, "MSFT.US")
	})

	t.Run("BatchGetISINsBySymbols with empty slice", func(t *testing.T) {
		mapping, err := repo.BatchGetISINsBySymbols([]string{})
		require.NoError(t, err)
		assert.Len(t, mapping, 0)
	})

	t.Run("GetISINBySymbol returns error for non-existent", func(t *testing.T) {
		_, err := repo.GetISINBySymbol("NONEXISTENT.US")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
