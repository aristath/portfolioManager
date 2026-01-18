//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aristath/sentinel/internal/clients/tradernet"
	"github.com/aristath/sentinel/internal/database"
	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetadataBatchSync_EndToEnd tests the complete batch metadata sync flow
// from repository ISINs retrieval through batch API call to database updates
func TestMetadataBatchSync_EndToEnd(t *testing.T) {
	apiKey := os.Getenv("TRADERNET_API_KEY")
	apiSecret := os.Getenv("TRADERNET_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping E2E test: TRADERNET_API_KEY and TRADERNET_API_SECRET not set")
	}

	// Setup temporary database
	tmpDir, err := os.MkdirTemp("", "sentinel-e2e-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "universe_test.db")

	// Initialize database
	db, err := database.New(dbPath, database.UniverseProfile, zerolog.New(os.Stderr).Level(zerolog.InfoLevel))
	require.NoError(t, err)
	defer db.Close()

	// Run migrations
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS securities (
			isin TEXT PRIMARY KEY,
			symbol TEXT NOT NULL,
			name TEXT,
			product_type TEXT,
			active INTEGER DEFAULT 1,
			data TEXT,
			last_synced INTEGER,
			created_at INTEGER DEFAULT (strftime('%s', 'now')),
			updated_at INTEGER DEFAULT (strftime('%s', 'now'))
		)`,
	}

	for _, migration := range migrations {
		_, err := db.Conn.Exec(migration)
		require.NoError(t, err)
	}

	// Insert test securities (non-index securities only)
	testSecurities := []struct {
		isin   string
		symbol string
		name   string
	}{
		{"US0378331005", "AAPL.US", "Apple Inc"},
		{"US5949181045", "MSFT.US", "Microsoft Corp"},
		{"US02079K3059", "GOOGL.US", "Alphabet Inc"},
		{"US0231351067", "AMZN.US", "Amazon.com Inc"},
		{"US88160R1014", "TSLA.US", "Tesla Inc"},
		{"US67066G1040", "NVDA.US", "NVIDIA Corp"},
		{"US30303M1027", "META.US", "Meta Platforms Inc"},
		{"US0846707026", "BRK.B.US", "Berkshire Hathaway Inc"},
		{"US46625H1005", "JPM.US", "JPMorgan Chase & Co"},
		{"US92826C8394", "V.US", "Visa Inc"},
	}

	for _, sec := range testSecurities {
		_, err := db.Conn.Exec(
			`INSERT INTO securities (isin, symbol, name, product_type, active, last_synced)
			 VALUES (?, ?, ?, ?, 1, NULL)`,
			sec.isin, sec.symbol, sec.name, "STOCK",
		)
		require.NoError(t, err)
	}

	// Initialize components
	log := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	brokerClient := tradernet.NewClient(apiKey, apiSecret, log)
	securityRepo := universe.NewSecurityRepository(db, log)
	metadataSyncService := universe.NewMetadataSyncService(securityRepo, brokerClient, log)

	// Verify initial state - no securities should be synced
	var syncedCount int
	err = db.Conn.QueryRow("SELECT COUNT(*) FROM securities WHERE last_synced IS NOT NULL").Scan(&syncedCount)
	require.NoError(t, err)
	assert.Equal(t, 0, syncedCount, "No securities should be synced initially")

	// Get all active ISINs
	isins := metadataSyncService.GetAllActiveISINs()
	require.Equal(t, len(testSecurities), len(isins), "Should get all test securities")

	// Execute batch sync
	successCount, err := metadataSyncService.SyncMetadataBatch(isins)
	require.NoError(t, err, "Batch sync should complete without error")
	assert.Greater(t, successCount, 0, "Should successfully sync at least some securities")

	// Verify database updates
	err = db.Conn.QueryRow("SELECT COUNT(*) FROM securities WHERE last_synced IS NOT NULL").Scan(&syncedCount)
	require.NoError(t, err)
	assert.Equal(t, successCount, syncedCount, "Synced count should match success count")

	// Verify last_synced timestamps are recent (within last 5 seconds)
	now := time.Now().Unix()
	var oldestSync int64
	err = db.Conn.QueryRow("SELECT MIN(last_synced) FROM securities WHERE last_synced IS NOT NULL").Scan(&oldestSync)
	require.NoError(t, err)
	assert.True(t, now-oldestSync < 5, "All synced timestamps should be recent (within 5 seconds)")

	// Verify data field is populated
	var dataCount int
	err = db.Conn.QueryRow("SELECT COUNT(*) FROM securities WHERE data IS NOT NULL AND data != ''").Scan(&dataCount)
	require.NoError(t, err)
	assert.Equal(t, successCount, dataCount, "All synced securities should have data populated")

	t.Logf("E2E Test Summary: Successfully synced %d/%d securities", successCount, len(isins))

	// Test batch sync again (simulates work processor execution of security:metadata:batch)
	// Reset last_synced to test second execution
	_, err = db.Conn.Exec("UPDATE securities SET last_synced = NULL")
	require.NoError(t, err)

	// Execute batch sync again (in production, this runs via security:metadata:batch work type every 24h)
	isins = metadataSyncService.GetAllActiveISINs()
	successCount2, err := metadataSyncService.SyncMetadataBatch(isins)
	assert.NoError(t, err, "Batch sync should execute without error")

	// Verify second execution populated data
	err = db.Conn.QueryRow("SELECT COUNT(*) FROM securities WHERE last_synced IS NOT NULL").Scan(&syncedCount)
	require.NoError(t, err)
	assert.Greater(t, syncedCount, 0, "Second batch sync should have synced securities")
	assert.Equal(t, successCount2, syncedCount, "Synced count should match second batch sync")

	t.Log("E2E test completed successfully - full stack verified")
}

// TestMetadataBatchSync_ExcludesIndices verifies that indices are excluded from batch sync
func TestMetadataBatchSync_ExcludesIndices(t *testing.T) {
	apiKey := os.Getenv("TRADERNET_API_KEY")
	apiSecret := os.Getenv("TRADERNET_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping E2E test: TRADERNET_API_KEY and TRADERNET_API_SECRET not set")
	}

	// Setup temporary database
	tmpDir, err := os.MkdirTemp("", "sentinel-e2e-indices-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "universe_test.db")

	// Initialize database
	db, err := database.New(dbPath, database.UniverseProfile, zerolog.New(os.Stderr).Level(zerolog.InfoLevel))
	require.NoError(t, err)
	defer db.Close()

	// Run migrations
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS securities (
			isin TEXT PRIMARY KEY,
			symbol TEXT NOT NULL,
			name TEXT,
			product_type TEXT,
			active INTEGER DEFAULT 1,
			data TEXT,
			last_synced INTEGER,
			created_at INTEGER DEFAULT (strftime('%s', 'now')),
			updated_at INTEGER DEFAULT (strftime('%s', 'now'))
		)`,
	}

	for _, migration := range migrations {
		_, err := db.Conn.Exec(migration)
		require.NoError(t, err)
	}

	// Insert test securities including indices
	testData := []struct {
		isin        string
		symbol      string
		name        string
		productType string
	}{
		// Regular securities
		{"US0378331005", "AAPL.US", "Apple Inc", "STOCK"},
		{"US5949181045", "MSFT.US", "Microsoft Corp", "STOCK"},
		// Indices (should be excluded)
		{"SP500INDEX", "SP500.IDX", "S&P 500 Index", "INDEX"},
		{"NASDAQINDEX", "NASDAQ.IDX", "NASDAQ Composite", "INDEX"},
	}

	for _, sec := range testData {
		_, err := db.Conn.Exec(
			`INSERT INTO securities (isin, symbol, name, product_type, active)
			 VALUES (?, ?, ?, ?, 1)`,
			sec.isin, sec.symbol, sec.name, sec.productType,
		)
		require.NoError(t, err)
	}

	// Initialize components
	log := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	brokerClient := tradernet.NewClient(apiKey, apiSecret, log)
	securityRepo := universe.NewSecurityRepository(db, log)
	metadataSyncService := universe.NewMetadataSyncService(securityRepo, brokerClient, log)

	// Get all active ISINs (should exclude indices)
	isins := metadataSyncService.GetAllActiveISINs()

	// Verify indices are excluded
	assert.Equal(t, 2, len(isins), "Should only get non-index securities")

	// Verify no .IDX symbols in ISINs
	for _, isin := range isins {
		security, err := securityRepo.GetByISIN(isin)
		require.NoError(t, err)
		assert.NotContains(t, security.Symbol, ".IDX", "ISIN list should not contain index symbols")
		assert.NotEqual(t, "INDEX", security.ProductType, "ISIN list should not contain indices")
	}

	t.Log("Index exclusion verified - only non-index securities included in batch")
}
