package server

import (
	"testing"

	"github.com/aristath/sentinel/internal/events"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestEnqueueEventDropsOldest(t *testing.T) {
	handler := &EventsStreamHandler{
		log: zerolog.Nop(),
	}

	eventChan := make(chan *events.Event, 2)

	event1 := &events.Event{Type: events.PriceUpdated}
	event2 := &events.Event{Type: events.ScoreUpdated}
	event3 := &events.Event{Type: events.PlanGenerated}

	handler.enqueueEvent(eventChan, event1)
	handler.enqueueEvent(eventChan, event2)
	handler.enqueueEvent(eventChan, event3)

	assert.Equal(t, 2, len(eventChan))

	first := <-eventChan
	second := <-eventChan

	assert.Equal(t, events.ScoreUpdated, first.Type)
	assert.Equal(t, events.PlanGenerated, second.Type)
}
