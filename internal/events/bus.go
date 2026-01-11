package events

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// EventHandler is a function that handles events
type EventHandler func(*Event)

// Subscription represents a registered event handler.
// It is used to unsubscribe when a consumer disconnects.
type Subscription struct {
	eventType EventType
	id        uint64
}

// Bus provides pub/sub event functionality
type Bus struct {
	subscribers map[EventType]map[uint64]EventHandler
	nextID      uint64
	mu          sync.RWMutex
	log         zerolog.Logger
}

// NewBus creates a new event bus
func NewBus(log zerolog.Logger) *Bus {
	return &Bus{
		subscribers: make(map[EventType]map[uint64]EventHandler),
		log:         log.With().Str("service", "events").Logger(),
	}
}

// Subscribe registers a handler for an event type
func (b *Bus) Subscribe(eventType EventType, handler EventHandler) Subscription {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.nextID++
	id := b.nextID

	if _, ok := b.subscribers[eventType]; !ok {
		b.subscribers[eventType] = make(map[uint64]EventHandler)
	}

	b.subscribers[eventType][id] = handler

	return Subscription{
		eventType: eventType,
		id:        id,
	}
}

// Unsubscribe removes a previously registered handler.
// It is safe to call multiple times.
func (b *Bus) Unsubscribe(sub Subscription) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if handlers, ok := b.subscribers[sub.eventType]; ok {
		delete(handlers, sub.id)
		if len(handlers) == 0 {
			delete(b.subscribers, sub.eventType)
		}
	}
}

// Emit publishes an event to all subscribers
func (b *Bus) Emit(eventType EventType, module string, data map[string]interface{}) {
	event := &Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
		Module:    module,
	}

	// Snapshot handlers to avoid holding the lock while invoking callbacks
	b.mu.RLock()
	var handlers []EventHandler
	if registered := b.subscribers[eventType]; len(registered) > 0 {
		handlers = make([]EventHandler, 0, len(registered))
		for _, handler := range registered {
			handlers = append(handlers, handler)
		}
	}
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
