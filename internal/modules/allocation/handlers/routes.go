package handlers

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all allocation routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/allocation", func(r chi.Router) {
		r.Get("/targets", h.HandleGetTargets)
		r.Get("/current", h.HandleGetCurrentAllocation)
		r.Get("/deviations", h.HandleGetDeviations)

		// Available options (for autocomplete)
		r.Get("/available/geographies", h.HandleGetAvailableGeographies)
		r.Get("/available/industries", h.HandleGetAvailableIndustries)

		// Set targets
		r.Put("/targets/geography", h.HandleSetGeographyTargets)
		r.Put("/targets/industry", h.HandleSetIndustryTargets)

		// Allocation Analytics (API extension)
		r.Get("/vs-targets", h.HandleGetAllocationVsTargets)
		r.Get("/rebalance-needs", h.HandleGetRebalanceNeeds)
		r.Get("/contribution", h.HandleGetGroupContribution)
	})
}
