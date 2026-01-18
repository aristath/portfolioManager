/**
 * Package di provides dependency injection for rebalancing service initialization.
 *
 * Step 12: Initialize Rebalancing Services
 * Rebalancing services handle portfolio rebalancing and negative balance correction.
 */
package di

import (
	"github.com/aristath/sentinel/internal/modules/cash_flows"
	"github.com/aristath/sentinel/internal/modules/rebalancing"
	"github.com/rs/zerolog"
)

// initializeRebalancingServices initializes rebalancing-related services.
func initializeRebalancingServices(container *Container, cashManager *cash_flows.CashManager, log zerolog.Logger) error {
	// Negative balance rebalancer
	// Handles emergency rebalancing when negative balances are detected
	// Sells positions to correct negative cash balances
	container.NegativeBalanceRebalancer = rebalancing.NewNegativeBalanceRebalancer(
		log,
		cashManager,
		container.BrokerClient,
		container.SecurityRepo,
		container.PositionRepo,
		container.SettingsRepo,
		container.CurrencyExchangeService,
		container.TradeExecutionService,
		container.RecommendationRepo,
	)

	// Rebalancing service
	// Orchestrates portfolio rebalancing based on allocation targets and triggers
	triggerChecker := rebalancing.NewTriggerChecker(log)
	container.RebalancingService = rebalancing.NewService(
		triggerChecker,
		container.NegativeBalanceRebalancer,
		container.PlannerService,
		container.PositionRepo,
		container.SecurityRepo,
		container.AllocRepo,
		cashManager,
		container.BrokerClient,
		container.PlannerConfigRepo,
		container.RecommendationRepo,
		container.OpportunityContextBuilder,
		container.SettingsRepo,
		log,
	)

	return nil
}
