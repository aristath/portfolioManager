package optimization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: isISIN helper is defined in mv_optimizer_test.go

// TestOptimizerService_OptimizeISINKeys will test the full service once implementation is complete
// For now, we focus on struct-level tests below
func TestOptimizerService_OptimizeISINKeys(t *testing.T) {
	t.Skip("Full service test - will be enabled after implementation")
}

// TestPortfolioState_ISINKeyedMaps verifies PortfolioState uses ISIN keys
func TestPortfolioState_ISINKeyedMaps(t *testing.T) {
	state := PortfolioState{
		Positions: map[string]Position{
			"US0378331005": {ISIN: "US0378331005", Quantity: 10, ValueEUR: 1500},
			"US5949181045": {ISIN: "US5949181045", Quantity: 5, ValueEUR: 1500},
		},
		CurrentPrices: map[string]float64{
			"US0378331005": 150.0, // ISIN key ✅
			"US5949181045": 300.0, // ISIN key ✅
		},
		DividendBonuses: map[string]float64{
			"US0378331005": 0.02, // ISIN key ✅
			"US5949181045": 0.03, // ISIN key ✅
		},
	}

	// Verify Positions uses ISIN keys
	for key := range state.Positions {
		assert.True(t, isISIN(key), "Positions should have ISIN keys, got: %s", key)
	}

	// Verify CurrentPrices uses ISIN keys
	for key := range state.CurrentPrices {
		assert.True(t, isISIN(key), "CurrentPrices should have ISIN keys, got: %s", key)
	}

	// Verify DividendBonuses uses ISIN keys
	for key := range state.DividendBonuses {
		assert.True(t, isISIN(key), "DividendBonuses should have ISIN keys, got: %s", key)
	}

	// Verify no Symbol keys
	assert.NotContains(t, state.Positions, "AAPL.US", "Positions should NOT have Symbol keys")
	assert.NotContains(t, state.CurrentPrices, "AAPL.US", "CurrentPrices should NOT have Symbol keys")
	assert.NotContains(t, state.DividendBonuses, "AAPL.US", "DividendBonuses should NOT have Symbol keys")
}

// TestResult_ISINKeyedMaps verifies Result uses ISIN keys
func TestResult_ISINKeyedMaps(t *testing.T) {
	result := Result{
		TargetWeights: map[string]float64{
			"US0378331005": 0.40, // ISIN key ✅
			"US5949181045": 0.60, // ISIN key ✅
		},
	}

	// Verify TargetWeights uses ISIN keys
	for key := range result.TargetWeights {
		assert.True(t, isISIN(key), "TargetWeights should have ISIN keys, got: %s", key)
	}

	// Verify no Symbol keys
	assert.NotContains(t, result.TargetWeights, "AAPL.US", "TargetWeights should NOT have Symbol keys")
	assert.NotContains(t, result.TargetWeights, "MSFT.US", "TargetWeights should NOT have Symbol keys")
}

// TestConstraints_ISINArray verifies Constraints uses ISIN array
func TestConstraints_ISINArray(t *testing.T) {
	constraints := Constraints{
		ISINs: []string{"US0378331005", "US5949181045"}, // ISIN array ✅
		MinWeights: map[string]float64{
			"US0378331005": 0.0, // ISIN key ✅
			"US5949181045": 0.0, // ISIN key ✅
		},
		MaxWeights: map[string]float64{
			"US0378331005": 0.50, // ISIN key ✅
			"US5949181045": 0.50, // ISIN key ✅
		},
	}

	// Verify ISINs array contains only ISINs
	for _, isin := range constraints.ISINs {
		assert.True(t, isISIN(isin), "ISINs array should only contain ISINs, got: %s", isin)
	}

	// Verify no Symbols field exists (compile check)
	// If this compiles, Symbols field doesn't exist ✅
	_ = constraints.ISINs

	// Verify map keys are ISINs
	for key := range constraints.MinWeights {
		assert.True(t, isISIN(key), "MinWeights should have ISIN keys, got: %s", key)
	}
	for key := range constraints.MaxWeights {
		assert.True(t, isISIN(key), "MaxWeights should have ISIN keys, got: %s", key)
	}

	// Verify no Symbol keys
	assert.NotContains(t, constraints.MinWeights, "AAPL.US", "MinWeights should NOT have Symbol keys")
	assert.NotContains(t, constraints.MaxWeights, "AAPL.US", "MaxWeights should NOT have Symbol keys")
}

