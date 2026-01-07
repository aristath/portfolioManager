package handlers

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all settings routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/settings", func(r chi.Router) {
		// GET /api/settings - Get all settings
		r.Get("/", h.HandleGetAll)

		// PUT /api/settings/{key} - Update a setting value
		r.Put("/{key}", h.HandleUpdate)

		// POST /api/settings/restart-service - Restart the systemd service
		r.Post("/restart-service", h.HandleRestartService)

		// POST /api/settings/restart - Trigger system reboot
		r.Post("/restart", h.HandleRestart)

		// POST /api/settings/reset-cache - Clear all cached data
		r.Post("/reset-cache", h.HandleResetCache)

		// GET /api/settings/cache-stats - Get cache statistics
		r.Get("/cache-stats", h.HandleGetCacheStats)

		// POST /api/settings/reschedule-jobs - Reschedule all jobs
		r.Post("/reschedule-jobs", h.HandleRescheduleJobs)

		// GET /api/settings/trading-mode - Get current trading mode
		r.Get("/trading-mode", h.HandleGetTradingMode)

		// POST /api/settings/trading-mode - Toggle trading mode
		r.Post("/trading-mode", h.HandleToggleTradingMode)
	})
}

