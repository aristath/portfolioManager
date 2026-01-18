/**
 * Package di provides dependency injection for client initialization.
 *
 * Step 1: Initialize Clients
 * External API clients must be initialized first as they are dependencies
 * for many services (currency exchange, price fetching, trade execution).
 */
package di

import (
	"os"

	"github.com/aristath/sentinel/internal/clients/tradernet"
	"github.com/aristath/sentinel/internal/config"
	"github.com/aristath/sentinel/internal/modules/display"
	"github.com/rs/zerolog"
)

// initializeClients initializes external API clients and display configuration.
func initializeClients(container *Container, cfg *config.Config, displayManager *display.StateManager, log zerolog.Logger) error {
	// Broker client (Tradernet adapter) - single external data source
	// Tradernet is the only external data source for prices, quotes, and trade execution
	container.BrokerClient = tradernet.NewTradernetBrokerAdapter(cfg.TradernetAPIKey, cfg.TradernetAPISecret, log)
	log.Info().Msg("Broker client initialized (Tradernet adapter)")

	// Configure display service (App Lab HTTP API on localhost:7000)
	// displayManager can be nil in tests - skip display configuration if nil
	if displayManager != nil {
		// Get display URL from settings or use default
		displayURL := display.DefaultDisplayURL
		if container.SettingsRepo != nil {
			if url, err := container.SettingsRepo.Get("display_url"); err == nil && url != nil && *url != "" {
				displayURL = *url
			}
		}
		displayManager.SetDisplayURL(displayURL)

		// Check if display should be enabled (default: true on Arduino hardware)
		// Display is enabled if running on Arduino Uno Q (check for arduino-router socket)
		displayEnabled := false
		if _, err := os.Stat("/var/run/arduino-router.sock"); err == nil {
			// Arduino router socket exists - we're on Arduino hardware
			displayEnabled = true
			log.Info().Str("url", displayURL).Msg("Arduino hardware detected, enabling display service")
		} else {
			// Allow manual override via settings (for testing/development)
			if container.SettingsRepo != nil {
				if enabled, err := container.SettingsRepo.Get("display_enabled"); err == nil && enabled != nil && *enabled == "true" {
					displayEnabled = true
					log.Info().Str("url", displayURL).Msg("Display service manually enabled via settings")
				}
			}
		}

		if displayEnabled {
			displayManager.Enable()
		} else {
			log.Info().Msg("Display service disabled (not on Arduino hardware)")
		}
	} else {
		log.Debug().Msg("Display manager not provided - skipping display configuration")
	}

	return nil
}
