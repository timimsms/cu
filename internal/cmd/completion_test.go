package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompletionCommand_Structure(t *testing.T) {
	// Test completion command
	t.Run("completion command exists", func(t *testing.T) {
		cmd := completionCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "completion [bash|zsh|fish|powershell]", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		assert.NotNil(t, cmd.Run)
		
		// Should accept exactly 1 argument
		args := cmd.Args
		assert.NotNil(t, args)
		
		// Check valid args
		assert.Contains(t, cmd.ValidArgs, "bash")
		assert.Contains(t, cmd.ValidArgs, "zsh")
		assert.Contains(t, cmd.ValidArgs, "fish")
		assert.Contains(t, cmd.ValidArgs, "powershell")
	})
}