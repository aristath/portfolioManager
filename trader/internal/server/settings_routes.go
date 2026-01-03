package server

import (
	"github.com/aristath/arduino-trader/internal/modules/settings"
	"github.com/go-chi/chi/v5"
)

// setupSettingsRoutes configures settings module routes
func (s *Server) setupSettingsRoutes(r chi.Router) {
	// Initialize settings repository
	settingsRepo := settings.NewRepository(s.configDB.Conn(), s.log)

	// Initialize settings service
	settingsService := settings.NewService(settingsRepo, s.log)

	// Initialize settings handler
	settingsHandler := settings.NewHandler(settingsService, s.log)

	// Register routes
	r.Route("/api/settings", func(r chi.Router) {
		// GET /api/settings - Get all settings
		r.Get("/", settingsHandler.HandleGetAll)

		// PUT /api/settings/{key} - Update a setting value
		r.Put("/{key}", settingsHandler.HandleUpdate)

		// POST /api/settings/restart-service - Restart the systemd service
		r.Post("/restart-service", settingsHandler.HandleRestartService)

		// POST /api/settings/restart - Trigger system reboot
		r.Post("/restart", settingsHandler.HandleRestart)

		// POST /api/settings/reset-cache - Clear all cached data
		r.Post("/reset-cache", settingsHandler.HandleResetCache)

		// GET /api/settings/cache-stats - Get cache statistics
		r.Get("/cache-stats", settingsHandler.HandleGetCacheStats)

		// POST /api/settings/reschedule-jobs - Reschedule all jobs
		r.Post("/reschedule-jobs", settingsHandler.HandleRescheduleJobs)

		// GET /api/settings/trading-mode - Get current trading mode
		r.Get("/trading-mode", settingsHandler.HandleGetTradingMode)

		// POST /api/settings/trading-mode - Toggle trading mode
		r.Post("/trading-mode", settingsHandler.HandleToggleTradingMode)
	})
}
