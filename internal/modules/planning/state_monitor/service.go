// Package state_monitor provides state change monitoring and event emission
package state_monitor

import (
	"sync"
	"time"

	"github.com/aristath/sentinel/internal/events"
	"github.com/aristath/sentinel/internal/modules/planning/hash"
	"github.com/rs/zerolog"
)

// StateMonitor monitors unified state hash and emits events on changes
type StateMonitor struct {
	stateHashSvc *hash.StateHashService
	eventManager *events.Manager
	log          zerolog.Logger

	// State tracking
	lastHash  string
	hashMutex sync.RWMutex

	// Lifecycle management
	ticker    *time.Ticker
	stopChan  chan struct{}
	stopOnce  sync.Once
	startOnce sync.Once
}

// NewStateMonitor creates a new state monitor
func NewStateMonitor(
	stateHashSvc *hash.StateHashService,
	eventManager *events.Manager,
	log zerolog.Logger,
) *StateMonitor {
	return &StateMonitor{
		stateHashSvc: stateHashSvc,
		eventManager: eventManager,
		log:          log.With().Str("component", "state_monitor").Logger(),
	}
}

// Start begins monitoring (runs every minute)
func (m *StateMonitor) Start() {
	m.startOnce.Do(func() {
		m.ticker = time.NewTicker(1 * time.Minute)
		m.stopChan = make(chan struct{})

		m.log.Info().Msg("State monitor started (checking every minute)")

		// Run immediately on start
		m.checkAndEmit()

		// Start periodic checks
		go m.run()
	})
}

// Stop stops monitoring
func (m *StateMonitor) Stop() {
	m.stopOnce.Do(func() {
		if m.ticker != nil {
			m.ticker.Stop()
		}
		if m.stopChan != nil {
			close(m.stopChan)
		}
		m.log.Info().Msg("State monitor stopped")
	})
}

// run executes the periodic check loop
func (m *StateMonitor) run() {
	for {
		select {
		case <-m.ticker.C:
			m.checkAndEmit()
		case <-m.stopChan:
			return
		}
	}
}

// checkAndEmit checks if state changed and emits event if so
func (m *StateMonitor) checkAndEmit() {
	currentHash, err := m.stateHashSvc.CalculateCurrentHash()
	if err != nil {
		m.log.Warn().Err(err).Msg("Failed to calculate state hash")
		return
	}

	m.hashMutex.RLock()
	lastHash := m.lastHash
	m.hashMutex.RUnlock()

	if lastHash == "" || currentHash != lastHash {
		if lastHash == "" {
			m.log.Info().
				Str("new_hash", currentHash).
				Msg("Initial state detected - emitting StateChanged event to bootstrap recommendations")
		} else {
			m.log.Info().
				Str("old_hash", lastHash).
				Str("new_hash", currentHash).
				Msg("State changed - emitting StateChanged event")
		}

		// Emit StateChanged event
		if m.eventManager != nil {
			m.eventManager.Emit(events.StateChanged, "state_monitor", map[string]interface{}{
				"old_hash": lastHash,
				"new_hash": currentHash,
			})
		}

		// Update hash
		m.hashMutex.Lock()
		m.lastHash = currentHash
		m.hashMutex.Unlock()
	}
}
