// Package events provides event management functionality.
package events

import (
	"encoding/json"
	"time"

	"github.com/rs/zerolog"
)

// Event represents a system event
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Module    string                 `json:"module"`
}

// Manager handles event emission and logging
type Manager struct {
	bus *Bus
	log zerolog.Logger
}

// NewManager creates a new event manager
func NewManager(bus *Bus, log zerolog.Logger) *Manager {
	return &Manager{
		bus: bus,
		log: log.With().Str("service", "events").Logger(),
	}
}

// Emit emits an event to the bus and logs it
func (m *Manager) Emit(eventType EventType, module string, data map[string]interface{}) {
	event := Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
		Module:    module,
	}

	// Publish to bus
	m.bus.Emit(eventType, module, data)

	// Log event
	eventJSON, _ := json.Marshal(event)
	m.log.Info().
		Str("event_type", string(eventType)).
		Str("module", module).
		RawJSON("event", eventJSON).
		Msg("Event emitted")
}

// EmitError emits an error event
func (m *Manager) EmitError(module string, err error, context map[string]interface{}) {
	data := map[string]interface{}{
		"error":   err.Error(),
		"context": context,
	}
	m.Emit(ErrorOccurred, module, data)
}
