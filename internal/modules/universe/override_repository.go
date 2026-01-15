package universe

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

// OverrideRepository handles security override database operations
// Uses EAV (Entity-Attribute-Value) pattern for flexible field overrides
type OverrideRepository struct {
	universeDB *sql.DB // universe.db - security_overrides table
	log        zerolog.Logger
}

// NewOverrideRepository creates a new override repository
func NewOverrideRepository(universeDB *sql.DB, log zerolog.Logger) *OverrideRepository {
	return &OverrideRepository{
		universeDB: universeDB,
		log:        log.With().Str("repo", "override").Logger(),
	}
}

// SetOverride sets or updates an override for a security field
// If value is empty, the override is deleted (falls back to default)
func (r *OverrideRepository) SetOverride(isin, field, value string) error {
	if isin == "" || field == "" {
		return fmt.Errorf("isin and field cannot be empty")
	}

	// If value is empty, delete the override
	if value == "" {
		return r.DeleteOverride(isin, field)
	}

	now := time.Now().Unix()

	// Use INSERT OR REPLACE to handle both insert and update
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

// GetOverrides returns all overrides for a security as a map of field -> value
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

// GetAllOverrides returns all overrides for all securities
// Returns a map of ISIN -> field -> value
// Used for efficient batch reads when loading multiple securities
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

// DeleteOverride removes an override for a specific field
// Idempotent - does not error if override doesn't exist
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

// DeleteAllOverrides removes all overrides for a security
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
