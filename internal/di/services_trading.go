/**
 * Package di provides dependency injection for trading service initialization.
 *
 * Step 4: Initialize Trading Services
 * Trading services handle trade validation, execution, and safety checks.
 */
package di

import (
	"github.com/aristath/sentinel/internal/modules/cash_flows"
	"github.com/aristath/sentinel/internal/modules/trading"
	"github.com/aristath/sentinel/internal/services"
	"github.com/rs/zerolog"
)

// initializeTradingServices initializes trading-related services.
func initializeTradingServices(container *Container, cashManager *cash_flows.CashManager, log zerolog.Logger) error {
	// Trade safety service with all validation layers
	// Validates trades against frequency limits, cooloff periods, market hours, etc.
	container.TradeSafetyService = trading.NewTradeSafetyService(
		container.TradeRepo,
		container.PositionRepo,
		container.SecurityRepo,
		container.SettingsService,
		container.MarketHoursService,
		log,
	)

	// Trading service
	// Orchestrates trading operations (validation, execution, event publishing)
	container.TradingService = trading.NewTradingService(
		container.TradeRepo,
		container.BrokerClient,
		container.TradeSafetyService,
		container.EventManager,
		log,
	)

	// Trade execution service - uses market orders for simplicity
	// Executes trades via broker API, updates positions, manages cash, publishes events
	container.TradeExecutionService = services.NewTradeExecutionService(
		container.BrokerClient,
		container.TradeRepo,
		container.PositionRepo,
		cashManager, // Use concrete type for now, will be interface later
		container.CurrencyExchangeService,
		container.EventManager,
		container.SettingsService,
		container.PlannerConfigRepo,
		container.HistoryDB.Conn(),
		container.SecurityRepo,
		container.MarketHoursService, // Market hours validation
		log,
	)

	return nil
}
