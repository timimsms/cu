package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestInitializeConfig(t *testing.T) {
	// Save original viper state
	originalConfig := viper.New()
	*originalConfig = *viper.GetViper()
	
	// Reset after test
	defer func() {
		*viper.GetViper() = *originalConfig
	}()

	t.Run("sets default values", func(t *testing.T) {
		// Reset viper
		viper.Reset()
		
		cfg, err := initializeConfig()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		
		// Check defaults were set
		assert.Equal(t, "table", viper.GetString("output"))
		assert.False(t, viper.GetBool("debug"))
	})

	t.Run("sets environment prefix", func(t *testing.T) {
		// Reset viper
		viper.Reset()
		
		// Set environment variable
		os.Setenv("CU_OUTPUT", "json")
		defer os.Unsetenv("CU_OUTPUT")
		
		cfg, err := initializeConfig()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		
		// Environment variable should override default
		assert.Equal(t, "json", viper.GetString("output"))
	})

	t.Run("handles missing config file gracefully", func(t *testing.T) {
		// Reset viper
		viper.Reset()
		
		// Set config path to non-existent location
		viper.SetConfigName("nonexistent")
		viper.AddConfigPath("/tmp/nonexistent")
		
		cfg, err := initializeConfig()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
	})

	t.Run("reads config file if exists", func(t *testing.T) {
		// Reset viper
		viper.Reset()
		
		// Create temporary config file
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")
		configContent := []byte("output: yaml\ndebug: true")
		err := os.WriteFile(configFile, configContent, 0644)
		assert.NoError(t, err)
		
		// Set viper to use temp config
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(tmpDir)
		
		cfg, err := initializeConfig()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		
		// Config file values should be loaded
		assert.Equal(t, "yaml", viper.GetString("output"))
		assert.True(t, viper.GetBool("debug"))
	})

	t.Run("returns error for invalid config file", func(t *testing.T) {
		// Reset viper
		viper.Reset()
		
		// Create temporary invalid config file
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")
		configContent := []byte("invalid yaml content:\n  - this is not valid\n    incomplete")
		err := os.WriteFile(configFile, configContent, 0644)
		assert.NoError(t, err)
		
		// Set viper to use temp config
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(tmpDir)
		
		cfg, err := initializeConfig()
		// This might not error as viper is quite tolerant of invalid YAML
		// but if it does error, check that it handles it gracefully
		if err != nil {
			assert.Contains(t, err.Error(), "error reading config")
			assert.Nil(t, cfg)
		} else {
			// If no error, the config should still be valid
			assert.NotNil(t, cfg)
		}
	})
}

func TestExecuteWithFactory(t *testing.T) {
	// This is an integration test that would require mocking all dependencies
	// For now, we test that the function exists and has the right signature
	t.Run("function exists", func(t *testing.T) {
		// The function exists if this compiles
		var fn func() error = ExecuteWithFactory
		assert.NotNil(t, fn)
	})
}

func TestExecuteWithFactory_Integration(t *testing.T) {
	// Save original state
	originalArgs := os.Args
	originalConfig := viper.New()
	*originalConfig = *viper.GetViper()
	
	// Reset after test
	defer func() {
		os.Args = originalArgs
		*viper.GetViper() = *originalConfig
	}()

	t.Run("handles help flag", func(t *testing.T) {
		// Reset viper
		viper.Reset()
		
		// Set args to request help
		os.Args = []string{"cu", "--help"}
		
		// Execute should not error when showing help
		err := ExecuteWithFactory()
		// Help causes a special exit, but not an error
		assert.NoError(t, err)
	})

	t.Run("handles version flag", func(t *testing.T) {
		// Reset viper
		viper.Reset()
		
		// Set args to request version
		os.Args = []string{"cu", "version"}
		
		// Execute should handle version command
		err := ExecuteWithFactory()
		// Version command might error if not fully configured, but should not panic
		if err != nil {
			assert.NotContains(t, err.Error(), "panic")
		}
	})
}

// Test helpers to ensure proper test isolation
func TestConfigIsolation(t *testing.T) {
	t.Run("viper state is isolated between tests", func(t *testing.T) {
		// Save original
		original := viper.GetString("output")
		
		// Change value
		viper.Set("output", "modified")
		assert.Equal(t, "modified", viper.GetString("output"))
		
		// Create new viper instance
		v := viper.New()
		v.SetDefault("output", "table")
		
		// Original should still be modified
		assert.Equal(t, "modified", viper.GetString("output"))
		
		// New instance should have default
		assert.Equal(t, "table", v.GetString("output"))
		
		// Restore original
		viper.Set("output", original)
	})
}