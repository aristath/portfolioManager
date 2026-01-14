package market_regime

import (
	"database/sql"
	"time"

	"github.com/rs/zerolog"
)

// IndexRepository provides database operations for market indices
// Uses MarketIndex type defined in index_service.go
type IndexRepository struct {
	db  *sql.DB
	log zerolog.Logger
}

// NewIndexRepository creates a new index repository
func NewIndexRepository(db *sql.DB, log zerolog.Logger) *IndexRepository {
	return &IndexRepository{
		db:  db,
		log: log.With().Str("repository", "index").Logger(),
	}
}

// Upsert creates or updates a market index
func (r *IndexRepository) Upsert(index MarketIndex) error {
	now := time.Now().Unix()

	query := `
		INSERT INTO market_indices (symbol, name, market_code, region, index_type, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(symbol) DO UPDATE SET
			name = excluded.name,
			market_code = excluded.market_code,
			region = excluded.region,
			index_type = excluded.index_type,
			enabled = excluded.enabled,
			updated_at = excluded.updated_at
	`

	enabled := 0
	if index.Enabled {
		enabled = 1
	}

	_, err := r.db.Exec(query, index.Symbol, index.Name, index.MarketCode, index.Region, index.IndexType, enabled, now, now)
	if err != nil {
		r.log.Error().Err(err).Str("symbol", index.Symbol).Msg("Failed to upsert market index")
		return err
	}

	return nil
}

// GetBySymbol retrieves a market index by symbol
func (r *IndexRepository) GetBySymbol(symbol string) (*MarketIndex, error) {
	query := `
		SELECT symbol, name, market_code, region, index_type, enabled, created_at, updated_at
		FROM market_indices
		WHERE symbol = ?
	`

	var index MarketIndex
	var enabled int

	err := r.db.QueryRow(query, symbol).Scan(
		&index.Symbol, &index.Name, &index.MarketCode, &index.Region,
		&index.IndexType, &enabled, &index.CreatedAt, &index.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error().Err(err).Str("symbol", symbol).Msg("Failed to get market index by symbol")
		return nil, err
	}

	index.Enabled = enabled == 1
	return &index, nil
}

// GetEnabledByRegion retrieves all enabled indices for a region
func (r *IndexRepository) GetEnabledByRegion(region string) ([]MarketIndex, error) {
	query := `
		SELECT symbol, name, market_code, region, index_type, enabled, created_at, updated_at
		FROM market_indices
		WHERE region = ? AND enabled = 1
		ORDER BY symbol
	`

	return r.queryIndices(query, region)
}

// GetEnabledPriceByRegion retrieves only enabled PRICE-type indices for a region
// Excludes VOLATILITY indices (VIX) from the result
func (r *IndexRepository) GetEnabledPriceByRegion(region string) ([]MarketIndex, error) {
	query := `
		SELECT symbol, name, market_code, region, index_type, enabled, created_at, updated_at
		FROM market_indices
		WHERE region = ? AND enabled = 1 AND index_type = 'PRICE'
		ORDER BY symbol
	`

	return r.queryIndices(query, region)
}

// GetAllEnabled retrieves all enabled indices across all regions
func (r *IndexRepository) GetAllEnabled() ([]MarketIndex, error) {
	query := `
		SELECT symbol, name, market_code, region, index_type, enabled, created_at, updated_at
		FROM market_indices
		WHERE enabled = 1
		ORDER BY region, symbol
	`

	return r.queryIndices(query)
}

// SyncFromKnownIndices ensures all known indices are in the database
// This is idempotent - running multiple times won't duplicate data
func (r *IndexRepository) SyncFromKnownIndices() error {
	knownIndices := GetKnownIndices()

	for _, known := range knownIndices {
		index := MarketIndex{
			Symbol:     known.Symbol,
			Name:       known.Name,
			MarketCode: known.MarketCode,
			Region:     known.Region,
			IndexType:  known.IndexType,
			Enabled:    true, // All known indices are enabled by default
		}

		if err := r.Upsert(index); err != nil {
			return err
		}
	}

	r.log.Info().Int("count", len(knownIndices)).Msg("Synced known indices to database")
	return nil
}

// Delete removes a market index by symbol
func (r *IndexRepository) Delete(symbol string) error {
	query := `DELETE FROM market_indices WHERE symbol = ?`

	_, err := r.db.Exec(query, symbol)
	if err != nil {
		r.log.Error().Err(err).Str("symbol", symbol).Msg("Failed to delete market index")
		return err
	}

	return nil
}

// queryIndices is a helper that runs a query and returns indices
func (r *IndexRepository) queryIndices(query string, args ...interface{}) ([]MarketIndex, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		r.log.Error().Err(err).Msg("Failed to query market indices")
		return nil, err
	}
	defer rows.Close()

	var indices []MarketIndex
	for rows.Next() {
		var index MarketIndex
		var enabled int

		err := rows.Scan(
			&index.Symbol, &index.Name, &index.MarketCode, &index.Region,
			&index.IndexType, &enabled, &index.CreatedAt, &index.UpdatedAt,
		)
		if err != nil {
			r.log.Error().Err(err).Msg("Failed to scan market index row")
			return nil, err
		}

		index.Enabled = enabled == 1
		indices = append(indices, index)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return indices, nil
}
