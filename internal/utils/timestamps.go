// Package utils provides utility functions for timestamp and date conversions.
package utils

import (
	"fmt"
	"time"
)

// ToUnix converts time.Time to Unix timestamp (seconds since epoch)
func ToUnix(t time.Time) int64 {
	return t.Unix()
}

// FromUnix converts Unix timestamp to time.Time (UTC)
func FromUnix(ts int64) time.Time {
	return time.Unix(ts, 0).UTC()
}

// DateToUnix converts YYYY-MM-DD string to Unix timestamp at midnight UTC
func DateToUnix(dateStr string) (int64, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0, fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
	}
	// Ensure it's at midnight UTC
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return t.Unix(), nil
}

// UnixToDate converts Unix timestamp to YYYY-MM-DD string
func UnixToDate(ts int64) string {
	t := FromUnix(ts)
	return t.Format("2006-01-02")
}
