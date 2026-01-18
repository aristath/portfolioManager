package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DataDir_DefaultWhenNotSet(t *testing.T) {
	// Save original environment
	originalTraderDataDir := os.Getenv("TRADER_DATA_DIR")
	originalDataDir := os.Getenv("DATA_DIR")
	defer func() {
		if originalTraderDataDir != "" {
			os.Setenv("TRADER_DATA_DIR", originalTraderDataDir)
		} else {
			os.Unsetenv("TRADER_DATA_DIR")
		}
		if originalDataDir != "" {
			os.Setenv("DATA_DIR", originalDataDir)
		} else {
			os.Unsetenv("DATA_DIR")
		}
	}()

	// Clear environment variables
	os.Unsetenv("TRADER_DATA_DIR")
	os.Unsetenv("DATA_DIR")

	// Use a temporary directory that we can actually create
	// This tests the default behavior without requiring /home/arduino to exist
	tmpDir := t.TempDir()
	os.Setenv("TRADER_DATA_DIR", tmpDir)

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Should use the temporary directory, resolved to absolute
	absPath, err := filepath.Abs(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, absPath, cfg.DataDir)

	// Now test the actual default by clearing TRADER_DATA_DIR
	// On macOS, this will fail to create /home/arduino/data, which is expected
	os.Unsetenv("TRADER_DATA_DIR")

	// This will fail on macOS but succeed on Linux (target system)
	_, err = Load()
	// On macOS, we expect this to fail because /home/arduino doesn't exist
	// On Linux, it should succeed
	if err != nil {
		// Verify the error is about directory creation (expected on macOS)
		assert.Contains(t, err.Error(), "failed to create data directory")
	} else {
		// On Linux, verify the default path is used
		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "/home/arduino/data", cfg.DataDir)
	}
}

func TestLoad_DataDir_FromTRADER_DATA_DIR(t *testing.T) {
	// Save original environment
	originalTraderDataDir := os.Getenv("TRADER_DATA_DIR")
	originalDataDir := os.Getenv("DATA_DIR")
	defer func() {
		if originalTraderDataDir != "" {
			os.Setenv("TRADER_DATA_DIR", originalTraderDataDir)
		} else {
			os.Unsetenv("TRADER_DATA_DIR")
		}
		if originalDataDir != "" {
			os.Setenv("DATA_DIR", originalDataDir)
		} else {
			os.Unsetenv("DATA_DIR")
		}
	}()

	// Set TRADER_DATA_DIR to a test path
	testPath := "/tmp/test-trader-data"
	os.Setenv("TRADER_DATA_DIR", testPath)
	os.Unsetenv("DATA_DIR")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Should use the value from TRADER_DATA_DIR, resolved to absolute
	absPath, err := filepath.Abs(testPath)
	require.NoError(t, err)
	assert.Equal(t, absPath, cfg.DataDir)
}

func TestLoad_DataDir_IgnoresOldDATA_DIR(t *testing.T) {
	// Save original environment
	originalTraderDataDir := os.Getenv("TRADER_DATA_DIR")
	originalDataDir := os.Getenv("DATA_DIR")
	defer func() {
		if originalTraderDataDir != "" {
			os.Setenv("TRADER_DATA_DIR", originalTraderDataDir)
		} else {
			os.Unsetenv("TRADER_DATA_DIR")
		}
		if originalDataDir != "" {
			os.Setenv("DATA_DIR", originalDataDir)
		} else {
			os.Unsetenv("DATA_DIR")
		}
	}()

	// Set old DATA_DIR but not TRADER_DATA_DIR
	// Use a temporary directory that we can actually create
	tmpDir := t.TempDir()
	os.Setenv("DATA_DIR", tmpDir)
	os.Unsetenv("TRADER_DATA_DIR")

	// On macOS, this will fail to create /home/arduino/data, which is expected
	_, err := Load()
	// On macOS, we expect this to fail because /home/arduino doesn't exist
	// On Linux, it should succeed
	if err != nil {
		// Verify the error is about directory creation (expected on macOS)
		assert.Contains(t, err.Error(), "failed to create data directory")
		// Verify that DATA_DIR was ignored (the error is about /home/arduino/data, not the tmp dir)
		assert.NotContains(t, err.Error(), tmpDir)
	} else {
		// On Linux, verify the default path is used and DATA_DIR is ignored
		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "/home/arduino/data", cfg.DataDir)
		assert.NotEqual(t, tmpDir, cfg.DataDir)
	}
}

