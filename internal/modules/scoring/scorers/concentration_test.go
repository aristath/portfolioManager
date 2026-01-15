package scorers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckConcentration_WithinLimits verifies that positions within thresholds pass
func TestCheckConcentration_WithinLimits(t *testing.T) {
	scorer := NewConcentrationScorer()
	thresholds := DefaultConcentrationThresholds()

	tests := []struct {
		name             string
		isin             string
		geography        string
		proposedValueEUR float64
		context          *ConcentrationContext
	}{
		{
			name:             "New position within all limits",
			isin:             "US0378331005", // AAPL
			geography:        "US",
			proposedValueEUR: 1000.0,
			context: &ConcentrationContext{
				Positions:  map[string]float64{},
				TotalValue: 10000.0,
				GeographyAllocations: map[string]float64{
					"US": 0.30, // 30% already in US
				},
			},
		},
		{
			name:             "Existing position still within limits",
			isin:             "US0378331005",
			geography:        "US",
			proposedValueEUR: 500.0,
			context: &ConcentrationContext{
				Positions: map[string]float64{
					"US0378331005": 1500.0, // 15% position
				},
				TotalValue: 10000.0,
				GeographyAllocations: map[string]float64{
					"US": 0.35,
				},
			},
		},
		{
			name:             "Position at exactly threshold (passes)",
			isin:             "US0378331005",
			geography:        "US",
			proposedValueEUR: 0.0, // Just checking current state
			context: &ConcentrationContext{
				Positions: map[string]float64{
					"US0378331005": 2000.0, // Exactly 20%
				},
				TotalValue: 10000.0,
				GeographyAllocations: map[string]float64{
					"US": 0.20,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.CheckConcentration(
				tt.isin,
				tt.geography,
				tt.proposedValueEUR,
				tt.context,
				thresholds,
			)

			assert.True(t, result.Passes, "Expected concentration check to pass: %s", result.Reason)
			assert.Empty(t, result.Reason, "Expected no reason when passing")
		})
	}
}

// TestCheckConcentration_PositionExceedsThreshold verifies that oversized positions are blocked
func TestCheckConcentration_PositionExceedsThreshold(t *testing.T) {
	scorer := NewConcentrationScorer()
	thresholds := DefaultConcentrationThresholds()

	tests := []struct {
		name             string
		isin             string
		geography        string
		proposedValueEUR float64
		context          *ConcentrationContext
		wantReason       string
	}{
		{
			name:             "New position would exceed 20% threshold",
			isin:             "US0378331005",
			geography:        "US",
			proposedValueEUR: 3000.0, // Would be 3000/13000 = 23% > 20%
			context: &ConcentrationContext{
				Positions:            map[string]float64{},
				TotalValue:           10000.0,
				GeographyAllocations: map[string]float64{},
			},
			wantReason: "position",
		},
		{
			name:             "Existing position already exceeds threshold",
			isin:             "US0378331005",
			geography:        "US",
			proposedValueEUR: 100.0,
			context: &ConcentrationContext{
				Positions: map[string]float64{
					"US0378331005": 2200.0, // Already 22%
				},
				TotalValue:           10000.0,
				GeographyAllocations: map[string]float64{},
			},
			wantReason: "position",
		},
		{
			name:             "Adding to position would exceed threshold",
			isin:             "US0378331005",
			geography:        "US",
			proposedValueEUR: 1000.0, // Would push from 18% to 28%
			context: &ConcentrationContext{
				Positions: map[string]float64{
					"US0378331005": 1800.0, // Currently 18%
				},
				TotalValue:           10000.0,
				GeographyAllocations: map[string]float64{},
			},
			wantReason: "position",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.CheckConcentration(
				tt.isin,
				tt.geography,
				tt.proposedValueEUR,
				tt.context,
				thresholds,
			)

			assert.False(t, result.Passes, "Expected concentration check to fail")
			assert.Contains(t, result.Reason, tt.wantReason, "Reason should mention position concentration")
		})
	}
}

