package portfolio

import (
	"testing"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestConvertAllocationsToDomain(t *testing.T) {
	tests := []struct {
		name     string
		input    []AllocationStatus
		expected []domain.PortfolioAllocation
	}{
		{
			name:     "empty slice",
			input:    []AllocationStatus{},
			expected: []domain.PortfolioAllocation{},
		},
		{
			name: "single allocation",
			input: []AllocationStatus{
				{
					Name:         "US",
					TargetPct:    0.5,
					CurrentPct:   0.45,
					CurrentValue: 4500.0,
					Deviation:    -0.05,
				},
			},
			expected: []domain.PortfolioAllocation{
				{
					Name:         "US",
					TargetPct:    0.5,
					CurrentPct:   0.45,
					CurrentValue: 4500.0,
					Deviation:    -0.05,
				},
			},
		},
		{
			name: "multiple allocations",
			input: []AllocationStatus{
				{
					Name:         "US",
					TargetPct:    0.5,
					CurrentPct:   0.45,
					CurrentValue: 4500.0,
					Deviation:    -0.05,
				},
				{
					Name:         "EU",
					TargetPct:    0.3,
					CurrentPct:   0.35,
					CurrentValue: 3500.0,
					Deviation:    0.05,
				},
				{
					Name:         "Asia",
					TargetPct:    0.2,
					CurrentPct:   0.20,
					CurrentValue: 2000.0,
					Deviation:    0.0,
				},
			},
			expected: []domain.PortfolioAllocation{
				{
					Name:         "US",
					TargetPct:    0.5,
					CurrentPct:   0.45,
					CurrentValue: 4500.0,
					Deviation:    -0.05,
				},
				{
					Name:         "EU",
					TargetPct:    0.3,
					CurrentPct:   0.35,
					CurrentValue: 3500.0,
					Deviation:    0.05,
				},
				{
					Name:         "Asia",
					TargetPct:    0.2,
					CurrentPct:   0.20,
					CurrentValue: 2000.0,
					Deviation:    0.0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertAllocationsToDomain(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
