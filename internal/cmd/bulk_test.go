package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestBulkCommand_Structure(t *testing.T) {
	// Test main bulk command
	t.Run("bulk command exists", func(t *testing.T) {
		cmd := bulkCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "bulk", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		
		// Should have subcommands
		assert.NotEmpty(t, cmd.Commands())
	})
	
	// Test bulk subcommands
	t.Run("bulk subcommands exist", func(t *testing.T) {
		subcommandNames := make(map[string]bool)
		for _, subcmd := range bulkCmd.Commands() {
			// Extract the base command name (before space)
			baseName := strings.Split(subcmd.Use, " ")[0]
			subcommandNames[baseName] = true
		}
		
		// Check for expected subcommands
		expectedSubcommands := []string{"update", "close", "delete"}
		for _, expected := range expectedSubcommands {
			assert.True(t, subcommandNames[expected], "Expected subcommand '%s' to exist", expected)
		}
	})
	
	// Test bulk update command uses
	t.Run("bulk update command uses", func(t *testing.T) {
		// Find the update subcommand
		var updateCmd *cobra.Command
		for _, subcmd := range bulkCmd.Commands() {
			if strings.HasPrefix(subcmd.Use, "update") {
				updateCmd = subcmd
				break
			}
		}
		
		if assert.NotNil(t, updateCmd, "update subcommand should exist") {
			assert.NotEmpty(t, updateCmd.Short)
			assert.NotNil(t, updateCmd.Run)
		}
	})
	
	// Test bulk close command
	t.Run("bulk close command", func(t *testing.T) {
		// Find the close subcommand
		var closeCmd *cobra.Command
		for _, subcmd := range bulkCmd.Commands() {
			if strings.HasPrefix(subcmd.Use, "close") {
				closeCmd = subcmd
				break
			}
		}
		
		if assert.NotNil(t, closeCmd, "close subcommand should exist") {
			assert.NotEmpty(t, closeCmd.Short)
			assert.NotNil(t, closeCmd.Run)
		}
	})
	
	// Test bulk delete command
	t.Run("bulk delete command", func(t *testing.T) {
		// Find the delete subcommand
		var deleteCmd *cobra.Command
		for _, subcmd := range bulkCmd.Commands() {
			if strings.HasPrefix(subcmd.Use, "delete") {
				deleteCmd = subcmd
				break
			}
		}
		
		if assert.NotNil(t, deleteCmd, "delete subcommand should exist") {
			assert.NotEmpty(t, deleteCmd.Short)
			assert.NotNil(t, deleteCmd.Run)
		}
	})
}