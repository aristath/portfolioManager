package events

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestBus_SubscribeAndEmit(t *testing.T) {
	bus := NewBus(zerolog.Nop())

	var receivedEvent *Event
	var receivedData map[string]interface{}

	handler := func(event *Event) {
		receivedEvent = event
		receivedData = event.Data
	}

	bus.Subscribe(PortfolioChanged, handler)

	data := map[string]interface{}{
		"portfolio_hash": "abc123",
		"change_type":    "position_added",
	}

	bus.Emit(PortfolioChanged, "portfolio", data)

	// Give handler time to execute
	time.Sleep(10 * time.Millisecond)

	assert.NotNil(t, receivedEvent)
	assert.Equal(t, PortfolioChanged, receivedEvent.Type)
	assert.Equal(t, "portfolio", receivedEvent.Module)
	assert.Equal(t, "abc123", receivedData["portfolio_hash"])
	assert.Equal(t, "position_added", receivedData["change_type"])
}

func TestBus_MultipleSubscribers(t *testing.T) {
	bus := NewBus(zerolog.Nop())

	var callCount1, callCount2 int

	handler1 := func(*Event) { callCount1++ }
	handler2 := func(*Event) { callCount2++ }

	bus.Subscribe(PortfolioChanged, handler1)
	bus.Subscribe(PortfolioChanged, handler2)

	bus.Emit(PortfolioChanged, "test", map[string]interface{}{})

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, callCount1)
	assert.Equal(t, 1, callCount2)
}

func TestBus_NoSubscribers(t *testing.T) {
	bus := NewBus(zerolog.Nop())

	// Should not panic
	bus.Emit(PortfolioChanged, "test", map[string]interface{}{})
}

func TestBus_DifferentEventTypes(t *testing.T) {
	bus := NewBus(zerolog.Nop())

	var portfolioCount, priceCount int

	bus.Subscribe(PortfolioChanged, func(*Event) { portfolioCount++ })
	bus.Subscribe(PriceUpdated, func(*Event) { priceCount++ })

	bus.Emit(PortfolioChanged, "test", map[string]interface{}{})
	bus.Emit(PriceUpdated, "test", map[string]interface{}{})

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, portfolioCount)
	assert.Equal(t, 1, priceCount)
}
