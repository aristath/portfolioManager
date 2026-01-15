package universe

// OverrideRepositoryInterface defines the contract for override repository operations
type OverrideRepositoryInterface interface {
	// SetOverride sets or updates an override for a security field
	// If value is empty, the override is deleted (falls back to default)
	SetOverride(isin, field, value string) error

	// GetOverrides returns all overrides for a security as a map of field -> value
	GetOverrides(isin string) (map[string]string, error)

	// GetAllOverrides returns all overrides for all securities
	// Returns a map of ISIN -> field -> value
	GetAllOverrides() (map[string]map[string]string, error)

	// DeleteOverride removes an override for a specific field
	DeleteOverride(isin, field string) error

	// DeleteAllOverrides removes all overrides for a security
	DeleteAllOverrides(isin string) error
}

// Compile-time check that OverrideRepository implements OverrideRepositoryInterface
var _ OverrideRepositoryInterface = (*OverrideRepository)(nil)