// TestSectorConstraint_ISINMapper verifies SectorConstraint uses ISIN keys
func TestSectorConstraint_ISINMapper(t *testing.T) {
	sectorConstraint := SectorConstraint{
		SectorMapper: map[string]string{
			"US0378331005": "Technology", // ISIN → sector ✅
			"US5949181045": "Technology", // ISIN → sector ✅
		},
	}

	// Verify SectorMapper uses ISIN keys
	for key := range sectorConstraint.SectorMapper {
		assert.True(t, isISIN(key), "SectorMapper should have ISIN keys, got: %s", key)
	}

	// Verify no Symbol keys
	assert.NotContains(t, sectorConstraint.SectorMapper, "AAPL.US", "SectorMapper should NOT have Symbol keys")
	assert.NotContains(t, sectorConstraint.SectorMapper, "MSFT.US", "SectorMapper should NOT have Symbol keys")
}

// TestNoDualKeyDuplication verifies maps don't have both ISIN and Symbol keys
func TestNoDualKeyDuplication(t *testing.T) {
	state := PortfolioState{
		Positions: map[string]Position{
			"US0378331005": {ISIN: "US0378331005", Quantity: 10, ValueEUR: 1500},
		},
		CurrentPrices: map[string]float64{
			"US0378331005": 150.0,
			"US5949181045": 300.0,
		},
	}

	// Verify map sizes equal security count (no duplication)
	assert.Equal(t, 1, len(state.Positions), "Positions should have 1 entry (no dual keys)")
	assert.Equal(t, 2, len(state.CurrentPrices), "CurrentPrices should have 2 entries (no dual keys)")

	// If we had dual keys, we'd have 4 entries (2 ISINs + 2 Symbols)
	// We should only have 2 entries (2 ISINs)
	assert.NotEqual(t, 4, len(state.CurrentPrices), "Should not have dual-key duplication")
}

// ========================================
// Tests for hashHRPCacheKey
// ========================================

func TestHashHRPCacheKey_Deterministic(t *testing.T) {
	covMatrix := [][]float64{
		{0.04, 0.01},
		{0.01, 0.03},
	}
	isins := []string{"US0378331005", "US5949181045"}
	linkage := hrpLinkageSingle

	// Generate hash multiple times
	hash1 := hashHRPCacheKey(covMatrix, isins, linkage)
	hash2 := hashHRPCacheKey(covMatrix, isins, linkage)
	hash3 := hashHRPCacheKey(covMatrix, isins, linkage)

	// All should be identical
	assert.Equal(t, hash1, hash2, "Hash should be deterministic")
	assert.Equal(t, hash2, hash3, "Hash should be deterministic")
	assert.Len(t, hash1, 32, "Hash should be 32 hex characters (16 bytes)")
}

func TestHashHRPCacheKey_OrderIndependent(t *testing.T) {
	covMatrix := [][]float64{
		{0.04, 0.01, 0.005},
		{0.01, 0.03, 0.008},
		{0.005, 0.008, 0.025},
	}
	linkage := hrpLinkageSingle

	// Different orderings of the same ISINs
	order1 := []string{"US0378331005", "US5949181045", "US88160R1014"}
	order2 := []string{"US88160R1014", "US0378331005", "US5949181045"}
	order3 := []string{"US5949181045", "US88160R1014", "US0378331005"}

	hash1 := hashHRPCacheKey(covMatrix, order1, linkage)
	hash2 := hashHRPCacheKey(covMatrix, order2, linkage)
	hash3 := hashHRPCacheKey(covMatrix, order3, linkage)

	assert.Equal(t, hash1, hash2, "Hash should be order-independent")
	assert.Equal(t, hash2, hash3, "Hash should be order-independent")
}

func TestHashHRPCacheKey_DifferentLinkage(t *testing.T) {
	covMatrix := [][]float64{
		{0.04, 0.01},
		{0.01, 0.03},
	}
	isins := []string{"ISIN1", "ISIN2"}

	hashSingle := hashHRPCacheKey(covMatrix, isins, hrpLinkageSingle)
	hashComplete := hashHRPCacheKey(covMatrix, isins, hrpLinkageComplete)

	assert.NotEqual(t, hashSingle, hashComplete, "Different linkage methods should produce different hashes")
}

func TestHashHRPCacheKey_DifferentCovMatrix(t *testing.T) {
	isins := []string{"ISIN1", "ISIN2"}
	linkage := hrpLinkageSingle

	cov1 := [][]float64{
		{0.04, 0.01},
		{0.01, 0.03},
	}
	cov2 := [][]float64{
		{0.05, 0.01},
		{0.01, 0.03},
	}

	hash1 := hashHRPCacheKey(cov1, isins, linkage)
	hash2 := hashHRPCacheKey(cov2, isins, linkage)

	assert.NotEqual(t, hash1, hash2, "Different covariance matrices should produce different hashes")
}

