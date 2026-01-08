package opportunities

import (
	"testing"

	"github.com/aristath/sentinel/internal/modules/planning/domain"
	"github.com/stretchr/testify/assert"
)

func TestSortByPriority(t *testing.T) {
	tests := []struct {
		name       string
		candidates []domain.ActionCandidate
		expected   []domain.ActionCandidate
		desc       string
	}{
		{
			name:       "empty slice",
			candidates: []domain.ActionCandidate{},
			expected:   []domain.ActionCandidate{},
			desc:       "Empty slice should remain empty",
		},
		{
			name: "single candidate",
			candidates: []domain.ActionCandidate{
				{Symbol: "AAPL", Priority: 0.8},
			},
			expected: []domain.ActionCandidate{
				{Symbol: "AAPL", Priority: 0.8},
			},
			desc: "Single candidate should remain unchanged",
		},
		{
			name: "already sorted descending",
			candidates: []domain.ActionCandidate{
				{Symbol: "AAPL", Priority: 0.9},
				{Symbol: "MSFT", Priority: 0.7},
				{Symbol: "GOOGL", Priority: 0.5},
			},
			expected: []domain.ActionCandidate{
				{Symbol: "AAPL", Priority: 0.9},
				{Symbol: "MSFT", Priority: 0.7},
				{Symbol: "GOOGL", Priority: 0.5},
			},
			desc: "Already sorted should remain sorted",
		},
		{
			name: "reverse order",
			candidates: []domain.ActionCandidate{
				{Symbol: "GOOGL", Priority: 0.5},
				{Symbol: "MSFT", Priority: 0.7},
				{Symbol: "AAPL", Priority: 0.9},
			},
			expected: []domain.ActionCandidate{
				{Symbol: "AAPL", Priority: 0.9},
				{Symbol: "MSFT", Priority: 0.7},
				{Symbol: "GOOGL", Priority: 0.5},
			},
			desc: "Reverse order should be sorted descending",
		},
		{
			name: "unsorted",
			candidates: []domain.ActionCandidate{
				{Symbol: "MSFT", Priority: 0.7},
				{Symbol: "AAPL", Priority: 0.9},
				{Symbol: "GOOGL", Priority: 0.5},
			},
			expected: []domain.ActionCandidate{
				{Symbol: "AAPL", Priority: 0.9},
				{Symbol: "MSFT", Priority: 0.7},
				{Symbol: "GOOGL", Priority: 0.5},
			},
			desc: "Unsorted should be sorted descending",
		},
		{
			name: "equal priorities",
			candidates: []domain.ActionCandidate{
				{Symbol: "AAPL", Priority: 0.7},
				{Symbol: "MSFT", Priority: 0.7},
				{Symbol: "GOOGL", Priority: 0.7},
			},
			expected: []domain.ActionCandidate{
				{Symbol: "AAPL", Priority: 0.7},
				{Symbol: "MSFT", Priority: 0.7},
				{Symbol: "GOOGL", Priority: 0.7},
			},
			desc: "Equal priorities should remain in original order (stable sort)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying the original
			candidates := make([]domain.ActionCandidate, len(tt.candidates))
			copy(candidates, tt.candidates)

			sortByPriority(candidates)

			// Check that priorities are in descending order
			for i := 0; i < len(candidates); i++ {
				assert.Equal(t, tt.expected[i].Priority, candidates[i].Priority, "%s: Priority mismatch at index %d", tt.desc, i)
			}

			// Verify it's actually sorted (descending)
			for i := 0; i < len(candidates)-1; i++ {
				assert.GreaterOrEqual(t, candidates[i].Priority, candidates[i+1].Priority,
					"Priorities should be in descending order at index %d and %d", i, i+1)
			}
		})
	}
}
