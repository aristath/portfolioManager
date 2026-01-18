/**
 * Package di provides dependency injection for display service initialization.
 *
 * Step 13: Initialize Ticker and Display Services
 * Display services handle LED ticker text generation and portfolio health visualization.
 */
package di

import (
	"fmt"
	"os"
	"time"

	"github.com/aristath/sentinel/internal/modules/cash_flows"
	"github.com/aristath/sentinel/internal/modules/display"
	"github.com/aristath/sentinel/internal/ticker"
	"github.com/rs/zerolog"
)

// initializeDisplayServices initializes display-related services.
func initializeDisplayServices(container *Container, cashManager *cash_flows.CashManager, displayManager *display.StateManager, log zerolog.Logger) error {
	// Ticker content service (generates ticker text)
	// Generates scrolling text for LED display (portfolio value, cash, next actions)
	container.TickerContentService = ticker.NewTickerContentService(
		container.PortfolioDB.Conn(),
		container.ConfigDB.Conn(),
		container.CacheDB.Conn(),
		cashManager,
		log,
	)
	log.Info().Msg("Ticker content service initialized")

	// Health calculator (calculates portfolio health scores)
	// Calculates health scores for each security in the portfolio
	// Health scores are used for LED display visualization
	container.HealthCalculator = display.NewHealthCalculator(
		container.PortfolioDB.Conn(),
		container.HistoryDBClient,
		container.ConfigDB.Conn(),
		log,
	)
	log.Info().Msg("Health calculator initialized")

	// Health updater (periodically sends health scores to display)
	// Sends portfolio health data to LED display for animated visualization
	displayURL := "http://localhost:7000"
	if envURL := os.Getenv("DISPLAY_URL"); envURL != "" {
		displayURL = envURL
	}
	updateInterval := 30 * time.Minute // Default 30 minutes
	if intervalSetting, err := container.SettingsRepo.Get("display_health_update_interval"); err == nil && intervalSetting != nil {
		// Parse string to float
		var intervalFloat float64
		if _, err := fmt.Sscanf(*intervalSetting, "%f", &intervalFloat); err == nil {
			updateInterval = time.Duration(intervalFloat) * time.Second
		}
	}
	container.HealthUpdater = display.NewHealthUpdater(
		container.HealthCalculator,
		displayURL,
		updateInterval,
		log,
	)
	log.Info().Dur("interval", updateInterval).Msg("Health updater initialized")

	// Mode manager (switches between display modes)
	// Manages LED display modes: TEXT (ticker), HEALTH (animated visualization), STATS (pixel count)
	if displayManager != nil {
		container.ModeManager = display.NewModeManager(
			displayManager,
			container.HealthUpdater,
			container.TickerContentService,
			log,
		)
		log.Info().Msg("Display mode manager initialized")

		// Apply display mode from settings (if configured)
		// Display mode can be set via Settings UI
		if container.SettingsRepo != nil {
			if mode, err := container.SettingsRepo.Get("display_mode"); err == nil && mode != nil && *mode != "" {
				if err := container.ModeManager.SetMode(display.DisplayMode(*mode)); err != nil {
					log.Warn().Err(err).Str("mode", *mode).Msg("Failed to set display mode from settings, using default")
				} else {
					log.Info().Str("mode", *mode).Msg("Applied display mode from settings")
				}
			}
		}
	}

	return nil
}
