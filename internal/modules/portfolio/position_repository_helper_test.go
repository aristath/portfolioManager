package portfolio

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		{
			name:     "large value",
			input:    1000000.0,
			expected: sql.NullFloat64{Float64: 1000000.0, Valid: true},
		},
		{
			name:     "small value",
			input:    0.0001,
			expected: sql.NullFloat64{Float64: 0.0001, Valid: true},
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
		{
			name:     "special characters",
			input:    "test@123",
			expected: sql.NullString{String: "test@123", Valid: true},
		},
		{
			name:     "unicode string",
			input:    "测试",
			expected: sql.NullString{String: "测试", Valid: true},
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