func TestLoad_DataDir_TRADER_DATA_DIRTakesPrecedence(t *testing.T) {
	// Save original environment
	originalTraderDataDir := os.Getenv("TRADER_DATA_DIR")
	originalDataDir := os.Getenv("DATA_DIR")
	defer func() {
		if originalTraderDataDir != "" {
			os.Setenv("TRADER_DATA_DIR", originalTraderDataDir)
		} else {
			os.Unsetenv("TRADER_DATA_DIR")
		}
		if originalDataDir != "" {
			os.Setenv("DATA_DIR", originalDataDir)
		} else {
			os.Unsetenv("DATA_DIR")
		}
	}()

	// Set both, TRADER_DATA_DIR should take precedence
	traderDataDir := "/tmp/trader-data-dir"
	oldDataDir := "/tmp/old-data-dir"
	os.Setenv("TRADER_DATA_DIR", traderDataDir)
	os.Setenv("DATA_DIR", oldDataDir)

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Should use TRADER_DATA_DIR, not DATA_DIR
	absPath, err := filepath.Abs(traderDataDir)
	require.NoError(t, err)
	assert.Equal(t, absPath, cfg.DataDir)
	assert.NotEqual(t, oldDataDir, cfg.DataDir)
}

func TestLoad_DataDir_ResolvesRelativeToAbsolute(t *testing.T) {
	// Save original environment
	originalTraderDataDir := os.Getenv("TRADER_DATA_DIR")
	originalDataDir := os.Getenv("DATA_DIR")
	defer func() {
		if originalTraderDataDir != "" {
			os.Setenv("TRADER_DATA_DIR", originalTraderDataDir)
		} else {
			os.Unsetenv("TRADER_DATA_DIR")
		}
		if originalDataDir != "" {
			os.Setenv("DATA_DIR", originalDataDir)
		} else {
			os.Unsetenv("DATA_DIR")
		}
	}()

	// Set relative path
	os.Setenv("TRADER_DATA_DIR", "./relative/path")
	os.Unsetenv("DATA_DIR")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Should be resolved to absolute path
	assert.True(t, filepath.IsAbs(cfg.DataDir), "DataDir should be absolute")

	// Verify it resolves correctly
	expectedAbs, err := filepath.Abs("./relative/path")
	require.NoError(t, err)
	assert.Equal(t, expectedAbs, cfg.DataDir)
}

func TestLoad_DataDir_ResolvesAbsolutePath(t *testing.T) {
	// Save original environment
	originalTraderDataDir := os.Getenv("TRADER_DATA_DIR")
	originalDataDir := os.Getenv("DATA_DIR")
	defer func() {
		if originalTraderDataDir != "" {
			os.Setenv("TRADER_DATA_DIR", originalTraderDataDir)
		} else {
			os.Unsetenv("TRADER_DATA_DIR")
		}
		if originalDataDir != "" {
			os.Setenv("DATA_DIR", originalDataDir)
		} else {
			os.Unsetenv("DATA_DIR")
		}
	}()

	// Set absolute path
	absPath := "/tmp/absolute-test-path"
	os.Setenv("TRADER_DATA_DIR", absPath)
	os.Unsetenv("DATA_DIR")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Should remain absolute (already absolute)
	assert.True(t, filepath.IsAbs(cfg.DataDir), "DataDir should be absolute")
	assert.Equal(t, absPath, cfg.DataDir)
}

