package handlers

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all settings routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/settings", func(r chi.Router) {
		r.Get("/", h.HandleGetSettings)
		r.Put("/", h.HandleUpdateSettings)
		r.Get("/onboarding", h.HandleGetOnboardingStatus)
		r.Post("/onboarding/complete", h.HandleCompleteOnboarding)
	})
}

