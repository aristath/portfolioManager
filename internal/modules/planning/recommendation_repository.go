// Package planning provides portfolio planning and strategy generation functionality.
package planning

import (
	"strings"
	"time"
)

// Recommendation represents a stored recommendation
type Recommendation struct {
	UUID                  string
	Symbol                string
	Name                  string
	Side                  string
	Quantity              float64
	EstimatedPrice        float64
	EstimatedValue        float64
	Reason                string
	Currency              string
	Priority              float64
	CurrentPortfolioScore float64
	NewPortfolioScore     float64
	ScoreChange           float64
	Status                string // "pending", "executed", "rejected", "expired", "failed"
	PortfolioHash         string
	CreatedAt             time.Time
	UpdatedAt             time.Time
	ExecutedAt            *time.Time
	RetryCount            int        // Number of execution attempts
	LastAttemptAt         *time.Time // Last execution attempt timestamp
	FailureReason         string     // Reason for last failure
}

// IsEmergencyReason determines if a recommendation reason indicates an emergency
func IsEmergencyReason(reason string) bool {
	emergencyPatterns := []string{
		"emergency",
		"negative balance",
		"margin call",
		"urgent",
		"critical",
		"risk limit",
		"stop loss",
		"forced",
	}

	reasonLower := strings.ToLower(reason)
	for _, pattern := range emergencyPatterns {
		if strings.Contains(reasonLower, pattern) {
			return true
		}
	}
	return false
}