func TestLoad_DataDir_CreatesDirectoryIfNeeded(t *testing.T) {
	// Save original environment
	originalTraderDataDir := os.Getenv("TRADER_DATA_DIR")
	originalDataDir := os.Getenv("DATA_DIR")
	defer func() {
		if originalTraderDataDir != "" {
			os.Setenv("TRADER_DATA_DIR", originalTraderDataDir)
		} else {
			os.Unsetenv("TRADER_DATA_DIR")
		}
		if originalDataDir != "" {
			os.Setenv("DATA_DIR", originalDataDir)
		} else {
			os.Unsetenv("DATA_DIR")
		}
	}()

	// Use temporary directory that doesn't exist
	tmpDir := filepath.Join(t.TempDir(), "new-data-dir")
	os.Setenv("TRADER_DATA_DIR", tmpDir)
	os.Unsetenv("DATA_DIR")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Directory should be created
	absPath, err := filepath.Abs(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, absPath, cfg.DataDir)

	// Verify directory exists
	info, err := os.Stat(cfg.DataDir)
	require.NoError(t, err, "Directory should be created")
	assert.True(t, info.IsDir(), "Should be a directory")
}

func TestLoad_DataDir_WithTemporaryDirectory(t *testing.T) {
	// Save original environment
	originalTraderDataDir := os.Getenv("TRADER_DATA_DIR")
	originalDataDir := os.Getenv("DATA_DIR")
	defer func() {
		if originalTraderDataDir != "" {
			os.Setenv("TRADER_DATA_DIR", originalTraderDataDir)
		} else {
			os.Unsetenv("TRADER_DATA_DIR")
		}
		if originalDataDir != "" {
			os.Setenv("DATA_DIR", originalDataDir)
		} else {
			os.Unsetenv("DATA_DIR")
		}
	}()

	// Use temporary directory
	tmpDir := t.TempDir()
	os.Setenv("TRADER_DATA_DIR", tmpDir)
	os.Unsetenv("DATA_DIR")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Should use the temporary directory, resolved to absolute
	absPath, err := filepath.Abs(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, absPath, cfg.DataDir)
}

func TestLoad_DataDir_CLIFlagTakesPrecedence(t *testing.T) {
	// Save original environment
	originalTraderDataDir := os.Getenv("TRADER_DATA_DIR")
	originalDataDir := os.Getenv("DATA_DIR")
	defer func() {
		if originalTraderDataDir != "" {
			os.Setenv("TRADER_DATA_DIR", originalTraderDataDir)
		} else {
			os.Unsetenv("TRADER_DATA_DIR")
		}
		if originalDataDir != "" {
			os.Setenv("DATA_DIR", originalDataDir)
		} else {
			os.Unsetenv("DATA_DIR")
		}
	}()

	// Set TRADER_DATA_DIR environment variable
	envDataDir := t.TempDir()
	os.Setenv("TRADER_DATA_DIR", envDataDir)
	os.Unsetenv("DATA_DIR")

	// Test with CLI override
	cliDataDir := t.TempDir()
	cfg, err := Load(cliDataDir)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// CLI flag should take precedence over environment variable
	absPath, err := filepath.Abs(cliDataDir)
	require.NoError(t, err)
	assert.Equal(t, absPath, cfg.DataDir)
	assert.NotEqual(t, envDataDir, cfg.DataDir)
}

func TestLoad_DataDir_CLIFlagEmptyString(t *testing.T) {
	// Save original environment
	originalTraderDataDir := os.Getenv("TRADER_DATA_DIR")
	originalDataDir := os.Getenv("DATA_DIR")
	defer func() {
		if originalTraderDataDir != "" {
			os.Setenv("TRADER_DATA_DIR", originalTraderDataDir)
		} else {
			os.Unsetenv("TRADER_DATA_DIR")
		}
		if originalDataDir != "" {
			os.Setenv("DATA_DIR", originalDataDir)
		} else {
			os.Unsetenv("DATA_DIR")
		}
	}()

	// Set TRADER_DATA_DIR environment variable
	envDataDir := t.TempDir()
	os.Setenv("TRADER_DATA_DIR", envDataDir)
	os.Unsetenv("DATA_DIR")

	// Test with empty string CLI override (should fall back to env var)
	cfg, err := Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Should fall back to environment variable when CLI flag is empty
	absPath, err := filepath.Abs(envDataDir)
	require.NoError(t, err)
	assert.Equal(t, absPath, cfg.DataDir)
}

