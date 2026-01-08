package database

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		profile  DatabaseProfile
		contains []string // Strings that should be present in connection string
	}{
		{
			name:    "standard profile",
			path:    "/path/to/db.sqlite",
			profile: ProfileStandard,
			contains: []string{
				"/path/to/db.sqlite",
				"journal_mode(WAL)",
				"synchronous(NORMAL)",
				"auto_vacuum(INCREMENTAL)",
				"temp_store(MEMORY)",
				"foreign_keys(1)",
				"wal_autocheckpoint(1000)",
				"cache_size(-64000)",
			},
		},
		{
			name:    "ledger profile",
			path:    "/path/to/ledger.sqlite",
			profile: ProfileLedger,
			contains: []string{
				"/path/to/ledger.sqlite",
				"journal_mode(WAL)",
				"synchronous(FULL)",
				"auto_vacuum(NONE)",
				"foreign_keys(1)",
			},
		},
		{
			name:    "cache profile",
			path:    "/path/to/cache.sqlite",
			profile: ProfileCache,
			contains: []string{
				"/path/to/cache.sqlite",
				"journal_mode(WAL)",
				"synchronous(OFF)",
				"auto_vacuum(FULL)",
				"temp_store(MEMORY)",
				"foreign_keys(1)",
			},
		},
		{
			name:    "empty profile defaults",
			path:    "/path/to/db.sqlite",
			profile: "",
			contains: []string{
				"/path/to/db.sqlite",
				"journal_mode(WAL)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildConnectionString(tt.path, tt.profile)

			// Should start with the path
			assert.True(t, strings.HasPrefix(result, tt.path), "Connection string should start with path")

			// Should contain all expected strings
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected, "Connection string should contain %s", expected)
			}

			// Should not contain conflicting settings
			if tt.profile == ProfileLedger {
				assert.NotContains(t, result, "synchronous(OFF)", "Ledger should not have synchronous(OFF)")
				assert.NotContains(t, result, "synchronous(NORMAL)", "Ledger should not have synchronous(NORMAL)")
			}

			if tt.profile == ProfileCache {
				assert.NotContains(t, result, "synchronous(FULL)", "Cache should not have synchronous(FULL)")
				assert.NotContains(t, result, "synchronous(NORMAL)", "Cache should not have synchronous(NORMAL)")
			}

			if tt.profile == ProfileStandard {
				assert.NotContains(t, result, "synchronous(OFF)", "Standard should not have synchronous(OFF)")
				assert.NotContains(t, result, "synchronous(FULL)", "Standard should not have synchronous(FULL)")
			}
		})
	}
}
