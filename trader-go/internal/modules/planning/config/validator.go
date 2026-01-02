package config

import (
	"fmt"
	"strings"

	"github.com/aristath/arduino-trader/internal/modules/planning/domain"
)

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// Validator validates planner configurations.
type Validator struct{}

// NewValidator creates a new configuration validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates a planner configuration.
// Returns ValidationErrors if the configuration is invalid.
func (v *Validator) Validate(config *domain.PlannerConfiguration) error {
	var errors ValidationErrors

	// Validate basic fields
	if config.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "name is required",
		})
	}

	// Validate numeric ranges
	if config.MaxDepth <= 0 {
		errors = append(errors, ValidationError{
			Field:   "max_depth",
			Message: "must be greater than 0",
		})
	}

	if config.MaxDepth > 10 {
		errors = append(errors, ValidationError{
			Field:   "max_depth",
			Message: "must be <= 10 (higher values can cause performance issues)",
		})
	}

	if config.MaxOpportunitiesPerCategory <= 0 {
		errors = append(errors, ValidationError{
			Field:   "max_opportunities_per_category",
			Message: "must be greater than 0",
		})
	}

	if config.PriorityThreshold < 0.0 || config.PriorityThreshold > 1.0 {
		errors = append(errors, ValidationError{
			Field:   "priority_threshold",
			Message: "must be between 0.0 and 1.0",
		})
	}

	if config.BeamWidth <= 0 {
		errors = append(errors, ValidationError{
			Field:   "beam_width",
			Message: "must be greater than 0",
		})
	}

	if config.DiversityWeight < 0.0 || config.DiversityWeight > 1.0 {
		errors = append(errors, ValidationError{
			Field:   "diversity_weight",
			Message: "must be between 0.0 and 1.0",
		})
	}

	if config.TransactionCostFixed < 0.0 {
		errors = append(errors, ValidationError{
			Field:   "transaction_cost_fixed",
			Message: "must be >= 0.0",
		})
	}

	if config.TransactionCostPercent < 0.0 {
		errors = append(errors, ValidationError{
			Field:   "transaction_cost_percent",
			Message: "must be >= 0.0",
		})
	}

	// Validate that at least one module is enabled in each category
	enabledCalculators := config.GetEnabledCalculators()
	if len(enabledCalculators) == 0 {
		errors = append(errors, ValidationError{
			Field:   "opportunity_calculators",
			Message: "at least one opportunity calculator must be enabled",
		})
	}

	enabledPatterns := config.GetEnabledPatterns()
	if len(enabledPatterns) == 0 {
		errors = append(errors, ValidationError{
			Field:   "pattern_generators",
			Message: "at least one pattern generator must be enabled",
		})
	}

	enabledGenerators := config.GetEnabledGenerators()
	if len(enabledGenerators) == 0 {
		errors = append(errors, ValidationError{
			Field:   "sequence_generators",
			Message: "at least one sequence generator must be enabled",
		})
	}

	enabledFilters := config.GetEnabledFilters()
	if len(enabledFilters) == 0 {
		errors = append(errors, ValidationError{
			Field:   "filters",
			Message: "at least one filter must be enabled",
		})
	}

	// Validate buy/sell permissions
	if !config.AllowBuy && !config.AllowSell {
		errors = append(errors, ValidationError{
			Field:   "allow_buy/allow_sell",
			Message: "at least one of allow_buy or allow_sell must be true",
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidateQuick performs basic validation without deep checks.
// Useful for quick validation during loading.
func (v *Validator) ValidateQuick(config *domain.PlannerConfiguration) error {
	var errors ValidationErrors

	if config.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "name is required",
		})
	}

	if config.MaxDepth <= 0 {
		errors = append(errors, ValidationError{
			Field:   "max_depth",
			Message: "must be greater than 0",
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidateParams validates module-specific parameters.
// This is a placeholder for future parameter validation logic.
func (v *Validator) ValidateParams(moduleName string, params map[string]interface{}) error {
	// TODO: Implement module-specific parameter validation
	// For now, we accept any parameters
	return nil
}
