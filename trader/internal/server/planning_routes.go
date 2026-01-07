package server

import (
	"github.com/aristath/portfolioManager/internal/modules/planning/config"
	"github.com/aristath/portfolioManager/internal/modules/planning/handlers"
	"github.com/aristath/portfolioManager/internal/modules/planning/planner"
	"github.com/aristath/portfolioManager/internal/modules/planning/repository"
	"github.com/go-chi/chi/v5"
)

// setupPlanningRoutes configures planning module routes.
func (s *Server) setupPlanningRoutes(r chi.Router) {
	// Use services from container (single source of truth)
	planningService := s.container.PlanningService
	configRepo := s.container.PlannerConfigRepo
	corePlanner := s.container.PlannerService

	// Initialize planner repository (uses agentsDB for sequences/evaluations)
	plannerRepo := repository.NewPlannerRepository(s.agentsDB, s.log)

	// Initialize config validator
	validator := config.NewValidator()

	// Initialize incremental planner (batch generation)
	incrementalPlanner := planner.NewIncrementalPlanner(
		corePlanner,
		plannerRepo,
		s.log,
	)

	// Initialize event broadcaster for SSE streaming
	eventBroadcaster := handlers.NewEventBroadcaster(s.log)

	// Initialize handlers
	recommendationsHandler := handlers.NewRecommendationsHandler(planningService, s.log)
	configHandler := handlers.NewConfigHandler(configRepo, validator, s.log)
	batchHandler := handlers.NewBatchHandler(incrementalPlanner, configRepo, s.log)
	executeHandler := handlers.NewExecuteHandler(plannerRepo, nil, s.log) // TODO: Pass trade executor
	statusHandler := handlers.NewStatusHandler(plannerRepo, s.log)
	streamHandler := handlers.NewStreamHandler(eventBroadcaster, s.log)

	// Register routes
	r.Route("/planning", func(r chi.Router) {
		// Recommendations (main planning endpoint)
		r.Get("/recommendations", recommendationsHandler.ServeHTTP)
		r.Post("/recommendations", recommendationsHandler.ServeHTTP)

		// Configuration management
		r.Get("/configs", configHandler.ServeHTTP)
		r.Post("/configs", configHandler.ServeHTTP)
		r.Get("/configs/{id}", configHandler.ServeHTTP)
		r.Put("/configs/{id}", configHandler.ServeHTTP)
		r.Delete("/configs/{id}", configHandler.ServeHTTP)
		r.Post("/configs/validate", configHandler.ServeHTTP)
		r.Get("/configs/{id}/history", configHandler.ServeHTTP)

		// Batch generation
		r.Post("/batch", batchHandler.ServeHTTP)

		// Plan execution
		r.Post("/execute", executeHandler.ServeHTTP)

		// Status monitoring
		r.Get("/status", statusHandler.ServeHTTP)

		// SSE streaming
		r.Get("/stream", streamHandler.ServeHTTP)
	})

	// Note: /api/trades/recommendations endpoints are handled by trading routes
	// (see setupTradingRoutes - TODO: add recommendations endpoints there)
}
