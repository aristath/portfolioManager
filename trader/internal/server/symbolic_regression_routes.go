package server

import (
	"github.com/aristath/portfolioManager/internal/modules/symbolic_regression"
	"github.com/go-chi/chi/v5"
)

// setupSymbolicRegressionRoutes configures symbolic regression module routes
func (s *Server) setupSymbolicRegressionRoutes(r chi.Router) {
	// Initialize formula storage
	formulaStorage := symbolic_regression.NewFormulaStorage(s.configDB.Conn(), s.log)

	// Initialize data preparation
	dataPrep := symbolic_regression.NewDataPrep(
		s.historyDB.Conn(),
		s.portfolioDB.Conn(),
		s.configDB.Conn(),
		s.universeDB.Conn(),
		s.log,
	)

	// Initialize discovery service
	discoveryService := symbolic_regression.NewDiscoveryService(
		dataPrep,
		formulaStorage,
		s.log,
	)

	// Initialize handlers
	handlers := symbolic_regression.NewHandlers(
		formulaStorage,
		discoveryService,
		dataPrep,
		s.log,
	)

	// Register routes
	handlers.RegisterRoutes(r)
}
