package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand_Structure(t *testing.T) {
	// Test root command
	t.Run("root command exists", func(t *testing.T) {
		cmd := rootCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "cu", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		
		// Should have subcommands
		assert.NotEmpty(t, cmd.Commands())
	})
	
	// Test root command has all expected subcommands
	t.Run("root has expected subcommands", func(t *testing.T) {
		subcommandNames := make(map[string]bool)
		for _, subcmd := range rootCmd.Commands() {
			subcommandNames[subcmd.Name()] = true
		}
		
		// Check for major command categories
		expectedCommands := []string{
			"auth",
			"task",
			"list",
			"space",
			"user",
			"config",
			"api",
			"bulk",
			"export",
			"interactive",
			"me",
			"version",
			"completion",
			"comment",
			"cache",
		}
		
		for _, expected := range expectedCommands {
			assert.True(t, subcommandNames[expected], "Expected command '%s' to exist", expected)
		}
	})
	
	// Test persistent flags
	t.Run("root persistent flags", func(t *testing.T) {
		// Check for config flag
		configFlag := rootCmd.PersistentFlags().Lookup("config")
		assert.NotNil(t, configFlag, "config persistent flag should exist")
		
		// Check for debug flag
		debugFlag := rootCmd.PersistentFlags().Lookup("debug")
		assert.NotNil(t, debugFlag, "debug persistent flag should exist")
		
		// Check for output flag
		outputFlag := rootCmd.PersistentFlags().Lookup("output")
		assert.NotNil(t, outputFlag, "output persistent flag should exist")
		assert.Equal(t, "o", outputFlag.Shorthand)
	})
}