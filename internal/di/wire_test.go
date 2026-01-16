package di

import (
	"testing"
	"time"

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

	// Wire everything (settings loading and deployment manager creation happen inside)
	container, err := Wire(cfg, log, displayManager)
	require.NoError(t, err)
	require.NotNil(t, container)

	// Verify container is fully populated
	assert.NotNil(t, container.UniverseDB)
	assert.NotNil(t, container.PortfolioService)
	assert.NotNil(t, container.TradingService)
	assert.NotNil(t, container.CashManager)

	// Verify work processor is initialized
	assert.NotNil(t, container.WorkComponents)
	assert.NotNil(t, container.WorkComponents.Processor)
	assert.NotNil(t, container.WorkComponents.Registry)

	// Cleanup - stop services and close databases
	// Note: Don't stop the work processor since Run() was never called in this test
	t.Cleanup(func() {
		if container != nil {
			// Stop WebSocket client if running
			if container.MarketStatusWS != nil {
				_ = container.MarketStatusWS.Stop()
			}

			// Give goroutines time to stop before closing databases
			time.Sleep(100 * time.Millisecond)

			// Close databases
			container.UniverseDB.Close()
			container.ConfigDB.Close()
			container.LedgerDB.Close()
			container.PortfolioDB.Close()
			container.HistoryDB.Close()
			container.CacheDB.Close()
			container.ClientDataDB.Close()
		}
	})
}
