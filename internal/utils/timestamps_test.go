package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToUnix(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected int64
	}{
		{
			name:     "epoch time",
			input:    time.Unix(0, 0).UTC(),
			expected: 0,
		},
		{
			name:     "specific timestamp",
			input:    time.Unix(1704067200, 0).UTC(), // 2024-01-01 00:00:00 UTC
			expected: 1704067200,
		},
		{
			name:     "current time",
			input:    time.Now().UTC(),
			expected: time.Now().UTC().Unix(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToUnix(tt.input)
			if tt.name == "current time" {
				// For current time, just verify it's a reasonable value
				assert.Greater(t, result, int64(1600000000)) // After 2020
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFromUnix(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected time.Time
	}{
		{
			name:     "epoch time",
			input:    0,
			expected: time.Unix(0, 0).UTC(),
		},
		{
			name:     "specific timestamp",
			input:    1704067200, // 2024-01-01 00:00:00 UTC
			expected: time.Unix(1704067200, 0).UTC(),
		},
		{
			name:     "negative timestamp",
			input:    -1,
			expected: time.Unix(-1, 0).UTC(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromUnix(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, time.UTC, result.Location())
		})
	}
}

func TestDateToUnix(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  int64
		wantError bool
	}{
		{
			name:      "valid date",
			input:     "2024-01-01",
			expected:  1704067200, // 2024-01-01 00:00:00 UTC
			wantError: false,
		},
		{
			name:      "valid date 2",
			input:     "2023-12-25",
			expected:  1703462400, // 2023-12-25 00:00:00 UTC
			wantError: false,
		},
		{
			name:      "invalid format",
			input:     "2024/01/01",
			wantError: true,
		},
		{
			name:      "invalid date",
			input:     "2024-13-01",
			wantError: true,
		},
		{
			name:      "empty string",
			input:     "",
			wantError: true,
		},
		{
			name:      "invalid format 2",
			input:     "not-a-date",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DateToUnix(tt.input)
			if tt.wantError {
				assert.Error(t, err)
				assert.Zero(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestUnixToDate(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{
			name:     "epoch time",
			input:    0,
			expected: "1970-01-01",
		},
		{
			name:     "specific timestamp",
			input:    1704067200, // 2024-01-01 00:00:00 UTC
			expected: "2024-01-01",
		},
		{
			name:     "another date",
			input:    1703462400, // 2023-12-25 00:00:00 UTC
			expected: "2023-12-25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnixToDate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that ToUnix and FromUnix are inverse operations
	original := time.Now().UTC()
	unix := ToUnix(original)
	restored := FromUnix(unix)

	// Should be equal to the second (Unix timestamps are second precision)
	assert.Equal(t, original.Unix(), restored.Unix())
}

func TestDateRoundTrip(t *testing.T) {
	// Test that DateToUnix and UnixToDate are inverse operations
	dateStr := "2024-01-15"
	unix, err := DateToUnix(dateStr)
	require.NoError(t, err)

	restored := UnixToDate(unix)
	assert.Equal(t, dateStr, restored)
}

func TestDateToUnixMidnightUTC(t *testing.T) {
	// Verify that DateToUnix sets time to midnight UTC
	dateStr := "2024-01-01"
	unix, err := DateToUnix(dateStr)
	require.NoError(t, err)

	tm := FromUnix(unix)
	assert.Equal(t, 0, tm.Hour())
	assert.Equal(t, 0, tm.Minute())
	assert.Equal(t, 0, tm.Second())
	assert.Equal(t, time.UTC, tm.Location())
}
