/**
 * Package di provides dependency injection for repository implementations.
 *
 * This package initializes all repository instances with their database connections.
 * Repositories abstract data access and provide clean interfaces for services.
 *
 * Repository initialization order matters due to dependencies:
 * - OverrideRepo must be created before SecurityRepo (SecurityRepo uses it for override merging)
 * - SecurityRepo must be created before repositories that need security lookups
 * - Adapters are created to bridge interface mismatches between repositories
 */
package di

import (
	"fmt"

	"github.com/aristath/sentinel/internal/clientdata"
	"github.com/aristath/sentinel/internal/modules/allocation"
	"github.com/aristath/sentinel/internal/modules/cash_flows"
	"github.com/aristath/sentinel/internal/modules/dividends"
	"github.com/aristath/sentinel/internal/modules/planning"
	planningrepo "github.com/aristath/sentinel/internal/modules/planning/repository"
	"github.com/aristath/sentinel/internal/modules/portfolio"
	"github.com/aristath/sentinel/internal/modules/settings"
	"github.com/aristath/sentinel/internal/modules/trading"
	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/rs/zerolog"
)

/**
 * InitializeRepositories creates all repositories and stores them in the container.
 *
 * Repositories are initialized in dependency order:
 * 1. OverrideRepo (no dependencies)
 * 2. SecurityRepo (depends on OverrideRepo)
 * 3. PositionRepo, ScoreRepo, TradeRepo, DividendRepo (depend on SecurityRepo via adapters)
 * 4. Other repositories (CashRepo, AllocRepo, SettingsRepo, etc.)
 * 5. In-memory repositories (RecommendationRepo, PlannerRepo - ephemeral data)
 * 6. HistoryDBClient (with price filter for read-time anomaly filtering)
 * 7. ClientDataRepo (cache repository)
 *
 * Adapters are used to bridge interface mismatches:
 * - SecurityProviderAdapter: Adapts SecurityRepo for PositionRepo
 * - TradingSecurityProviderAdapter: Adapts SecurityRepo for TradeRepo, DividendRepo, ScoreRepo
 * - AllocationSecurityProviderAdapter: Adapts SecurityRepo for AllocRepo
 *
 * @param container - Container to store repository instances (must not be nil)
 * @param log - Structured logger instance
 * @returns error - Error if repository initialization fails
 */
func InitializeRepositories(container *Container, log zerolog.Logger) error {
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}

	// Override repository (needs universeDB) - must be created before SecurityRepository
	// Stores user-configurable security overrides (min_lot, allow_buy, allow_sell, etc.)
	container.OverrideRepo = universe.NewOverrideRepository(
		container.UniverseDB.Conn(),
		log,
	)

	// Security repository (needs universeDB and OverrideRepo for override merging)
	// Merges base security data with user overrides automatically
	container.SecurityRepo = universe.NewSecurityRepositoryWithOverrides(
		container.UniverseDB.Conn(),
		container.OverrideRepo,
		log,
	)

	// Security provider adapter (wraps SecurityRepo for PositionRepo)
	// Bridges interface mismatch between SecurityRepo and PositionRepo's security provider interface
	securityProvider := NewSecurityProviderAdapter(container.SecurityRepo)

	// Position repository (needs portfolioDB, universeDB, and securityProvider)
	// Manages portfolio positions with security lookups via adapter
	container.PositionRepo = portfolio.NewPositionRepository(
		container.PortfolioDB.Conn(),
		container.UniverseDB.Conn(),
		securityProvider,
		log,
	)

	// Score repository (needs portfolioDB and securityProvider for GetBySymbol)
	// Stores security scores and metrics with security lookups via adapter
	scoreSecurityProvider := NewTradingSecurityProviderAdapter(container.SecurityRepo)
	container.ScoreRepo = universe.NewScoreRepositoryWithUniverse(
		container.PortfolioDB.Conn(),
		scoreSecurityProvider,
		log,
	)

	// Dividend repository (needs ledgerDB and security provider for ISIN lookup)
	// Stores dividend transactions with security lookups via adapter
	dividendSecurityProvider := NewTradingSecurityProviderAdapter(container.SecurityRepo)
	container.DividendRepo = dividends.NewDividendRepository(
		container.LedgerDB.Conn(),
		dividendSecurityProvider,
		log,
	)

	// Cash repository (needs portfolioDB)
	// Manages cash balances (cash-as-balances architecture)
	container.CashRepo = cash_flows.NewCashRepository(
		container.PortfolioDB.Conn(),
		log,
	)

	// Trade repository (needs ledgerDB and security provider for ISIN lookup)
	// Stores trade transactions (immutable audit trail) with security lookups via adapter
	tradingSecurityProvider := NewTradingSecurityProviderAdapter(container.SecurityRepo)
	container.TradeRepo = trading.NewTradeRepository(
		container.LedgerDB.Conn(),
		tradingSecurityProvider,
		log,
	)

	// Allocation repository (needs configDB and securityProvider)
	// Manages allocation targets (geography, industry) with security lookups via adapter
	allocSecurityProvider := NewAllocationSecurityProviderAdapter(container.SecurityRepo)
	container.AllocRepo = allocation.NewRepository(
		container.ConfigDB.Conn(),
		allocSecurityProvider,
		log,
	)
	// Set universeDB for allocation repository (needed for security lookups)
	container.AllocRepo.SetUniverseDB(container.UniverseDB.Conn())

	// Settings repository (needs configDB)
	// Manages application settings (credentials, configuration)
	container.SettingsRepo = settings.NewRepository(
		container.ConfigDB.Conn(),
		log,
	)

	// Cash flows repository (needs ledgerDB)
	// Stores cash flow transactions (deposits, withdrawals)
	container.CashFlowsRepo = cash_flows.NewRepository(
		container.LedgerDB.Conn(),
		log,
	)

	// Planning recommendation repository (IN-MEMORY - ephemeral data)
	// Recommendations are generated on-demand and don't need persistence
	container.RecommendationRepo = planning.NewInMemoryRecommendationRepository(log)

	// Planner config repository (needs configDB)
	// Stores planner configuration (opportunity calculators, filters, etc.)
	// Wrapped with settings override to apply min_hold_days and sell_cooldown_days from settings
	rawConfigRepo := planningrepo.NewConfigRepository(
		container.ConfigDB,
		log,
	)
	container.PlannerConfigRepo = NewPlannerConfigWithSettingsOverride(
		rawConfigRepo,
		container.SettingsRepo,
	)

	// Planner repository (IN-MEMORY - ephemeral sequences/evaluations/best_results)
	// Sequences and evaluations are generated on-demand and don't need persistence
	container.PlannerRepo = planningrepo.NewInMemoryPlannerRepository(log)

	// History DB client with price filter for read-time anomaly filtering
	// PriceFilter removes anomalous prices (outliers, data errors) at read time
	priceFilter := universe.NewPriceFilter(log)
	container.HistoryDBClient = universe.NewHistoryDB(
		container.HistoryDB.Conn(),
		priceFilter,
		log,
	)

	// Client data repository (needs clientDataDB)
	// Caches client-specific data (symbol mappings, exchange rates, current prices)
	container.ClientDataRepo = clientdata.NewRepository(
		container.ClientDataDB.Conn(),
	)

	log.Info().Msg("All repositories initialized")

	return nil
}
