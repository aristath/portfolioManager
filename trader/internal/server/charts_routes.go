// Package server provides HTTP server and routing functionality.
package server

import (
	"github.com/aristath/portfolioManager/internal/modules/charts"
	"github.com/go-chi/chi/v5"
)

// setupChartsRoutes configures charts module routes
func (s *Server) setupChartsRoutes(r chi.Router) {
	// Use services from container (single source of truth)
	securityRepo := s.container.SecurityRepo

	// Initialize charts service
	chartsService := charts.NewService(
		s.historyDB.Conn(),
		securityRepo,
		s.universeDB.Conn(),
		s.log,
	)

	// Initialize charts handler
	chartsHandler := charts.NewHandler(chartsService, s.log)

	// Register routes
	r.Route("/charts", func(r chi.Router) {
		// GET /api/charts/sparklines - Get 1-year sparkline data for all active securities
		r.Get("/sparklines", chartsHandler.HandleGetSparklines)

		// GET /api/charts/securities/{isin} - Get historical price data for a specific security
		r.Get("/securities/{isin}", chartsHandler.HandleGetSecurityChart)
	})
}
