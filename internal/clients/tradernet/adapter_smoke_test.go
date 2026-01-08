package tradernet

import (
	"testing"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/rs/zerolog"
)

// TestTradernetBrokerAdapterCompiles verifies the adapter compiles and implements domain.BrokerClient
func TestTradernetBrokerAdapterCompiles(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	adapter := NewTradernetBrokerAdapter("test-key", "test-secret", log)

	// Type assertion to verify it implements domain.BrokerClient
	var _ domain.BrokerClient = adapter

	// Verify adapter is not nil
	if adapter == nil {
		t.Fatal("adapter should not be nil")
	}

	// Verify adapter has a client
	if adapter.client == nil {
		t.Fatal("adapter.client should not be nil")
	}
}
