package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestUserCommand_Structure(t *testing.T) {
	// Test main user command
	t.Run("user command exists", func(t *testing.T) {
		cmd := userCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "user", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		
		// Should have list subcommand at minimum
		assert.NotEmpty(t, cmd.Commands())
	})
	
	// Test user list command
	t.Run("user list command", func(t *testing.T) {
		// Find the list subcommand
		var listCmd *cobra.Command
		for _, subcmd := range userCmd.Commands() {
			if subcmd.Use == "list" {
				listCmd = subcmd
				break
			}
		}
		
		if assert.NotNil(t, listCmd, "list subcommand should exist") {
			assert.NotEmpty(t, listCmd.Short)
			assert.NotNil(t, listCmd.Run)
			
			// Check for common flags
			assert.NotNil(t, listCmd.Flags())
		}
	})
}