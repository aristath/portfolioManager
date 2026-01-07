package patterns

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"

	"github.com/aristath/sentinel/internal/modules/planning/domain"
	"github.com/rs/zerolog"
)

// PatternGenerator is the interface that all pattern generators must implement.
// Each generator creates action sequences from identified opportunities using
// different strategic patterns (direct buy, rebalance, profit-taking, etc.).
type PatternGenerator interface {
	// Name returns the unique identifier for this pattern generator.
	Name() string

	// Generate creates action sequences from the given opportunities.
	// Returns a list of sequences with associated metadata.
	Generate(opportunities domain.OpportunitiesByCategory, params map[string]interface{}) ([]domain.ActionSequence, error)
}

// BasePattern provides common functionality for all pattern generators.
type BasePattern struct {
	log zerolog.Logger
}

// NewBasePattern creates a new base pattern with logging.
func NewBasePattern(log zerolog.Logger, name string) *BasePattern {
	return &BasePattern{
		log: log.With().Str("pattern", name).Logger(),
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

// CreateSequence is a helper to create an ActionSequence from a list of actions.
func CreateSequence(actions []domain.ActionCandidate, patternType string) domain.ActionSequence {
	// Calculate aggregate priority (average)
	priority := 0.0
	if len(actions) > 0 {
		for _, action := range actions {
			priority += action.Priority
		}
		priority /= float64(len(actions))
	}

	// Generate sequence hash (simple implementation for now)
	sequenceHash := generateSequenceHash(actions)

	return domain.ActionSequence{
		Actions:      actions,
		Priority:     priority,
		Depth:        len(actions),
		PatternType:  patternType,
		SequenceHash: sequenceHash,
	}
}

// generateSequenceHash creates a deterministic MD5 hash for a sequence.
// Matches evaluation service hashSequence() and legacy Python implementation.
// Based on: (symbol, side, quantity) tuples, order-dependent
func generateSequenceHash(actions []domain.ActionCandidate) string {
	// Import crypto/md5 and encoding/hex at package level
	type tuple struct {
		Symbol   string `json:"symbol"`
		Side     string `json:"side"`
		Quantity int    `json:"quantity"`
	}

	// Create tuples matching Python: [(c.symbol, c.side, c.quantity) for c in sequence]
	tuples := make([]tuple, len(actions))
	for i, action := range actions {
		tuples[i] = tuple{
			Symbol:   action.Symbol,
			Side:     action.Side,
			Quantity: action.Quantity,
		}
	}

	// JSON marshal (Go's json.Marshal preserves order by default, like sort_keys=False)
	jsonBytes, err := json.Marshal(tuples)
	if err != nil {
		// Fallback: should not happen, but handle gracefully
		return ""
	}

	// MD5 hash and return hex digest (matches hashlib.md5().hexdigest())
	hash := md5.Sum(jsonBytes)
	return hex.EncodeToString(hash[:])
}
