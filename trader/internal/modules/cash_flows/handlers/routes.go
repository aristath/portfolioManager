package handlers

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all cash flows routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	// Cash-flows routes (faithful translation of Python routes)
	r.Route("/cash-flows", func(r chi.Router) {
		r.Get("/", h.HandleGetCashFlows)      // GET / - list cash flows with filters
		r.Get("/sync", h.HandleSyncCashFlows) // GET /sync - sync from Tradernet
		r.Get("/summary", h.HandleGetSummary) // GET /summary - aggregate statistics
	})
}
