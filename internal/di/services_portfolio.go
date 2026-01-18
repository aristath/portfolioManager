/**
 * Package di provides dependency injection for portfolio service initialization.
 *
 * Step 6: Initialize Portfolio Service
 * Portfolio service manages portfolio state and operations.
 */
package di

import (
	"github.com/aristath/sentinel/internal/modules/cash_flows"
	"github.com/aristath/sentinel/internal/modules/portfolio"
	"github.com/rs/zerolog"
)

// initializePortfolioService initializes the portfolio service.
func initializePortfolioService(container *Container, cashManager *cash_flows.CashManager, log zerolog.Logger) error {
	// Create adapter for SecuritySetupService to match portfolio.SecuritySetupServiceInterface
	// This bridges the interface mismatch between universe and portfolio packages
	setupServiceAdapter := &securitySetupServiceAdapter{service: container.SetupService}

	// Portfolio service (with SecuritySetupService adapter for auto-adding missing securities)
	// Manages portfolio state, positions, and provides portfolio-level operations
	// Auto-adds missing securities via SecuritySetupService adapter
	portfolioSecurityProvider := NewSecurityProviderAdapter(container.SecurityRepo)
	container.PortfolioService = portfolio.NewPortfolioService(
		container.PositionRepo,
		container.AllocRepo,
		cashManager, // Use concrete type
		container.UniverseDB.Conn(),
		portfolioSecurityProvider,
		container.BrokerClient,
		container.CurrencyExchangeService,
		container.ExchangeRateCacheService,
		container.SettingsService,
		setupServiceAdapter, // Use adapter to match interface
		log,
	)

	return nil
}
