package events

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// EventHandler is a function that handles events
type EventHandler func(*Event)

// Bus provides pub/sub event functionality
type Bus struct {
	subscribers map[EventType][]EventHandler
	mu          sync.RWMutex
	log         zerolog.Logger
}

// NewBus creates a new event bus
func NewBus(log zerolog.Logger) *Bus {
	return &Bus{
		subscribers: make(map[EventType][]EventHandler),
		log:         log.With().Str("service", "events").Logger(),
	}
}

// Subscribe registers a handler for an event type
func (b *Bus) Subscribe(eventType EventType, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
}

// Emit publishes an event to all subscribers
func (b *Bus) Emit(eventType EventType, module string, data map[string]interface{}) {
	event := &Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
		Module:    module,
	}

	b.mu.RLock()
	handlers := b.subscribers[eventType]
	b.mu.RUnlock()

	// Execute handlers asynchronously
	for _, handler := range handlers {
		go handler(event)
	}

	b.log.Debug().
		Str("event_type", string(eventType)).
		Str("module", module).
		Int("subscribers", len(handlers)).
		Msg("Event emitted")
}
