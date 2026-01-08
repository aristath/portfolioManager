package universe

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringValue(t *testing.T) {
	tests := []struct {
		name     string
		ptr      *string
		expected string
	}{
		{
			name:     "nil pointer",
			ptr:      nil,
			expected: "",
		},
		{
			name:     "empty string",
			ptr:      stringPtr(""),
			expected: "",
		},
		{
			name:     "non-empty string",
			ptr:      stringPtr("test"),
			expected: "test",
		},
		{
			name:     "string with spaces",
			ptr:      stringPtr("test string"),
			expected: "test string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringValue(tt.ptr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNullString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected sql.NullString
	}{
		{
			name:     "empty string",
			input:    "",
			expected: sql.NullString{Valid: false},
		},
		{
			name:     "non-empty string",
			input:    "test",
			expected: sql.NullString{String: "test", Valid: true},
		},
		{
			name:     "string with spaces",
			input:    "test string",
			expected: sql.NullString{String: "test string", Valid: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nullString(tt.input)
			assert.Equal(t, tt.expected.Valid, result.Valid)
			if tt.expected.Valid {
				assert.Equal(t, tt.expected.String, result.String)
			}
		})
	}
}

func TestNullFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected sql.NullFloat64
	}{
		{
			name:     "zero value",
			input:    0.0,
			expected: sql.NullFloat64{Valid: false},
		},
		{
			name:     "positive value",
			input:    3.14,
			expected: sql.NullFloat64{Float64: 3.14, Valid: true},
		},
		{
			name:     "negative value",
			input:    -5.5,
			expected: sql.NullFloat64{Float64: -5.5, Valid: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nullFloat64(tt.input)
			assert.Equal(t, tt.expected.Valid, result.Valid)
			if tt.expected.Valid {
				assert.InDelta(t, tt.expected.Float64, result.Float64, 0.0001)
			}
		})
	}
}

func TestBoolToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected int
	}{
		{
			name:     "true",
			input:    true,
			expected: 1,
		},
		{
			name:     "false",
			input:    false,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := boolToInt(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetScore(t *testing.T) {
	tests := []struct {
		name     string
		scores   map[string]float64
		key      string
		expected float64
	}{
		{
			name:     "nil map",
			scores:   nil,
			key:      "key1",
			expected: 0.0,
		},
		{
			name:     "empty map",
			scores:   map[string]float64{},
			key:      "key1",
			expected: 0.0,
		},
		{
			name:     "key exists",
			scores:   map[string]float64{"key1": 0.85, "key2": 0.65},
			key:      "key1",
			expected: 0.85,
		},
		{
			name:     "key does not exist",
			scores:   map[string]float64{"key1": 0.85},
			key:      "key2",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getScore(tt.scores, tt.key)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}

func TestGetSubScore(t *testing.T) {
	tests := []struct {
		name      string
		subScores map[string]map[string]float64
		group     string
		key       string
		expected  float64
	}{
		{
			name:      "nil map",
			subScores: nil,
			group:     "group1",
			key:       "key1",
			expected:  0.0,
		},
		{
			name:      "group does not exist",
			subScores: map[string]map[string]float64{"group1": {"key1": 0.85}},
			group:     "group2",
			key:       "key1",
			expected:  0.0,
		},
		{
			name:      "key does not exist in group",
			subScores: map[string]map[string]float64{"group1": {"key1": 0.85}},
			group:     "group1",
			key:       "key2",
			expected:  0.0,
		},
		{
			name:      "key exists",
			subScores: map[string]map[string]float64{"group1": {"key1": 0.85, "key2": 0.65}},
			group:     "group1",
			key:       "key1",
			expected:  0.85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSubScore(tt.subScores, tt.group, tt.key)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}

func TestCalculateBelow52wHighPct(t *testing.T) {
	tests := []struct {
		name         string
		currentPrice *float64
		price52wHigh *float64
		expected     float64
	}{
		{
			name:         "nil current price",
			currentPrice: nil,
			price52wHigh: floatPtr(100.0),
			expected:     0.0,
		},
		{
			name:         "nil 52w high",
			currentPrice: floatPtr(90.0),
			price52wHigh: nil,
			expected:     0.0,
		},
		{
			name:         "zero 52w high",
			currentPrice: floatPtr(90.0),
			price52wHigh: floatPtr(0.0),
			expected:     0.0,
		},
		{
			name:         "at 52w high",
			currentPrice: floatPtr(100.0),
			price52wHigh: floatPtr(100.0),
			expected:     0.0,
		},
		{
			name:         "above 52w high",
			currentPrice: floatPtr(105.0),
			price52wHigh: floatPtr(100.0),
			expected:     0.0,
		},
		{
			name:         "10% below 52w high",
			currentPrice: floatPtr(90.0),
			price52wHigh: floatPtr(100.0),
			expected:     10.0,
		},
		{
			name:         "20% below 52w high",
			currentPrice: floatPtr(80.0),
			price52wHigh: floatPtr(100.0),
			expected:     20.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateBelow52wHighPct(tt.currentPrice, tt.price52wHigh)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

func TestIsPoorRiskAdjusted(t *testing.T) {
	tests := []struct {
		name         string
		sharpeRatio  float64
		sortinoRatio float64
		expected     bool
		desc         string
	}{
		{
			name:         "both zero",
			sharpeRatio:  0.0,
			sortinoRatio: 0.0,
			expected:     false,
			desc:         "No data means can't assess - not poor",
		},
		{
			name:         "good Sharpe ratio",
			sharpeRatio:  1.5,
			sortinoRatio: 0.0,
			expected:     false,
			desc:         "Good Sharpe ratio is not poor",
		},
		{
			name:         "poor Sharpe ratio",
			sharpeRatio:  0.3,
			sortinoRatio: 0.0,
			expected:     true,
			desc:         "Poor Sharpe ratio (< 0.5) is poor",
		},
		{
			name:         "good Sortino ratio",
			sharpeRatio:  0.0,
			sortinoRatio: 1.2,
			expected:     false,
			desc:         "Good Sortino ratio is not poor",
		},
		{
			name:         "poor Sortino ratio",
			sharpeRatio:  0.0,
			sortinoRatio: 0.4,
			expected:     true,
			desc:         "Poor Sortino ratio (< 0.5) is poor",
		},
		{
			name:         "Sharpe takes precedence when both present",
			sharpeRatio:  0.3,
			sortinoRatio: 1.0,
			expected:     true,
			desc:         "Poor Sharpe overrides good Sortino",
		},
		{
			name:         "exact threshold Sharpe",
			sharpeRatio:  0.5,
			sortinoRatio: 0.0,
			expected:     false,
			desc:         "Exactly 0.5 Sharpe is not poor",
		},
		{
			name:         "exact threshold Sortino",
			sharpeRatio:  0.0,
			sortinoRatio: 0.5,
			expected:     false,
			desc:         "Exactly 0.5 Sortino is not poor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPoorRiskAdjusted(tt.sharpeRatio, tt.sortinoRatio)
			assert.Equal(t, tt.expected, result, tt.desc)
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "no duplicates",
			input:    []string{"tag1", "tag2", "tag3"},
			expected: []string{"tag1", "tag2", "tag3"},
		},
		{
			name:     "with duplicates",
			input:    []string{"tag1", "tag2", "tag1", "tag3", "tag2"},
			expected: []string{"tag1", "tag2", "tag3"},
		},
		{
			name:     "all duplicates",
			input:    []string{"tag1", "tag1", "tag1"},
			expected: []string{"tag1"},
		},
		{
			name:     "preserves order",
			input:    []string{"tag1", "tag2", "tag1", "tag3"},
			expected: []string{"tag1", "tag2", "tag3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeDuplicates(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}