// TestToDeploymentConfig tests the ToDeploymentConfig conversion
func TestToDeploymentConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         *DeploymentConfig
		expectedBranch string
	}{
		{
			name: "with GitHubBranch set",
			config: &DeploymentConfig{
				Enabled:                true,
				GitBranch:              "main",
				GitHubBranch:           "develop",
				TraderBinaryName:       "sentinel",
				TraderServiceName:      "sentinel",
				LockTimeout:            120,
				HealthCheckTimeout:     10,
				HealthCheckMaxAttempts: 3,
			},
			expectedBranch: "develop",
		},
		{
			name: "without GitHubBranch uses GitBranch",
			config: &DeploymentConfig{
				Enabled:                true,
				GitBranch:              "main",
				GitHubBranch:           "",
				TraderBinaryName:       "sentinel",
				TraderServiceName:      "sentinel",
				LockTimeout:            120,
				HealthCheckTimeout:     10,
				HealthCheckMaxAttempts: 3,
			},
			expectedBranch: "main",
		},
		{
			name: "GitHub artifact settings",
			config: &DeploymentConfig{
				Enabled:                true,
				UseGitHubArtifacts:     true,
				GitHubWorkflowName:     "build-go.yml",
				GitHubArtifactName:     "sentinel-arm64",
				GitHubBranch:           "release",
				GitHubRepo:             "test/repo",
				TraderBinaryName:       "sentinel",
				TraderServiceName:      "sentinel",
				LockTimeout:            60,
				HealthCheckTimeout:     5,
				HealthCheckMaxAttempts: 5,
			},
			expectedBranch: "release",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pass test GitHub token
			result := tt.config.ToDeploymentConfig("test_github_token")

			assert.NotNil(t, result)
			assert.Equal(t, tt.config.Enabled, result.Enabled)
			assert.Equal(t, tt.config.DeployDir, result.DeployDir)
			assert.Equal(t, tt.config.APIPort, result.APIPort)
			assert.Equal(t, tt.config.APIHost, result.APIHost)
			assert.Equal(t, time.Duration(tt.config.LockTimeout)*time.Second, result.LockTimeout)
			assert.Equal(t, time.Duration(tt.config.HealthCheckTimeout)*time.Second, result.HealthCheckTimeout)
			assert.Equal(t, tt.config.HealthCheckMaxAttempts, result.HealthCheckMaxAttempts)
			assert.Equal(t, tt.config.GitBranch, result.GitBranch)
			assert.Equal(t, tt.expectedBranch, result.GitHubBranch)
			assert.Equal(t, tt.config.TraderBinaryName, result.TraderConfig.BinaryName)
			assert.Equal(t, tt.config.TraderServiceName, result.TraderConfig.ServiceName)
			assert.Equal(t, tt.config.UseGitHubArtifacts, result.UseGitHubArtifacts)
			assert.Equal(t, tt.config.GitHubWorkflowName, result.GitHubWorkflowName)
			assert.Equal(t, tt.config.GitHubArtifactName, result.GitHubArtifactName)
			assert.Equal(t, tt.config.GitHubRepo, result.GitHubRepo)
		})
	}
}