// TestCheckConcentration_GeographyExceedsThreshold verifies geography concentration is blocked for NEW positions
func TestCheckConcentration_GeographyExceedsThreshold(t *testing.T) {
	scorer := NewConcentrationScorer()
	thresholds := DefaultConcentrationThresholds()

	tests := []struct {
		name             string
		isin             string
		geography        string
		proposedValueEUR float64
		context          *ConcentrationContext
		wantPasses       bool
		wantReason       string
	}{
		{
			name:             "NEW position in over-concentrated geography is blocked",
			isin:             "NEW_ISIN_123",
			geography:        "US",
			proposedValueEUR: 500.0,
			context: &ConcentrationContext{
				Positions:  map[string]float64{}, // No existing position
				TotalValue: 10000.0,
				GeographyAllocations: map[string]float64{
					"US": 0.45, // Already 45% in US, over 40% threshold
				},
			},
			wantPasses: false,
			wantReason: "geography",
		},
		{
			name:             "EXISTING position in over-concentrated geography is allowed",
			isin:             "US0378331005",
			geography:        "US",
			proposedValueEUR: 500.0,
			context: &ConcentrationContext{
				Positions: map[string]float64{
					"US0378331005": 1000.0, // Already own this
				},
				TotalValue: 10000.0,
				GeographyAllocations: map[string]float64{
					"US": 0.45, // 45% in US
				},
			},
			wantPasses: true, // Allowed because we already own it
			wantReason: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.CheckConcentration(
				tt.isin,
				tt.geography,
				tt.proposedValueEUR,
				tt.context,
				thresholds,
			)

			assert.Equal(t, tt.wantPasses, result.Passes, "Passes mismatch")
			if tt.wantReason != "" {
				assert.Contains(t, result.Reason, tt.wantReason)
			}
		})
	}
}

// TestCheckConcentration_MultipleGeographies verifies handling of comma-separated geographies
func TestCheckConcentration_MultipleGeographies(t *testing.T) {
	scorer := NewConcentrationScorer()
	thresholds := DefaultConcentrationThresholds()

	tests := []struct {
		name             string
		isin             string
		geography        string
		proposedValueEUR float64
		context          *ConcentrationContext
		wantPasses       bool
	}{
		{
			name:             "Multi-geo where one geography is over-concentrated",
			isin:             "NEW_ISIN",
			geography:        "US, EU", // Comma-separated
			proposedValueEUR: 500.0,
			context: &ConcentrationContext{
				Positions:  map[string]float64{},
				TotalValue: 10000.0,
				GeographyAllocations: map[string]float64{
					"US": 0.45, // Over threshold
					"EU": 0.20, // Under threshold
				},
			},
			wantPasses: false, // Should fail because US is over-concentrated
		},
		{
			name:             "Multi-geo where all geographies are within limits",
			isin:             "NEW_ISIN",
			geography:        "US, EU",
			proposedValueEUR: 500.0,
			context: &ConcentrationContext{
				Positions:  map[string]float64{},
				TotalValue: 10000.0,
				GeographyAllocations: map[string]float64{
					"US": 0.30,
					"EU": 0.20,
				},
			},
			wantPasses: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.CheckConcentration(
				tt.isin,
				tt.geography,
				tt.proposedValueEUR,
				tt.context,
				thresholds,
			)

			assert.Equal(t, tt.wantPasses, result.Passes, "Passes mismatch for %s", tt.name)
		})
	}
}

// TestCheckConcentration_NilContext verifies graceful handling of nil context
func TestCheckConcentration_NilContext(t *testing.T) {
	scorer := NewConcentrationScorer()
	thresholds := DefaultConcentrationThresholds()

	result := scorer.CheckConcentration(
		"US0378331005",
		"US",
		1000.0,
		nil, // Nil context
		thresholds,
	)

	// Should pass by default when no context available (no data to check against)
	assert.True(t, result.Passes, "Should pass with nil context")
}

// TestCheckConcentration_EmptyPortfolio verifies handling of empty portfolio
func TestCheckConcentration_EmptyPortfolio(t *testing.T) {
	scorer := NewConcentrationScorer()
	thresholds := DefaultConcentrationThresholds()

	result := scorer.CheckConcentration(
		"US0378331005",
		"US",
		1000.0,
		&ConcentrationContext{
			Positions:            map[string]float64{},
			TotalValue:           0.0, // Empty portfolio
			GeographyAllocations: map[string]float64{},
		},
		thresholds,
	)

	// First position in empty portfolio should always pass
	assert.True(t, result.Passes, "First position in empty portfolio should pass")
}

// TestConcentrationThresholds_Defaults verifies default threshold values
func TestConcentrationThresholds_Defaults(t *testing.T) {
	thresholds := DefaultConcentrationThresholds()

	assert.Equal(t, 0.20, thresholds.MaxPositionWeight, "Default max position weight should be 20%")
	assert.Equal(t, 0.40, thresholds.MaxGeographyWeight, "Default max geography weight should be 40%")
}

// TestConcentrationScorer_Constructor verifies scorer creation
func TestConcentrationScorer_Constructor(t *testing.T) {
	scorer := NewConcentrationScorer()
	require.NotNil(t, scorer, "Constructor should return non-nil scorer")
}
