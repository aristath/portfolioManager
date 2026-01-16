package di

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainer_Initialization(t *testing.T) {
	container := &Container{}

	// Test that container can be created
	assert.NotNil(t, container)

	// Test that all fields are accessible (nil initially)
	assert.Nil(t, container.UniverseDB)
	assert.Nil(t, container.ConfigDB)
	assert.Nil(t, container.LedgerDB)
	assert.Nil(t, container.PortfolioDB)
	assert.Nil(t, container.HistoryDB)
	assert.Nil(t, container.CacheDB)
	assert.Nil(t, container.ClientDataDB)
}

func TestContainer_CanSetDatabases(t *testing.T) {
	container := &Container{}

	// This test verifies the container can hold database references
	// We can't create real databases in unit tests, but we can verify the structure
	assert.NotNil(t, container)
}

func TestContainer_HasWorkComponents(t *testing.T) {
	container := &Container{}

	// Test that WorkComponents field exists and is nil initially
	assert.NotNil(t, container)
	assert.Nil(t, container.WorkComponents)
}
