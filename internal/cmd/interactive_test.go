package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInteractiveCommand_Structure(t *testing.T) {
	// Test main interactive command
	t.Run("interactive command exists", func(t *testing.T) {
		cmd := interactiveCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "interactive", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		assert.NotNil(t, cmd.Run)
		
		// No aliases for interactive command
	})
	
	// Test interactive command has no specific flags
	t.Run("interactive command flags", func(t *testing.T) {
		// Interactive command doesn't define its own flags
		// It uses global flags from root command
		assert.NotNil(t, interactiveCmd.Flags())
	})
}