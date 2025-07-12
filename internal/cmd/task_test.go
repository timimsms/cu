package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaskCommands_Structure(t *testing.T) {
	// Test main task command
	t.Run("task command exists", func(t *testing.T) {
		cmd := taskCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "task", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		
		// Should have subcommands
		assert.NotEmpty(t, cmd.Commands())
	})
	
	// Test task list command
	t.Run("task list command", func(t *testing.T) {
		cmd := taskListCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "list", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotNil(t, cmd.Run)
		
		// Should have some flags
		assert.NotNil(t, cmd.Flags())
	})
	
	// Test task create command
	t.Run("task create command", func(t *testing.T) {
		cmd := taskCreateCmd
		assert.NotNil(t, cmd)
		assert.Contains(t, cmd.Use, "create")
		assert.NotEmpty(t, cmd.Short)
		assert.NotNil(t, cmd.Run)
	})
	
	// Test task update command
	t.Run("task update command", func(t *testing.T) {
		cmd := taskUpdateCmd
		assert.NotNil(t, cmd)
		assert.Contains(t, cmd.Use, "update")
		assert.NotEmpty(t, cmd.Short)
		assert.NotNil(t, cmd.Run)
	})
	
	// Test task view command
	t.Run("task view command", func(t *testing.T) {
		cmd := taskViewCmd
		assert.NotNil(t, cmd)
		assert.Contains(t, cmd.Use, "view")
		assert.NotEmpty(t, cmd.Short)
		assert.NotNil(t, cmd.Run)
	})
	
	// Test other task commands exist
	t.Run("other task commands", func(t *testing.T) {
		assert.NotNil(t, taskCloseCmd)
		assert.NotNil(t, taskReopenCmd)
		assert.NotNil(t, taskSearchCmd)
	})
}