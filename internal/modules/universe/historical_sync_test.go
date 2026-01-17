package universe

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestHistoricalSyncServiceCreation(t *testing.T) {
	log := zerolog.Nop()

	service := NewHistoricalSyncService(
		nil, // brokerClient
		nil, // securityRepo
		nil, // historyDB
		log,
	)

	assert.NotNil(t, service)
}

func TestHistoricalSyncService_SyncWithoutClients(t *testing.T) {
	log := zerolog.Nop()

	service := NewHistoricalSyncService(
		nil, // brokerClient
		nil, // securityRepo
		nil, // historyDB
		log,
	)

	// Should panic with nil security repo (nil pointer dereference)
	assert.Panics(t, func() {
		_ = service.SyncHistoricalPrices("AAPL.US")
	})
}

// Note: Full integration tests with real Tradernet and database
// should be in integration test suite. These are unit tests focusing
// on service logic without external dependencies.
