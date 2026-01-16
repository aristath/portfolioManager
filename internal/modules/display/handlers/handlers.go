// Package handlers provides HTTP handlers for LED display management.
package handlers

import (
	"github.com/aristath/sentinel/internal/modules/display"
	"github.com/aristath/sentinel/internal/modules/settings"
)

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
)

// Handlers provides HTTP handlers for display module
type Handlers struct {
	stateManager    *display.StateManager
	modeManager     *display.ModeManager
	healthCalc      *display.HealthCalculator
	healthUpdater   *display.HealthUpdater
	settingsService *settings.Service
	log             zerolog.Logger
}

// NewHandlers creates a new display handlers instance
func NewHandlers(
	stateManager *display.StateManager,
	modeManager *display.ModeManager,
	healthCalc *display.HealthCalculator,
	healthUpdater *display.HealthUpdater,
	settingsService *settings.Service,
	log zerolog.Logger,
) *Handlers {
	return &Handlers{
		stateManager:    stateManager,
		modeManager:     modeManager,
		healthCalc:      healthCalc,
		healthUpdater:   healthUpdater,
		settingsService: settingsService,
		log:             log.With().Str("module", "display_handlers").Logger(),
	}
}

// HandleGetState handles GET /api/display/state
// Returns current display state (text, LED3, LED4)
func (h *Handlers) HandleGetState(w http.ResponseWriter, r *http.Request) {
	state := h.stateManager.GetState()

	h.writeJSON(w, state)
}

// SetTextRequest represents the request to set display text
type SetTextRequest struct {
	Text string `json:"text"`
}

// HandleSetText handles POST /api/display/text
// Sets the display text
func (h *Handlers) HandleSetText(w http.ResponseWriter, r *http.Request) {
	var req SetTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error().Err(err).Msg("Failed to decode set text request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.stateManager.SetText(req.Text)

	h.writeJSON(w, map[string]string{"status": "ok", "text": req.Text})
}

// SetLEDRequest represents the request to set LED color
type SetLEDRequest struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

// HandleSetLED3 handles POST /api/display/led3
// Sets LED 3 RGB color
func (h *Handlers) HandleSetLED3(w http.ResponseWriter, r *http.Request) {
	var req SetLEDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error().Err(err).Msg("Failed to decode set LED3 request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.stateManager.SetLED3(req.R, req.G, req.B)

	h.writeJSON(w, map[string]string{"status": "ok"})
}

// HandleSetLED4 handles POST /api/display/led4
// Sets LED 4 RGB color
func (h *Handlers) HandleSetLED4(w http.ResponseWriter, r *http.Request) {
	var req SetLEDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error().Err(err).Msg("Failed to decode set LED4 request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.stateManager.SetLED4(req.R, req.G, req.B)

	h.writeJSON(w, map[string]string{"status": "ok"})
}

// SetModeRequest represents the request to set display mode
type SetModeRequest struct {
	Mode string `json:"mode"`
}

// HandleGetMode handles GET /api/display/mode
// Returns current display mode
func (h *Handlers) HandleGetMode(w http.ResponseWriter, r *http.Request) {
	mode := h.modeManager.GetMode()
	h.writeJSON(w, map[string]string{"mode": string(mode)})
}

// HandleSetMode handles POST /api/display/mode
// Sets the display mode (TEXT, HEALTH, STATS)
func (h *Handlers) HandleSetMode(w http.ResponseWriter, r *http.Request) {
	var req SetModeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error().Err(err).Msg("Failed to decode set mode request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	mode := display.DisplayMode(req.Mode)

	// Validate mode
	if mode != display.ModeText && mode != display.ModeHealth && mode != display.ModeStats {
		h.log.Warn().Str("mode", req.Mode).Msg("Invalid display mode")
		http.Error(w, "Invalid mode. Must be TEXT, HEALTH, or STATS", http.StatusBadRequest)
		return
	}

	if err := h.modeManager.SetMode(mode); err != nil {
		h.log.Error().Err(err).Str("mode", req.Mode).Msg("Failed to set display mode")
		http.Error(w, "Failed to set mode", http.StatusInternalServerError)
		return
	}

	// Persist the mode change to settings database
	if _, err := h.settingsService.Set("display_mode", req.Mode); err != nil {
		h.log.Warn().Err(err).Str("mode", req.Mode).Msg("Failed to persist display mode to settings")
		// Don't fail the request - the mode was already applied successfully
	} else {
		h.log.Debug().Str("mode", req.Mode).Msg("Display mode persisted to settings")
	}

	h.log.Info().Str("mode", req.Mode).Msg("Display mode changed")
	h.writeJSON(w, map[string]string{"status": "ok", "mode": req.Mode})
}

// HandleGetPortfolioHealth handles GET /api/display/portfolio-health/preview
// Returns current portfolio health scores (for debugging)
func (h *Handlers) HandleGetPortfolioHealth(w http.ResponseWriter, r *http.Request) {
	healthData, err := h.healthCalc.CalculateAllHealth()
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to calculate portfolio health")
		http.Error(w, "Failed to calculate health", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, healthData)
}

// HandleTriggerHealthUpdate handles POST /api/display/portfolio-health/trigger
// Manually triggers a health update (for testing)
func (h *Handlers) HandleTriggerHealthUpdate(w http.ResponseWriter, r *http.Request) {
	h.healthUpdater.TriggerUpdate()
	h.writeJSON(w, map[string]string{"status": "ok", "message": "Health update triggered"})
}

// writeJSON writes a JSON response
func (h *Handlers) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Error().Err(err).Msg("Failed to encode JSON response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
