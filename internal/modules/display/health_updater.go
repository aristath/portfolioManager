package display

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// HealthUpdater periodically calculates and sends health scores to the display
type HealthUpdater struct {
	healthCalc     *HealthCalculator
	displayURL     string
	updateInterval time.Duration
	httpClient     *http.Client
	log            zerolog.Logger
	stopCh         chan struct{}
	running        bool
	mu             sync.RWMutex
}

// NewHealthUpdater creates a new health updater service
func NewHealthUpdater(
	healthCalc *HealthCalculator,
	displayURL string,
	updateInterval time.Duration,
	log zerolog.Logger,
) *HealthUpdater {
	return &HealthUpdater{
		healthCalc:     healthCalc,
		displayURL:     displayURL,
		updateInterval: updateInterval,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		log:    log.With().Str("service", "health_updater").Logger(),
		stopCh: make(chan struct{}),
	}
}

// Start starts the health updater background service
func (h *HealthUpdater) Start() {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		h.log.Warn().Msg("Health updater already running")
		return
	}
	h.running = true
	h.stopCh = make(chan struct{})
	h.mu.Unlock()

	h.log.Info().
		Dur("interval", h.updateInterval).
		Msg("Starting health updater service")

	go func() {
		// Send initial update immediately
		h.sendHealthUpdate()

		ticker := time.NewTicker(h.updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				h.sendHealthUpdate()
			case <-h.stopCh:
				h.log.Info().Msg("Health updater stopped")
				return
			}
		}
	}()
}

// Stop stops the health updater background service
func (h *HealthUpdater) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return
	}

	h.log.Info().Msg("Stopping health updater service")
	close(h.stopCh)
	h.running = false
}

// IsRunning returns whether the updater is currently running
func (h *HealthUpdater) IsRunning() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.running
}

// sendHealthUpdate calculates and sends health update to display
func (h *HealthUpdater) sendHealthUpdate() {
	h.log.Debug().Msg("Calculating portfolio health")

	// Calculate health for all holdings
	healthData, err := h.healthCalc.CalculateAllHealth()
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to calculate health")
		return
	}

	if len(healthData.Securities) == 0 {
		h.log.Warn().Msg("No securities to display")
		return
	}

	// Send to Python app
	jsonData, err := json.Marshal(healthData)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to marshal health data")
		return
	}

	url := fmt.Sprintf("%s/portfolio-health", h.displayURL)
	resp, err := h.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		h.log.Debug().Err(err).Msg("Failed to send health update to display (display may be offline)")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		h.log.Warn().
			Int("status", resp.StatusCode).
			Msg("Display returned error status")
		return
	}

	h.log.Info().
		Int("securities", len(healthData.Securities)).
		Msg("Sent health update to display")
}

// TriggerUpdate triggers an immediate health update (for testing/manual refresh)
func (h *HealthUpdater) TriggerUpdate() {
	h.log.Info().Msg("Manual health update triggered")
	go h.sendHealthUpdate()
}
