package market_regime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRegionForMarketCode(t *testing.T) {
	tests := []struct {
		name       string
		marketCode string
		expected   string
	}{
		// US markets
		{"FIX returns US", "FIX", RegionUS},

		// European markets
		{"EU returns EU", "EU", RegionEU},
		{"ATHEX returns EU", "ATHEX", RegionEU},

		// Asian markets
		{"HKEX returns ASIA", "HKEX", RegionAsia},
		{"HKG returns ASIA", "HKG", RegionAsia},

		// Russian markets
		{"FORTS returns RUSSIA", "FORTS", RegionRussia},
		{"MCX returns RUSSIA", "MCX", RegionRussia},

		// Middle East
		{"TABADUL returns MIDDLE_EAST", "TABADUL", RegionMiddleEast},

		// Central Asia
		{"KASE returns CENTRAL_ASIA", "KASE", RegionCentralAsia},

		// Unknown
		{"unknown returns UNKNOWN", "UNKNOWN_MKT", RegionUnknown},
		{"empty string returns UNKNOWN", "", RegionUnknown},
		{"lowercase returns UNKNOWN", "fix", RegionUnknown}, // case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRegionForMarketCode(tt.marketCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegionHasIndices(t *testing.T) {
	tests := []struct {
		name     string
		region   string
		expected bool
	}{
		// Regions with indices
		{"US has indices", RegionUS, true},
		{"EU has indices", RegionEU, true},
		{"ASIA has indices", RegionAsia, true},

		// Regions without indices (use global average fallback)
		{"RUSSIA has no indices", RegionRussia, false},
		{"MIDDLE_EAST has no indices", RegionMiddleEast, false},
		{"CENTRAL_ASIA has no indices", RegionCentralAsia, false},
		{"UNKNOWN has no indices", RegionUnknown, false},

		// Edge cases
		{"empty string has no indices", "", false},
		{"arbitrary string has no indices", "INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RegionHasIndices(tt.region)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegionConstants(t *testing.T) {
	// Verify constants have expected values
	assert.Equal(t, "US", RegionUS)
	assert.Equal(t, "EU", RegionEU)
	assert.Equal(t, "ASIA", RegionAsia)
	assert.Equal(t, "RUSSIA", RegionRussia)
	assert.Equal(t, "MIDDLE_EAST", RegionMiddleEast)
	assert.Equal(t, "CENTRAL_ASIA", RegionCentralAsia)
	assert.Equal(t, "UNKNOWN", RegionUnknown)
}

func TestGetAllRegionsWithIndices(t *testing.T) {
	regions := GetAllRegionsWithIndices()

	// Should contain exactly US, EU, ASIA
	assert.Len(t, regions, 3)
	assert.Contains(t, regions, RegionUS)
	assert.Contains(t, regions, RegionEU)
	assert.Contains(t, regions, RegionAsia)

	// Should not contain regions without indices
	assert.NotContains(t, regions, RegionRussia)
	assert.NotContains(t, regions, RegionMiddleEast)
	assert.NotContains(t, regions, RegionCentralAsia)
	assert.NotContains(t, regions, RegionUnknown)
}

func TestCalculateGlobalAverage(t *testing.T) {
	tests := []struct {
		name         string
		regionScores map[string]float64
		expected     float64
	}{
		{
			name: "all regions positive",
			regionScores: map[string]float64{
				RegionUS:   0.5,
				RegionEU:   0.3,
				RegionAsia: 0.1,
			},
			expected: 0.3, // (0.5 + 0.3 + 0.1) / 3
		},
		{
			name: "mixed positive and negative",
			regionScores: map[string]float64{
				RegionUS:   0.6,
				RegionEU:   0.0,
				RegionAsia: -0.3,
			},
			expected: 0.1, // (0.6 + 0.0 + -0.3) / 3
		},
		{
			name: "all negative",
			regionScores: map[string]float64{
				RegionUS:   -0.3,
				RegionEU:   -0.6,
				RegionAsia: -0.3,
			},
			expected: -0.4, // (-0.3 + -0.6 + -0.3) / 3
		},
		{
			name: "ignores regions without indices",
			regionScores: map[string]float64{
				RegionUS:         0.6,
				RegionEU:         0.3,
				RegionAsia:       0.0,
				RegionRussia:     0.9, // Should be ignored
				RegionMiddleEast: 0.8, // Should be ignored
			},
			expected: 0.3, // (0.6 + 0.3 + 0.0) / 3
		},
		{
			name: "only two regions available",
			regionScores: map[string]float64{
				RegionUS: 0.4,
				RegionEU: 0.2,
				// Asia missing
			},
			expected: 0.3, // (0.4 + 0.2) / 2
		},
		{
			name: "only one region available",
			regionScores: map[string]float64{
				RegionUS: 0.5,
			},
			expected: 0.5, // 0.5 / 1
		},
		{
			name:         "empty scores returns zero",
			regionScores: map[string]float64{},
			expected:     0.0, // True neutral when no data
		},
		{
			name: "only unknown regions returns zero",
			regionScores: map[string]float64{
				RegionRussia:     0.5,
				RegionMiddleEast: 0.3,
			},
			expected: 0.0, // No regions with indices
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateGlobalAverage(tt.regionScores)
			assert.InDelta(t, tt.expected, result, 0.001)
		})
	}
}
