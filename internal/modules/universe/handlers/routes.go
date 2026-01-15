package handlers

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all universe routes
func (h *UniverseHandlers) RegisterRoutes(r chi.Router) {
	// Universe/Securities routes (faithful translation of Python routes)
	r.Route("/securities", func(r chi.Router) {
		// GET endpoints (implemented in Go)
		r.Get("/", h.HandleGetStocks)      // List all securities with scores
		r.Get("/{isin}", h.HandleGetStock) // Get security detail by ISIN

		// POST endpoints (proxied to Python for complex operations)
		r.Post("/", h.HandleCreateStock)                           // Create security
		r.Post("/add-by-identifier", h.HandleAddStockByIdentifier) // Auto-setup by symbol/ISIN
		r.Post("/refresh-all", h.HandleRefreshAllScores)           // Recalculate all scores

		// Security-specific POST endpoints
		r.Post("/{isin}/refresh-data", h.HandleRefreshSecurityData) // Full data refresh
		r.Post("/{isin}/refresh", h.HandleRefreshStockScore)        // Quick score refresh

		// PUT/DELETE endpoints
		r.Put("/{isin}", h.HandleUpdateStock)    // Update security (requires score recalc)
		r.Delete("/{isin}", h.HandleDeleteStock) // Soft delete (implemented in Go)

		// Override endpoints (for user customizations separate from Tradernet defaults)
		r.Get("/{isin}/overrides", h.HandleGetSecurityOverrides)              // Get all overrides for security
		r.Put("/{isin}/overrides/{field}", h.HandleSetSecurityOverride)       // Set specific override
		r.Delete("/{isin}/overrides/{field}", h.HandleDeleteSecurityOverride) // Reset field to default
	})
}