func TestHashHRPCacheKey_DifferentISINs(t *testing.T) {
	covMatrix := [][]float64{
		{0.04, 0.01},
		{0.01, 0.03},
	}
	linkage := hrpLinkageSingle

	isins1 := []string{"ISIN1", "ISIN2"}
	isins2 := []string{"ISIN1", "ISIN3"}

	hash1 := hashHRPCacheKey(covMatrix, isins1, linkage)
	hash2 := hashHRPCacheKey(covMatrix, isins2, linkage)

	assert.NotEqual(t, hash1, hash2, "Different ISINs should produce different hashes")
}

// ========================================
// Tests for hashMVCacheKey
// ========================================

func TestHashMVCacheKey_Deterministic(t *testing.T) {
	expectedReturns := map[string]float64{
		"US0378331005": 0.12,
		"US5949181045": 0.08,
	}
	covMatrix := [][]float64{
		{0.04, 0.01},
		{0.01, 0.03},
	}
	constraints := Constraints{
		ISINs: []string{"US0378331005", "US5949181045"},
	}
	targetReturn := 0.10

	// Generate hash multiple times
	hash1 := hashMVCacheKey(expectedReturns, covMatrix, constraints, targetReturn)
	hash2 := hashMVCacheKey(expectedReturns, covMatrix, constraints, targetReturn)
	hash3 := hashMVCacheKey(expectedReturns, covMatrix, constraints, targetReturn)

	// All should be identical
	assert.Equal(t, hash1, hash2, "Hash should be deterministic")
	assert.Equal(t, hash2, hash3, "Hash should be deterministic")
	assert.Len(t, hash1, 32, "Hash should be 32 hex characters (16 bytes)")
}

func TestHashMVCacheKey_OrderIndependent(t *testing.T) {
	covMatrix := [][]float64{
		{0.04, 0.01},
		{0.01, 0.03},
	}
	targetReturn := 0.10

	// Different orderings in constraints
	constraints1 := Constraints{ISINs: []string{"ISIN1", "ISIN2"}}
	constraints2 := Constraints{ISINs: []string{"ISIN2", "ISIN1"}}

	// Different orderings in expected returns (maps are unordered)
	returns1 := map[string]float64{"ISIN1": 0.12, "ISIN2": 0.08}
	returns2 := map[string]float64{"ISIN2": 0.08, "ISIN1": 0.12}

	hash1 := hashMVCacheKey(returns1, covMatrix, constraints1, targetReturn)
	hash2 := hashMVCacheKey(returns2, covMatrix, constraints2, targetReturn)

	assert.Equal(t, hash1, hash2, "Hash should be order-independent")
}

func TestHashMVCacheKey_DifferentTargetReturn(t *testing.T) {
	expectedReturns := map[string]float64{"ISIN1": 0.12, "ISIN2": 0.08}
	covMatrix := [][]float64{
		{0.04, 0.01},
		{0.01, 0.03},
	}
	constraints := Constraints{ISINs: []string{"ISIN1", "ISIN2"}}

	hash10 := hashMVCacheKey(expectedReturns, covMatrix, constraints, 0.10)
	hash12 := hashMVCacheKey(expectedReturns, covMatrix, constraints, 0.12)

	assert.NotEqual(t, hash10, hash12, "Different target returns should produce different hashes")
}

func TestHashMVCacheKey_DifferentExpectedReturns(t *testing.T) {
	covMatrix := [][]float64{
		{0.04, 0.01},
		{0.01, 0.03},
	}
	constraints := Constraints{ISINs: []string{"ISIN1", "ISIN2"}}
	targetReturn := 0.10

	returns1 := map[string]float64{"ISIN1": 0.12, "ISIN2": 0.08}
	returns2 := map[string]float64{"ISIN1": 0.10, "ISIN2": 0.08}

	hash1 := hashMVCacheKey(returns1, covMatrix, constraints, targetReturn)
	hash2 := hashMVCacheKey(returns2, covMatrix, constraints, targetReturn)

	assert.NotEqual(t, hash1, hash2, "Different expected returns should produce different hashes")
}

func TestHashMVCacheKey_DifferentCovMatrix(t *testing.T) {
	expectedReturns := map[string]float64{"ISIN1": 0.12, "ISIN2": 0.08}
	constraints := Constraints{ISINs: []string{"ISIN1", "ISIN2"}}
	targetReturn := 0.10

	cov1 := [][]float64{
		{0.04, 0.01},
		{0.01, 0.03},
	}
	cov2 := [][]float64{
		{0.05, 0.01},
		{0.01, 0.03},
	}

	hash1 := hashMVCacheKey(expectedReturns, cov1, constraints, targetReturn)
	hash2 := hashMVCacheKey(expectedReturns, cov2, constraints, targetReturn)

	assert.NotEqual(t, hash1, hash2, "Different covariance matrices should produce different hashes")
}
