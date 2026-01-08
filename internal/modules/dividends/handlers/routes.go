package handlers

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all dividend routes
func (h *DividendHandlers) RegisterRoutes(r chi.Router) {
	// Dividend routes (faithful translation of Python repository to HTTP API)
	r.Route("/dividends", func(r chi.Router) {
		// List endpoints
		r.Get("/", h.HandleGetDividends)                        // List all dividends
		r.Get("/{id}", h.HandleGetDividendByID)                 // Get dividend by ID
		r.Get("/symbol/{symbol}", h.HandleGetDividendsBySymbol) // Get dividends by symbol

		// CRITICAL: Endpoints used by dividend_reinvestment.py job
		r.Get("/unreinvested", h.HandleGetUnreinvestedDividends) // Get unreinvested dividends
		r.Post("/{id}/pending-bonus", h.HandleSetPendingBonus)   // Set pending bonus
		r.Post("/{id}/mark-reinvested", h.HandleMarkReinvested)  // Mark as reinvested

		// Management endpoints
		r.Post("/", h.HandleCreateDividend)                  // Create dividend
		r.Post("/clear-bonus/{symbol}", h.HandleClearBonus)  // Clear bonus by symbol
		r.Get("/pending-bonuses", h.HandleGetPendingBonuses) // Get all pending bonuses

		// Analytics endpoints
		r.Get("/analytics/total", h.HandleGetTotalDividendsBySymbol)       // Total dividends by symbol
		r.Get("/analytics/reinvestment-rate", h.HandleGetReinvestmentRate) // Overall reinvestment rate
	})
}
