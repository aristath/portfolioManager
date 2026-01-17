package universe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestApplyDefaultsDoesNotOverwriteMinLot verifies that ApplyDefaults()
// does NOT overwrite MinLot when it's already set from JSON data
func TestApplyDefaultsDoesNotOverwriteMinLot(t *testing.T) {
	// Simulate security parsed from JSON with lot size from Tradernet API
	security := &Security{
		ISIN:   "CNE100006WS8",
		Symbol: "CAT.3750.AS",
		MinLot: 100, // ← From quotes.x_lot in JSON data
	}

	// ApplyDefaults should NOT overwrite MinLot
	ApplyDefaults(security)

	// Assert MinLot is preserved
	assert.Equal(t, 100, security.MinLot, "ApplyDefaults must not overwrite MinLot from JSON data")

	// Assert other defaults are applied
	assert.True(t, security.AllowBuy, "AllowBuy should be defaulted to true")
	assert.True(t, security.AllowSell, "AllowSell should be defaulted to true")
	assert.Equal(t, 1.0, security.PriorityMultiplier, "PriorityMultiplier should be defaulted to 1.0")
}

// TestApplyDefaultsWithZeroMinLot verifies that ApplyDefaults()
// sets MinLot to 1 when it's 0 (not set from JSON)
func TestApplyDefaultsWithZeroMinLot(t *testing.T) {
	// Simulate security where JSON extraction failed or didn't provide MinLot
	security := &Security{
		ISIN:   "TEST123456789",
		Symbol: "TEST.US",
		MinLot: 0, // ← Not provided by API or extraction failed
	}

	// ApplyDefaults should set MinLot to default only if it's 0
	ApplyDefaults(security)

	// Assert MinLot gets default value
	assert.Equal(t, 1, security.MinLot, "ApplyDefaults should set MinLot to 1 when it's 0")

	// Assert other defaults are applied
	assert.True(t, security.AllowBuy, "AllowBuy should be defaulted to true")
	assert.True(t, security.AllowSell, "AllowSell should be defaulted to true")
	assert.Equal(t, 1.0, security.PriorityMultiplier, "PriorityMultiplier should be defaulted to 1.0")
}

// TestApplyDefaultsWithNilSecurity verifies that ApplyDefaults()
// handles nil security gracefully
func TestApplyDefaultsWithNilSecurity(t *testing.T) {
	// Should not panic
	assert.NotPanics(t, func() {
		ApplyDefaults(nil)
	}, "ApplyDefaults should handle nil security without panicking")
}

// TestApplyOverridesPreservesNonOverriddenFields verifies that ApplyOverrides()
// only modifies fields that have overrides, leaving others unchanged
func TestApplyOverridesPreservesNonOverriddenFields(t *testing.T) {
	security := &Security{
		ISIN:               "TEST123456789",
		Symbol:             "TEST.US",
		MinLot:             100, // From JSON data
		AllowBuy:           true,
		AllowSell:          true,
		PriorityMultiplier: 1.0,
	}

	// Apply override only for AllowBuy
	overrides := map[string]string{
		"allow_buy": "false",
	}

	ApplyOverrides(security, overrides)

	// Assert only AllowBuy was changed
	assert.False(t, security.AllowBuy, "AllowBuy should be overridden to false")

	// Assert other fields remain unchanged
	assert.Equal(t, 100, security.MinLot, "MinLot should not be changed by unrelated override")
	assert.True(t, security.AllowSell, "AllowSell should remain true")
	assert.Equal(t, 1.0, security.PriorityMultiplier, "PriorityMultiplier should remain 1.0")
}

// TestApplyOverridesMinLot verifies that MinLot can be overridden when explicitly set
func TestApplyOverridesMinLot(t *testing.T) {
	security := &Security{
		ISIN:   "TEST123456789",
		Symbol: "TEST.US",
		MinLot: 100, // From JSON data
	}

	// Explicitly override MinLot
	overrides := map[string]string{
		"min_lot": "50",
	}

	ApplyOverrides(security, overrides)

	// Assert MinLot was overridden
	assert.Equal(t, 50, security.MinLot, "MinLot should be overridden to 50")
}

// TestApplyOverridesWithEmptyValue verifies that empty override values are skipped
func TestApplyOverridesWithEmptyValue(t *testing.T) {
	security := &Security{
		ISIN:     "TEST123456789",
		Symbol:   "TEST.US",
		AllowBuy: true,
	}

	// Empty value should be skipped (use default)
	overrides := map[string]string{
		"allow_buy": "",
	}

	ApplyOverrides(security, overrides)

	// Assert AllowBuy remains unchanged
	assert.True(t, security.AllowBuy, "AllowBuy should remain true when override value is empty")
}

// TestApplyOverridesWithNilSecurity verifies that ApplyOverrides()
// handles nil security gracefully
func TestApplyOverridesWithNilSecurity(t *testing.T) {
	overrides := map[string]string{
		"allow_buy": "false",
	}

	// Should not panic
	assert.NotPanics(t, func() {
		ApplyOverrides(nil, overrides)
	}, "ApplyOverrides should handle nil security without panicking")
}

// TestApplyOverridesWithNilOverrides verifies that ApplyOverrides()
// handles nil overrides map gracefully
func TestApplyOverridesWithNilOverrides(t *testing.T) {
	security := &Security{
		ISIN:   "TEST123456789",
		Symbol: "TEST.US",
	}

	// Should not panic
	assert.NotPanics(t, func() {
		ApplyOverrides(security, nil)
	}, "ApplyOverrides should handle nil overrides without panicking")
}
