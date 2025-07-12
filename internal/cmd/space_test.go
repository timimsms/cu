package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSpaceCommand_Structure(t *testing.T) {
	// Test main space command
	t.Run("space command exists", func(t *testing.T) {
		cmd := spaceCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "space", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		
		// Should have list subcommand at minimum
		assert.NotEmpty(t, cmd.Commands())
	})
	
	// Test space list command
	t.Run("space list command", func(t *testing.T) {
		// Find the list subcommand
		var listCmd *cobra.Command
		for _, subcmd := range spaceCmd.Commands() {
			if subcmd.Use == "list" {
				listCmd = subcmd
				break
			}
		}
		
		if assert.NotNil(t, listCmd, "list subcommand should exist") {
			assert.NotEmpty(t, listCmd.Short)
			assert.NotNil(t, listCmd.Run)
		}
	})
}