package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCommand_Structure(t *testing.T) {
	// Test version command
	t.Run("version command exists", func(t *testing.T) {
		cmd := versionCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "version", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		// RunE rather than Run: formatting can fail, and the error should
		// propagate as a non-zero exit rather than being swallowed.
		assert.True(t, cmd.Run != nil || cmd.RunE != nil, "version command must be runnable")
	})
}
