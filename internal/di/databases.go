/**
 * Package di provides dependency injection for database connections.
 *
 * This package initializes all 8 databases in the Sentinel architecture.
 * Each database uses SQLite with WAL mode and profile-specific PRAGMAs
 * for optimal performance and safety.
 */
package di

import (
	"fmt"

	"github.com/aristath/sentinel/internal/config"
	"github.com/aristath/sentinel/internal/database"
	"github.com/rs/zerolog"
)

/**
 * InitializeDatabases initializes all 8 databases and applies schemas.
 *
 * This function creates database connections for the 8-database architecture:
 * 1. universe.db - Investment universe (securities, groups)
 * 2. config.db - Application configuration (settings, allocation targets)
 * 3. ledger.db - Immutable financial audit trail (trades, cash flows, dividends)
 * 4. portfolio.db - Current portfolio state (positions, scores, metrics, snapshots)
 * 5. history.db - Historical time-series data (prices, rates, cleanup tracking)
 * 6. cache.db - Ephemeral operational data (job history)
 * 7. client_data.db - Cache for exchange rates and current prices
 * 8. calculations.db - Calculation cache (technical indicators, optimizer results)
 *
 * Each database uses a profile-specific configuration:
 * - ProfileStandard: Balanced performance and safety (universe, config, portfolio, history)
 * - ProfileLedger: Maximum safety for immutable audit trail (ledger)
 * - ProfileCache: Maximum speed for ephemeral data (cache, client_data, calculations)
 *
 * The function ensures proper cleanup on error by closing all databases
 * that were successfully initialized before the error occurred.
 *
 * @param cfg - Application configuration (contains DataDir path)
 * @param log - Structured logger instance
 * @returns *Container - Container with initialized database connections
 * @returns error - Error if database initialization fails
 */
func InitializeDatabases(cfg *config.Config, log zerolog.Logger) (*Container, error) {
	container := &Container{}

	// 1. universe.db - Investment universe (securities, groups)
	universeDB, err := database.New(database.Config{
		Path:    cfg.DataDir + "/universe.db",
		Profile: database.ProfileStandard,
		Name:    "universe",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize universe database: %w", err)
	}
	container.UniverseDB = universeDB

	// 2. config.db - Application configuration (settings, allocation targets)
	configDB, err := database.New(database.Config{
		Path:    cfg.DataDir + "/config.db",
		Profile: database.ProfileStandard,
		Name:    "config",
	})
	if err != nil {
		universeDB.Close()
		return nil, fmt.Errorf("failed to initialize config database: %w", err)
	}
	container.ConfigDB = configDB

	// 3. ledger.db - Immutable financial audit trail (trades, cash flows, dividends)
	// Uses ProfileLedger for maximum safety (synchronous writes, full integrity checks)
	ledgerDB, err := database.New(database.Config{
		Path:    cfg.DataDir + "/ledger.db",
		Profile: database.ProfileLedger, // Maximum safety for immutable audit trail
		Name:    "ledger",
	})
	if err != nil {
		universeDB.Close()
		configDB.Close()
		return nil, fmt.Errorf("failed to initialize ledger database: %w", err)
	}
	container.LedgerDB = ledgerDB

	// 4. portfolio.db - Current portfolio state (positions, scores, metrics, snapshots)
	portfolioDB, err := database.New(database.Config{
		Path:    cfg.DataDir + "/portfolio.db",
		Profile: database.ProfileStandard,
		Name:    "portfolio",
	})
	if err != nil {
		universeDB.Close()
		configDB.Close()
		ledgerDB.Close()
		return nil, fmt.Errorf("failed to initialize portfolio database: %w", err)
	}
	container.PortfolioDB = portfolioDB

	// 5. history.db - Historical time-series data (prices, rates, cleanup tracking)
	historyDB, err := database.New(database.Config{
		Path:    cfg.DataDir + "/history.db",
		Profile: database.ProfileStandard,
		Name:    "history",
	})
	if err != nil {
		universeDB.Close()
		configDB.Close()
		ledgerDB.Close()
		portfolioDB.Close()
		return nil, fmt.Errorf("failed to initialize history database: %w", err)
	}
	container.HistoryDB = historyDB

	// 6. cache.db - Ephemeral operational data (job history)
	// Uses ProfileCache for maximum speed (can tolerate data loss)
	cacheDB, err := database.New(database.Config{
		Path:    cfg.DataDir + "/cache.db",
		Profile: database.ProfileCache, // Maximum speed for ephemeral data
		Name:    "cache",
	})
	if err != nil {
		universeDB.Close()
		configDB.Close()
		ledgerDB.Close()
		portfolioDB.Close()
		historyDB.Close()
		return nil, fmt.Errorf("failed to initialize cache database: %w", err)
	}
	container.CacheDB = cacheDB

	// 7. client_data.db - Client-specific symbol mappings and cached data (Tradernet)
	// Uses ProfileCache for maximum speed (cache can be rebuilt)
	clientDataDB, err := database.New(database.Config{
		Path:    cfg.DataDir + "/client_data.db",
		Profile: database.ProfileCache, // Maximum speed for cache data
		Name:    "client_data",
	})
	if err != nil {
		universeDB.Close()
		configDB.Close()
		ledgerDB.Close()
		portfolioDB.Close()
		historyDB.Close()
		cacheDB.Close()
		return nil, fmt.Errorf("failed to initialize client_data database: %w", err)
	}
	container.ClientDataDB = clientDataDB

	// 8. calculations.db - Calculation cache (technical indicators, optimizer results)
	// Uses ProfileCache for maximum speed (calculations can be recomputed)
	calculationsDB, err := database.New(database.Config{
		Path:    cfg.DataDir + "/calculations.db",
		Profile: database.ProfileCache, // Maximum speed for cache data
		Name:    "calculations",
	})
	if err != nil {
		universeDB.Close()
		configDB.Close()
		ledgerDB.Close()
		portfolioDB.Close()
		historyDB.Close()
		cacheDB.Close()
		clientDataDB.Close()
		return nil, fmt.Errorf("failed to initialize calculations database: %w", err)
	}
	container.CalculationsDB = calculationsDB

	// Apply schemas to all databases (single source of truth)
	// Migration system ensures all databases have the correct schema version
	for _, db := range []*database.DB{universeDB, configDB, ledgerDB, portfolioDB, historyDB, cacheDB, clientDataDB, calculationsDB} {
		if err := db.Migrate(); err != nil {
			// Cleanup on error - close all databases that were successfully opened
			universeDB.Close()
			configDB.Close()
			ledgerDB.Close()
			portfolioDB.Close()
			historyDB.Close()
			cacheDB.Close()
			clientDataDB.Close()
			calculationsDB.Close()
			return nil, fmt.Errorf("failed to apply schema to %s: %w", db.Name(), err)
		}
	}

	log.Info().Msg("All databases initialized and schemas applied")

	return container, nil
}
