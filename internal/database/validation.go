// Package database provides validation functionality for database migrations.
// Specifically, it validates ISIN (International Securities Identification Number)
// requirements before migrating from symbol-based to ISIN-based primary keys.
package database

import (
	"database/sql"
	"fmt"
	"strings"
)

// ISINValidator validates ISIN requirements before migration.
// It checks that all securities have ISINs, no duplicates exist, and all
// foreign key references are valid. Used during the migration from symbol-based
// to ISIN-based primary keys (migration 031).
type ISINValidator struct {
	db *sql.DB // Database connection for validation queries
}

// ValidationResult contains the results of all validation checks.
// Used to report validation status and any issues found.
type ValidationResult struct {
	IsValid            bool     // True if all validations pass
	MissingISINs       []string // Securities without ISIN (symbols)
	DuplicateISINs     []string // Duplicate ISIN values found
	OrphanedReferences []string // Foreign key references to non-existent securities (format: "table:column:value")
}

// NewISINValidator creates a new ISIN validator.
// The validator uses the provided database connection to perform validation queries.
//
// Parameters:
//   - db: Database connection (typically universe.db)
//
// Returns:
//   - *ISINValidator: Initialized validator
func NewISINValidator(db *sql.DB) *ISINValidator {
	return &ISINValidator{
		db: db,
	}
}

// ValidateAllSecuritiesHaveISIN checks that all securities have a non-empty ISIN.
// This is required before migration because ISIN becomes the primary key.
// Returns list of symbols that are missing ISIN.
//
// Returns:
//   - []string: List of symbols missing ISIN (empty if all have ISIN)
//   - error: Error if query fails
func (v *ISINValidator) ValidateAllSecuritiesHaveISIN() ([]string, error) {
	// Query for securities with NULL, empty, or whitespace-only ISIN
	query := `
		SELECT symbol
		FROM securities
		WHERE isin IS NULL OR isin = '' OR TRIM(isin) = ''
	`

	rows, err := v.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query securities: %w", err)
	}
	defer rows.Close()

	var missingISINs []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, fmt.Errorf("failed to scan symbol: %w", err)
		}
		missingISINs = append(missingISINs, symbol)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return missingISINs, nil
}

// ValidateNoDuplicateISINs checks that no two securities share the same ISIN.
// This is required because ISIN becomes the primary key (must be unique).
// Returns list of duplicate ISIN values found.
//
// Returns:
//   - []string: List of duplicate ISIN values (empty if no duplicates)
//   - error: Error if query fails
func (v *ISINValidator) ValidateNoDuplicateISINs() ([]string, error) {
	// Query for ISINs that appear more than once
	query := `
		SELECT isin, COUNT(*) as count
		FROM securities
		WHERE isin IS NOT NULL AND isin != '' AND TRIM(isin) != ''
		GROUP BY isin
		HAVING COUNT(*) > 1
	`

	rows, err := v.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query duplicate ISINs: %w", err)
	}
	defer rows.Close()

	var duplicateISINs []string
	for rows.Next() {
		var isin string
		var count int
		if err := rows.Scan(&isin, &count); err != nil {
			return nil, fmt.Errorf("failed to scan duplicate ISIN: %w", err)
		}
		duplicateISINs = append(duplicateISINs, isin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return duplicateISINs, nil
}

// ValidateForeignKeys checks that all foreign key references point to existing securities.
// This validates that scores, positions, trades, etc. reference valid securities by ISIN.
// This is critical because foreign key constraints will be enforced after migration.
// Returns list of orphaned references (format: "table:column:value").
//
// Returns:
//   - []string: List of orphaned references (empty if all references are valid)
//   - error: Error if query fails
func (v *ISINValidator) ValidateForeignKeys() ([]string, error) {
	var errors []string

	// Check scores table (now uses isin as PRIMARY KEY)
	// Find scores that reference non-existent securities
	scoreQuery := `
		SELECT s.isin
		FROM scores s
		LEFT JOIN securities sec ON s.isin = sec.isin
		WHERE sec.isin IS NULL
	`
	rows, err := v.db.Query(scoreQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query orphaned scores: %w", err)
	}
	for rows.Next() {
		var isin string
		if err := rows.Scan(&isin); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan orphaned score: %w", err)
		}
		errors = append(errors, fmt.Sprintf("scores:isin:%s", isin))
	}
	rows.Close()

	// Check positions table (now uses isin as PRIMARY KEY)
	// Find positions that reference non-existent securities
	positionQuery := `
		SELECT p.isin
		FROM positions p
		LEFT JOIN securities sec ON p.isin = sec.isin
		WHERE sec.isin IS NULL
	`
	rows, err = v.db.Query(positionQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query orphaned positions: %w", err)
	}
	for rows.Next() {
		var isin string
		if err := rows.Scan(&isin); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan orphaned position: %w", err)
		}
		errors = append(errors, fmt.Sprintf("positions:isin:%s", isin))
	}
	rows.Close()

	return errors, nil
}

// ValidateAll runs all validation checks and returns a comprehensive result.
// This is the main entry point for validation before migration.
// It runs all three validation checks and aggregates the results.
//
// Returns:
//   - *ValidationResult: Comprehensive validation results
//   - error: Error if any validation check fails to execute
func (v *ISINValidator) ValidateAll() (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:            true,
		MissingISINs:       []string{},
		DuplicateISINs:     []string{},
		OrphanedReferences: []string{},
	}

	// Check for missing ISINs
	// All securities must have ISIN before migration
	missingISINs, err := v.ValidateAllSecuritiesHaveISIN()
	if err != nil {
		return nil, fmt.Errorf("failed to validate ISIN presence: %w", err)
	}
	result.MissingISINs = missingISINs
	if len(missingISINs) > 0 {
		result.IsValid = false
	}

	// Check for duplicate ISINs
	// ISIN must be unique because it becomes the primary key
	duplicateISINs, err := v.ValidateNoDuplicateISINs()
	if err != nil {
		return nil, fmt.Errorf("failed to validate duplicate ISINs: %w", err)
	}
	result.DuplicateISINs = duplicateISINs
	if len(duplicateISINs) > 0 {
		result.IsValid = false
	}

	// Check foreign keys
	// All foreign key references must point to existing securities
	orphanedRefs, err := v.ValidateForeignKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to validate foreign keys: %w", err)
	}
	result.OrphanedReferences = orphanedRefs
	if len(orphanedRefs) > 0 {
		result.IsValid = false
	}

	return result, nil
}

// FormatErrors formats validation errors for display.
// Returns a human-readable string summarizing all validation issues found.
// If all validations passed, returns a success message.
//
// Returns:
//   - string: Formatted error message or success message
func (r *ValidationResult) FormatErrors() string {
	if r.IsValid {
		return "All validations passed"
	}

	var parts []string

	// Format missing ISINs
	if len(r.MissingISINs) > 0 {
		parts = append(parts, fmt.Sprintf("Missing ISINs (%d): %s", len(r.MissingISINs), strings.Join(r.MissingISINs, ", ")))
	}

	// Format duplicate ISINs
	if len(r.DuplicateISINs) > 0 {
		parts = append(parts, fmt.Sprintf("Duplicate ISINs (%d): %s", len(r.DuplicateISINs), strings.Join(r.DuplicateISINs, ", ")))
	}

	// Format orphaned references
	if len(r.OrphanedReferences) > 0 {
		parts = append(parts, fmt.Sprintf("Orphaned references (%d): %s", len(r.OrphanedReferences), strings.Join(r.OrphanedReferences, ", ")))
	}

	return strings.Join(parts, "\n")
}
