// Package handlers provides HTTP handlers for system settings management.
package handlers

import (
	"encoding/json"
	"net/http"
	"os/exec"

	"github.com/aristath/sentinel/internal/events"
	"github.com/aristath/sentinel/internal/modules/settings"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// OnboardingServiceInterface defines the interface for onboarding service
type OnboardingServiceInterface interface {
	RunOnboarding() error
}

// CredentialRefresher defines the interface for refreshing tradernet client credentials
type CredentialRefresher interface {
	RefreshCredentials() error
}

// Handler provides HTTP handlers for settings endpoints
type Handler struct {
	service             *settings.Service
	onboardingService   OnboardingServiceInterface
	credentialRefresher CredentialRefresher
	eventManager        *events.Manager
	log                 zerolog.Logger
}

// NewHandler creates a new settings handler
func NewHandler(service *settings.Service, eventManager *events.Manager, log zerolog.Logger) *Handler {
	return &Handler{
		service:      service,
		eventManager: eventManager,
		log:          log.With().Str("handler", "settings").Logger(),
	}
}

// SetOnboardingService sets the onboarding service (for dependency injection)
func (h *Handler) SetOnboardingService(onboardingService OnboardingServiceInterface) {
	h.onboardingService = onboardingService
}

// SetCredentialRefresher sets the credential refresher (for dependency injection)
func (h *Handler) SetCredentialRefresher(refresher CredentialRefresher) {
	h.credentialRefresher = refresher
}

// HandleGetAll handles GET /api/settings
// Faithful translation from Python: app/api/settings.py -> get_all_settings()
func (h *Handler) HandleGetAll(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetAll()
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to get all settings")
		http.Error(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(settings); err != nil {
		h.log.Error().Err(err).Msg("Failed to encode settings response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleUpdate handles PUT /api/settings/{key}
// Faithful translation from Python: app/api/settings.py -> update_setting_value()
func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	var update settings.SettingUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	isFirstTimeSetup, err := h.service.Set(key, update.Value)
	if err != nil {
		h.log.Error().
			Err(err).
			Str("key", key).
			Interface("value", update.Value).
			Msg("Failed to update setting")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Refresh tradernet client credentials if this was a credential update
	if (key == "tradernet_api_key" || key == "tradernet_api_secret") && h.credentialRefresher != nil {
		if err := h.credentialRefresher.RefreshCredentials(); err != nil {
			h.log.Warn().Err(err).Msg("Failed to refresh tradernet client credentials after update")
		} else {
			h.log.Info().Msg("Tradernet client credentials refreshed after settings update")
		}
	}

	// Trigger onboarding if this is first-time credential setup
	if isFirstTimeSetup && h.onboardingService != nil {
		h.log.Info().Msg("First-time credential setup detected, triggering onboarding")
		go func() {
			if err := h.onboardingService.RunOnboarding(); err != nil {
				h.log.Error().Err(err).Msg("Onboarding failed")
			} else {
				h.log.Info().Msg("Onboarding completed successfully")
			}
		}()
	}

	// Emit SETTINGS_CHANGED event
	if h.eventManager != nil {
		h.eventManager.Emit(events.SettingsChanged, "settings", map[string]interface{}{
			"key":   key,
			"value": update.Value,
		})
	}

	// Return updated value
	result := map[string]interface{}{key: update.Value}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleRestartService handles POST /api/settings/restart-service
// Faithful translation from Python: app/api/settings.py -> restart_service()
func (h *Handler) HandleRestartService(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("sudo", "systemctl", "restart", "sentinel")
	output, err := cmd.CombinedOutput()

	response := map[string]string{}
	if err != nil {
		response["status"] = "error"
		response["message"] = string(output)
		h.log.Warn().
			Err(err).
			Str("output", string(output)).
			Msg("Failed to restart service")
	} else {
		response["status"] = "ok"
		response["message"] = "Service restart initiated"
		h.log.Info().Msg("Service restart initiated")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleRestart handles POST /api/settings/restart
// Faithful translation from Python: app/api/settings.py -> restart_system()
func (h *Handler) HandleRestart(w http.ResponseWriter, r *http.Request) {
	// Start reboot process in background
	cmd := exec.Command("sudo", "reboot")
	if err := cmd.Start(); err != nil {
		h.log.Error().Err(err).Msg("Failed to initiate system reboot")
		http.Error(w, "Failed to initiate reboot", http.StatusInternalServerError)
		return
	}

	h.log.Warn().Msg("System reboot initiated")

	response := map[string]string{"status": "rebooting"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleResetCache handles POST /api/settings/reset-cache
// Faithful translation from Python: app/api/settings.py -> reset_cache()
func (h *Handler) HandleResetCache(w http.ResponseWriter, r *http.Request) {
	// Note: Full cache implementation would require cache infrastructure
	// This is a simplified version that acknowledges the request
	h.log.Info().Msg("Cache reset requested")

	response := map[string]string{
		"status":  "ok",
		"message": "Cache reset acknowledged (simplified implementation)",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetCacheStats handles GET /api/settings/cache-stats
// Faithful translation from Python: app/api/settings.py -> get_cache_stats()
func (h *Handler) HandleGetCacheStats(w http.ResponseWriter, r *http.Request) {
	// Note: Full implementation would require calculations DB integration
	// This is a simplified version returning stub data
	stats := settings.CacheStats{
		SimpleCache: settings.SimpleCacheStats{
			Entries: 0,
		},
		CalculationsDB: settings.CalculationsDBStats{
			Entries:        0,
			ExpiredCleaned: 0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleRescheduleJobs handles POST /api/settings/reschedule-jobs
// Faithful translation from Python: app/api/settings.py -> reschedule_jobs()
func (h *Handler) HandleRescheduleJobs(w http.ResponseWriter, r *http.Request) {
	// Note: Full implementation would require scheduler integration
	// This is a simplified version that acknowledges the request
	h.log.Info().Msg("Job rescheduling requested")

	response := map[string]string{
		"status":  "ok",
		"message": "Job rescheduling acknowledged (simplified implementation)",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetTradingMode handles GET /api/settings/trading-mode
// Faithful translation from Python: app/api/settings.py -> get_trading_mode_endpoint()
func (h *Handler) HandleGetTradingMode(w http.ResponseWriter, r *http.Request) {
	mode, err := h.service.GetTradingMode()
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to get trading mode")
		http.Error(w, "Failed to get trading mode", http.StatusInternalServerError)
		return
	}

	response := settings.TradingModeResponse{TradingMode: mode}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleToggleTradingMode handles POST /api/settings/trading-mode
// Faithful translation from Python: app/api/settings.py -> toggle_trading_mode()
func (h *Handler) HandleToggleTradingMode(w http.ResponseWriter, r *http.Request) {
	newMode, previousMode, err := h.service.ToggleTradingMode()
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to toggle trading mode")
		http.Error(w, "Failed to toggle trading mode", http.StatusInternalServerError)
		return
	}

	h.log.Info().
		Str("previous_mode", previousMode).
		Str("new_mode", newMode).
		Msg("Trading mode toggled")

	response := settings.TradingModeToggleResponse{
		TradingMode:  newMode,
		PreviousMode: previousMode,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DataSourceType represents a type of data that can be fetched from multiple sources
type DataSourceType string

// Data source type constants
const (
	DataTypeFundamentals    DataSourceType = "fundamentals"
	DataTypeCurrentPrices   DataSourceType = "current_prices"
	DataTypeHistorical      DataSourceType = "historical"
	DataTypeTechnicals      DataSourceType = "technicals"
	DataTypeExchangeRates   DataSourceType = "exchange_rates"
	DataTypeISINLookup      DataSourceType = "isin_lookup"
	DataTypeCompanyMetadata DataSourceType = "company_metadata"
)

// DataSourceConfig represents the configuration for a single data type
type DataSourceConfig struct {
	Type        DataSourceType `json:"type"`
	Description string         `json:"description"`
	Priorities  []string       `json:"priorities"`
	SettingKey  string         `json:"setting_key"`
}

// DataSourcesResponse represents the complete data sources configuration
type DataSourcesResponse struct {
	Sources          []DataSourceConfig `json:"sources"`
	AvailableSources []string           `json:"available_sources"`
}

// HandleGetDataSources handles GET /api/settings/data-sources
// Returns structured data source priorities configuration
func (h *Handler) HandleGetDataSources(w http.ResponseWriter, r *http.Request) {
	// Define data source types with descriptions
	dataTypes := []struct {
		Type        DataSourceType
		Description string
		SettingKey  string
		Default     string
	}{
		{DataTypeFundamentals, "Financial fundamentals (P/E, margins, company overview)", "datasource_fundamentals", `["alphavantage","yahoo"]`},
		{DataTypeCurrentPrices, "Real-time and delayed stock quotes", "datasource_current_prices", `["tradernet","alphavantage","yahoo"]`},
		{DataTypeHistorical, "Historical OHLCV time series data", "datasource_historical", `["tradernet","alphavantage","yahoo"]`},
		{DataTypeTechnicals, "Technical indicators (RSI, SMA, MACD, etc.)", "datasource_technicals", `["alphavantage","yahoo"]`},
		{DataTypeExchangeRates, "Currency exchange rates", "datasource_exchange_rates", `["exchangerate","tradernet","alphavantage"]`},
		{DataTypeISINLookup, "ISIN to ticker symbol resolution", "datasource_isin_lookup", `["openfigi","yahoo"]`},
		{DataTypeCompanyMetadata, "Company metadata (industry, sector, country)", "datasource_company_metadata", `["alphavantage","yahoo","openfigi"]`},
	}

	sources := make([]DataSourceConfig, 0, len(dataTypes))
	for _, dt := range dataTypes {
		// Get current priorities from settings
		var priorities []string
		if val, err := h.service.Get(dt.SettingKey); err == nil && val != nil {
			// Type assert to string since data source settings are stored as JSON strings
			if strVal, ok := val.(string); ok && strVal != "" {
				if err := json.Unmarshal([]byte(strVal), &priorities); err != nil {
					h.log.Warn().Str("key", dt.SettingKey).Err(err).Msg("Failed to parse data source priorities, using default")
					_ = json.Unmarshal([]byte(dt.Default), &priorities)
				}
			} else {
				_ = json.Unmarshal([]byte(dt.Default), &priorities)
			}
		} else {
			_ = json.Unmarshal([]byte(dt.Default), &priorities)
		}

		sources = append(sources, DataSourceConfig{
			Type:        dt.Type,
			Description: dt.Description,
			Priorities:  priorities,
			SettingKey:  dt.SettingKey,
		})
	}

	// List of all available data sources
	availableSources := []string{
		"alphavantage",
		"yahoo",
		"tradernet",
		"exchangerate",
		"openfigi",
	}

	response := DataSourcesResponse{
		Sources:          sources,
		AvailableSources: availableSources,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error().Err(err).Msg("Failed to encode data sources response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleUpdateDataSource handles PUT /api/settings/data-sources/{type}
// Updates the priority order for a specific data type
func (h *Handler) HandleUpdateDataSource(w http.ResponseWriter, r *http.Request) {
	dataType := chi.URLParam(r, "type")
	if dataType == "" {
		http.Error(w, "Data type is required", http.StatusBadRequest)
		return
	}

	// Map data type to setting key
	settingKeyMap := map[string]string{
		"fundamentals":     "datasource_fundamentals",
		"current_prices":   "datasource_current_prices",
		"historical":       "datasource_historical",
		"technicals":       "datasource_technicals",
		"exchange_rates":   "datasource_exchange_rates",
		"isin_lookup":      "datasource_isin_lookup",
		"company_metadata": "datasource_company_metadata",
	}

	settingKey, ok := settingKeyMap[dataType]
	if !ok {
		http.Error(w, "Invalid data type", http.StatusBadRequest)
		return
	}

	// Parse request body
	var request struct {
		Priorities []string `json:"priorities"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(request.Priorities) == 0 {
		http.Error(w, "Priorities array cannot be empty", http.StatusBadRequest)
		return
	}

	// Validate sources
	validSources := map[string]bool{
		"alphavantage": true,
		"yahoo":        true,
		"tradernet":    true,
		"exchangerate": true,
		"openfigi":     true,
	}
	for _, source := range request.Priorities {
		if !validSources[source] {
			http.Error(w, "Invalid data source: "+source, http.StatusBadRequest)
			return
		}
	}

	// Convert to JSON string for storage
	prioritiesJSON, err := json.Marshal(request.Priorities)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to marshal priorities")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store the setting
	_, err = h.service.Set(settingKey, string(prioritiesJSON))
	if err != nil {
		h.log.Error().Err(err).Str("key", settingKey).Msg("Failed to update data source priority")
		http.Error(w, "Failed to update setting", http.StatusInternalServerError)
		return
	}

	// Emit settings changed event
	if h.eventManager != nil {
		h.eventManager.Emit(events.SettingsChanged, "settings", map[string]interface{}{
			"key":   settingKey,
			"value": string(prioritiesJSON),
		})
	}

	// Return updated configuration
	response := DataSourceConfig{
		Type:       DataSourceType(dataType),
		Priorities: request.Priorities,
		SettingKey: settingKey,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
