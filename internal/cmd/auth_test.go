package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthCommand_Structure(t *testing.T) {
	// Test main auth command
	t.Run("auth command exists", func(t *testing.T) {
		cmd := authCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "auth", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		
		// Should have subcommands
		assert.NotEmpty(t, cmd.Commands())
	})
	
	// Test auth subcommands
	t.Run("auth subcommands exist", func(t *testing.T) {
		subcommandNames := make(map[string]bool)
		for _, subcmd := range authCmd.Commands() {
			subcommandNames[subcmd.Use] = true
		}
		
		// Check for expected subcommands
		expectedSubcommands := []string{"login", "logout", "status"}
		for _, expected := range expectedSubcommands {
			assert.True(t, subcommandNames[expected], "Expected subcommand '%s' to exist", expected)
		}
	})
	
	// Test auth login command
	t.Run("auth login command", func(t *testing.T) {
		cmd := authLoginCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "login", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotNil(t, cmd.Run)
		
		// Check for token flag
		tokenFlag := cmd.Flag("token")
		assert.NotNil(t, tokenFlag, "token flag should exist")
		
		// Check for workspace flag
		workspaceFlag := cmd.Flag("workspace")
		assert.NotNil(t, workspaceFlag, "workspace flag should exist")
	})
	
	// Test auth logout command
	t.Run("auth logout command", func(t *testing.T) {
		cmd := authLogoutCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "logout", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotNil(t, cmd.Run)
	})
	
	// Test auth status command
	t.Run("auth status command", func(t *testing.T) {
		cmd := authStatusCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "status", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotNil(t, cmd.Run)
	})
}