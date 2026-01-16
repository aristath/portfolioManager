// Package universe provides repository implementations for managing the investment universe.
// This file implements the OverrideRepository, which handles user-configurable security overrides
// using an EAV (Entity-Attribute-Value) pattern for flexible field customization.
package universe

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

// OverrideRepository handles security override database operations.
// It uses an EAV (Entity-Attribute-Value) pattern where each override is stored as a row
// in the security_overrides table with columns: isin, field, value.
//
// Supported override fields:
//   - allow_buy: Boolean string ("true"/"false") - whether security can be bought
//   - allow_sell: Boolean string ("true"/"false") - whether security can be sold
//   - min_lot: Integer string - minimum lot size for trading
//   - priority_multiplier: Float string - multiplier for priority scoring
//   - min_portfolio_target: Float string - minimum target allocation percentage (0-100)
//   - max_portfolio_target: Float string - maximum target allocation percentage (0-100)
//
// Overrides are automatically merged into Security objects by SecurityRepository when
// an OverrideRepository is provided.
type OverrideRepository struct {
	universeDB *sql.DB // universe.db - security_overrides table
	log        zerolog.Logger
}

// NewOverrideRepository creates a new override repository.
// The repository manages user-configurable security overrides stored in the security_overrides table.
//
// Parameters:
//   - universeDB: Database connection to universe.db
//   - log: Structured logger
//
// Returns:
//   - *OverrideRepository: Initialized repository instance
func NewOverrideRepository(universeDB *sql.DB, log zerolog.Logger) *OverrideRepository {
	return &OverrideRepository{
		universeDB: universeDB,
		log:        log.With().Str("repo", "override").Logger(),
	}
}

// SetOverride sets or updates an override for a security field.
// If value is empty, the override is deleted (falls back to default value).
// Uses INSERT OR REPLACE to handle both insert and update in a single operation.
//
// Parameters:
//   - isin: Security ISIN (primary key)
//   - field: Override field name (e.g., "allow_buy", "min_lot", "priority_multiplier")
//   - value: Override value as string (will be parsed by ApplyOverrides function)
//
// Returns:
//   - error: Error if database operation fails
func (r *OverrideRepository) SetOverride(isin, field, value string) error {
	if isin == "" || field == "" {
		return fmt.Errorf("isin and field cannot be empty")
	}

	// If value is empty, delete the override (revert to default)
	// This allows users to "unset" an override by passing an empty string
	if value == "" {
		return r.DeleteOverride(isin, field)
	}

	now := time.Now().Unix()

	// Use INSERT OR REPLACE to handle both insert and update
	// ON CONFLICT clause updates existing override if (isin, field) combination already exists
	query := `
		INSERT INTO security_overrides (isin, field, value, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(isin, field) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at
	`

	_, err := r.universeDB.Exec(query, isin, field, value, now, now)
	if err != nil {
		return fmt.Errorf("failed to set override: %w", err)
	}

	r.log.Debug().
		Str("isin", isin).
		Str("field", field).
		Str("value", value).
		Msg("Override set")

	return nil
}

// GetOverrides returns all overrides for a security as a map of field -> value.
// This is used by SecurityRepository to merge overrides into Security objects.
//
// Parameters:
//   - isin: Security ISIN
//
// Returns:
//   - map[string]string: Map of field name to override value (empty map if no overrides)
//   - error: Error if query fails
func (r *OverrideRepository) GetOverrides(isin string) (map[string]string, error) {
	overrides := make(map[string]string)

	query := "SELECT field, value FROM security_overrides WHERE isin = ?"

	rows, err := r.universeDB.Query(query, isin)
	if err != nil {
		return nil, fmt.Errorf("failed to query overrides: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var field, value string
		if err := rows.Scan(&field, &value); err != nil {
			return nil, fmt.Errorf("failed to scan override: %w", err)
		}
		overrides[field] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating overrides: %w", err)
	}

	return overrides, nil
}

// GetAllOverrides returns all overrides for all securities.
// Returns a nested map: ISIN -> field -> value.
// This is used for efficient batch reads when loading multiple securities,
// avoiding N+1 query problems by fetching all overrides in a single query.
//
// Returns:
//   - map[string]map[string]string: Nested map of ISIN to field to value
//   - error: Error if query fails
func (r *OverrideRepository) GetAllOverrides() (map[string]map[string]string, error) {
	allOverrides := make(map[string]map[string]string)

	query := "SELECT isin, field, value FROM security_overrides ORDER BY isin, field"

	rows, err := r.universeDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all overrides: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var isin, field, value string
		if err := rows.Scan(&isin, &field, &value); err != nil {
			return nil, fmt.Errorf("failed to scan override: %w", err)
		}

		// Initialize inner map if this is the first override for this ISIN
		if allOverrides[isin] == nil {
			allOverrides[isin] = make(map[string]string)
		}
		allOverrides[isin][field] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating overrides: %w", err)
	}

	return allOverrides, nil
}

// DeleteOverride removes an override for a specific field.
// This reverts the field to its default value. The operation is idempotent -
// it does not error if the override doesn't exist.
//
// Parameters:
//   - isin: Security ISIN
//   - field: Override field name to delete
//
// Returns:
//   - error: Error if database operation fails
func (r *OverrideRepository) DeleteOverride(isin, field string) error {
	query := "DELETE FROM security_overrides WHERE isin = ? AND field = ?"

	_, err := r.universeDB.Exec(query, isin, field)
	if err != nil {
		return fmt.Errorf("failed to delete override: %w", err)
	}

	r.log.Debug().
		Str("isin", isin).
		Str("field", field).
		Msg("Override deleted")

	return nil
}

// DeleteAllOverrides removes all overrides for a security.
// This reverts all user-configurable fields to their default values.
// Useful when resetting a security to its original configuration.
//
// Parameters:
//   - isin: Security ISIN
//
// Returns:
//   - error: Error if database operation fails
func (r *OverrideRepository) DeleteAllOverrides(isin string) error {
	query := "DELETE FROM security_overrides WHERE isin = ?"

	result, err := r.universeDB.Exec(query, isin)
	if err != nil {
		return fmt.Errorf("failed to delete all overrides: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	r.log.Debug().
		Str("isin", isin).
		Int64("deleted", rowsAffected).
		Msg("All overrides deleted for security")

	return nil
}
