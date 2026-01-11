// Package server provides the HTTP server and routing for Sentinel.
package server

import (
	"time"

	"github.com/aristath/sentinel/internal/events"
	"github.com/rs/zerolog"
)

// StatusMonitor periodically checks system statuses and emits events on changes
type StatusMonitor struct {
	eventManager   *events.Manager
	systemHandlers *SystemHandlers
	log            zerolog.Logger

	// Track previous states
	lastSystemStatus    *SystemStatusResponse
	lastTradernetStatus bool

	// Dependency injection for testing
	getSystemStatus func() (SystemStatusResponse, error)
}

// NewStatusMonitor creates a new status monitor
func NewStatusMonitor(
	eventManager *events.Manager,
	systemHandlers *SystemHandlers,
	log zerolog.Logger,
) *StatusMonitor {
	return &StatusMonitor{
		eventManager:   eventManager,
		systemHandlers: systemHandlers,
		log:            log.With().Str("component", "status_monitor").Logger(),
		getSystemStatus: func() (SystemStatusResponse, error) {
			if systemHandlers == nil {
				return SystemStatusResponse{}, nil
			}
			return systemHandlers.GetSystemStatusSnapshot()
		},
	}
}

// Start begins periodic status monitoring
func (m *StatusMonitor) Start(interval time.Duration) {
	go m.monitor(interval)
}

// monitor runs the periodic monitoring loop
func (m *StatusMonitor) monitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Do initial check
	m.checkStatuses()

	for range ticker.C {
		m.checkStatuses()
	}
}

// checkStatuses checks all monitored statuses and emits events on changes
func (m *StatusMonitor) checkStatuses() {
	// Check system status (simplified - just check if active positions count changed)
	// Full system status check would be expensive, so we do minimal checks
	m.checkSystemStatus()

	// Check tradernet status
	m.checkTradernetStatus()
}

// checkSystemStatus checks if system status has changed
func (m *StatusMonitor) checkSystemStatus() {
	if m.eventManager == nil || m.getSystemStatus == nil {
		return
	}

	status, err := m.getSystemStatus()
	if err != nil {
		m.log.Warn().Err(err).Msg("Unable to get system status snapshot")
		return
	}

	if m.lastSystemStatus != nil && systemStatusEqual(*m.lastSystemStatus, status) {
		return
	}

	m.eventManager.Emit(events.SystemStatusChanged, "status_monitor", map[string]interface{}{
		"status":             status.Status,
		"cash_balance_eur":   status.CashBalanceEUR,
		"cash_balance":       status.CashBalance,
		"cash_balance_total": status.CashBalanceTotal,
		"security_count":     status.SecurityCount,
		"position_count":     status.PositionCount,
		"active_positions":   status.ActivePositions,
		"last_sync":          status.LastSync,
		"universe_active":    status.UniverseActive,
		"timestamp":          time.Now().Format(time.RFC3339),
	})

	m.lastSystemStatus = &status
}

// checkTradernetStatus checks if tradernet connection status has changed
func (m *StatusMonitor) checkTradernetStatus() {
	if m.systemHandlers == nil || m.systemHandlers.brokerClient == nil {
		return
	}

	// Check connection status
	connected := m.systemHandlers.brokerClient.IsConnected()

	// Emit event if status changed
	if connected != m.lastTradernetStatus {
		if m.eventManager != nil {
			m.eventManager.Emit(events.TradernetStatusChanged, "status_monitor", map[string]interface{}{
				"connected": connected,
				"timestamp": time.Now().Format(time.RFC3339),
			})
		}
		m.lastTradernetStatus = connected
	}
}

func systemStatusEqual(a, b SystemStatusResponse) bool {
	return a.Status == b.Status &&
		a.CashBalanceEUR == b.CashBalanceEUR &&
		a.CashBalanceTotal == b.CashBalanceTotal &&
		a.CashBalance == b.CashBalance &&
		a.SecurityCount == b.SecurityCount &&
		a.PositionCount == b.PositionCount &&
		a.ActivePositions == b.ActivePositions &&
		a.LastSync == b.LastSync &&
		a.UniverseActive == b.UniverseActive
}
