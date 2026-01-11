package queue

import (
	"testing"
	"time"

	"github.com/aristath/sentinel/internal/events"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterListeners(t *testing.T) {
	bus := events.NewBus(zerolog.Nop())
	queue := NewMemoryQueue()
	history := NewHistory(nil) // No DB for this test
	manager := NewManager(queue, history)
	registry := NewRegistry()

	RegisterListeners(bus, manager, registry, zerolog.Nop())

	// Emit state changed event (triggers planner_batch)
	bus.Emit(events.StateChanged, "test", map[string]interface{}{
		"old_hash": "abc123",
		"new_hash": "def456",
	})

	// Give listener time to process
	time.Sleep(50 * time.Millisecond)

	// Should have enqueued planner_batch job
	assert.Equal(t, 1, manager.Size())

	job, err := manager.Dequeue()
	require.NoError(t, err)
	assert.Equal(t, JobTypePlannerBatch, job.Type)
	assert.Equal(t, PriorityCritical, job.Priority)
}

func TestListeners_MultipleEvents(t *testing.T) {
	bus := events.NewBus(zerolog.Nop())
	queue := NewMemoryQueue()
	history := NewHistory(nil)
	manager := NewManager(queue, history)
	registry := NewRegistry()

	RegisterListeners(bus, manager, registry, zerolog.Nop())

	// Emit multiple events
	bus.Emit(events.StateChanged, "test", map[string]interface{}{"old_hash": "a", "new_hash": "b"})
	bus.Emit(events.PriceUpdated, "test", map[string]interface{}{})
	bus.Emit(events.RecommendationsReady, "test", map[string]interface{}{})

	time.Sleep(50 * time.Millisecond)

	// Should have enqueued multiple jobs
	assert.GreaterOrEqual(t, manager.Size(), 2)
}
