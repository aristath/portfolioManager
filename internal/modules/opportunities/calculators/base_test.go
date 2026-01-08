package calculators

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestGetFloatParam(t *testing.T) {
	tests := []struct {
		name         string
		params       map[string]interface{}
		key          string
		defaultValue float64
		expected     float64
	}{
		{
			name:         "float64 value exists",
			params:       map[string]interface{}{"key": 123.45},
			key:          "key",
			defaultValue: 0.0,
			expected:     123.45,
		},
		{
			name:         "int value exists (converted to float64)",
			params:       map[string]interface{}{"key": 100},
			key:          "key",
			defaultValue: 0.0,
			expected:     100.0,
		},
		{
			name:         "key does not exist",
			params:       map[string]interface{}{"other": 123.45},
			key:          "key",
			defaultValue: 99.99,
			expected:     99.99,
		},
		{
			name:         "nil params",
			params:       nil,
			key:          "key",
			defaultValue: 99.99,
			expected:     99.99,
		},
		{
			name:         "wrong type (string)",
			params:       map[string]interface{}{"key": "not a number"},
			key:          "key",
			defaultValue: 99.99,
			expected:     99.99,
		},
		{
			name:         "wrong type (bool)",
			params:       map[string]interface{}{"key": true},
			key:          "key",
			defaultValue: 99.99,
			expected:     99.99,
		},
		{
			name:         "zero float64 value",
			params:       map[string]interface{}{"key": 0.0},
			key:          "key",
			defaultValue: 99.99,
			expected:     0.0,
		},
		{
			name:         "negative float64 value",
			params:       map[string]interface{}{"key": -123.45},
			key:          "key",
			defaultValue: 0.0,
			expected:     -123.45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFloatParam(tt.params, tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIntParam(t *testing.T) {
	tests := []struct {
		name         string
		params       map[string]interface{}
		key          string
		defaultValue int
		expected     int
	}{
		{
			name:         "int value exists",
			params:       map[string]interface{}{"key": 100},
			key:          "key",
			defaultValue: 0,
			expected:     100,
		},
		{
			name:         "float64 value exists (converted to int)",
			params:       map[string]interface{}{"key": 123.45},
			key:          "key",
			defaultValue: 0,
			expected:     123,
		},
		{
			name:         "float64 value with decimal (truncated)",
			params:       map[string]interface{}{"key": 123.99},
			key:          "key",
			defaultValue: 0,
			expected:     123,
		},
		{
			name:         "key does not exist",
			params:       map[string]interface{}{"other": 100},
			key:          "key",
			defaultValue: 99,
			expected:     99,
		},
		{
			name:         "nil params",
			params:       nil,
			key:          "key",
			defaultValue: 99,
			expected:     99,
		},
		{
			name:         "wrong type (string)",
			params:       map[string]interface{}{"key": "not a number"},
			key:          "key",
			defaultValue: 99,
			expected:     99,
		},
		{
			name:         "wrong type (bool)",
			params:       map[string]interface{}{"key": true},
			key:          "key",
			defaultValue: 99,
			expected:     99,
		},
		{
			name:         "zero int value",
			params:       map[string]interface{}{"key": 0},
			key:          "key",
			defaultValue: 99,
			expected:     0,
		},
		{
			name:         "negative int value",
			params:       map[string]interface{}{"key": -100},
			key:          "key",
			defaultValue: 0,
			expected:     -100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetIntParam(tt.params, tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetBoolParam(t *testing.T) {
	tests := []struct {
		name         string
		params       map[string]interface{}
		key          string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "bool true exists",
			params:       map[string]interface{}{"key": true},
			key:          "key",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "bool false exists",
			params:       map[string]interface{}{"key": false},
			key:          "key",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "key does not exist",
			params:       map[string]interface{}{"other": true},
			key:          "key",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "nil params",
			params:       nil,
			key:          "key",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "wrong type (string)",
			params:       map[string]interface{}{"key": "true"},
			key:          "key",
			defaultValue: false,
			expected:     false,
		},
		{
			name:         "wrong type (int)",
			params:       map[string]interface{}{"key": 1},
			key:          "key",
			defaultValue: false,
			expected:     false,
		},
		{
			name:         "wrong type (float64)",
			params:       map[string]interface{}{"key": 1.0},
			key:          "key",
			defaultValue: false,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBoolParam(tt.params, tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewBaseCalculator(t *testing.T) {
	log := zerolog.Nop()
	calc := NewBaseCalculator(log, "test_calculator")

	assert.NotNil(t, calc)
	assert.NotNil(t, calc.log)
}
