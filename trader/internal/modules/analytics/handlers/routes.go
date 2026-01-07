package handlers

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all analytics routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/analytics", func(r chi.Router) {
		r.Get("/factor-exposure", h.HandleGetFactorExposure)
		r.Get("/risk-metrics", h.HandleGetRiskMetrics)
		r.Get("/correlation", h.HandleGetCorrelation)
	})
}