// TestGetEnv tests the getEnv helper function indirectly through Load
// Since getEnv is package-private, we test it indirectly
func TestConfig_EnvironmentVariables(t *testing.T) {
	// Save original environment
	originalPort := os.Getenv("GO_PORT")
	originalDevMode := os.Getenv("DEV_MODE")
	originalLogLevel := os.Getenv("LOG_LEVEL")
	originalEvalURL := os.Getenv("EVALUATOR_SERVICE_URL")
	originalTraderDataDir := os.Getenv("TRADER_DATA_DIR")
	defer func() {
		if originalPort != "" {
			os.Setenv("GO_PORT", originalPort)
		} else {
			os.Unsetenv("GO_PORT")
		}
		if originalDevMode != "" {
			os.Setenv("DEV_MODE", originalDevMode)
		} else {
			os.Unsetenv("DEV_MODE")
		}
		if originalLogLevel != "" {
			os.Setenv("LOG_LEVEL", originalLogLevel)
		} else {
			os.Unsetenv("LOG_LEVEL")
		}
		if originalEvalURL != "" {
			os.Setenv("EVALUATOR_SERVICE_URL", originalEvalURL)
		} else {
			os.Unsetenv("EVALUATOR_SERVICE_URL")
		}
		if originalTraderDataDir != "" {
			os.Setenv("TRADER_DATA_DIR", originalTraderDataDir)
		} else {
			os.Unsetenv("TRADER_DATA_DIR")
		}
	}()

	tmpDir := t.TempDir()
	os.Setenv("TRADER_DATA_DIR", tmpDir)

	t.Run("GO_PORT as int", func(t *testing.T) {
		os.Setenv("GO_PORT", "9000")
		os.Unsetenv("DEV_MODE")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("EVALUATOR_SERVICE_URL")

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, 9000, cfg.Port)
	})

	t.Run("GO_PORT invalid defaults", func(t *testing.T) {
		os.Setenv("GO_PORT", "invalid")
		os.Unsetenv("DEV_MODE")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("EVALUATOR_SERVICE_URL")

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, 8001, cfg.Port) // Should default to 8001
	})

	t.Run("DEV_MODE as bool", func(t *testing.T) {
		os.Unsetenv("GO_PORT")
		os.Setenv("DEV_MODE", "true")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("EVALUATOR_SERVICE_URL")

		cfg, err := Load()
		require.NoError(t, err)
		assert.True(t, cfg.DevMode)
	})

	t.Run("DEV_MODE false", func(t *testing.T) {
		os.Unsetenv("GO_PORT")
		os.Setenv("DEV_MODE", "false")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("EVALUATOR_SERVICE_URL")

		cfg, err := Load()
		require.NoError(t, err)
		assert.False(t, cfg.DevMode)
	})

	t.Run("DEV_MODE invalid defaults to false", func(t *testing.T) {
		os.Unsetenv("GO_PORT")
		os.Setenv("DEV_MODE", "invalid")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("EVALUATOR_SERVICE_URL")

		cfg, err := Load()
		require.NoError(t, err)
		assert.False(t, cfg.DevMode) // Should default to false
	})

	t.Run("LOG_LEVEL from env", func(t *testing.T) {
		os.Unsetenv("GO_PORT")
		os.Unsetenv("DEV_MODE")
		os.Setenv("LOG_LEVEL", "debug")
		os.Unsetenv("EVALUATOR_SERVICE_URL")

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "debug", cfg.LogLevel)
	})

	t.Run("LOG_LEVEL defaults to info", func(t *testing.T) {
		os.Unsetenv("GO_PORT")
		os.Unsetenv("DEV_MODE")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("EVALUATOR_SERVICE_URL")

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "info", cfg.LogLevel) // Should default to "info"
	})

	t.Run("EVALUATOR_SERVICE_URL from env", func(t *testing.T) {
		os.Unsetenv("GO_PORT")
		os.Unsetenv("DEV_MODE")
		os.Unsetenv("LOG_LEVEL")
		os.Setenv("EVALUATOR_SERVICE_URL", "http://custom:9999")

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "http://custom:9999", cfg.EvaluatorServiceURL)
	})

	t.Run("EVALUATOR_SERVICE_URL defaults", func(t *testing.T) {
		os.Unsetenv("GO_PORT")
		os.Unsetenv("DEV_MODE")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("EVALUATOR_SERVICE_URL")

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:9000", cfg.EvaluatorServiceURL) // Should default
	})
}
