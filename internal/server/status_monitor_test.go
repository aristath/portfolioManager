package server

import (
	"testing"
	"time"

	"github.com/aristath/sentinel/internal/events"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestStatusMonitorEmitsOnlyOnChange(t *testing.T) {
	log := zerolog.Nop()
	bus := events.NewBus(log)
	manager := events.NewManager(bus, log)

	monitor := &StatusMonitor{
		eventManager: manager,
		log:          log,
		getSystemStatus: func() (SystemStatusResponse, error) {
			return SystemStatusResponse{
				Status:           "healthy",
				CashBalanceEUR:   10,
				CashBalanceTotal: 10,
				CashBalance:      10,
				SecurityCount:    1,
				PositionCount:    1,
				ActivePositions:  1,
				UniverseActive:   1,
				LastSync:         "2024-01-01 10:00",
			}, nil
		},
	}

	eventsChan := make(chan events.Event, 5)
	_ = bus.Subscribe(events.SystemStatusChanged, func(event *events.Event) {
		eventsChan <- *event
	})

	monitor.checkSystemStatus()

	select {
	case <-eventsChan:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected first system status event")
	}

	// Same snapshot should not emit again
	monitor.checkSystemStatus()

	select {
	case evt := <-eventsChan:
		t.Fatalf("unexpected extra event: %+v", evt)
	case <-time.After(100 * time.Millisecond):
	}

	// Change snapshot to trigger a new event
	monitor.getSystemStatus = func() (SystemStatusResponse, error) {
		return SystemStatusResponse{
			Status:           "healthy",
			CashBalanceEUR:   20,
			CashBalanceTotal: 20,
			CashBalance:      20,
			SecurityCount:    1,
			PositionCount:    1,
			ActivePositions:  1,
			UniverseActive:   1,
			LastSync:         "2024-01-01 10:00",
		}, nil
	}

	monitor.checkSystemStatus()

	select {
	case <-eventsChan:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected system status change event")
	}

	// Ensure last snapshot updated
	assert.NotNil(t, monitor.lastSystemStatus)
	assert.Equal(t, 20.0, monitor.lastSystemStatus.CashBalanceEUR)
}
