package market_regime

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

// IndexSyncService ensures market indices exist in the securities table
// and have historical price data available for regime calculation.
type IndexSyncService struct {
	universeDB *sql.DB
	configDB   *sql.DB
	log        zerolog.Logger
}

// NewIndexSyncService creates a new index sync service
func NewIndexSyncService(universeDB, configDB *sql.DB, log zerolog.Logger) *IndexSyncService {
	return &IndexSyncService{
		universeDB: universeDB,
		configDB:   configDB,
		log:        log.With().Str("component", "index_sync_service").Logger(),
	}
}

// SyncIndicesToSecurities ensures all known indices exist in the securities table.
// This is idempotent - running multiple times won't duplicate data.
// Indices are created with:
//   - ISIN: "INDEX-{SYMBOL}" format
//   - product_type: "INDEX"
//   - active: 1
//   - allow_buy: 0 (indices are not tradeable)
//   - allow_sell: 0
//   - market_code: from the known index config
func (s *IndexSyncService) SyncIndicesToSecurities() error {
	knownIndices := GetKnownIndices()
	now := time.Now().Unix()

	synced := 0
	skipped := 0

	for _, idx := range knownIndices {
		// Generate ISIN for index (indices don't have real ISINs)
		isin := fmt.Sprintf("INDEX-%s", idx.Symbol)

		// Check if already exists
		var exists bool
		err := s.universeDB.QueryRow(`
			SELECT COUNT(*) > 0 FROM securities
			WHERE isin = ?
		`, isin).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check index existence for %s: %w", idx.Symbol, err)
		}

		if exists {
			skipped++
			continue
		}

		// Create index in securities table
		_, err = s.universeDB.Exec(`
			INSERT INTO securities
			(isin, symbol, name, product_type, market_code, active, allow_buy, allow_sell, created_at, updated_at)
			VALUES (?, ?, ?, 'INDEX', ?, 1, 0, 0, ?, ?)
		`, isin, idx.Symbol, idx.Name, idx.MarketCode, now, now)
		if err != nil {
			return fmt.Errorf("failed to create index %s in securities: %w", idx.Symbol, err)
		}

		synced++
		s.log.Debug().
			Str("symbol", idx.Symbol).
			Str("isin", isin).
			Str("region", idx.Region).
			Msg("Created index in securities table")
	}

	s.log.Info().
		Int("synced", synced).
		Int("skipped", skipped).
		Int("total", len(knownIndices)).
		Msg("Synced indices to securities table")

	return nil
}

// SyncAll performs a full sync:
// 1. Syncs known indices to market_indices table (config DB)
// 2. Syncs indices to securities table (universe DB)
func (s *IndexSyncService) SyncAll() error {
	// Step 1: Sync to market_indices (config DB)
	if s.configDB != nil {
		indexRepo := NewIndexRepository(s.configDB, s.log)
		if err := indexRepo.SyncFromKnownIndices(); err != nil {
			return fmt.Errorf("failed to sync to market_indices: %w", err)
		}
	}

	// Step 2: Sync to securities (universe DB)
	if err := s.SyncIndicesToSecurities(); err != nil {
		return fmt.Errorf("failed to sync to securities: %w", err)
	}

	return nil
}

// GetIndicesWithISIN returns all known indices with their generated ISINs
func (s *IndexSyncService) GetIndicesWithISIN() []struct {
	Symbol string
	ISIN   string
	Region string
} {
	knownIndices := GetKnownIndices()
	result := make([]struct {
		Symbol string
		ISIN   string
		Region string
	}, 0, len(knownIndices))

	for _, idx := range knownIndices {
		// Only include PRICE indices, not VOLATILITY
		if idx.IndexType != IndexTypePrice {
			continue
		}
		result = append(result, struct {
			Symbol string
			ISIN   string
			Region string
		}{
			Symbol: idx.Symbol,
			ISIN:   fmt.Sprintf("INDEX-%s", idx.Symbol),
			Region: idx.Region,
		})
	}

	return result
}

// HistoricalPriceSyncer interface for syncing historical prices
// This allows injecting HistoricalSyncService without circular dependency
type HistoricalPriceSyncer interface {
	SyncHistoricalPrices(symbol string) error
}

// SyncHistoricalPricesForIndices syncs historical price data for all known PRICE indices.
// This must be called after SyncAll() to ensure indices exist in the securities table.
// Uses the provided syncer (typically HistoricalSyncService) to fetch prices from broker API.
func (s *IndexSyncService) SyncHistoricalPricesForIndices(syncer HistoricalPriceSyncer) error {
	if syncer == nil {
		return fmt.Errorf("historical price syncer is nil")
	}

	knownIndices := GetKnownIndices()
	synced := 0
	failed := 0

	for _, idx := range knownIndices {
		// Only sync PRICE indices (not VOLATILITY like VIX)
		if idx.IndexType != IndexTypePrice {
			continue
		}

		s.log.Debug().
			Str("symbol", idx.Symbol).
			Str("region", idx.Region).
			Msg("Syncing historical prices for index")

		if err := syncer.SyncHistoricalPrices(idx.Symbol); err != nil {
			s.log.Warn().
				Err(err).
				Str("symbol", idx.Symbol).
				Msg("Failed to sync historical prices for index")
			failed++
			continue
		}

		synced++
	}

	s.log.Info().
		Int("synced", synced).
		Int("failed", failed).
		Int("total", len(knownIndices)).
		Msg("Completed historical price sync for indices")

	// Return error only if ALL indices failed
	if synced == 0 && failed > 0 {
		return fmt.Errorf("failed to sync historical prices for any index")
	}

	return nil
}

// EnsureIndexExists checks if a specific index exists in securities table,
// creating it if necessary. Returns the ISIN.
func (s *IndexSyncService) EnsureIndexExists(symbol string) (string, error) {
	// Find the index in known indices
	var foundIdx *KnownIndex
	for _, idx := range GetKnownIndices() {
		if idx.Symbol == symbol {
			foundIdx = &idx
			break
		}
	}

	if foundIdx == nil {
		return "", fmt.Errorf("unknown index symbol: %s", symbol)
	}

	isin := fmt.Sprintf("INDEX-%s", symbol)
	now := time.Now().Unix()

	// Upsert the index
	_, err := s.universeDB.Exec(`
		INSERT INTO securities
		(isin, symbol, name, product_type, market_code, active, allow_buy, allow_sell, created_at, updated_at)
		VALUES (?, ?, ?, 'INDEX', ?, 1, 0, 0, ?, ?)
		ON CONFLICT(isin) DO UPDATE SET
			name = excluded.name,
			market_code = excluded.market_code,
			updated_at = excluded.updated_at
	`, isin, symbol, foundIdx.Name, foundIdx.MarketCode, now, now)
	if err != nil {
		return "", fmt.Errorf("failed to upsert index %s: %w", symbol, err)
	}

	return isin, nil
}
