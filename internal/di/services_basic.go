/**
 * Package di provides dependency injection for basic service initialization.
 *
 * Step 2: Initialize Basic Services
 * Foundational services that other services depend on.
 */
package di

import (
	"github.com/aristath/sentinel/internal/clients/tradernet"
	"github.com/aristath/sentinel/internal/events"
	"github.com/aristath/sentinel/internal/market_regime"
	"github.com/aristath/sentinel/internal/modules/market_hours"
	"github.com/aristath/sentinel/internal/modules/settings"
	"github.com/aristath/sentinel/internal/services"
	"github.com/rs/zerolog"
)

// initializeBasicServices initializes foundational services.
func initializeBasicServices(container *Container, log zerolog.Logger) error {
	// Currency exchange service
	// Fetches exchange rates from Tradernet API
	container.CurrencyExchangeService = services.NewCurrencyExchangeService(container.BrokerClient, log)

	// Market hours service
	// Provides market hours and holiday information for all exchanges
	container.MarketHoursService = market_hours.NewMarketHoursService()

	// Market state detector (for market-aware scheduling)
	// Detects whether markets are open/closed for work processor scheduling
	container.MarketStateDetector = market_regime.NewMarketStateDetector(
		container.SecurityRepo,
		container.MarketHoursService,
		log,
	)

	// Event system (bus-based architecture)
	// EventBus provides pub/sub for system-wide events
	// EventManager wraps the bus with additional functionality
	container.EventBus = events.NewBus(log)
	container.EventManager = events.NewManager(container.EventBus, log)

	// Market status WebSocket client
	// Connects to Tradernet WebSocket for real-time market status updates
	// Publishes events to EventBus when market status changes
	container.MarketStatusWS = tradernet.NewMarketStatusWebSocket(
		"wss://wss.tradernet.com/",
		"", // Empty string for demo mode (SID not required)
		container.EventBus,
		log,
	)

	// Start WebSocket connection (non-blocking, will auto-retry)
	// Connection failures don't fail startup - reconnect loop handles retries
	if err := container.MarketStatusWS.Start(); err != nil {
		log.Warn().Err(err).Msg("Market status WebSocket connection failed, will auto-retry")
		// Don't fail startup - reconnect loop will handle it
	}

	// Settings service (needed for trade safety and other services)
	// Provides access to application settings with temperament-aware adjustments
	container.SettingsService = settings.NewService(container.SettingsRepo, log)

	// Exchange rate cache service (Tradernet + DB cache)
	// Primary: Fetches from Tradernet API
	// Secondary: Falls back to DB cache if API unavailable
	container.ExchangeRateCacheService = services.NewExchangeRateCacheService(
		container.CurrencyExchangeService, // Tradernet (primary)
		container.HistoryDBClient,         // DB cache (secondary)
		container.SettingsService,
		log,
	)

	// Price conversion service (converts native currency prices to EUR)
	// Converts prices from security's native currency to EUR for portfolio calculations
	container.PriceConversionService = services.NewPriceConversionService(
		container.CurrencyExchangeService,
		log,
	)

	return nil
}
