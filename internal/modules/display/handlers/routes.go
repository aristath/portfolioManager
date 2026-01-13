package handlers

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all display routes
func (h *Handlers) RegisterRoutes(r chi.Router) {
	// Display routes
	r.Route("/display", func(r chi.Router) {
		r.Get("/state", h.HandleGetState) // Get current display state
		r.Post("/text", h.HandleSetText)  // Set display text
		r.Post("/led3", h.HandleSetLED3)  // Set LED3 color
		r.Post("/led4", h.HandleSetLED4)  // Set LED4 color

		// Mode management
		r.Get("/mode", h.HandleGetMode)  // Get current display mode
		r.Post("/mode", h.HandleSetMode) // Set display mode

		// Portfolio health
		r.Route("/portfolio-health", func(r chi.Router) {
			r.Get("/preview", h.HandleGetPortfolioHealth)   // Preview health scores
			r.Post("/trigger", h.HandleTriggerHealthUpdate) // Manually trigger update
		})
	})
}
