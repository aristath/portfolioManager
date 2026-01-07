package di

import (
	"testing"

	"github.com/aristath/sentinel/internal/config"
	"github.com/aristath/sentinel/internal/modules/display"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWire(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		DataDir:            tmpDir,
		TradernetAPIKey:    "test-key",
		TradernetAPISecret: "test-secret",
	}
	log := zerolog.Nop()

	// Create display manager
	displayManager := display.NewStateManager(log)

	// Wire everything
	container, jobs, err := Wire(cfg, log, displayManager)
	require.NoError(t, err)
	require.NotNil(t, container)
	require.NotNil(t, jobs)

	// Verify container is fully populated
	assert.NotNil(t, container.UniverseDB)
	assert.NotNil(t, container.PortfolioService)
	assert.NotNil(t, container.TradingService)
	assert.NotNil(t, container.CashManager)

	// Verify jobs are registered
	assert.NotNil(t, jobs.HealthCheck)
	assert.NotNil(t, jobs.SyncCycle)
	assert.NotNil(t, jobs.DividendReinvest)

	// Cleanup
	container.UniverseDB.Close()
	container.ConfigDB.Close()
	container.LedgerDB.Close()
	container.PortfolioDB.Close()
	container.AgentsDB.Close()
	container.HistoryDB.Close()
	container.CacheDB.Close()
}
