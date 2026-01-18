package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSettingDefaults_LimitOrderBuffer(t *testing.T) {
	// Verify default exists
	val, exists := SettingDefaults["limit_order_buffer_percent"]
	assert.True(t, exists, "limit_order_buffer_percent must exist in defaults")

	// Verify default value is 0.05 (5%)
	floatVal, ok := val.(float64)
	assert.True(t, ok, "limit_order_buffer_percent must be float64")
	assert.Equal(t, 0.05, floatVal, "default should be 5%")
}

func TestSettingDescriptions_LimitOrderBuffer(t *testing.T) {
	// Verify description exists
	desc, exists := SettingDescriptions["limit_order_buffer_percent"]
	assert.True(t, exists, "limit_order_buffer_percent description must exist")
	assert.NotEmpty(t, desc, "description must not be empty")
	assert.Contains(t, desc, "Buffer", "description should mention buffer")
}

// ============================================================================
// TEMPERAMENT SETTINGS TESTS
// ============================================================================

func TestSettingDefaults_TemperamentRiskTolerance(t *testing.T) {
	// Verify default exists
	val, exists := SettingDefaults["risk_tolerance"]
	assert.True(t, exists, "risk_tolerance must exist in defaults")

	// Verify default value is 0.5 (balanced)
	floatVal, ok := val.(float64)
	assert.True(t, ok, "risk_tolerance must be float64")
	assert.Equal(t, 0.5, floatVal, "default should be 0.5 (balanced)")
}

func TestSettingDefaults_TemperamentAggression(t *testing.T) {
	// Verify default exists
	val, exists := SettingDefaults["temperament_aggression"]
	assert.True(t, exists, "temperament_aggression must exist in defaults")

	// Verify default value is 0.5 (balanced)
	floatVal, ok := val.(float64)
	assert.True(t, ok, "temperament_aggression must be float64")
	assert.Equal(t, 0.5, floatVal, "default should be 0.5 (balanced)")
}

func TestSettingDefaults_TemperamentPatience(t *testing.T) {
	// Verify default exists
	val, exists := SettingDefaults["temperament_patience"]
	assert.True(t, exists, "temperament_patience must exist in defaults")

	// Verify default value is 0.5 (balanced)
	floatVal, ok := val.(float64)
	assert.True(t, ok, "temperament_patience must be float64")
	assert.Equal(t, 0.5, floatVal, "default should be 0.5 (balanced)")
}

func TestSettingDescriptions_TemperamentSettings(t *testing.T) {
	// Verify descriptions exist for all temperament settings
	temperamentSettings := []string{
		"risk_tolerance",
		"temperament_aggression",
		"temperament_patience",
	}

	for _, setting := range temperamentSettings {
		desc, exists := SettingDescriptions[setting]
		assert.True(t, exists, "%s description must exist", setting)
		assert.NotEmpty(t, desc, "%s description must not be empty", setting)
	}
}

func TestTemperamentSettingsInValidRange(t *testing.T) {
	// All temperament settings should default to a value between 0 and 1
	temperamentSettings := []string{
		"risk_tolerance",
		"temperament_aggression",
		"temperament_patience",
	}

	for _, setting := range temperamentSettings {
		val, exists := SettingDefaults[setting]
		assert.True(t, exists, "%s must exist", setting)

		floatVal, ok := val.(float64)
		assert.True(t, ok, "%s must be float64", setting)
		assert.GreaterOrEqual(t, floatVal, 0.0, "%s must be >= 0", setting)
		assert.LessOrEqual(t, floatVal, 1.0, "%s must be <= 1", setting)
	}
}

// ============================================================================
// JOB SCHEDULING SETTINGS TESTS
// ============================================================================

func TestRemovedJobSettingsDoNotExist(t *testing.T) {
	// Settings removed as part of work type interval simplification
	removedSettings := []string{
		"job_sync_cycle_minutes",
		"job_maintenance_hour",
	}

	for _, key := range removedSettings {
		_, exists := SettingDefaults[key]
		assert.False(t, exists, "Setting %s should NOT exist (removed)", key)
	}
}

func TestJobAutoDeployMinutesExists(t *testing.T) {
	// job_auto_deploy_minutes is the only configurable job interval
	val, exists := SettingDefaults["job_auto_deploy_minutes"]
	assert.True(t, exists, "job_auto_deploy_minutes must exist")

	floatVal, ok := val.(float64)
	assert.True(t, ok, "job_auto_deploy_minutes must be float64")
	assert.Equal(t, 5.0, floatVal, "default should be 5.0 minutes")
}

func TestJobAutoDeployMinutesHasDescription(t *testing.T) {
	desc, exists := SettingDescriptions["job_auto_deploy_minutes"]
	assert.True(t, exists, "job_auto_deploy_minutes description must exist")
	assert.NotEmpty(t, desc, "description must not be empty")
	assert.Contains(t, desc, "user-configurable", "description should mention it's user-configurable")
}

// ============================================================================
// COOLOFF PERIOD SETTINGS TESTS
// ============================================================================

func TestSettingDefaults_ContainsCooloffSettings(t *testing.T) {
	// Verify all cooloff settings exist with correct defaults
	assert.Contains(t, SettingDefaults, "buy_cooldown_days", "buy_cooldown_days must exist")
	assert.Contains(t, SettingDefaults, "sell_cooldown_days", "sell_cooldown_days must exist")
	assert.Contains(t, SettingDefaults, "min_hold_days", "min_hold_days must exist")

	// Verify default values
	assert.Equal(t, 30.0, SettingDefaults["buy_cooldown_days"], "buy_cooldown_days default should be 30")
	assert.Equal(t, 180.0, SettingDefaults["sell_cooldown_days"], "sell_cooldown_days default should be 180")
	assert.Equal(t, 90.0, SettingDefaults["min_hold_days"], "min_hold_days default should be 90")
}

func TestCooloffSettingsAreFloats(t *testing.T) {
	// All cooloff settings should be stored as float64
	cooloffSettings := []string{
		"buy_cooldown_days",
		"sell_cooldown_days",
		"min_hold_days",
	}

	for _, setting := range cooloffSettings {
		val, exists := SettingDefaults[setting]
		assert.True(t, exists, "%s must exist", setting)

		_, ok := val.(float64)
		assert.True(t, ok, "%s must be float64", setting)
	}
}

func TestCooloffSettingsInValidRange(t *testing.T) {
	// All cooloff settings should have reasonable default values
	cooloffSettings := map[string]struct {
		min float64
		max float64
	}{
		"buy_cooldown_days":  {0, 365},
		"sell_cooldown_days": {0, 730},
		"min_hold_days":      {0, 365},
	}

	for setting, bounds := range cooloffSettings {
		val, exists := SettingDefaults[setting]
		assert.True(t, exists, "%s must exist", setting)

		floatVal, ok := val.(float64)
		assert.True(t, ok, "%s must be float64", setting)
		assert.GreaterOrEqual(t, floatVal, bounds.min, "%s must be >= %.0f", setting, bounds.min)
		assert.LessOrEqual(t, floatVal, bounds.max, "%s must be <= %.0f", setting, bounds.max)
	}
}
