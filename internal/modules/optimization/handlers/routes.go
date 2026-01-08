package handlers

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all optimization routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	// Optimization routes (faithful translation of Python routes)
	r.Route("/optimizer", func(r chi.Router) {
		r.Get("/", h.HandleGetStatus) // Get optimizer status and last run
		r.Post("/run", h.HandleRun)   // Run optimization
	})
}
