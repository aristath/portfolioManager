package server

import (
	"github.com/aristath/portfolioManager/internal/modules/settings"
	"github.com/aristath/portfolioManager/internal/modules/universe"
	"github.com/go-chi/chi/v5"
)

// setupSettingsRoutes configures settings module routes
func (s *Server) setupSettingsRoutes(r chi.Router) {
	// Use services from container (single source of truth)
	settingsService := s.container.SettingsService
	tradernetClient := s.container.TradernetClient
	currencyExchangeService := s.container.CurrencyExchangeService
	positionRepo := s.container.PositionRepo
	securityRepo := s.container.SecurityRepo
	portfolioService := s.container.PortfolioService
	scoreRepo := s.container.ScoreRepo
	yahooClient := s.container.YahooClient
	historyDB := s.container.HistoryDBClient
	setupService := s.container.SetupService
	syncService := s.container.SyncService
	securityScorer := s.container.SecurityScorer
	tradingService := s.container.TradingService

	// Create a simple score calculator that uses UniverseHandlers pattern
	// For onboarding, we'll create a minimal handler just for score calculation
	universeHandler := universe.NewUniverseHandlers(
		securityRepo,
		scoreRepo,
		s.portfolioDB.Conn(),
		positionRepo,
		securityScorer,
		yahooClient,
		historyDB,
		setupService,
		syncService,
		currencyExchangeService,
		s.log,
	)

	// Wire score calculator (already done in services.go, but ensure it's set here too)
	setupService.SetScoreCalculator(universeHandler)
	syncService.SetScoreCalculator(universeHandler)

	// Create onboarding service
	onboardingService := settings.NewOnboardingService(
		portfolioService,
		syncService,
		tradingService,
		tradernetClient,
		s.log,
	)

	// Initialize settings handler
	settingsHandler := settings.NewHandler(settingsService, s.log)
	settingsHandler.SetOnboardingService(onboardingService)

	// Set credential refresher to refresh system handlers' tradernet client
	if s.systemHandlers != nil {
		settingsHandler.SetCredentialRefresher(s.systemHandlers)
	}

	// Register routes
	// Note: r is already under /api route group, so use /settings not /api/settings
	r.Route("/settings", func(r chi.Router) {
		// GET /api/settings - Get all settings
		r.Get("/", settingsHandler.HandleGetAll)

		// PUT /api/settings/{key} - Update a setting value
		r.Put("/{key}", settingsHandler.HandleUpdate)

		// POST /api/settings/restart-service - Restart the systemd service
		r.Post("/restart-service", settingsHandler.HandleRestartService)

		// POST /api/settings/restart - Trigger system reboot
		r.Post("/restart", settingsHandler.HandleRestart)

		// POST /api/settings/reset-cache - Clear all cached data
		r.Post("/reset-cache", settingsHandler.HandleResetCache)

		// GET /api/settings/cache-stats - Get cache statistics
		r.Get("/cache-stats", settingsHandler.HandleGetCacheStats)

		// POST /api/settings/reschedule-jobs - Reschedule all jobs
		r.Post("/reschedule-jobs", settingsHandler.HandleRescheduleJobs)

		// GET /api/settings/trading-mode - Get current trading mode
		r.Get("/trading-mode", settingsHandler.HandleGetTradingMode)

		// POST /api/settings/trading-mode - Toggle trading mode
		r.Post("/trading-mode", settingsHandler.HandleToggleTradingMode)
	})
}
