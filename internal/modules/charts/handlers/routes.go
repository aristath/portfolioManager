package handlers

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all charts routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/charts", func(r chi.Router) {
		r.Get("/sparklines", h.HandleGetSparklines)
		r.Get("/securities/{isin}", h.HandleGetSecurityChart)
	})
}
