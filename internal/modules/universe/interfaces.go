package universe

import (
	"database/sql"
	"time"
)

// ProgressReporter interface for progress reporting (avoids import cycle)
type ProgressReporter interface {
	Report(current, total int, message string)
}

// SyncServiceInterface defines the contract for sync service operations
// Used by UniverseService to enable testing with mocks
type SyncServiceInterface interface {
	SyncAllPrices() (int, error)
	SyncAllPricesWithReporter(reporter ProgressReporter) (int, error)
	SyncPricesForSymbols(symbolMap map[string]*string) (int, error)
}

// SecurityRepositoryInterface defines the contract for security repository operations
// Used by UniverseService to enable testing with mocks
type SecurityRepositoryInterface interface {
	// GetBySymbol returns a security by symbol
	GetBySymbol(symbol string) (*Security, error)

	// GetByISIN returns a security by ISIN
	GetByISIN(isin string) (*Security, error)

	// GetByIdentifier returns a security by symbol or ISIN (smart lookup)
	GetByIdentifier(identifier string) (*Security, error)

	// GetAllActive returns all active securities
	GetAllActive() ([]Security, error)

	// GetDistinctExchanges returns all distinct exchange names
	GetDistinctExchanges() ([]string, error)

	// GetAllActiveTradable returns all active and tradable securities
	GetAllActiveTradable() ([]Security, error)

	// GetAll returns all securities (active and inactive)
	GetAll() ([]Security, error)

	// Create creates a new security
	Create(security Security) error

	// Update updates a security by ISIN
	Update(isin string, updates map[string]interface{}) error

	// Delete deletes a security by ISIN
	Delete(isin string) error

	// GetWithScores returns securities with their scores joined
	GetWithScores(portfolioDB *sql.DB) ([]SecurityWithScore, error)

	// SetTagsForSecurity replaces all tags for a security (deletes existing, inserts new)
	// symbol parameter is kept for backward compatibility, but we look up ISIN internally
	SetTagsForSecurity(symbol string, tagIDs []string) error

	// GetTagsForSecurity returns all tag IDs for a security (public method)
	// symbol parameter is kept for backward compatibility, but we look up ISIN internally
	GetTagsForSecurity(symbol string) ([]string, error)

	// GetTagsWithUpdateTimes returns all tags for a security with their last update times
	// symbol parameter is kept for backward compatibility, but we look up ISIN internally
	GetTagsWithUpdateTimes(symbol string) (map[string]time.Time, error)

	// UpdateSpecificTags updates only the specified tags for a security, preserving other tags
	// symbol parameter is kept for backward compatibility, but we look up ISIN internally
	UpdateSpecificTags(symbol string, tagIDs []string) error

	// GetByTags returns active securities matching any of the provided tags
	GetByTags(tagIDs []string) ([]Security, error)

	// GetPositionsByTags returns securities that are in the provided position symbols AND have the specified tags
	GetPositionsByTags(positionSymbols []string, tagIDs []string) ([]Security, error)

	// HardDelete permanently removes a security and all related data from universe.db
	HardDelete(isin string) error
}

// Compile-time check that SecurityRepository implements SecurityRepositoryInterface
var _ SecurityRepositoryInterface = (*SecurityRepository)(nil)

// SecurityDeletionServiceInterface defines the contract for hard deletion operations
type SecurityDeletionServiceInterface interface {
	// HardDelete permanently removes a security and all related data across databases
	// Returns error if security has open positions, pending orders, or does not exist
	HardDelete(isin string) error
}

// Compile-time check that SecurityDeletionService implements SecurityDeletionServiceInterface
var _ SecurityDeletionServiceInterface = (*SecurityDeletionService)(nil)

// DBExecutor defines the contract for database execution operations
// Used by SyncService to enable testing with mocks
type DBExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// Note: CurrencyExchangeServiceInterface has been moved to domain/interfaces.go
// It is now available as domain.CurrencyExchangeServiceInterface
