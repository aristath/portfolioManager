package repository

import (
	"github.com/aristath/sentinel/internal/modules/planning/domain"
)

// ConfigRepositoryInterface defines the contract for config repository operations
type ConfigRepositoryInterface interface {
	// GetDefaultConfig retrieves the planner configuration (single config exists)
	GetDefaultConfig() (*domain.PlannerConfiguration, error)

	// GetConfig retrieves a configuration by ID (always returns the single config)
	GetConfig(id int64) (*domain.PlannerConfiguration, error)

	// GetConfigByName retrieves a configuration by name (always returns the single config)
	GetConfigByName(name string) (*domain.PlannerConfiguration, error)

	// UpdateConfig updates the planner configuration (single config exists)
	UpdateConfig(
		id int64,
		cfg *domain.PlannerConfiguration,
		changedBy string,
		changeNote string,
	) error

	// CreateConfig creates a new configuration (actually updates the single config)
	CreateConfig(
		cfg *domain.PlannerConfiguration,
		isDefault bool,
	) (int64, error)

	// ListConfigs returns a list of configurations (always returns single config)
	ListConfigs() ([]ConfigRecord, error)

	// DeleteConfig deletes a configuration (resets to defaults)
	DeleteConfig(id int64) error

	// SetDefaultConfig sets a configuration as default (no-op, single config exists)
	SetDefaultConfig(id int64) error

	// GetConfigHistory returns configuration history (empty, no history table)
	GetConfigHistory(configID int64, limit int) ([]ConfigHistoryRecord, error)
}

// Compile-time check that ConfigRepository implements ConfigRepositoryInterface
var _ ConfigRepositoryInterface = (*ConfigRepository)(nil)
