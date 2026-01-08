package calculators

import (
	"github.com/aristath/sentinel/internal/modules/planning/domain"
	"github.com/rs/zerolog"
)

// OpportunityCalculator is the interface that all opportunity calculators must implement.
// Each calculator identifies trading opportunities of a specific type (profit taking,
// averaging down, rebalancing, etc.) based on current portfolio state.
type OpportunityCalculator interface {
	// Name returns the unique identifier for this calculator.
	Name() string

	// Calculate identifies trading opportunities based on the opportunity context.
	// Returns a list of action candidates with priorities and reasons.
	Calculate(ctx *domain.OpportunityContext, params map[string]interface{}) ([]domain.ActionCandidate, error)

	// Category returns the opportunity category this calculator produces.
	Category() domain.OpportunityCategory
}

// BaseCalculator provides common functionality for all calculators.
type BaseCalculator struct {
	log zerolog.Logger
}

// NewBaseCalculator creates a new base calculator with logging.
func NewBaseCalculator(log zerolog.Logger, name string) *BaseCalculator {
	return &BaseCalculator{
		log: log.With().Str("calculator", name).Logger(),
	}
}

// GetFloatParam retrieves a float parameter with a default value.
func GetFloatParam(params map[string]interface{}, key string, defaultValue float64) float64 {
	if params == nil {
		return defaultValue
	}
	if val, ok := params[key]; ok {
		if floatVal, ok := val.(float64); ok {
			return floatVal
		}
		if intVal, ok := val.(int); ok {
			return float64(intVal)
		}
	}
	return defaultValue
}

// GetIntParam retrieves an int parameter with a default value.
func GetIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if params == nil {
		return defaultValue
	}
	if val, ok := params[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
		if floatVal, ok := val.(float64); ok {
			return int(floatVal)
		}
	}
	return defaultValue
}

// GetBoolParam retrieves a bool parameter with a default value.
func GetBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if params == nil {
		return defaultValue
	}
	if val, ok := params[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

// RoundToLotSize intelligently rounds quantity to lot size
// Strategy:
//  1. Try rounding down: floor(quantity/lotSize) * lotSize
//  2. If result is 0 or invalid, try rounding up: ceil(quantity/lotSize) * lotSize
//  3. Return the valid rounded quantity, or 0 if both fail
func RoundToLotSize(quantity int, lotSize int) int {
	if lotSize <= 0 {
		return quantity // No rounding needed
	}

	// Strategy 1: Round down
	roundedDown := (quantity / lotSize) * lotSize

	// If rounding down gives valid result (>= lotSize), use it
	if roundedDown >= lotSize {
		return roundedDown
	}

	// Strategy 2: Round up (only if rounding down failed)
	// Using ceiling: (quantity + lotSize - 1) / lotSize * lotSize
	roundedUp := ((quantity + lotSize - 1) / lotSize) * lotSize

	// Use rounded up if it's valid, otherwise return 0
	if roundedUp >= lotSize {
		return roundedUp
	}

	return 0 // Cannot make valid
}
