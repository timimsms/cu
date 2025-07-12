package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListCommands_Structure(t *testing.T) {
	// Test main list command
	t.Run("list command exists", func(t *testing.T) {
		cmd := listCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "list", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		
		// Should have subcommands
		assert.NotEmpty(t, cmd.Commands())
	})
	
	// Test list lists command
	t.Run("list lists command", func(t *testing.T) {
		cmd := listListCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "list", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotNil(t, cmd.Run)
	})
	
	// Test list default command
	t.Run("list default command", func(t *testing.T) {
		cmd := listDefaultCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "default <list-id>", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotNil(t, cmd.Run)
	})
}