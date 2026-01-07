package events

import (
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestBus_SubscribeAndEmit(t *testing.T) {
	bus := NewBus(zerolog.Nop())

	var receivedEvent *Event
	var receivedData map[string]interface{}
	var mu sync.Mutex

	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(event *Event) {
		mu.Lock()
		receivedEvent = event
		receivedData = event.Data
		mu.Unlock()
		wg.Done()
	}

	bus.Subscribe(PortfolioChanged, handler)

	data := map[string]interface{}{
		"portfolio_hash": "abc123",
		"change_type":    "position_added",
	}

	bus.Emit(PortfolioChanged, "portfolio", data)

	wg.Wait()

	mu.Lock()
	assert.NotNil(t, receivedEvent)
	assert.Equal(t, PortfolioChanged, receivedEvent.Type)
	assert.Equal(t, "portfolio", receivedEvent.Module)
	assert.Equal(t, "abc123", receivedData["portfolio_hash"])
	assert.Equal(t, "position_added", receivedData["change_type"])
	mu.Unlock()
}

func TestBus_MultipleSubscribers(t *testing.T) {
	bus := NewBus(zerolog.Nop())

	var callCount1, callCount2 int
	var mu1, mu2 sync.Mutex

	var wg sync.WaitGroup
	wg.Add(2)

	handler1 := func(*Event) {
		mu1.Lock()
		callCount1++
		mu1.Unlock()
		wg.Done()
	}
	handler2 := func(*Event) {
		mu2.Lock()
		callCount2++
		mu2.Unlock()
		wg.Done()
	}

	bus.Subscribe(PortfolioChanged, handler1)
	bus.Subscribe(PortfolioChanged, handler2)

	bus.Emit(PortfolioChanged, "test", map[string]interface{}{})

	wg.Wait()

	mu1.Lock()
	mu2.Lock()
	assert.Equal(t, 1, callCount1)
	assert.Equal(t, 1, callCount2)
	mu2.Unlock()
	mu1.Unlock()
}

func TestBus_NoSubscribers(t *testing.T) {
	bus := NewBus(zerolog.Nop())

	// Should not panic
	bus.Emit(PortfolioChanged, "test", map[string]interface{}{})
}

func TestBus_DifferentEventTypes(t *testing.T) {
	bus := NewBus(zerolog.Nop())

	var portfolioCount, priceCount int
	var mu sync.Mutex

	var wg sync.WaitGroup
	wg.Add(2)

	bus.Subscribe(PortfolioChanged, func(*Event) {
		mu.Lock()
		portfolioCount++
		mu.Unlock()
		wg.Done()
	})
	bus.Subscribe(PriceUpdated, func(*Event) {
		mu.Lock()
		priceCount++
		mu.Unlock()
		wg.Done()
	})

	bus.Emit(PortfolioChanged, "test", map[string]interface{}{})
	bus.Emit(PriceUpdated, "test", map[string]interface{}{})

	wg.Wait()

	mu.Lock()
	assert.Equal(t, 1, portfolioCount)
	assert.Equal(t, 1, priceCount)
	mu.Unlock()
}
